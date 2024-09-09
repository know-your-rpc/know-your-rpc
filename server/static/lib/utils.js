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

export function getLastTimeRangeStr() {
    return localStorage.getItem("range_time") || "48h";
}

export function dateRangeToTimestamp(dateRangeStr) {
    const now = Math.round(Date.now() / 1000);
    switch (dateRangeStr) {
        case "30min": {
            return [now - (30 * 60), now];
        }
        case "48h": {
            return [now - (48 * 3600), now]
        }
        case "7d": {
            return [now - (7 * 24 * 3600), now]
        }
        case "14d": {
            return [now - (14 * 24 * 3600), now]
        }
        default: throw new Error(`unknown dataRangeStr=${dateRangeStr}`)
    }
}

