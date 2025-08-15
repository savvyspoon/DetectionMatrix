// Enhanced chart configurations and utilities
class ChartManager {
    constructor() {
        this.charts = {};
        this.colors = {
            primary: '#2c3e50',
            secondary: '#3498db',
            success: '#10b981',
            warning: '#f59e0b',
            danger: '#ef4444',
            info: '#0ea5e9',
            gradient: {
                primary: ['#2c3e50', '#34495e'],
                secondary: ['#3498db', '#2980b9'],
                success: ['#10b981', '#059669'],
                warning: ['#f59e0b', '#d97706'],
                danger: ['#ef4444', '#dc2626']
            }
        };
    }

    createGradient(ctx, colorSet) {
        const gradient = ctx.createLinearGradient(0, 0, 0, 400);
        gradient.addColorStop(0, colorSet[0]);
        gradient.addColorStop(1, colorSet[1]);
        return gradient;
    }

    // Risk score trend chart with sparkline effect
    createRiskTrendChart(canvasId, data) {
        const ctx = document.getElementById(canvasId).getContext('2d');
        
        if (this.charts[canvasId]) {
            this.charts[canvasId].destroy();
        }

        this.charts[canvasId] = new Chart(ctx, {
            type: 'line',
            data: {
                labels: data.labels,
                datasets: [{
                    label: 'Risk Score',
                    data: data.values,
                    borderColor: this.colors.danger,
                    backgroundColor: this.createGradient(ctx, this.colors.gradient.danger),
                    borderWidth: 3,
                    tension: 0.4,
                    fill: true,
                    pointRadius: 0,
                    pointHoverRadius: 6,
                    pointBackgroundColor: '#fff',
                    pointBorderColor: this.colors.danger,
                    pointBorderWidth: 2
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                interaction: {
                    intersect: false,
                    mode: 'index'
                },
                plugins: {
                    legend: {
                        display: false
                    },
                    tooltip: {
                        backgroundColor: 'rgba(0, 0, 0, 0.8)',
                        titleColor: '#fff',
                        bodyColor: '#fff',
                        borderColor: this.colors.danger,
                        borderWidth: 1,
                        cornerRadius: 8,
                        padding: 12,
                        displayColors: false,
                        callbacks: {
                            label: (context) => `Score: ${context.parsed.y}`
                        }
                    }
                },
                scales: {
                    x: {
                        display: false
                    },
                    y: {
                        display: false,
                        beginAtZero: true
                    }
                }
            }
        });

        return this.charts[canvasId];
    }

    // MITRE ATT&CK coverage heatmap
    createMitreHeatmap(canvasId, data) {
        const ctx = document.getElementById(canvasId).getContext('2d');
        
        if (this.charts[canvasId]) {
            this.charts[canvasId].destroy();
        }

        // Process data for heatmap
        const heatmapData = data.map((tactic, x) => 
            tactic.techniques.map((technique, y) => ({
                x: x,
                y: y,
                v: technique.coverage,
                label: technique.name
            }))
        ).flat();

        this.charts[canvasId] = new Chart(ctx, {
            type: 'scatter',
            data: {
                datasets: [{
                    label: 'Coverage',
                    data: heatmapData,
                    backgroundColor: (context) => {
                        const value = context.raw.v;
                        const alpha = value / 100;
                        return `rgba(52, 152, 219, ${alpha})`;
                    },
                    borderColor: 'rgba(52, 152, 219, 0.8)',
                    borderWidth: 1,
                    pointRadius: 20,
                    pointHoverRadius: 22
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: {
                    legend: {
                        display: false
                    },
                    tooltip: {
                        callbacks: {
                            label: (context) => {
                                return `${context.raw.label}: ${context.raw.v}% coverage`;
                            }
                        }
                    }
                },
                scales: {
                    x: {
                        type: 'category',
                        labels: data.map(t => t.name),
                        grid: {
                            display: false
                        }
                    },
                    y: {
                        type: 'linear',
                        min: 0,
                        max: data[0]?.techniques.length || 10,
                        ticks: {
                            stepSize: 1
                        },
                        grid: {
                            display: false
                        }
                    }
                }
            }
        });

        return this.charts[canvasId];
    }

    // Detection efficacy radar chart
    createEfficacyRadar(canvasId, data) {
        const ctx = document.getElementById(canvasId).getContext('2d');
        
        if (this.charts[canvasId]) {
            this.charts[canvasId].destroy();
        }

        this.charts[canvasId] = new Chart(ctx, {
            type: 'radar',
            data: {
                labels: ['Coverage', 'Accuracy', 'Enrichment', 'Actionability', 'Performance'],
                datasets: [{
                    label: 'Current',
                    data: data.current,
                    borderColor: this.colors.secondary,
                    backgroundColor: 'rgba(52, 152, 219, 0.1)',
                    borderWidth: 2,
                    pointBackgroundColor: this.colors.secondary,
                    pointBorderColor: '#fff',
                    pointBorderWidth: 2,
                    pointRadius: 4,
                    pointHoverRadius: 6
                }, {
                    label: 'Target',
                    data: data.target,
                    borderColor: this.colors.success,
                    backgroundColor: 'rgba(16, 185, 129, 0.1)',
                    borderWidth: 2,
                    borderDash: [5, 5],
                    pointBackgroundColor: this.colors.success,
                    pointBorderColor: '#fff',
                    pointBorderWidth: 2,
                    pointRadius: 4,
                    pointHoverRadius: 6
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: {
                    legend: {
                        position: 'bottom',
                        labels: {
                            padding: 20,
                            usePointStyle: true
                        }
                    },
                    tooltip: {
                        callbacks: {
                            label: (context) => {
                                return `${context.dataset.label}: ${context.parsed.r}%`;
                            }
                        }
                    }
                },
                scales: {
                    r: {
                        angleLines: {
                            display: true,
                            color: 'rgba(0, 0, 0, 0.1)'
                        },
                        suggestedMin: 0,
                        suggestedMax: 100,
                        ticks: {
                            stepSize: 20,
                            callback: (value) => value + '%'
                        },
                        pointLabels: {
                            font: {
                                size: 12,
                                weight: 'bold'
                            }
                        }
                    }
                }
            }
        });

        return this.charts[canvasId];
    }

    // Event timeline chart
    createEventTimeline(canvasId, data) {
        const ctx = document.getElementById(canvasId).getContext('2d');
        
        if (this.charts[canvasId]) {
            this.charts[canvasId].destroy();
        }

        const datasets = Object.keys(data.series).map((key, index) => {
            const colors = [this.colors.primary, this.colors.secondary, this.colors.warning];
            return {
                label: key,
                data: data.series[key],
                borderColor: colors[index % colors.length],
                backgroundColor: colors[index % colors.length] + '20',
                borderWidth: 2,
                tension: 0.4,
                fill: true
            };
        });

        this.charts[canvasId] = new Chart(ctx, {
            type: 'line',
            data: {
                labels: data.labels,
                datasets: datasets
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                interaction: {
                    mode: 'index',
                    intersect: false
                },
                plugins: {
                    legend: {
                        position: 'bottom',
                        labels: {
                            padding: 15,
                            usePointStyle: true
                        }
                    },
                    tooltip: {
                        backgroundColor: 'rgba(0, 0, 0, 0.8)',
                        cornerRadius: 8,
                        padding: 12
                    }
                },
                scales: {
                    x: {
                        grid: {
                            display: false
                        }
                    },
                    y: {
                        beginAtZero: true,
                        grid: {
                            color: 'rgba(0, 0, 0, 0.05)'
                        }
                    }
                }
            }
        });

        return this.charts[canvasId];
    }

    // Real-time risk gauge
    createRiskGauge(canvasId, value, maxValue = 100) {
        const ctx = document.getElementById(canvasId).getContext('2d');
        
        if (this.charts[canvasId]) {
            this.charts[canvasId].destroy();
        }

        const getColor = (val) => {
            if (val < 30) return this.colors.success;
            if (val < 70) return this.colors.warning;
            return this.colors.danger;
        };

        this.charts[canvasId] = new Chart(ctx, {
            type: 'doughnut',
            data: {
                datasets: [{
                    data: [value, maxValue - value],
                    backgroundColor: [
                        getColor(value),
                        'rgba(0, 0, 0, 0.05)'
                    ],
                    borderWidth: 0
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                circumference: 180,
                rotation: 270,
                cutout: '80%',
                plugins: {
                    legend: {
                        display: false
                    },
                    tooltip: {
                        enabled: false
                    }
                }
            },
            plugins: [{
                id: 'gaugeText',
                afterDraw: (chart) => {
                    const ctx = chart.ctx;
                    ctx.save();
                    ctx.textAlign = 'center';
                    ctx.textBaseline = 'middle';
                    ctx.font = 'bold 24px sans-serif';
                    ctx.fillStyle = getColor(value);
                    ctx.fillText(value, chart.width / 2, chart.height - 20);
                    ctx.font = '12px sans-serif';
                    ctx.fillStyle = '#666';
                    ctx.fillText('Risk Score', chart.width / 2, chart.height - 5);
                    ctx.restore();
                }
            }]
        });

        return this.charts[canvasId];
    }

    // Update all charts with animation
    updateChart(canvasId, newData) {
        if (this.charts[canvasId]) {
            this.charts[canvasId].data = newData;
            this.charts[canvasId].update('active');
        }
    }

    // Destroy chart
    destroyChart(canvasId) {
        if (this.charts[canvasId]) {
            this.charts[canvasId].destroy();
            delete this.charts[canvasId];
        }
    }

    // Destroy all charts
    destroyAll() {
        Object.keys(this.charts).forEach(id => this.destroyChart(id));
    }
}

// Initialize global chart manager
window.chartManager = new ChartManager();

// Auto-initialize charts on page load
document.addEventListener('DOMContentLoaded', () => {
    // Look for chart containers with data attributes
    document.querySelectorAll('[data-chart-type]').forEach(container => {
        const canvasId = container.querySelector('canvas')?.id;
        if (!canvasId) return;

        const chartType = container.dataset.chartType;
        const chartData = container.dataset.chartData ? 
            JSON.parse(container.dataset.chartData) : null;

        if (chartData) {
            switch(chartType) {
                case 'risk-trend':
                    window.chartManager.createRiskTrendChart(canvasId, chartData);
                    break;
                case 'mitre-heatmap':
                    window.chartManager.createMitreHeatmap(canvasId, chartData);
                    break;
                case 'efficacy-radar':
                    window.chartManager.createEfficacyRadar(canvasId, chartData);
                    break;
                case 'event-timeline':
                    window.chartManager.createEventTimeline(canvasId, chartData);
                    break;
                case 'risk-gauge':
                    window.chartManager.createRiskGauge(canvasId, chartData.value, chartData.max);
                    break;
            }
        }
    });
});