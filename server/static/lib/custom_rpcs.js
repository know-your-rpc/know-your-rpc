import { requireAuthorization } from "./auth.js";
import { dateRangeToTimestamp, getLastChainId, getLastTimeRangeStr, getRequest, postRequest } from "./utils.js";

window.addEventListener('DOMContentLoaded', async () => {
    try {
        await requireAuthorization();
    } catch (e) {
        alert("You have to be logged in!")
        window.location.href = "/";
    }
    const [from, to] = dateRangeToTimestamp(getLastTimeRangeStr());
    renderCustomRpcTable(getLastChainId(), from, to);
});

const TABLE_BODY_ID = "custom_rpcs_table_body";


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

async function removeAllCustomRpcs(chainId) {
    try {
        await postRequest("/api/custom-rpc/remove-all", { chainId, rpcUrl: "https://mock.com" });
        console.log('All custom RPCs removed successfully');
        renderCustomRpcTable(chainId);
    } catch (error) {
        console.error('Failed to remove all custom RPCs:', error);
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


async function renderCustomRpcTable(chainId, from, to) {
    const topRpcResponse = await fetchTopRpcs(chainId, from, to);

    const rows = topRpcResponse.map(tr).join("");

    const table = window.document.getElementById(TABLE_BODY_ID);
    // @ts-ignore
    table.innerHTML = rows;
    // @ts-ignore
    document.getElementById('custom-rpc-btn').addEventListener('click', () => saveCustomRpc(chainId));

    //@ts-ignore
    document.getElementById('remove-all-custom-rpcs-btn').addEventListener('click', () => removeAllCustomRpcs(chainId));

    //@ts-ignore
    document.querySelectorAll('[data-rpc-url]').forEach(button => {
        // @ts-ignore
        button.addEventListener('click', () => removeCustomRpc(button.dataset.rpcUrl, chainId));
    });
}

async function fetchTopRpcs(chainId, from, to) {
    try {
        return await getRequest("/api/stats/top-rpcs", { chainId, all: true, from, to });
    } catch (e) {
        console.error("Failed to fetch top RPCs", e);
        return [];
    }
}

// @ts-ignore
window.addEventListener("_update_chain_id", ({ detail: { chainId } }) => {
    const [from, to] = dateRangeToTimestamp(getLastTimeRangeStr());
    renderCustomRpcTable(chainId, from, to);
});

// @ts-ignore
window.addEventListener("_update_time_range", ({ detail: { range } }) => {
    const [from, to] = dateRangeToTimestamp(range);
    renderCustomRpcTable(getLastChainId(), from, to);
});
