/**
 * Copyright (c) 2022 Gitpod GmbH. All rights reserved.
 * Licensed under the GNU Affero General Public License (AGPL).
 * See License-AGPL.txt in the project root for license information.
 */

import { useContext, useEffect, useState } from "react";
import { useLocation } from "react-router";
import { PageWithSubMenu } from "../components/PageWithSubMenu";
import { getCurrentTeam, TeamsContext } from "./teams-context";
import { getTeamSettingsMenu } from "./TeamSettings";
import { PaymentContext } from "../payment-context";
import { getGitpodService } from "../service/service";
import { BillableSession } from "@gitpod/gitpod-protocol/lib/usage";
import { Item, ItemField, ItemsList } from "../components/ItemsList";

function TeamUsage() {
    const { teams } = useContext(TeamsContext);
    const { showPaymentUI } = useContext(PaymentContext);
    const location = useLocation();
    const team = getCurrentTeam(location, teams);
    const [billedUsage, setBilledUsage] = useState<BillableSession[]>([]);

    useEffect(() => {
        if (!team) {
            return;
        }
        (async () => {
            const billedUsageResult = await getGitpodService().server.getBilledUsage("some-attribution-id");
            setBilledUsage(billedUsageResult);
        })();
    }, [team]);

    return (
        <PageWithSubMenu
            subMenu={getTeamSettingsMenu({ team, showPaymentUI })}
            title="Usage"
            subtitle="Manage team usage."
        >
            <div className="app-container">
                <ItemsList className="mt-2">
                    <Item header={true} className="grid grid-cols-3">
                        <ItemField className="my-auto">
                            <span className="pl-14">Workspace</span>
                        </ItemField>
                        <ItemField className="my-auto">
                            <span className="pl-14">Class</span>
                        </ItemField>
                        <ItemField className="my-auto">
                            <span className="pl-14">Credits Used</span>
                        </ItemField>
                    </Item>
                </ItemsList>
            </div>
        </PageWithSubMenu>
    );
}

export default TeamUsage;
