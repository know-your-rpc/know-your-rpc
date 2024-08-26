import { requireAuthorization } from "./auth.js";
import { getRequest, postRequest } from "./utils.js";

window.addEventListener('DOMContentLoaded', async () => {
    try {
        await requireAuthorization();
    } catch (e) {
        window.location.href = "/";
    }
    renderCustomRpcTable("1");
});

const TABLE_BODY_ID = "custom_rpcs_table_body";

function lastTr() {
    return `<tr>
                <td colspan="5" scope="row"><input type="text" placeholder="Add new custom RPC URL" id="custom-rpc-input"></td>
                <td>
                    <button class="small-button" id="custom-rpc-btn">
                        <svg width="20" height="20" viewBox="0 0 16 16" fill="currentColor" xmlns="http://www.w3.org/2000/svg">
                            <path d="M8 4a.5.5 0 0 1 .5.5v3h3a.5.5 0 0 1 0 1h-3v3a.5.5 0 0 1-1 0v-3h-3a.5.5 0 0 1 0-1h3v-3A.5.5 0 0 1 8 4z"/>
                        </svg>
                    </button>
                </td>
            </tr>`;
}

async function saveCustomRpc(currentChainId) {
    // @ts-ignore
    const customRpcUrl = document.getElementById('custom-rpc-input').value;
    if (customRpcUrl) {
        console.log('Custom RPC URL to save:', customRpcUrl);
        // @ts-ignore
        document.getElementById('custom-rpc-input').value = '';

        await postRequest("/api/custom-rpc/add", { rpcUrl: customRpcUrl, chainId: currentChainId });
        // @ts-ignore
        renderCustomRpcTable(currentChainId);
    }
}

async function removeCustomRpc(rpcUrl, currentChainId) {
    if (rpcUrl) {
        console.log('Custom RPC URL to save:', rpcUrl);

        await postRequest("/api/custom-rpc/remove", { rpcUrl, chainId: currentChainId });
        // @ts-ignore
        renderCustomRpcTable(currentChainId);
    }
}


function tr({ avgDiffFromMedian, avgRequestDuration, errorRate, rpcUrl }, index) {
    return `<tr>
                <td>${index + 1}</td>
                <th scope="row">${rpcUrl}</th>
                <td>${errorRate.toFixed(2)}</td>
                <td>${avgRequestDuration.toFixed(2)}</td>
                <td>${avgDiffFromMedian.toFixed(2)}</td>
                <td>
                    <button class="small-button" data-rpc-url="${rpcUrl}">
                        <svg width="20" height="20" viewBox="0 0 16 16" fill="currentColor" xmlns="http://www.w3.org/2000/svg">
                            <path d="M5.5 5.5A.5.5 0 0 1 6 6v6a.5.5 0 0 1-1 0V6a.5.5 0 0 1 .5-.5zm2.5 0a.5.5 0 0 1 .5.5v6a.5.5 0 0 1-1 0V6a.5.5 0 0 1 .5-.5zm3 .5a.5.5 0 0 0-1 0v6a.5.5 0 0 0 1 0V6z"/>
                            <path fill-rule="evenodd" d="M14.5 3a1 1 0 0 1-1 1H13v9a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V4h-.5a1 1 0 0 1-1-1V2a1 1 0 0 1 1-1H6a1 1 0 0 1 1-1h2a1 1 0 0 1 1 1h3.5a1 1 0 0 1 1 1v1zM4.118 4 4 4.059V13a1 1 0 0 0 1 1h6a1 1 0 0 0 1-1V4.059L11.882 4H4.118zM2.5 3V2h11v1h-11z"/>
                        </svg>
                    </button>
                </td>
            </tr>`;
}


export async function renderCustomRpcTable(chainId) {
    const topRpcResponse = await getRequest("/api/stats/top-rpcs", { chainId, all: true });

    const rows = lastTr() + topRpcResponse.map(tr).join("");

    const table = window.document.getElementById(TABLE_BODY_ID);
    // @ts-ignore
    table.innerHTML = rows;
    // @ts-ignore
    document.getElementById('custom-rpc-btn').addEventListener('click', () => saveCustomRpc(chainId));
    //@ts-ignore
    document.querySelectorAll('[data-rpc-url]').forEach(button => {
        // @ts-ignore
        button.addEventListener('click', () => removeCustomRpc(button.dataset.rpcUrl, chainId));
    });
}

// @ts-ignore
window.addEventListener("_update_chain_id", ({ detail: { chainId } }) => {
    renderCustomRpcTable(chainId);
});
