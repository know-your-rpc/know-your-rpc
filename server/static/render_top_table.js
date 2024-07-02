import { getRequest } from "./utils.js";

window.addEventListener('DOMContentLoaded', renderTopTable);

const MAX_RPC_COUNT = 8;
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

export async function renderTopTable() {
    const topRpcResponse = await getRequest("/api/stats/top-rpcs");
    console.log("dupa")

    const rows = topRpcResponse.slice(0, MAX_RPC_COUNT).reverse().map(tr).join("");

    // @ts-ignore
    window.document.getElementById(TABLE_BODY_ID).innerHTML = rows;
}