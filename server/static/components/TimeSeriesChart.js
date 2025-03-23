//@ts-nocheck
import { getLastChainId, getLastTimeRange, getRequest, periodToTimeRange } from '../lib/utils.js';

const autocolors = window['chartjs-plugin-autocolors'];
Chart.register(autocolors);

// @ts-nocheck
async function fetchDataSet({ url, from, to, chainId }) {
    const data = await getRequest(url, { from: Math.round(from / 1_000), to: Math.round(to / 1_000), chainId });

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

function now() {
    return Math.round(Date.now() / 1000)
}

function getQueryParams() {
    const params = new URLSearchParams(window.location.search);
    return {
        from: params.get('from'),
        to: params.get('to'),
        chainId: params.get('chainId')
    };
}

function updateQueryParams(params) {
    const url = new URL(window.location);
    Object.entries(params).forEach(([key, value]) => {
        if (value) {
            url.searchParams.set(key, value);
        }
    });
    window.history.replaceState({}, '', url);
}

class TimeSeriesChart extends HTMLElement {
    chainId;

    constructor() {
        super();

        const params = getQueryParams();
        this.chainId = params.chainId || getLastChainId();

        let [from, to] = params.from && params.to
            ? [parseInt(params.from), parseInt(params.to)]
            : getLastTimeRange();

        this.currentStartX = from;
        this.currentEndX = to;

        updateQueryParams({
            chainId: this.chainId,
            from: from,
            to: to,
        });

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
            </article>
        `;
        this.appendChild(this.container);
        this.attachListeners();

        this.resetZoom = _.debounce(this.resetZoom.bind(this), 1_000);
    }

    attachListeners() {
        window.document.getElementById(`${this.canvasId}-btn-toggle`).addEventListener('click', this.toggleVisibilityAll.bind(this));
        window.document.getElementById(`${this.canvasId}-btn-reset-zoom`).addEventListener('click', this.resetZoom.bind(this));
        // @ts-ignore
        window.addEventListener("_update_chain_id", ({ detail: { chainId } }) => {
            this.chart.destroy();
            this.chainId = chainId;
            updateQueryParams({ chainId: this.chainId });
            this.connectedCallback()
        });
        window.addEventListener("_update_time_range", ({ detail: { range } }) => {
            this.chart.destroy();
            const [start, end] = range;
            this.currentStartX = start;
            this.currentEndX = end;
            updateQueryParams({
                from: start,
                to: end
            });
            this.connectedCallback()
        });
    }

    async fetchDataSets() {
        return await fetchDataSet({ url: this.dataset.url, from: this.currentStartX, to: this.currentEndX, chainId: this.chainId });
    }

    async connectedCallback() {
        this.options = this.createTimeSeriesOpts();
        const fetchedDataSets = await this.fetchDataSets();

        this.chart = createChart(this.canvasId, this.options, fetchedDataSets);
    }

    async onZoom() {
        const { min, max } = this.chart.scales.x;
        this.currentStartX = min;
        this.currentEndX = max;

        await this.updateDataSetsWithNewData();
    }

    async updateDataSetsWithNewData() {
        for (const newDataSet of await this.fetchDataSets()) {
            const matchingDataset = this.chart.data.datasets.find(dt => dt.label === newDataSet.label);
            if (matchingDataset) {
                matchingDataset.data = newDataSet.data;
            }
        }
        this.chart.stop(); // make sure animations are not running
        this.chart.update('none');
    }

    async resetZoom() {
        const [from, to] = getLastTimeRange();
        this.currentStartX = from;
        this.currentEndX = to;
        await this.updateDataSetsWithNewData();
        this.chart.resetZoom()
    }

    toggleVisibilityAll() {
        for (let i = 0; i < this.chart.data.datasets.length; i += 1) {
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
            datasets: {
                line: {
                    pointRadius: 1,  // Makes points smaller (default is 3)
                    borderWidth: 2,  // Makes lines thinner (default is 3)
                    tension: 0.15,    // Adds slight smoothing to lines (0 = no smoothing, 1 = maximum smoothing)
                    fill: false,     // Don't fill area under the line
                }
            },
            plugins: {
                autocolors: {
                    mode: 'dataset',
                    enabled: true
                },
                legend: {
                    position: 'bottom',
                    labels: {
                        usePointStyle: true,  // Makes legend items use points instead of rectangles
                        padding: 20
                    }
                },
                zoom: {
                    limits: {
                        x: { min: 'original', max: 'original', minRange: 60 * 1000 },
                        y: { min: -20 }
                    },
                    pan: {
                        enabled: true,
                        mode: 'x',
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
                        mode: 'x',
                        onZoomComplete: this.onZoom.bind(this)
                    }
                }
            },
            scales: {
                x: {
                    type: 'timeseries',
                    position: 'bottom',
                    time: {
                        displayFormats: {
                            millisecond: 'HH:mm:ss.SSS',
                            second: 'HH:mm:ss',
                            minute: 'HH:mm',
                            hour: 'MMM D, HH:mm',
                            day: 'MMM D',
                            week: 'MMM D',
                            month: 'MMM YYYY',
                            quarter: 'MMM YYYY',
                            year: 'YYYY'
                        },
                        tooltipFormat: 'MMM D, HH:mm',
                    },
                    grid: {
                        display: true,
                        color: 'rgba(0,0,0,0.05)'
                    },
                    ticks: {
                        maxRotation: 0,  // Prevents label rotation
                    }
                },
                y: {
                    type: 'linear',
                    max: Number(this.dataset.max),
                    min: Number(this.dataset.min),
                    ticks: {
                        stepSize: this.dataset.stepsize,
                        callback: function (value) {
                            // Format large numbers with K/M/B suffixes
                            if (Math.abs(value) >= 1000000000) {
                                return (value / 1000000000) + 'B';
                            }
                            if (Math.abs(value) >= 1000000) {
                                return (value / 1000000) + 'M';
                            }
                            if (Math.abs(value) >= 1000) {
                                return (value / 1000) + 'K';
                            }
                            return value;
                        }
                    },
                    grid: {
                        color: 'rgba(0,0,0,0.05)'
                    }
                }
            },
            interaction: {
                mode: 'nearest',
                intersect: false,
            }
        }
    }

}

window.customElements.define('timeseries-chart', TimeSeriesChart);
