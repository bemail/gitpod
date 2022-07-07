/**
 * Copyright (c) 2022 Gitpod GmbH. All rights reserved.
 * Licensed under the GNU Affero General Public License (AGPL).
 * See License-AGPL.txt in the project root for license information.
 */

export interface Usage {
    workspaceId: string;
    workspaceInstanceId: string;

    // "workspace" or "prebuild"
    workpaceType: string;

    // "standard" or "XL"
    workspaceClass: string;

    // Is this when the workspace started, or when the report started?
    startTime: number;

    // Same question as above: is it the workspace or the usage report?
    endTime: number;

    // The credits used for this workspace ID
    usedCredits: number;

    // Relevant for workspace type. When prebuild, shows "prebuild"
    userId: string;

    // The project or repo
    project: string;
}
