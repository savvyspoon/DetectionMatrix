// MITRE ATT&CK JavaScript functionality
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
        
        async init() {
            await Promise.all([
                this.fetchTechniques(),
                this.fetchCoverage()
            ]);
            this.initWatchers();
            this.initChart();
        },
        
        async fetchTechniques() {
            try {
                this.techniques = await MitreAPI.fetchTechniques();
                
                if (this.techniques.length === 0) {
                    console.warn('No MITRE techniques found. Database may need to be populated.');
                    // Show a helpful message to the user
                    UIUtils.showAlert('No MITRE techniques found. Please import MITRE data first.', 'warning');
                }
                
                this.applyFilters();
                await this.fetchDetectionCounts();
            } catch (error) {
                console.error('Error fetching techniques:', error);
                UIUtils.showAlert('Error loading MITRE techniques. Please check the server connection.', 'error');
                this.techniques = [];
            }
        },
        
        async fetchCoverage() {
            try {
                this.coverage = await MitreAPI.fetchCoverage();
            } catch (error) {
                console.error('Error fetching coverage:', error);
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
            this.$nextTick(() => {
                const ctx = document.getElementById('coverageChart');
                if (ctx && Object.keys(this.coverage).length > 0) {
                    new Chart(ctx, {
                        type: 'bar',
                        data: {
                            labels: Object.keys(this.coverage),
                            datasets: [{
                                label: 'Coverage Percentage',
                                data: Object.values(this.coverage),
                                backgroundColor: 'rgba(33, 150, 243, 0.8)'
                            }]
                        },
                        options: {
                            responsive: true,
                            maintainAspectRatio: false,
                            plugins: {
                                title: {
                                    display: true,
                                    text: 'MITRE ATT&CK Coverage by Tactic'
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
                                }
                            }
                        }
                    });
                }
            });
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