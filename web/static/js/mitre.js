// MITRE ATT&CK JavaScript functionality

// Global flag to prevent duplicate initialization
window.mitreInitialized = false;
window.mitreChart = null;

// Clean up charts when page becomes hidden
document.addEventListener('visibilitychange', function() {
    if (document.hidden) {
        if (window.mitreChart) {
            try {
                window.mitreChart.destroy();
            } catch (e) {
                console.warn('Error destroying mitreChart on visibility change:', e);
            }
            window.mitreChart = null;
        }
        if (window.coverageChart) {
            try {
                window.coverageChart.destroy();
            } catch (e) {
                console.warn('Error destroying coverageChart on visibility change:', e);
            }
            window.coverageChart = null;
        }
        window.mitreInitialized = false;
    }
});

class MitreAPI {
    static async fetchTechniques() {
        return await APIUtils.fetchAPI('/api/mitre/techniques');
    }

    static async fetchCoverage() {
        return await APIUtils.fetchAPI('/api/mitre/coverage');
    }

    static async fetchDetectionsByTechnique(techniqueId) {
        return await APIUtils.fetchAPI(`/api/mitre/techniques/${techniqueId}/detections`);
    }
}

// Alpine.js MITRE data function
function mitreData() {
    return {
        techniques: [],
        filteredTechniques: [],
        tacticFilter: 'all',
        searchTerm: '',
        coverage: {},
        selectedTechnique: null,
        detectionsByTechnique: [],
        techniqueDetectionCounts: {},
        loading: true,
        error: null,
        
        async init() {
            // Prevent duplicate initialization using global flag
            if (window.mitreInitialized) {
                console.log('MITRE component already initialized globally, skipping...');
                return;
            }
            window.mitreInitialized = true;
            
            console.log('Initializing MITRE data component...');
            this.loading = true;
            this.error = null;
            
            try {
                // Fetch techniques first, then coverage
                await this.fetchTechniques();
                await this.fetchCoverage();
                console.log('Data fetching completed, initializing UI components...');
                this.initWatchers();
                
                console.log('MITRE component initialization complete');
                this.loading = false;
                
                // Initialize chart after loading is set to false and DOM updates
                this.$nextTick(() => {
                    console.log('About to initialize chart after DOM update...');
                    setTimeout(() => {
                        if (!window.mitreChart) {  // Only init if chart doesn't exist
                            this.initChart();
                        }
                    }, 200);
                });
                
            } catch (error) {
                console.error('Error during MITRE component initialization:', error);
                this.error = error.message;
                this.loading = false;
                UIUtils.showAlert('Failed to initialize MITRE component', 'error');
            }
        },
        
        async fetchTechniques() {
            try {
                console.log('Fetching MITRE techniques from API...');
                this.techniques = await MitreAPI.fetchTechniques();
                console.log(`Successfully loaded ${this.techniques.length} MITRE techniques`);
                
                if (this.techniques.length === 0) {
                    console.warn('No MITRE techniques found. Database may need to be populated.');
                    UIUtils.showAlert('No MITRE techniques found. Please import MITRE data first.', 'warning');
                } else {
                    UIUtils.showAlert(`Successfully loaded ${this.techniques.length} MITRE techniques`, 'success');
                }
                
                this.applyFilters();
                await this.fetchDetectionCounts();
            } catch (error) {
                console.error('Error fetching techniques:', error);
                console.error('Error details:', {
                    message: error.message,
                    stack: error.stack,
                    url: '/api/mitre/techniques'
                });
                UIUtils.showAlert(`Error loading MITRE techniques: ${error.message}`, 'error');
                this.techniques = [];
            }
        },
        
        async fetchCoverage() {
            try {
                console.log('Fetching MITRE coverage data...');
                this.coverage = await MitreAPI.fetchCoverage();
                console.log('Coverage data loaded:', this.coverage);
            } catch (error) {
                console.error('Error fetching coverage:', error);
                console.error('Coverage API error details:', {
                    message: error.message,
                    stack: error.stack,
                    url: '/api/mitre/coverage'
                });
                this.coverage = {};
            }
        },
        
        async fetchDetectionCounts() {
            try {
                // Fetch detection counts for each technique
                const counts = {};
                await Promise.all(
                    this.techniques.map(async (technique) => {
                        try {
                            const detections = await MitreAPI.fetchDetectionsByTechnique(technique.id);
                            counts[technique.id] = detections.length;
                        } catch (error) {
                            counts[technique.id] = 0;
                        }
                    })
                );
                this.techniqueDetectionCounts = counts;
            } catch (error) {
                console.error('Error fetching detection counts:', error);
            }
        },
        
        async fetchDetectionsByTechnique(techniqueId) {
            try {
                this.detectionsByTechnique = await MitreAPI.fetchDetectionsByTechnique(techniqueId);
            } catch (error) {
                console.error('Error fetching detections for technique:', error);
                this.detectionsByTechnique = [];
            }
        },
        
        applyFilters() {
            let filtered = [...this.techniques];
            
            // Apply tactic filter
            if (this.tacticFilter !== 'all') {
                filtered = filtered.filter(technique => 
                    technique.tactic && technique.tactic.toLowerCase() === this.tacticFilter.toLowerCase()
                );
            }
            
            // Apply search filter
            if (this.searchTerm) {
                const searchLower = this.searchTerm.toLowerCase();
                filtered = filtered.filter(technique =>
                    technique.id.toLowerCase().includes(searchLower) ||
                    technique.name.toLowerCase().includes(searchLower) ||
                    (technique.description && technique.description.toLowerCase().includes(searchLower))
                );
            }
            
            this.filteredTechniques = filtered;
        },
        
        async selectTechnique(technique) {
            this.selectedTechnique = technique;
            await this.fetchDetectionsByTechnique(technique.id);
        },
        
        closeModal() {
            this.selectedTechnique = null;
            this.detectionsByTechnique = [];
        },
        
        getTechniqueClass(technique) {
            const count = this.techniqueDetectionCounts[technique.id] || 0;
            if (count === 0) return 'technique-uncovered';
            if (count >= 3) return 'technique-well-covered';
            return 'technique-covered';
        },
        
        getDetectionCount(techniqueId) {
            return this.techniqueDetectionCounts[techniqueId] || 0;
        },
        
        getCoverageForTactic(tactic) {
            return this.coverage[tactic] || { covered: 0, total: 0 };
        },
        
        getUniqueTactics() {
            const tactics = [...new Set(this.techniques.map(t => t.tactic).filter(Boolean))];
            return tactics.sort();
        },
        
        viewDetection(detectionId) {
            UIUtils.navigateTo(`detections-detail.html?id=${detectionId}`);
        },
        
        // Watchers for filter changes
        renderMatrixView() {
            const tacticMap = {};
            
            this.techniques.forEach(technique => {
                if (!technique.tactic) {
                    return;
                }
                
                if (!tacticMap[technique.tactic]) {
                    tacticMap[technique.tactic] = [];
                }
                
                tacticMap[technique.tactic].push({
                    id: technique.id,
                    name: technique.name,
                    description: technique.description
                });
            });
            
            return tacticMap;
        },
        
        getTechniqueColor(techniqueId) {
            const count = this.techniqueDetectionCounts[techniqueId] || 0;
            if (count === 0) return '#f8f9fa'; // Light gray for no coverage
            if (count >= 3) return '#e8f5e9'; // Light green for good coverage
            return '#fff3e0'; // Light orange for some coverage
        },
        
        async selectTechnique(techniqueId) {
            const technique = this.techniques.find(t => t.id === techniqueId);
            if (technique) {
                this.selectedTechnique = technique;
                await this.fetchDetectionsByTechnique(techniqueId);
            }
        },
        
        initChart() {
            console.log('=== CHART INIT START ===');
            console.log('Current line number check - this should be around line 251');
            console.log('Coverage data:', this.coverage);
            
            const ctx = document.getElementById('coverageChart');
            if (!ctx) {
                console.error('Canvas element not found!');
                return;
            }
            
            if (Object.keys(this.coverage).length === 0) {
                console.warn('No coverage data available for chart');
                return;
            }
            
            // Complete chart cleanup - destroy everything
            console.log('=== CLEANING UP CHARTS ===');
            
            // Get all possible chart references
            const possibleCharts = ['mitreChart', 'coverageChart', 'chart', 'myChart'];
            
            possibleCharts.forEach(chartName => {
                if (window[chartName]) {
                    console.log(`Found chart: ${chartName}, attempting to destroy...`);
                    try {
                        if (typeof window[chartName].destroy === 'function') {
                            window[chartName].destroy();
                            console.log(`Successfully destroyed ${chartName}`);
                        } else {
                            console.log(`${chartName} does not have destroy method`);
                        }
                    } catch (e) {
                        console.warn(`Error destroying ${chartName}:`, e);
                    }
                    window[chartName] = null;
                }
            });
            
            // Clear any Chart.js instances from the canvas
            if (ctx.chart) {
                console.log('Found chart instance on canvas, destroying...');
                try {
                    ctx.chart.destroy();
                } catch (e) {
                    console.warn('Error destroying canvas chart:', e);
                }
            }
            
            console.log('=== CREATING NEW CHART ===');
            
            try {
                const chartData = {
                    labels: Object.keys(this.coverage),
                    datasets: [{
                        label: 'Coverage Percentage',
                        data: Object.values(this.coverage),
                        backgroundColor: 'rgba(33, 150, 243, 0.8)',
                        borderColor: 'rgba(33, 150, 243, 1)',
                        borderWidth: 1
                    }]
                };
                
                const chartOptions = {
                    responsive: true,
                    maintainAspectRatio: false,
                    plugins: {
                        title: {
                            display: true,
                            text: 'MITRE ATT&CK Coverage by Tactic',
                            font: { size: 16 }
                        },
                        legend: { display: true }
                    },
                    scales: {
                        y: {
                            beginAtZero: true,
                            max: 100,
                            ticks: {
                                callback: function(value) {
                                    return value + '%';
                                }
                            },
                            title: {
                                display: true,
                                text: 'Coverage %'
                            }
                        },
                        x: {
                            title: {
                                display: true,
                                text: 'MITRE ATT&CK Tactics'
                            }
                        }
                    }
                };
                
                console.log('Chart data:', chartData);
                console.log('About to create Chart instance...');
                
                window.mitreChart = new Chart(ctx, {
                    type: 'bar',
                    data: chartData,
                    options: chartOptions
                });
                
                console.log('✅ Chart created successfully!', window.mitreChart);
                console.log('=== CHART INIT COMPLETE ===');
                
            } catch (error) {
                console.error('❌ Error creating chart:', error);
                console.error('Error stack:', error.stack);
                console.error('Error message:', error.message);
            }
        },
        
        initWatchers() {
            this.$nextTick(() => {
                this.$watch('tacticFilter', () => {
                    this.applyFilters();
                });
                
                this.$watch('searchTerm', () => {
                    this.applyFilters();
                });
            });
        }
    };
}