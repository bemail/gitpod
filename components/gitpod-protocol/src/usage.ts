/**
 * Copyright (c) 2022 Gitpod GmbH. All rights reserved.
 * Licensed under the GNU Affero General Public License (AGPL).
 * See License-AGPL.txt in the project root for license information.
 */

import { WorkspaceType } from "@gitpod/gitpod-protocol";

export interface BillableSession {
    // The id of the one paying the bill
    attributionId: string;

    // Relevant for workspace type. When prebuild, shows "prebuild"
    userId?: string;
    teamId?: string;

    instanceId: string;

    workspaceId: string;

    workspaceType: WorkspaceType;

    // "standard" or "XL"
    workspaceClass: string;

    // When the workspace started
    startTime: string;

    // When the workspace ended
    endTime: string;

    // The credits used for this session
    credits: number;

    // TODO - maybe
    projectId?: string;
}

export const billableSessionDummyData: BillableSession[] = [
    {
        attributionId: "some-attribution-id",
        userId: "prebuild",
        teamId: "prebuild",
        instanceId: "some-instance-id",
        workspaceId: "some-workspace-id",
        workspaceType: "prebuild",
        workspaceClass: "XL",
        startTime: "string",
        endTime: "string",
        credits: 320,
        projectId: "project-123",
    },
    {
        attributionId: "some-attribution-id2",
        userId: "some-user",
        teamId: "some-team",
        instanceId: "some-instance-id2",
        workspaceId: "some-workspace-id2",
        workspaceType: "regular",
        workspaceClass: "standard",
        startTime: "string",
        endTime: "string",
        credits: 130,
        projectId: "project-123",
    },
    {
        attributionId: "some-attribution-id3",
        userId: "some-other-user",
        instanceId: "some-instance-id3",
        workspaceId: "some-workspace-id3",
        workspaceType: "regular",
        workspaceClass: "XL",
        startTime: "string",
        endTime: "string",
        credits: 150,
        projectId: "project-134",
    },
];
