// @ts-nocheck
async function getRequest(url, queryParams) {
    const response = await fetch(url + "?" + new URLSearchParams(queryParams), { method: "GET" })

    if (response.ok) {
        return await response.json();
    }

    throw new Error(`Request failed statusCode=${response.status}`)
}

async function fetchDataSet({ url, query }) {
    const data = await getRequest(url, query);

    return data;
}

function createChart(canvasId, options, datasets) {
    const ctx = document.getElementById(canvasId);
    const chart = new Chart(ctx, {
        type: 'line',
        data: {
            datasets
        },
        options
    });
    return chart;
}

function createTimeSeriesOpts(opts = {}) {
    return {
        parse: false,
        animation: false,
        normalize: false,
        responsive: true,
        scales: {
            x: {
                type: 'timeseries',
                position: 'bottom',
                time: { unit: 'second' }
            },
            y: {
                type: 'linear',
                max: Number(opts.max),
                min: Number(opts.min),
                ticks: {
                    stepSize: opts.stepsize,
                }
            }
        }
    }
}


// TODO: should look like this https://velodata.app/futures/BTC
// TODO: fix colors
// TODO: add zoom
// TODO: add downsample?
// TODO: cap top and bottom?
class TimeSeriesChart extends HTMLElement {
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
                    <small>${this.dataset.info}</small>
                </header>
                <canvas id="${this.canvasId}"></canvas>                
                <footer>
                    <div class="grid">
                        <div style="text-align: left;">
                            <button id="${this.canvasId}-btn-toggle" class="small-button">Toggle all</button>
                        </div>
                        <div style="text-align: right;">
                            <button class="small-button">Toggle all</button>
                        </div>
                    </div>
                </footer>
            </article>
        `;
        this.appendChild(this.container);
        this.attachListeners();
    }

    attachListeners() {
        window.document.getElementById(`${this.canvasId}-btn-toggle`).addEventListener('click', this.toggleVisibilityAll.bind(this));
    }

    attr(name) {
        const attributeValue = this.getAttribute(name);
        if (!attributeValue) {
            throw new Error(`Missing attribute name=${name} component-id=${this.id}`)
        }
        return attributeValue;
    }

    async connectedCallback() {
        this.options = createTimeSeriesOpts(this.dataset);
        this.chartDataSets = await fetchDataSet(this.dataset);
        this.chart = createChart(this.canvasId, this.options, this.chartDataSets);
        console.log(this.chart)
    }

    toggleVisibilityAll() {
        for (let i = 0; i < this.chartDataSets.length; i += 1) {
            console.log("click")
            const currentState = this.chart.getDatasetMeta(i).hidden;
            this.chart.setDatasetVisibility(i, currentState);
        }
        this.chart.update();
    }
}

window.customElements.define('timeseries-chart', TimeSeriesChart);

// TODO: choose regin
// TODO: choose chainId