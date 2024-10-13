import { getRequest } from "../lib/utils.js";

class ChainSearch extends HTMLElement {
    supportedChains = [];
    inputElement;
    optionsElement;
    filter = [];

    constructor() {
        super();
        this.container = document.createElement("container");
        this.container.innerHTML = `
<details class="dropdown" name="search-chain" id="search-chain" placeholder="Chain ID or name"
            style="margin-top: 5em;" aria-label="Chain ID or name" >
  <summary id="search-chain-output"></summary>
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
            window.onkeydown = undefined;
            this.filter = [];
            this.optionsElement.innerHTML = this.supportedChains.map(({ chainId, name }) => `<li><a data-chainid=${chainId} data-name="${name}" href="#" style="text-align: justify">name=${name} id=${chainId}</a></li>`).join("\n");
            const chainName = event.target.dataset.name;
            if (!chainId) {
                return;
            }

            // @ts-ignore
            window.dispatchEvent(new CustomEvent("_update_chain_id", { detail: { chainId } }));
            localStorage.setItem("last_chain_id", chainId);
            this.updateChainInTitle();
            this.inputElement?.removeAttribute("open")
        } else {
            this.inputElement?.setAttribute("open", "open")


            window.onkeydown = (event) => {
                if (event.key === "Escape") {
                    this.inputElement?.removeAttribute("open")
                    this.filter = [];
                    window.onkeydown = undefined;
                    return;
                }

                if (event.key === "Backspace") {
                    this.filter.pop();
                } else if (/^[a-zA-Z0-9]$/.test(event.key)) {
                    this.filter.push(event.key);
                } else {
                    return;
                }

                window.document.getElementById("search-chain-output").innerText = this.filter.join("");

                this.optionsElement.innerHTML = this.supportedChains
                    .filter(({ name, chainId }) => name.toLowerCase().includes(this.filter.join("")) || chainId.toString().includes(this.filter.join("")))
                    .map(({ chainId, name }) => `<li><a data-chainid=${chainId} data-name="${name}" href="#" style="text-align: justify">name=${name} id=${chainId}</a></li>`).join("\n");
            }
        }
    }

    updateChainInTitle() {
        const choosenChainId = localStorage.getItem("last_chain_id") || "1";
        const chainName = this.supportedChains.find(({ chainId }) => chainId.toString() === choosenChainId).name;
        // @ts-ignore
        window.document.getElementById("search-chain-output").innerText = `name=${chainName} id=${choosenChainId}`;
    }

    async connectedCallback() {
        this.supportedChains = await getRequest("/api/supported-chains");
        this.updateChainInTitle();
        // @ts-ignore
        this.optionsElement.innerHTML = this.supportedChains.map(({ chainId, name }) => `<li><a data-chainid=${chainId} data-name="${name}" href="#" style="text-align: justify">name=${name} id=${chainId}</a></li>`).join("\n");
    }

}
window.customElements.define('chain-search', ChainSearch);
