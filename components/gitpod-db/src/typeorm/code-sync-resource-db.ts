/**
 * Copyright (c) 2020 Gitpod GmbH. All rights reserved.
 * Licensed under the GNU Affero General Public License (AGPL).
 * See License-AGPL.txt in the project root for license information.
 */

import { inject, injectable } from "inversify";
import { EntityManager } from "typeorm";
import * as uuid from "uuid";
import {
    DBCodeSyncResource,
    IUserDataCollectionManifest,
    IUserDataManifest,
    IUserDataResourceManifest,
    ServerResource,
} from "./entity/db-code-sync-resource";
import { TypeORM } from "./typeorm";

export interface CodeSyncInsertOptions {
    latestRev?: string;
    revLimit?: number;
    overwrite?: boolean;
}

@injectable()
export class CodeSyncResourceDB {
    @inject(TypeORM)
    private readonly typeORM: TypeORM;

    async getManifest(userId: string): Promise<IUserDataManifest> {
        const connection = await this.typeORM.getConnection();
        const result = await connection.manager
            .createQueryBuilder(DBCodeSyncResource, "resource")
            .where(
                "resource.userId = :userId AND resource.kind != 'editSessions' AND resource.collection = :collection AND resource.deleted = 0",
                {
                    userId,
                    collection: uuid.NIL,
                },
            )
            .andWhere((qb) => {
                const subQuery = qb
                    .subQuery()
                    .select("resource2.userId")
                    .addSelect("resource2.kind")
                    .addSelect("max(resource2.created)")
                    .from(DBCodeSyncResource, "resource2")
                    .where(
                        "resource2.userId = :userId AND resource2.kind != 'editSessions' AND resource2.collection = :collection AND resource2.deleted = 0",
                        { userId, collection: uuid.NIL },
                    )
                    .groupBy("resource2.kind")
                    .orderBy("resource2.created", "DESC")
                    .getQuery();
                return "(resource.userId,resource.kind,resource.created) IN " + subQuery;
            })
            .getMany();
        const latest: IUserDataResourceManifest = Object.create(null);
        for (const resource of result) {
            latest[resource.kind] = resource.rev;
        }

        const collectionsResult = await connection.manager
            .createQueryBuilder(DBCodeSyncResource, "resource")
            .where(
                "resource.userId = :userId AND resource.kind != 'editSessions' AND resource.collection != :collection AND resource.deleted = 0",
                {
                    userId,
                    collection: uuid.NIL,
                },
            )
            .andWhere((qb) => {
                const subQuery = qb
                    .subQuery()
                    .select("resource2.userId")
                    .addSelect("resource2.kind")
                    .addSelect("resource2.collection")
                    .addSelect("max(resource2.created)")
                    .from(DBCodeSyncResource, "resource2")
                    .where(
                        "resource2.userId = :userId AND resource2.kind != 'editSessions' AND resource2.collection != :collection AND resource2.deleted = 0",
                        { userId, collection: uuid.NIL },
                    )
                    .groupBy("resource2.kind")
                    .addGroupBy("resource2.collection")
                    .orderBy("resource2.created", "DESC")
                    .getQuery();
                return "(resource.userId,resource.kind,resource.collection,resource.created) IN " + subQuery;
            })
            .getMany();
        const collections: IUserDataCollectionManifest = Object.create(null);
        for (const resource of collectionsResult) {
            if (!collections[resource.collection]) {
                collections[resource.collection] = { latest: Object.create(null) };
            }
            collections[resource.collection].latest[resource.kind] = resource.rev;
        }

        return { session: userId, latest, collections };
    }

    async getResource(userId: string, kind: ServerResource, rev: string): Promise<DBCodeSyncResource | undefined> {
        const connection = await this.typeORM.getConnection();
        return this.doGetResource(connection.manager, userId, kind, rev);
    }

    async getResources(userId: string, kind: ServerResource): Promise<DBCodeSyncResource[]> {
        const connection = await this.typeORM.getConnection();
        return this.doGetResources(connection.manager, userId, kind);
    }

    async deleteSettingsSyncResources(userId: string, doDelete: () => Promise<void>): Promise<void> {
        const connection = await this.typeORM.getConnection();
        await connection.transaction(async (manager) => {
            await manager
                .createQueryBuilder()
                .update(DBCodeSyncResource)
                .set({ deleted: true })
                .where("userId = :userId AND kind != 'editSessions' AND deleted = 0", { userId })
                .execute();
            await doDelete();
        });
    }

