//@ts-nocheck
import { getLastChainId, getLastTimeRange, getRequest, periodToTimeRange } from '../lib/utils.js';

class HeatmapChart extends HTMLElement {
    chainId;

    constructor() {
        super();

        // Use current chain ID from local storage rather than URL params
        // This ensures we're always using the most current selection
        this.chainId = getLastChainId();

        let [from, to] = getLastTimeRange();
        this.currentStartX = from;
        this.currentEndX = to;

        // Update URL params to match actual state
        updateQueryParams({
            chainId: this.chainId,
            from: from,
            to: to,
        });

        this.container = document.createElement("container");
        this.chartId = `heatmap-${this.id}`
        this.container.innerHTML = `
            <article>
                <header>
                    <div style="text-align: center;">
                        <h3>${this.dataset.title}</h3>
                    </div>
                    
                    <div class="grid">
                        <div style="text-align: left;">
                            <button id="${this.chartId}-btn-reset-zoom" class="small-button">Reset zoom</button>
                        </div>

                        <div style="text-align: right;">
                            <button id="${this.chartId}-btn-toggle" class="small-button">Toggle all</button>
                        </div>                        
                    </div>
                </header>
                <div id="${this.chartId}" style="width:100%; min-height:350px;"></div>               
            </article>
        `;
        this.appendChild(this.container);
        this.attachListeners();

        this.resetZoom = _.debounce(this.resetZoom.bind(this), 1_000);
    }

    attachListeners() {
        window.document.getElementById(`${this.chartId}-btn-reset-zoom`).addEventListener('click', this.resetZoom.bind(this));
        
        window.addEventListener("_update_chain_id", ({ detail: { chainId } }) => {
            if (this.chart) {
                this.chart.destroy();
            }
            this.chainId = chainId;
            updateQueryParams({ chainId: this.chainId });
            this.connectedCallback();
        });
        
        window.addEventListener("_update_time_range", ({ detail: { range } }) => {
            if (this.chart) {
                this.chart.destroy();
            }
            const [start, end] = range;
            this.currentStartX = start;
            this.currentEndX = end;
            updateQueryParams({
                from: start,
                to: end
            });
            this.connectedCallback();
        });

        // Add theme change listener
        window.addEventListener("_theme_changed", () => {
            this.connectedCallback()
        });
    }

    async fetchDataSets() {
        return await fetchDataSet({ 
            url: this.dataset.url, 
            from: this.currentStartX, 
            to: this.currentEndX, 
            chainId: this.chainId 
        });
    }

    // Transform line chart data to heatmap format
    transformDataToHeatmap(datasets) {
        const rpcs = [];
        const timestamps = new Set();
        const dataMap = {};
        
        // First, collect all RPCs and timestamps
        datasets.forEach(dataset => {
            const rpcName = dataset.label;
            rpcs.push(rpcName);
            
            // Initialize data map for this RPC
            dataMap[rpcName] = {};
            
            // Collect timestamps and values
            dataset.data.forEach(point => {
                const timestamp = new Date(point.x).getTime();
                timestamps.add(timestamp);
                dataMap[rpcName][timestamp] = point.y;
            });
        });
        
        // Sort timestamps chronologically
        const sortedTimestamps = Array.from(timestamps).sort((a, b) => a - b);
        
        // Ensure all RPCs have values for all timestamps (fill gaps with zeros)
        rpcs.forEach(rpc => {
            sortedTimestamps.forEach(timestamp => {
                if (dataMap[rpc][timestamp] === undefined) {
                    dataMap[rpc][timestamp] = 0;
                }
            });
        });
        
        // Generate series data for ApexCharts
        const series = rpcs.map(rpc => {
            return {
                name: rpc,
                data: sortedTimestamps.map(timestamp => {
                    return {
                        x: new Date(timestamp).toISOString(),
                        y: dataMap[rpc][timestamp] !== undefined ? dataMap[rpc][timestamp].toFixed(2) : 0
                    };
                })
            };
        });
        
        console.log(`Processed ${rpcs.length} RPCs with ${sortedTimestamps.length} timestamps each`);
        
        return series;
    }

