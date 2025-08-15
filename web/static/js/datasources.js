// Data Sources JavaScript functionality
class DataSourcesAPI {
    static async fetchDataSources() {
        return await APIUtils.fetchAPI('/api/datasources');
    }

    static async createDataSource(dataSource) {
        return await APIUtils.postAPI('/api/datasources', dataSource);
    }

    static async updateDataSource(id, dataSource) {
        return await APIUtils.postAPI(`/api/datasources/${id}`, dataSource);
    }

    static async deleteDataSource(id) {
        return await APIUtils.deleteAPI(`/api/datasources/${id}`);
    }

    static async fetchDetectionsByDataSource(id) {
        return await APIUtils.fetchAPI(`/api/datasources/id/${id}/detections`);
    }

    static async fetchTechniquesByDataSource(id) {
        return await APIUtils.fetchAPI(`/api/datasources/id/${id}/techniques`);
    }
}

// Alpine.js data sources data function
function dataSourcesData() {
    return {
        dataSources: [],
        filteredDataSources: [],
        searchTerm: '',
        showForm: false,
        newDataSource: {
            name: '',
            description: '',
            log_format: ''
        },
        selectedDataSource: null,
        associatedDetections: [],
        associatedTechniques: [],
        utilization: {},
        
        async init() {
            await this.fetchDataSources();
            await this.fetchUtilization();
            this.initWatchers();
            this.initChart();
        },
        
        async fetchDataSources() {
            try {
                this.dataSources = await DataSourcesAPI.fetchDataSources();
                this.applyFilters();
            } catch (error) {
                console.error('Error fetching data sources:', error);
                UIUtils.showAlert('Error loading data sources');
            }
        },
        
        applyFilters() {
            let filtered = [...this.dataSources];
            
            if (this.searchTerm) {
                const searchLower = this.searchTerm.toLowerCase();
                filtered = filtered.filter(ds => {
                    const name = (ds.name || '').toLowerCase();
                    const description = (ds.description || '').toLowerCase();
                    const logFormat = (ds.log_format || '').toLowerCase();
                    return name.includes(searchLower) ||
                           description.includes(searchLower) ||
                           logFormat.includes(searchLower);
                });
            }
            
            this.filteredDataSources = filtered;
        },
        
        async createDataSource() {
            try {
                await DataSourcesAPI.createDataSource(this.newDataSource);
                await this.fetchDataSources();
                this.resetForm();
            } catch (error) {
                console.error('Error creating data source:', error);
                UIUtils.showAlert('Error creating data source');
            }
        },
        
        async updateDataSource() {
            if (!this.selectedDataSource) return;
            
            try {
                await DataSourcesAPI.updateDataSource(this.selectedDataSource.id, this.selectedDataSource);
                await this.fetchDataSources();
                this.selectedDataSource = null;
            } catch (error) {
                console.error('Error updating data source:', error);
                UIUtils.showAlert('Error updating data source');
            }
        },
        
        async deleteDataSource(id) {
            if (!UIUtils.confirmAction('Are you sure you want to delete this data source?')) {
                return;
            }
            
            try {
                await DataSourcesAPI.deleteDataSource(id);
                await this.fetchDataSources();
            } catch (error) {
                console.error('Error deleting data source:', error);
                UIUtils.showAlert('Error deleting data source');
            }
        },
        
        async selectDataSource(id) {
            const dataSource = this.dataSources.find(ds => ds.id === id);
            if (dataSource) {
                // Ensure all required properties exist with safe defaults
                this.selectedDataSource = {
                    id: dataSource.id,
                    name: dataSource.name || '',
                    description: dataSource.description || '',
                    log_format: dataSource.log_format || ''
                };
                await Promise.all([
                    this.fetchDetectionsByDataSource(id),
                    this.fetchTechniquesByDataSource(id)
                ]);
            } else {
                console.error('Data source not found:', id);
                this.selectedDataSource = null;
            }
        },
        
        async fetchDetectionsByDataSource(id) {
            try {
                this.associatedDetections = await DataSourcesAPI.fetchDetectionsByDataSource(id);
            } catch (error) {
                console.error('Error fetching detections:', error);
                this.associatedDetections = [];
            }
        },
        
        async fetchTechniquesByDataSource(id) {
            try {
                this.associatedTechniques = await DataSourcesAPI.fetchTechniquesByDataSource(id);
            } catch (error) {
                console.error('Error fetching techniques:', error);
                this.associatedTechniques = [];
            }
        },
        
        async fetchUtilization() {
            try {
                // Calculate utilization from data sources
                const utilization = {};
                this.dataSources.forEach(ds => {
                    utilization[ds.name] = Math.floor(Math.random() * 10); // Placeholder
                });
                this.utilization = utilization;
            } catch (error) {
                console.error('Error calculating utilization:', error);
                this.utilization = {};
            }
        },
        
        resetForm() {
            this.newDataSource = {
                name: '',
                description: '',
                log_format: ''
            };
            this.showForm = false;
        },
        
        closeModal() {
            this.selectedDataSource = null;
            this.associatedDetections = [];
            this.associatedTechniques = [];
        },
        
        initChart() {
            this.$nextTick(() => {
                const ctx = document.getElementById('utilizationChart');
                if (ctx && Object.keys(this.utilization).length > 0) {
                    new Chart(ctx, {
                        type: 'bar',
                        data: {
                            labels: Object.keys(this.utilization),
                            datasets: [{
                                label: 'Detections Using Data Source',
                                data: Object.values(this.utilization),
                                backgroundColor: 'rgba(33, 150, 243, 0.8)'
                            }]
                        },
                        options: {
                            responsive: true,
                            maintainAspectRatio: false,
                            plugins: {
                                title: {
                                    display: true,
                                    text: 'Data Source Utilization'
                                }
                            }
                        }
                    });
                }
            });
        },
        
        viewDetection(detectionId) {
            UIUtils.navigateTo(`detections-detail.html?id=${detectionId}`);
        },
        
        // Watchers for filter changes
        initWatchers() {
            this.$nextTick(() => {
                this.$watch('searchTerm', () => {
                    this.applyFilters();
                });
            });
        }
    };
}