    async deleteResource(
        userId: string,
        kind: ServerResource,
        rev: string | undefined,
        doDelete: (rev?: string) => Promise<void>,
    ): Promise<void> {
        const connection = await this.typeORM.getConnection();
        if (rev) {
            await connection.transaction(async (manager) => {
                await manager
                    .createQueryBuilder()
                    .delete()
                    .from(DBCodeSyncResource)
                    .where("userId = :userId AND kind = :kind AND rev = :rev", { userId, kind, rev })
                    .execute();
                await doDelete(rev);
            });
        } else {
            await connection.transaction(async (manager) => {
                await manager
                    .createQueryBuilder()
                    .update(DBCodeSyncResource)
                    .set({ deleted: true })
                    .where("userId = :userId AND kind = :kind", { userId, kind })
                    .execute();
                await doDelete();
            });
        }
    }

    async insert(
        userId: string,
        kind: ServerResource,
        doInsert: (rev: string, oldRevs?: string[]) => Promise<void>,
        params?: CodeSyncInsertOptions,
    ): Promise<string | undefined> {
        const connection = await this.typeORM.getConnection();
        return await connection.transaction(async (manager) => {
            let latest: DBCodeSyncResource | undefined;
            let toDeleted: DBCodeSyncResource[] | undefined;
            if (params?.revLimit) {
                const resources = await this.doGetResources(manager, userId, kind);
                latest = resources[0];
                if (resources.length >= params.revLimit) {
                    if (params.overwrite) {
                        // delete + 1 to insert instead of update
                        toDeleted = resources.splice(params?.revLimit - 1);
                    } else {
                        return undefined;
                    }
                }
            } else {
                latest = await this.doGetResource(manager, userId, kind, "latest");
            }
            // user setting always show with diff so we need to make sure it’s changed from prev revision
            if (params?.latestRev && latest?.rev !== params.latestRev) {
                return undefined;
            }
            const rev = uuid.v4();
            await manager
                .createQueryBuilder()
                .insert()
                .into(DBCodeSyncResource)
                .values({ userId, kind, rev })
                .execute();
            await doInsert(
                rev,
                toDeleted?.map((e) => e.rev),
            );
            return rev;
        });
    }

    private doGetResource(
        manager: EntityManager,
        userId: string,
        kind: ServerResource,
        rev: string | "latest",
    ): Promise<DBCodeSyncResource | undefined> {
        let qb = manager
            .createQueryBuilder(DBCodeSyncResource, "resource")
            .where("resource.userId = :userId AND resource.kind = :kind AND resource.deleted = 0", { userId, kind });
        if (rev === "latest") {
            qb.orderBy("resource.created", "DESC");
        } else {
            qb = qb.andWhere("resource.rev = :rev", { rev });
        }
        return qb.getOne();
    }

    private doGetResources(
        manager: EntityManager,
        userId: string,
        kind: ServerResource,
    ): Promise<DBCodeSyncResource[]> {
        return manager
            .getRepository(DBCodeSyncResource)
            .createQueryBuilder("resource")
            .where("resource.userId = :userId AND resource.kind = :kind AND resource.deleted = 0", { userId, kind })
            .orderBy("resource.created", "DESC")
            .getMany();
    }

    async getCollections(userId: string): Promise<string[]> {
        const connection = await this.typeORM.getConnection();
        const result = await connection.manager
            .createQueryBuilder(DBCodeSyncResource, "resource")
            .where("resource.userId = :userId AND resource.collection != :collection AND resource.deleted = 0", {
                userId,
                collection: uuid.NIL,
            })
            .distinctOn(["resource.collection"])
            .getMany();
        return result.map((r) => r.collection);
    }

    async deleteCollection(
        userId: string,
        collection: string | undefined,
        doDelete: (collection?: string) => Promise<void>,
    ): Promise<void> {
        const connection = await this.typeORM.getConnection();
        if (collection) {
            await connection.transaction(async (manager) => {
                await manager
                    .createQueryBuilder()
                    .update(DBCodeSyncResource)
                    .set({ deleted: true })
                    .where("userId = :userId AND collection = :collection", { userId, collection })
                    .execute();
                await doDelete(collection);
            });
        } else {
            await connection.transaction(async (manager) => {
                await manager
                    .createQueryBuilder()
                    .update(DBCodeSyncResource)
                    .set({ deleted: true })
                    .where("userId = :userId AND collection != :collection", { userId, collection: uuid.NIL })
                    .execute();
                await doDelete();
            });
        }
    }
}
