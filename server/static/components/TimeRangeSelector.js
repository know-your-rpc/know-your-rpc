import { getLastTimeRangeStr } from "../lib/utils.js";

class TimeRangeSelector extends HTMLElement {
    constructor() {
        super();
        this.container = document.createElement("container");
        this.render();
        this.attachListeners();
        this.appendChild(this.container);
    }

    render() {
        this.container.innerHTML = `
            <style>
                .time-range-button {
                    border: 1px solid #646b79;
                }
                .time-range-button.active {
                    background-color: white;
                }
            </style>
            <fieldset role="group" style="margin-bottom: 5rem" >
                <button class="time-range-button" data-range="30min">30 min</button>
                <button class="time-range-button" data-range="48h">48 h</button>
                <button class="time-range-button" data-range="7d">7 days</button>
                <button class="time-range-button" data-range="14d">14 days</button>
            </fieldset>
        `;
    }

    attachListeners() {
        const buttons = this.container.querySelectorAll('.time-range-button');
        buttons.forEach(button => {
            button.addEventListener('click', () => {
                buttons.forEach(btn => btn.classList.remove('active'));
                button.classList.add('active');

                const range = button.getAttribute('data-range');
                // @ts-ignore
                localStorage.setItem("range_time", range);
                this.dispatchEvent(new CustomEvent('_update_time_range', {
                    detail: { range },
                    bubbles: true,
                    composed: true
                }));


            });
        });
    }

    connectedCallback() {
        const defaultButton = this.container.querySelector(`[data-range="${getLastTimeRangeStr()}"]`);
        defaultButton?.classList.add("active");
    }
}


window.customElements.define('time-range-selector', TimeRangeSelector);