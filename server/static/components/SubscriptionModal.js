//@ts-nocheck
import { AUTH_SIGNATURE, authorize, requireAuthorization } from "../lib/auth.js";
import { payForSubscription } from "../lib/usdc.js";


class SubscriptionModal extends HTMLElement {
    constructor() {
        super();
        this.container = document.createElement("container");
        this.container.innerHTML = `<dialog id="subscription-modal">
            <article>
                <header>
                    <h3>Upgrade to Premium</h3>
                </header>
                <p>Unlock premium features to enhance your RPC monitoring experience:</p>
                <ul>
                    <li>Track custom RPCs</li>
                    <li>Coming soon: Alerts - get notified when your RPC is misbehaving</li>
                    <li>Coming soon: Votes - choose next feature</li>
                </ul>
                <footer>
                    <button class="contrast" id="subscription-modal-cancel-btn">Use free version</button>
                    <button class="primary" id="subscription-modal-subscribe-btn">Subscribe
                        10 USDC/month</button>
                </footer>
            </article>
        </dialog>`;
        this.appendChild(this.container);
        this.modal = document.getElementById('subscription-modal');
        this.cancelBtn = document.getElementById('subscription-modal-cancel-btn');
        this.subscribeBtn = document.getElementById('subscription-modal-subscribe-btn');
        this.attachListeners();
    }

    attachListeners() {
        this.cancelBtn.addEventListener('click', () => {
            console.log("cancel btn clicked");
            this.modal.close();
            localStorage.removeItem(AUTH_SIGNATURE);
            window.location.href = "/";
        });

        this.subscribeBtn.addEventListener('click', async () => {
            await payForSubscription().finally(() => {
                window.location.href = "/";
            });
        });
    }

    async connectedCallback() {
        try {
            if (this.dataset.mustauth === "true" || localStorage.getItem(AUTH_SIGNATURE)) {
                await requireAuthorization();
            }
        } catch (e) {
            console.log(e);
            this.modal.showModal();
        }
    }

}

window.customElements.define('subscription-modal', SubscriptionModal);