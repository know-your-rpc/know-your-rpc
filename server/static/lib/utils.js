import { AUTH_SIGNATURE } from "./auth.js";

export async function getRequest(url, queryParams) {
    const authorization = getAuthorization();

    // @ts-ignore
    const response = await fetch(url + "?" + new URLSearchParams(queryParams), { method: "GET", headers: { Authorization: authorization } })

    if (response.ok) {
        return await response.json();
    }

    throw new Error(`Request failed url=${url} statusCode=${response.status}`)
}

export async function postRequest(url, body) {
    const authorization = getAuthorization();

    // @ts-ignore
    const response = await fetch(url, { method: "POST", headers: { Authorization: authorization }, body: JSON.stringify(body) })

    if (response.ok) {
        return
    }

    throw new Error(`Request failed url=${url} statusCode=${response.status}`)
}

function getAuthorization() {
    const maybeAuthSignature = localStorage.getItem(AUTH_SIGNATURE)

    if (!maybeAuthSignature) {
        return undefined;
    }

    const { message, authorizationSignature } = JSON.parse(maybeAuthSignature);

    return `${authorizationSignature}#${message}`;
}

export function getLastChainId() {
    return localStorage.getItem("last_chain_id") || "1";
}

export function getLastPeriod() {
    return localStorage.getItem("period") || "48h";
}

export function getLastTimeRange() {
    if (localStorage.getItem("time_range")) {
        return localStorage.getItem("time_range")?.split("-").map(Number)
    } else {
        return periodToTimeRange(getLastPeriod());
    }
}

export function periodToTimeRange(period) {
    const now = Date.now();
    switch (period) {
        case "30min": {
            return [now - (30 * 60 * 1_000), now];
        }
        case "48h": {
            return [now - (48 * 3600 * 1_000), now]
        }
        case "7d": {
            return [now - (7 * 24 * 3600 * 1_000), now]
        }
        case "14d": {
            return [now - (14 * 24 * 3600 * 1_000), now]
        }
        default: throw new Error(`unknown period=${period}`)
    }
}

export function toastSuccess(text) {
    Toastify({
        text: text,
        duration: 3000,
        style: {
            background: "linear-gradient(to right, #00b09b, #96c93d)",
        }
    }).showToast();
}

export function toastError(text) {
    Toastify({
        text: `Removed ${rpcUrl} from chainId=${currentChainId}`,
        duration: 3000,
        style: {
            background: "linear-gradient(to right, #00b09b, #96c93d)",
        }
    }).showToast();

}   