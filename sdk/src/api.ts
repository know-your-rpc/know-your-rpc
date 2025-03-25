import { Wallet } from "@ethersproject/wallet";

const ONE_YEAR = 365 * 24 * 3600;

export type RpcInfo = {
  url: string;
  headers: Record<string, string>;
};

export class Api {
  #baseUrl: string;
  #privateKey: string;
  #authData?: { authToken: string; validUntil: number };

  constructor(baseUrl: string, privateKey: string) {
    this.#baseUrl = baseUrl;
    this.#privateKey = privateKey;
  }

  async #makeRequest(endpoint: string, data: any) {
    if (!this.#authData || this.#authData.validUntil < Date.now() / 1000) {
      this.#authData = await this.#generateAuthToken(
        new URL(this.#baseUrl).hostname,
      );
    }

    const response = await fetch(`${this.#baseUrl}${endpoint}`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: this.#authData.authToken,
      },
      body: JSON.stringify(data),
    });

    if (!response.ok) {
      throw new Error(`Request failed with status ${response.status}`);
    }

    return response;
  }

  async addCustomRpc(
    rpcUrl: string,
    chainId: string,
    headers?: Record<string, string>,
  ) {
    return this.#makeRequest("/api/custom-rpc/add", {
      rpcUrl,
      chainId,
      headers,
    });
  }

  async removeCustomRpc(rpcUrl: string, chainId: string) {
    return this.#makeRequest("/api/custom-rpc/remove", { rpcUrl, chainId });
  }

  async removeAllCustomRpcs(chainId: string) {
    return this.#makeRequest("/api/custom-rpc/remove-all", { chainId });
  }

  async syncCustomRpcs(rpcInfos: RpcInfo[], chainId: string) {
    return this.#makeRequest("/api/custom-rpc/sync", { chainId, rpcInfos });
  }

  async #generateAuthToken(domain: string, period = ONE_YEAR) {
    const validUntil = Math.round(Date.now() / 1000) + period;
    const message = `action=authorize_all version=0 domain=${domain} valid_until=${validUntil}`;

    const wallet = new Wallet(Buffer.from(this.#privateKey, "hex"));

    const authorizationSignature = await wallet.signMessage(message);

    const authToken = `${authorizationSignature}#${message}`;

    return { authToken, validUntil };
  }
}
