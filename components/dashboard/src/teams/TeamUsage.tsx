/**
 * Copyright (c) 2022 Gitpod GmbH. All rights reserved.
 * Licensed under the GNU Affero General Public License (AGPL).
 * See License-AGPL.txt in the project root for license information.
 */

import { useContext } from "react";
import { useLocation } from "react-router";
import { PageWithSubMenu } from "../components/PageWithSubMenu";
import { getCurrentTeam, TeamsContext } from "./teams-context";
import { getTeamSettingsMenu } from "./TeamSettings";
import { PaymentContext } from "../payment-context";

function TeamUsage() {
    const { teams } = useContext(TeamsContext);
    const { showPaymentUI } = useContext(PaymentContext);
    const location = useLocation();
    const team = getCurrentTeam(location, teams);

    return (
        <PageWithSubMenu
            subMenu={getTeamSettingsMenu({ team, showPaymentUI })}
            title="Usage"
            subtitle="Manage team usage."
        >
            <div className="flex flex-col space-y-2">
                <div className="px-6 py-3 flex justify-between text-sm text-gray-400 border-t border-b border-gray-200 dark:border-gray-800 mb-2">
                    <div className="w-7/12">Workspace</div>
                    <div className="w-5/12">Usage duration</div>
                    <div className="w-5/12">Credits used</div>
                    <div className="w-5/12">Last used</div>
                </div>
            </div>
        </PageWithSubMenu>
    );
}

export default TeamUsage;
