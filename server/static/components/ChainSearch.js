import { getRequest } from "../lib/utils.js";

class ChainSearch extends HTMLElement {
    supportedChains = [];
    inputElement;
    optionsElement;

    constructor() {
        super();
        this.container = document.createElement("container");
        this.container.innerHTML = `
<details class="dropdown" name="search-chain" id="search-chain" placeholder="Chain ID or name"
            style="margin-top: 5em;" aria-label="Chain ID or name" >
  <summary id="search-chain-output">Chain id or name eg. Ethereum</summary>
  <ul id="search-chain-options">
  </ul>
</details>
        `;
        this.appendChild(this.container);
        this.inputElement = window.document.getElementById("search-chain");
        this.optionsElement = window.document.getElementById("search-chain-options");
        this.attachListeners();
    }

    isActive() {
        return this.inputElement?.getAttribute("open") !== null;
    }

    attachListeners() {
        this.inputElement?.addEventListener('click', (event) => this.onClick(event))
    }

    onClick(event) {
        event.preventDefault();
        if (this.isActive()) {
            const chainId = event.target.dataset.chainid;
            const chainName = event.target.dataset.name;
            if (!chainId) {
                return;
            }

            // @ts-ignore
            window.document.getElementById("search-chain-output").innerText = `${chainName} (${chainId})`;
            window.dispatchEvent(new CustomEvent("_update_chain_id", { detail: { chainId } }));
            localStorage.setItem("last_chain_id", chainId);
            this.inputElement?.removeAttribute("open")
        } else {
            this.inputElement?.setAttribute("open", "open")
        }
    }

    async connectedCallback() {
        this.supportedChains = await getRequest("/api/supported-chains");
        // @ts-ignore
        this.optionsElement.innerHTML = this.supportedChains.map(({ chainId, name }) => `<li><a data-chainid=${chainId} data-name="${name}" href="#" style="text-align: justify">${name} (${chainId})</a></li>`).join("\n");
    }

}

window.customElements.define('chain-search', ChainSearch);