/**
 * Copyright (c) 2022 Gitpod GmbH. All rights reserved.
 * Licensed under the GNU Affero General Public License (AGPL).
 * See License-AGPL.txt in the project root for license information.
 */

import { MigrationInterface, QueryRunner } from "typeorm";
import { columnExists } from "./helper/helper";

export class CodeSyncCollectionDB1664913169612 implements MigrationInterface {
    public async up(queryRunner: QueryRunner): Promise<void> {
        if (!(await columnExists(queryRunner, "d_b_code_sync_resource", "collection"))) {
            await queryRunner.query(
                "ALTER TABLE d_b_code_sync_resource ADD COLUMN `collection` char(36) NOT NULL DEFAULT '00000000-0000-0000-0000-000000000000', DROP PRIMARY KEY, ADD PRIMARY KEY (`userId`, `kind`, `rev`, `collection`), ALGORITHM=INPLACE, LOCK=NONE",
            );
        }
    }

    public async down(queryRunner: QueryRunner): Promise<void> {}
}
