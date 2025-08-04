// Dashboard JavaScript functionality
class DashboardAPI {
    static async fetchDetectionCount() {
        try {
            const response = await fetch('/api/detections/count');
            if (response.ok) {
                const data = await response.json();
                return data.count;
            }
        } catch (error) {
            console.error('Error fetching detection count:', error);
        }
        return 0;
    }

    static async fetchDetectionCountByStatus() {
        try {
            const response = await fetch('/api/detections/count/status');
            if (response.ok) {
                return await response.json();
            }
        } catch (error) {
            console.error('Error fetching detection count by status:', error);
        }
        return [];
    }

    static async fetchMitreCoverageSummary() {
        try {
            const response = await fetch('/api/mitre/coverage/summary');
            if (response.ok) {
                return await response.json();
            }
        } catch (error) {
            console.error('Error fetching MITRE coverage summary:', error);
        }
        return {};
    }

    static async fetchMitreCoverageByTactic() {
        try {
            const response = await fetch('/api/mitre/coverage');
            if (response.ok) {
                return await response.json();
            }
        } catch (error) {
            console.error('Error fetching MITRE coverage by tactic:', error);
        }
        return {};
    }

    static async fetchHighRiskEntities() {
        try {
            const response = await fetch('/api/risk/high');
            if (response.ok) {
                const data = await response.json();
                return data ? data.length : 0;
            }
        } catch (error) {
            console.error('Error fetching high risk entities:', error);
        }
        return 0;
    }

    static async fetchActiveAlerts() {
        try {
            const response = await fetch('/api/risk/alerts?status=open');
            if (response.ok) {
                const data = await response.json();
                return data ? data.length : 0;
            }
        } catch (error) {
            console.error('Error fetching active alerts:', error);
        }
        return 0;
    }
}

class DashboardChart {
    static renderCoverageChart(mitreData) {
        const ctx = document.getElementById('coverageChart');
        if (!ctx || !mitreData) return null;

        const labels = Object.keys(mitreData);
        const coveredData = labels.map(tactic => mitreData[tactic]?.covered || 0);
        const totalData = labels.map(tactic => mitreData[tactic]?.total || 0);

        return new Chart(ctx, {
            type: 'bar',
            data: {
                labels: labels,
                datasets: [
                    {
                        label: 'Covered',
                        data: coveredData,
                        backgroundColor: '#3b82f6',
                        borderColor: '#2563eb',
                        borderWidth: 1
                    },
                    {
                        label: 'Total',
                        data: totalData,
                        backgroundColor: '#e5e7eb',
                        borderColor: '#d1d5db',
                        borderWidth: 1
                    }
                ]
            },
            options: {
                responsive: true,
                plugins: {
                    title: {
                        display: true,
                        text: 'MITRE ATT&CK Coverage by Tactic'
                    },
                    legend: {
                        position: 'top',
                    }
                },
                scales: {
                    y: {
                        beginAtZero: true,
                        ticks: {
                            precision: 0
                        }
                    },
                    x: {
                        ticks: {
                            maxRotation: 45,
                            minRotation: 45
                        }
                    }
                }
            }
        });
    }
}

// Alpine.js dashboard data function
function dashboardData() {
    return {
        lastUpdated: new Date(),
        stats: {
            totalDetections: 0,
            production: 0,
            testing: 0,
            draft: 0,
            idea: 0,
            techniquesCovered: 0,
            totalTechniques: 0,
            tacticsCovered: 0,
            totalTactics: 0,
            highRiskEntities: 0,
            activeAlerts: 0,
            eventsToday: 0,
            falsePositives: 0
        },
        mitreData: {},
        chart: null,

        async init() {
            this.$nextTick(async () => {
                await this.loadAllData();
            });
        },

        async loadAllData() {
            try {
                // Load detection data
                this.stats.totalDetections = await DashboardAPI.fetchDetectionCount();
                
                const statusData = await DashboardAPI.fetchDetectionCountByStatus();
                this.updateStatusCounts(statusData);

                // Load MITRE data
                const coverageSummary = await DashboardAPI.fetchMitreCoverageSummary();
                this.updateCoverageSummary(coverageSummary);

                this.mitreData = await DashboardAPI.fetchMitreCoverageByTactic();
                this.chart = DashboardChart.renderCoverageChart(this.mitreData);

                // Load risk data
                this.stats.highRiskEntities = await DashboardAPI.fetchHighRiskEntities();
                this.stats.activeAlerts = await DashboardAPI.fetchActiveAlerts();

                this.lastUpdated = new Date();
            } catch (error) {
                console.error('Error loading dashboard data:', error);
            }
        },

        updateStatusCounts(statusData) {
            if (Array.isArray(statusData)) {
                statusData.forEach(item => {
                    switch (item.status) {
                        case 'production':
                            this.stats.production = item.count;
                            break;
                        case 'test':
                            this.stats.testing = item.count;
                            break;
                        case 'draft':
                            this.stats.draft = item.count;
                            break;
                        case 'idea':
                            this.stats.idea = item.count;
                            break;
                    }
                });
            }
        },

        updateCoverageSummary(summary) {
            if (summary) {
                this.stats.techniquesCovered = summary.techniques_covered || 0;
                this.stats.totalTechniques = summary.total_techniques || 0;
                this.stats.tacticsCovered = summary.tactics_covered || 0;
                this.stats.totalTactics = summary.total_tactics || 0;
            }
        },

        getCoveragePercentage() {
            if (this.stats.totalTechniques === 0) return 0;
            return Math.round((this.stats.techniquesCovered / this.stats.totalTechniques) * 100);
        },

        getTacticCoveragePercentage() {
            if (this.stats.totalTactics === 0) return 0;
            return Math.round((this.stats.tacticsCovered / this.stats.totalTactics) * 100);
        }
    };
}