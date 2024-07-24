import { AUTH_SIGNATURE } from "./auth.js";

export async function getRequest(url, queryParams) {
    const authorization = getAuthorization();

    // @ts-ignore
    const response = await fetch(url + "?" + new URLSearchParams(queryParams), { method: "GET", headers: { Authorization: authorization } })

    if (response.ok) {
        return await response.json();
    }

    throw new Error(`Request failed statusCode=${response.status}`)
}

function getAuthorization() {
    const maybeAuthSignature = localStorage.getItem(AUTH_SIGNATURE)

    if (!maybeAuthSignature) {
        return undefined;
    }

    const { message, authorizationSignature } = JSON.parse(maybeAuthSignature);

    return `${authorizationSignature}#${message}`;
}