    async connectedCallback() {
        const fetchedDataSets = await this.fetchDataSets();
        const heatmapData = this.transformDataToHeatmap(fetchedDataSets);
        
        // Calculate dynamic height based on number of RPCs
        const rpcsCount = heatmapData.length;
        const baseHeight = 350;
        const heightPerRpc = 35;
        const dynamicHeight = Math.max(baseHeight, Math.min(1200, rpcsCount * heightPerRpc));
        
        document.getElementById(this.chartId).style.height = `${dynamicHeight}px`;
        
        const isDark = document.documentElement.getAttribute('data-theme') === 'dark';
        const options = {
            series: heatmapData,
            chart: {
                type: 'heatmap',
                height: dynamicHeight,
                fontFamily: 'JetBrains Mono, monospace',
                toolbar: {
                    show: false
                },
                animations: {
                    enabled: false
                },
                background: 'transparent',
                events: {
                    zoomed: this.onZoom.bind(this),
                },
                zoom: {
                    enabled: true,
                    type: 'x',
                    autoScaleYaxis: true,
                    allowMouseWheelZoom:false,
                    zoomedArea: {
                        fill: {
                            color: '#90CAF9',
                            opacity: 0.4
                        },
                        stroke: {
                            color: '#0D47A1',
                            opacity: 0.4,
                            width: 1
                        }
                    }
                }
            },
            theme: {
                mode: isDark ? 'dark' : 'light',
                palette: 'palette1'
            },
            dataLabels: {
                enabled: false
            },
            colors: ["#008FFB"],
            plotOptions: {
                heatmap: {
                    radius: 0,
                    enableShades: false,
                    distributed: false,
                    useFillColorAsStroke: false,
                    strokeWidth: 0.5,
                    colorScale: {
                        ranges: [
                            {
                                from: 0,
                                to: 1,
                                color: '#00E396',
                                name: 'low'
                            },
                            {
                                from: 1,
                                to: 5,
                                color: '#FEB019',
                                name: 'medium'
                            },
                            {
                                from: 5,
                                to: 100,
                                color: '#FF4560',
                                name: 'high'
                            }
                        ]
                    }
                }
            },
            xaxis: {
                type: 'datetime',
                labels: {
                    datetimeUTC: false,
                    format: 'MMM dd, HH:mm',
                    style: {
                        fontSize: '11px',
                        fontFamily: 'JetBrains Mono, monospace',
                        colors: isDark ? '#ffffff' : '#000000'
                    }
                },
                axisBorder: {
                    show: false,
                    color: isDark ? '#ffffff' : '#000000'
                },
                axisTicks: {
                    show: true,
                    color: isDark ? '#ffffff' : '#000000'
                }
            },
            yaxis: {
                labels: {
                    style: {
                        fontSize: '11px',
                        fontFamily: 'JetBrains Mono, monospace',
                        colors: isDark ? '#ffffff' : '#000000'
                    },
                    maxWidth: 300,
                    trim: false,
                    align: 'left',
                    offsetX: 15
                }
            },
            tooltip: {
                x: {
                    format: 'MMM dd, HH:mm'
                },
                y: {
                    formatter: function(value) {
                        return value + '%';
                    }
                },
                theme: isDark ? 'dark' : 'light',
                style: {
                    fontSize: '12px',
                    fontFamily: 'JetBrains Mono, monospace'
                }
            },
            legend: {
                position: 'bottom',
                horizontalAlign: 'center',
                fontSize: '11px',
                fontFamily: 'JetBrains Mono, monospace',
                labels: {
                    colors: isDark ? '#ffffff' : '#000000'
                },
                onItemClick: {
                    toggleDataSeries: true
                },
            }
        };

        // If the chart already exists, destroy it before creating a new one
        if (this.chart) {
            this.chart.destroy();
        }

        // Create the chart
        this.chart = new ApexCharts(document.getElementById(this.chartId), options);
        this.chart.render();        
    }

    onZoom(chartContext, { xaxis }) {
        this.currentStartX = new Date(xaxis.min).getTime();
        this.currentEndX = new Date(xaxis.max).getTime();
        
        this.updateDataSetsWithNewData();
    }

    async updateDataSetsWithNewData() {
        const fetchedDataSets = await this.fetchDataSets();
        const heatmapData = this.transformDataToHeatmap(fetchedDataSets);
        
        this.chart.updateSeries(heatmapData);
    }

    async resetZoom() {
        const [from, to] = getLastTimeRange();
        this.currentStartX = from;
        this.currentEndX = to;
        
        await this.updateDataSetsWithNewData();
        
        if (this.chart) {
            this.chart.zoomX(
                this.currentStartX,
                this.currentEndX
            );
        }
    }

    
}

// Helper functions (same as in TimeSeriesChart)
async function fetchDataSet({ url, from, to, chainId }) {
    const data = await getRequest(url, { 
        from: Math.round(from / 1_000), 
        to: Math.round(to / 1_000), 
        chainId 
    });

    return data;
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

window.customElements.define('heatmap-chart', HeatmapChart);
