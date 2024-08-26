//@ts-nocheck
import { AUTH_SIGNATURE, authorize } from "../lib/auth.js";


class LogInBtn extends HTMLElement {
    btn;
    isLoggedIn = false;

    constructor() {
        super();
        this.container = document.createElement("container");
        this.container.innerHTML = `<a id="login-button" href="_">Log In</a>`;
        this.appendChild(this.container);
        this.btn = document.getElementById("login-button")
        this.attachListeners();
        if (localStorage.getItem(AUTH_SIGNATURE)) {
            this.btn.textContent = 'Log Out';
            this.isLoggedIn = true;
        }
    }

    attachListeners() {
        this.btn?.addEventListener('click', (event) => this.onClick(event))
    }

    async onClick(event) {
        event.preventDefault();

        if (this.isLoggedIn) {
            this.btn.textContent = 'Log In';
            localStorage.removeItem(AUTH_SIGNATURE);
            this.isLoggedIn = false;

        } else {
            authorize().then(() => {
                this.isLoggedIn = true;
                this.btn.textContent = 'Log Out';
            }).catch(err => {
                console.error(err);
                alert("failed to execute personal_sign")
            })
        }
    }

    async connectedCallback() {

    }

}

window.customElements.define('log-in-btn', LogInBtn);