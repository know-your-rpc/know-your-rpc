import { getLastChainId, getLastTimeRange, getRequest } from "./utils.js";

window.addEventListener('DOMContentLoaded', () => {
    const [from, to] = getLastTimeRange();
    renderTopTable(getLastChainId(), from, to);
}
);

const MAX_RPC_COUNT = 100;
const TABLE_BODY_ID = "top_table_body";


function tr({ avgDiffFromMedian, avgRequestDuration, errorRate, rpcUrl }, index) {
    if (errorRate === -1 && avgDiffFromMedian === -1 && avgRequestDuration === -1) {
        return "";
    }

    return `<tr>
                <td>${index + 1}</td>
                <th scope="row">${rpcUrl}</th>
                <td>${errorRate.toFixed(2)}</td>
                <td>${avgRequestDuration.toFixed(2)}</td>
                <td>${avgDiffFromMedian.toFixed(2)}</td>
            </tr>`;
}

export async function renderTopTable(chainId, from, to) {
    const topRpcResponse = await getRequest("/api/stats/top-rpcs", { chainId, from: Math.round(from / 1_000), to: Math.round(to / 1_000) });

    const rows = topRpcResponse.slice(0, MAX_RPC_COUNT).map(tr).join("");

    // @ts-ignore
    window.document.getElementById(TABLE_BODY_ID).innerHTML = rows;
}

// @ts-ignore
window.addEventListener("_update_chain_id", ({ detail: { chainId } }) => {
    const [from, to] = getLastTimeRange();
    renderTopTable(chainId, from, to);
});


// @ts-ignore
window.addEventListener("_update_time_range", ({ detail: { range } }) => {
    const [from, to] = range;
    renderTopTable(getLastChainId(), from, to);
});
