const autocolors = window['chartjs-plugin-autocolors'];
Chart.register(autocolors);


// @ts-nocheck
async function getRequest(url, queryParams) {
    const response = await fetch(url + "?" + new URLSearchParams(queryParams), { method: "GET" })

    if (response.ok) {
        return await response.json();
    }

    throw new Error(`Request failed statusCode=${response.status}`)
}

async function fetchDataSet({ url, from, to }) {
    const data = await getRequest(url, { from, to });

    return data;
}

function createChart(canvasId, options, datasets) {
    const ctx = document.getElementById(canvasId);
    const chart = new Chart(ctx, {
        type: 'line',
        data: {
            datasets
        },
        options,
    });
    return chart;
}

function roundToFullHours(timestamp) {
    // TODO: maybe use something else then ceil
    return Math.ceil(timestamp / 3600) * 3600;
}

function now() {
    return Math.round(Date.now() / 1000)
}

class TimeSeriesChart extends HTMLElement {
    currentStartX = roundToFullHours(now() - 48 * 3600);
    currentEndX = roundToFullHours(now());

    _initialDataSets;

    constructor() {
        super();
        this.container = document.createElement("container");
        this.canvasId = `canvas-${this.id}`
        this.container.innerHTML = `
            <article>
                <header>
                    <div style="text-align: center;">
                        <h3>${this.dataset.title}</h3>
                    </div>
                    
                    <div class="grid">
                        <div style="text-align: left;">
                        <button id="${this.canvasId}-btn-reset-zoom" class="small-button">Reset zoom</button>
                        </div>
                        <div style="text-align: right;">
                        
                            <button id="${this.canvasId}-btn-toggle" class="small-button">Toggle all</button>
                        </div>
                        
                    </div>

                </header>
                <canvas id="${this.canvasId}"></canvas>                
                <footer>
                <small>${this.dataset.info}</small>
                </footer>
            </article>
        `;
        this.appendChild(this.container);
        this.attachListeners();

        this.resetZoom = _.debounce(this.resetZoom.bind(this), 1_000);
    }

    attachListeners() {
        window.document.getElementById(`${this.canvasId}-btn-toggle`).addEventListener('click', this.toggleVisibilityAll.bind(this));
        window.document.getElementById(`${this.canvasId}-btn-reset-zoom`).addEventListener('click', this.resetZoom.bind(this));
    }

    attr(name) {
        const attributeValue = this.getAttribute(name);
        if (!attributeValue) {
            throw new Error(`Missing attribute name=${name} component-id=${this.id}`)
        }
        return attributeValue;
    }

    async fetchData() {
        return await fetchDataSet({ url: this.dataset.url, from: this.currentStartX, to: this.currentEndX });
    }

    async connectedCallback() {
        this.options = this.createTimeSeriesOpts();
        this._initialDataSets = await this.fetchData();
        this.chart = createChart(this.canvasId, this.options, this._initialDataSets);
        console.log(this.chart)
    }

    async onZoom({ chart }) {
        const { min, max } = chart.scales.x;
        this.currentStartX = Math.round(min / 1000);
        this.currentEndX = Math.round(max / 1000);


        // clearTimeout(timer);
        // timer = setTimeout(() => {
        console.log('Fetched data between ' + new Date(this.currentStartX * 1000) + ' and ' + new Date(this.currentEndX * 1000));
        chart.data.datasets = await this.fetchData();
        chart.stop(); // make sure animations are not running
        chart.update('none');
        // }, 500);
    }

    resetZoom() {
        this.chart.data.datasets = this._initialDataSets;
        this.chart.stop(); // make sure animations are not running
        this.chart.update('none');
        this.chart.resetZoom()
    }

    toggleVisibilityAll() {
        for (let i = 0; i < this._initialDataSets.length; i += 1) {
            console.log("click")
            const currentState = this.chart.getDatasetMeta(i).hidden;
            this.chart.setDatasetVisibility(i, currentState);
        }
        this.chart.update();
    }

    createTimeSeriesOpts() {
        return {
            parse: false,
            animation: false,
            normalize: false,
            responsive: true,
            plugins: {
                autocolors: {
                    mode: 'dataset',
                    enabled: true
                },
                zoom: {
                    limits: {
                        x: { min: 'original', max: 'original', minRange: 60 * 1000 },
                    },
                    pan: {
                        enabled: true,
                        mode: 'xy',
                        modifierKey: 'ctrl',
                        // onPanComplete: startFetch
                    },
                    zoom: {
                        drag: {
                            enabled: true,
                            backgroundColor: 'rgba(225,225,225,0.5)'
                        },
                        pinch: {
                            enabled: true
                        },
                        mode: 'xy',
                        onZoomComplete: this.onZoom.bind(this)
                    }
                }
            },
            scales: {
                x: {
                    type: 'timeseries',
                    position: 'bottom',
                    time: { unit: 'second' }
                },
                y: {
                    type: 'linear',
                    max: Number(this.dataset.max),
                    min: Number(this.dataset.min),
                    ticks: {
                        stepSize: this.dataset.stepsize,
                    }
                }
            }
        }
    }
}

window.customElements.define('timeseries-chart', TimeSeriesChart);

// TODO: choose regin
// TODO: choose chainId