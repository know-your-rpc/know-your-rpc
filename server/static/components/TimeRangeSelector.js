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
            <button class="time-range-button" data-range="24h">30 min</button>
                <button class="time-range-button" data-range="24h">12 h</button>
                <button class="time-range-button" data-range="7d">7 days</button>
                <button class="time-range-button" data-range="1m">1 month</button>
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
                this.dispatchEvent(new CustomEvent('rangeSelected', {
                    detail: { range },
                    bubbles: true,
                    composed: true
                }));
            });
        });
    }

    connectedCallback() {
        const defaultButton = this.container.querySelector('[data-range="24h"]');
        if (defaultButton) {
            defaultButton.click();
        }
    }
}

window.customElements.define('time-range-selector', TimeRangeSelector);