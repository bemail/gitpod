/**
 * Copyright (c) 2020 Gitpod GmbH. All rights reserved.
 * Licensed under the GNU Affero General Public License (AGPL).
 * See License-AGPL.txt in the project root for license information.
 */

import { test, suite } from "mocha-typescript";
import * as chai from "chai";
import { PrebuiltWorkspace, PrebuiltWorkspaceState } from "./protocol";

const expect = chai.expect;

@suite
export class TestPrebuiltWorkspace {
    @test
    public testIsDone() {
        expect(PrebuiltWorkspace.isDone(newPrebuildWorkspaceWithState("queued"))).to.equal(false);
        expect(PrebuiltWorkspace.isDone(newPrebuildWorkspaceWithState("building"))).to.equal(false);
        expect(PrebuiltWorkspace.isDone(newPrebuildWorkspaceWithState("aborted"))).to.equal(false);
        expect(PrebuiltWorkspace.isDone(newPrebuildWorkspaceWithState("timeout"))).to.equal(true);
        expect(PrebuiltWorkspace.isDone(newPrebuildWorkspaceWithState("available"))).to.equal(true);
        expect(PrebuiltWorkspace.isDone(newPrebuildWorkspaceWithState("failed"))).to.equal(true);
        expect(PrebuiltWorkspace.isDone(newPrebuildWorkspaceWithState("random" as PrebuiltWorkspaceState))).to.equal(
            false,
        );
    }
}

function newPrebuildWorkspaceWithState(state: PrebuiltWorkspaceState): PrebuiltWorkspace {
    return {
        id: "1",
        cloneURL: "https://github.com/gitpod-io/gitpod.git",
        commit: "ffe6f48",
        buildWorkspaceId: "123",
        creationTime: new Date().toUTCString(),
        state,
    };
}
