/**
 * Copyright (c) 2020 Gitpod GmbH. All rights reserved.
 * Licensed under the GNU Affero General Public License (AGPL).
 * See License-AGPL.txt in the project root for license information.
 */

/**
 * Script that waits for a database to become available
 */
import "reflect-metadata";
import { Config } from "./config";
import { TypeORM } from ".";

const retryPeriod = 5000; // [ms]
const totalAttempts = 30;

async function connectOrReschedule(attempt: number) {
    const config = new Config();
    const typeorm = new TypeORM(config, { connectTimeoutMS: retryPeriod });
    try {
        await typeorm.connect();
        console.log("DB is available");
        process.exit(0);
    } catch(err) {
        rescheduleConnectionAttempt(attempt, err);
    }

}

function rescheduleConnectionAttempt(attempt: number, err: any) {
    if(attempt == totalAttempts) {
        console.log(`Could not connect within ${totalAttempts} attempts. Stopping.`)
        process.exit(1);
    }
    console.log(`Connection attempt ${attempt}/${totalAttempts} failed. Retrying in ${retryPeriod / 1000} seconds.`);
    setTimeout(() => connectOrReschedule(attempt + 1), retryPeriod);
}

connectOrReschedule(0);
