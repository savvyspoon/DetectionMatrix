// Dashboard JavaScript functionality
class DashboardAPI {
    static async fetchDetectionCount() {
        try {
            const response = await fetch('/api/detections/count');
            if (response.ok) {
                const data = await response.json();
                return data.count || 0;
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
                const data = await response.json();
                return data;
            }
        } catch (error) {
            console.error('Error fetching detection count by status:', error);
        }
        return {};
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
            const response = await fetch('/api/risk/alerts');
            if (response.ok) {
                const data = await response.json();
                const items = data.items || data || [];
                // Count alerts that are not "Closed"
                return items ? items.filter(alert => alert.status !== 'Closed').length : 0;
            }
        } catch (error) {
            console.error('Error fetching active alerts:', error);
        }
        return 0;
    }

    static async fetchEventsToday() {
        try {
            const response = await fetch('/api/events');
            if (response.ok) {
                const data = await response.json();
                const today = new Date().toISOString().split('T')[0];
                
                const items = data.items || [];
                return items.filter(event => event.timestamp.startsWith(today)).length;
            }
        } catch (error) {
            console.error('Error fetching events today:', error);
        }
        return 0;
    }

    static async fetchFalsePositives() {
        try {
            const response = await fetch('/api/events');
            if (response.ok) {
                const data = await response.json();
                const items = data.items || [];
                return items.filter(event => event.is_false_positive === true).length;
            }
        } catch (error) {
            console.error('Error fetching false positives:', error);
        }
        return 0;
    }
}

class DashboardChart {
    static renderCoverageChart(mitreData) {
        const ctx = document.getElementById('mitreChart');
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
window.dashboardData = function dashboardData() {
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

        init() {
            // Load data immediately and also use $nextTick for DOM readiness
            this.loadAllData().catch(err => console.error('Init load failed:', err));
        },

        async refreshData() {
            await this.loadAllData();
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
                this.renderChart();

                // Load risk data
                this.stats.highRiskEntities = await DashboardAPI.fetchHighRiskEntities();
                this.stats.activeAlerts = await DashboardAPI.fetchActiveAlerts();
                this.stats.eventsToday = await DashboardAPI.fetchEventsToday();
                this.stats.falsePositives = await DashboardAPI.fetchFalsePositives();

                this.lastUpdated = new Date();
            } catch (error) {
                console.error('Error loading dashboard data:', error);
            }
        },

        updateStatusCounts(statusData) {
            if (statusData && typeof statusData === 'object') {
                // Handle object format returned by API: {"draft": 3, "idea": 3, "production": 4, "test": 3}
                this.stats.production = statusData.production || 0;
                this.stats.testing = statusData.test || 0;  // API returns "test", we map to "testing"
                this.stats.draft = statusData.draft || 0;
                this.stats.idea = statusData.idea || 0;
            }
        },

        updateCoverageSummary(summary) {
            if (summary) {
                // API returns: {"coveredTactics":3,"coveredTechniques":4,"totalTactics":14,"totalTechniques":679}
                this.stats.techniquesCovered = summary.coveredTechniques || 0;
                this.stats.totalTechniques = summary.totalTechniques || 0;
                this.stats.tacticsCovered = summary.coveredTactics || 0;
                this.stats.totalTactics = summary.totalTactics || 0;
            }
        },

        getCoveragePercentage() {
            if (this.stats.totalTechniques === 0) return 0;
            return Math.round((this.stats.techniquesCovered / this.stats.totalTechniques) * 100);
        },

        getTacticCoveragePercentage() {
            if (this.stats.totalTactics === 0) return 0;
            return Math.round((this.stats.tacticsCovered / this.stats.totalTactics) * 100);
        },

        formatTime(date) {
            if (!date) return 'Never';
            return date.toLocaleString();
        },

        renderChart() {
            const ctx = document.getElementById('mitreChart');
            if (!ctx || !this.mitreData) return;

            // The API returns coverage percentages by tactic: {"Collection":7.5,"Command and Control":0,...}
            const labels = Object.keys(this.mitreData);
            const data = labels.map(tactic => this.mitreData[tactic] || 0);

            if (this.chart) {
                this.chart.destroy();
            }

            this.chart = new Chart(ctx, {
                type: 'bar',
                data: {
                    labels: labels,
                    datasets: [{
                        label: 'Coverage %',
                        data: data,
                        backgroundColor: '#3b82f6',
                        borderColor: '#2563eb',
                        borderWidth: 1
                    }]
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
                            max: 100,
                            ticks: {
                                callback: function(value) {
                                    return value + '%';
                                }
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
    };
}
