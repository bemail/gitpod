/**
 * Copyright (c) 2022 Gitpod GmbH. All rights reserved.
 * Licensed under the GNU Affero General Public License (AGPL).
 * See License-AGPL.txt in the project root for license information.
 */

export interface Usage {
    // Usage ID
    id: string;

    workspaceId: string;
    workspaceInstanceId: string;

    // "workspace" or "prebuild"
    workspaceType: string;

    // "standard" or "XL"
    workspaceClass: string;

    // Is this when the workspace started, or when the report started?
    startTime: string;

    // Same question as above: is it the workspace or the usage report?
    endTime: string;

    // The credits used for this workspace ID
    usedCredits: number;

    // Relevant for workspace type. When prebuild, shows "prebuild"
    userId: string;

    // The project or repo
    project: string;
}

export const usageDummyData: Usage[] = [
    {
        id: "some-usage-id",
        workspaceId: "some-workspace-id",
        workspaceInstanceId: "some-instance-id",
        workspaceType: "prebuild",
        workspaceClass: "XL",
        startTime: "string",
        endTime: "string",
        usedCredits: 320,
        userId: "prebuild",
        project: "project-123",
    },
    {
        id: "some-usage-id2",
        workspaceId: "some-workspace-id2",
        workspaceInstanceId: "some-instance-id2",
        workspaceType: "workspace",
        workspaceClass: "standard",
        startTime: "string",
        endTime: "string",
        usedCredits: 130,
        userId: "some-user",
        project: "project-123",
    },
    {
        id: "some-usage-id3",
        workspaceId: "some-workspace-id3",
        workspaceInstanceId: "some-instance-id3",
        workspaceType: "workspace",
        workspaceClass: "XL",
        startTime: "string",
        endTime: "string",
        usedCredits: 150,
        userId: "some-other-user",
        project: "project-134",
    },
];
