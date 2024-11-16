import { getLastPeriod, getLastTimeRange, periodToTimeRange } from "../lib/utils.js";


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
                
            </style>
            
            <div class="grid">
            <label>Period
                <select>
                    <option value="30min">30 min</option>
                    <option value="48h">48 h</option>
                    <option value="7d">7 days</option>
                    <option value="14d">14 days</option>
                </select>
                </label>
                <label>From
                <input type="datetime-local" id="datetime-from">
                </label>
                <label>To
                <input type="datetime-local" id="datetime-to">
                </label>
            </div>
        `;
    }

    attachListeners() {
        this.select = this.container.querySelector('select');
        this.fromInput = this.container.querySelector('#datetime-from');
        this.toInput = this.container.querySelector('#datetime-to');

        this.select.addEventListener('change', (e) => {
            const period = e.target.value;

            const [from, to] = periodToTimeRange(period);

            localStorage.setItem("period", period);
            localStorage.setItem("time_range", `${from}-${to}`)

            this.updatePickerState(period, from, to);

            this.dispatchEvent(new CustomEvent('_update_time_range', {
                detail: { range: [from, to] },
                bubbles: true,
                composed: true
            }));
        });

        [this.fromInput, this.toInput].forEach(input => {
            input.addEventListener('change', () => {
                const from = new Date(this.fromInput.value).getTime();
                const to = new Date(this.toInput.value).getTime();

                localStorage.setItem("time_range", `${from}-${to}`)

                this.dispatchEvent(new CustomEvent('_update_time_range', {
                    detail: { range: [from, to] },
                    bubbles: true,
                    composed: true
                }));
            });
        });
    }

    connectedCallback() {
        const period = getLastPeriod();
        const [from, to] = periodToTimeRange(period);

        this.updatePickerState(period, from, to);
    }

    updatePickerState(period, from, to) {
        const now = new Date();
        const twoWeeksAgo = new Date(Date.now() - 14 * 24 * 60 * 60 * 1000);
        this.select.value = period;
        this.fromInput.value = new Date(from).toISOString().slice(0, 16);
        this.toInput.value = new Date(to).toISOString().slice(0, 16);
        this.fromInput.min = twoWeeksAgo.toISOString().slice(0, 16);
        this.fromInput.max = now.toISOString().slice(0, 16);
        this.toInput.min = twoWeeksAgo.toISOString().slice(0, 16);
        this.toInput.max = now.toISOString().slice(0, 16);
    }
}


window.customElements.define('time-range-selector', TimeRangeSelector);