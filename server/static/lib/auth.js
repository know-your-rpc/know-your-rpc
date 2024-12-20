// @ts-nocheck
import buffer from 'https://cdn.jsdelivr.net/npm/buffer@6.0.3/+esm'
import { toastSuccess, getRequest } from "./utils.js";

const SEVEN_DAYS = 7 * 24 * 3600;

export const AUTH_SIGNATURE = "authorization_signature";

export async function authorize() {
    if (!window.ethereum) {
        alert("Could not detect window.ethereum, please install some web3 wallet for authorization")
    }

    await window.ethereum.enable();

    const validUntil = Math.round(Date.now() / 1000) + SEVEN_DAYS;
    const message = `action=authorize_all version=0 domain=${document.domain} valid_until=${validUntil}`

    const accounts = await window.ethereum.request({ method: 'eth_requestAccounts' });
    const from = accounts[0];
    const msg = `0x${buffer.Buffer.from(message, "utf8").toString("hex")}`;
    const authorizationSignature = await ethereum.request({
        method: "personal_sign",
        params: [msg, from],
        from
    });
    localStorage.setItem(AUTH_SIGNATURE, JSON.stringify({ authorizationSignature, message, validUntil }))
    window.dispatchEvent(new CustomEvent("_authorization_success"))
}

export async function requireAuthorization() {
    if (!localStorage.getItem(AUTH_SIGNATURE)) {
        toastSuccess("Please authorize to continue (checkout your wallet)");
        await authorize();
    }
    const payment = await getPaymentData();

    if (payment.validUntil < Date.now() / 1000) {
        throw new Error("Subscription expired");
    }

    return payment;
}

export async function getPaymentData() {
    return await getRequest("/api/payment");
}