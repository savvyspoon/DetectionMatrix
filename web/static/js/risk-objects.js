// Risk Objects JavaScript functionality
class RiskObjectsAPI {
    static async fetchRiskObjects(threshold = '') {
        let url = '/api/risk/objects';
        if (threshold && threshold !== '') {
            url += `?threshold=${threshold}`;
        }
        return await APIUtils.fetchAPI(url);
    }
}

// Alpine.js risk objects data function
function riskObjectsData() {
    return {
        riskObjects: [],
        filteredRiskObjects: [],
        entityTypeFilter: 'all',
        riskThreshold: '',
        searchTerm: '',
        sortBy: 'score',
        sortOrder: 'desc',
        
        async init() {
            await this.fetchRiskObjects();
            this.initWatchers();
            this.initChart();
        },
        
        async fetchRiskObjects() {
            try {
                const data = await RiskObjectsAPI.fetchRiskObjects(this.riskThreshold);
                // Handle ListResponse structure from API
                if (data && typeof data === 'object' && 'items' in data) {
                    this.riskObjects = Array.isArray(data.items) ? data.items : [];
                } else if (Array.isArray(data)) {
                    this.riskObjects = data;
                } else {
                    this.riskObjects = [];
                }
                this.applySortingAndFiltering();
            } catch (error) {
                console.error('Error fetching risk objects:', error);
                UIUtils.showAlert('Error loading risk objects');
            }
        },
        
        applyFilters() {
            this.applySortingAndFiltering();
        },
        
        applySortingAndFiltering() {
            // Ensure riskObjects is an array before spreading
            if (!Array.isArray(this.riskObjects)) {
                console.warn('RiskObjects is not an array:', this.riskObjects);
                this.filteredRiskObjects = [];
                return;
            }
            let filtered = [...this.riskObjects];
            
            // Apply entity type filter
            if (this.entityTypeFilter !== 'all') {
                filtered = filtered.filter(obj => obj.entity_type === this.entityTypeFilter);
            }
            
            // Apply search filter
            if (this.searchTerm) {
                const searchLower = this.searchTerm.toLowerCase();
                filtered = filtered.filter(obj => 
                    obj.entity_value.toLowerCase().includes(searchLower) ||
                    obj.entity_type.toLowerCase().includes(searchLower)
                );
            }
            
            // Apply sorting
            filtered.sort((a, b) => {
                let aVal, bVal;
                
                switch (this.sortBy) {
                    case 'score':
                        aVal = a.current_score;
                        bVal = b.current_score;
                        break;
                    case 'entity':
                        aVal = a.entity_value.toLowerCase();
                        bVal = b.entity_value.toLowerCase();
                        break;
                    case 'type':
                        aVal = a.entity_type.toLowerCase();
                        bVal = b.entity_type.toLowerCase();
                        break;
                    case 'lastSeen':
                        aVal = new Date(a.last_seen);
                        bVal = new Date(b.last_seen);
                        break;
                    default:
                        return 0;
                }
                
                if (this.sortOrder === 'asc') {
                    return aVal < bVal ? -1 : aVal > bVal ? 1 : 0;
                } else {
                    return aVal > bVal ? -1 : aVal < bVal ? 1 : 0;
                }
            });
            
            this.filteredRiskObjects = filtered;
        },
        
        setSortBy(field) {
            if (this.sortBy === field) {
                this.sortOrder = this.sortOrder === 'asc' ? 'desc' : 'asc';
            } else {
                this.sortBy = field;
                this.sortOrder = 'desc';
            }
            this.applySortingAndFiltering();
        },
        
        getRiskLevelClass(score) {
            if (score >= 80) return 'risk-critical';
            if (score >= 60) return 'risk-high';
            if (score >= 40) return 'risk-medium';
            if (score >= 20) return 'risk-low';
            return 'risk-minimal';
        },
        
        getRiskLevelText(score) {
            if (score >= 80) return 'Critical';
            if (score >= 60) return 'High';
            if (score >= 40) return 'Medium';
            if (score >= 20) return 'Low';
            return 'Minimal';
        },
        
        getEntityTypeIcon(entityType) {
            switch (entityType) {
                case 'user': return '👤';
                case 'host': return '💻';
                case 'ip': return '🌐';
                default: return '📍';
            }
        },
        
        formatTimestamp(timestamp) {
            return UIUtils.formatDate(timestamp);
        },
        
        getSortIcon(field) {
            if (this.sortBy !== field) return '↕️';
            return this.sortOrder === 'asc' ? '↑' : '↓';
        },
        
        viewRiskObject(objectId) {
            UIUtils.navigateTo(`risk-objects-detail.html?id=${objectId}`);
        },
        
        initChart() {
            this.$nextTick(() => {
                const ctx = document.getElementById('riskDistributionChart');
                if (ctx && this.riskObjects.length > 0) {
                    // Calculate risk distribution
                    const distribution = {
                        'Critical (80+)': 0,
                        'High (60-79)': 0,
                        'Medium (40-59)': 0,
                        'Low (20-39)': 0,
                        'Minimal (0-19)': 0
                    };
                    
                    this.riskObjects.forEach(obj => {
                        const score = obj.current_score;
                        if (score >= 80) distribution['Critical (80+)']++;
                        else if (score >= 60) distribution['High (60-79)']++;
                        else if (score >= 40) distribution['Medium (40-59)']++;
                        else if (score >= 20) distribution['Low (20-39)']++;
                        else distribution['Minimal (0-19)']++;
                    });
                    
                    new Chart(ctx, {
                        type: 'doughnut',
                        data: {
                            labels: Object.keys(distribution),
                            datasets: [{
                                data: Object.values(distribution),
                                backgroundColor: [
                                    '#dc3545', // Critical - red
                                    '#fd7e14', // High - orange
                                    '#ffc107', // Medium - yellow
                                    '#28a745', // Low - green
                                    '#6c757d'  // Minimal - gray
                                ]
                            }]
                        },
                        options: {
                            responsive: true,
                            maintainAspectRatio: false,
                            plugins: {
                                title: {
                                    display: true,
                                    text: 'Risk Level Distribution'
                                },
                                legend: {
                                    position: 'right'
                                }
                            }
                        }
                    });
                }
            });
        },
        
        // Watchers for filter changes
        initWatchers() {
            this.$nextTick(() => {
                this.$watch('entityTypeFilter', () => {
                    this.applySortingAndFiltering();
                });
                
                this.$watch('searchTerm', () => {
                    this.applySortingAndFiltering();
                });
                
                this.$watch('riskThreshold', async () => {
                    await this.fetchRiskObjects();
                });
            });
        }
    };
}
