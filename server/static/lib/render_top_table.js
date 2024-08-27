import { getLastChainId, getRequest } from "./utils.js";

window.addEventListener('DOMContentLoaded', () => renderTopTable(getLastChainId()));

const MAX_RPC_COUNT = 10;
const TABLE_BODY_ID = "top_table_body";



function tr({ avgDiffFromMedian, avgRequestDuration, errorRate, rpcUrl }, index) {
    return `<tr>
                <td>${index + 1}</td>
                <th scope="row">${rpcUrl}</th>
                <td>${errorRate.toFixed(2)}</td>
                <td>${avgRequestDuration.toFixed(2)}</td>
                <td>${avgDiffFromMedian.toFixed(2)}</td>
            </tr>`;
}

export async function renderTopTable(chainId) {
    const topRpcResponse = await getRequest("/api/stats/top-rpcs", { chainId });

    console.log({ topRpcResponse })

    const rows = topRpcResponse.slice(0, MAX_RPC_COUNT).map(tr).join("");

    // @ts-ignore
    window.document.getElementById(TABLE_BODY_ID).innerHTML = rows;
}

// @ts-ignore
window.addEventListener("_update_chain_id", ({ detail: { chainId } }) => {
    renderTopTable(chainId);
});
