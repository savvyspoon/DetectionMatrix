// Detections page business logic
function detectionsData() {
    return {
        detections: [],
        filteredDetections: [],
        dataSources: [],
        mitreTechniques: [],
        statusFilter: 'all',
        severityFilter: 'all',
        searchTerm: '',
        showForm: false,
        selectedDetection: null,
        selectedDataSource: null,
        selectedMitreTechnique: null,
        newDetection: {
            name: '',
            description: '',
            status: 'idea',
            severity: 'medium',
            risk_points: 50,
            playbook_link: '',
            owner: ''
        },
        
        async init() {
            await Promise.all([
                this.fetchDetections(),
                this.fetchDataSources(),
                this.fetchMitreTechniques()
            ]);
            this.initWatchers();
        },
        
        async fetchDetections() {
            try {
                const detections = await DetectionsAPI.fetchDetections();
                this.detections = detections;
                this.applyFilters();
            } catch (error) {
                console.error('Error fetching detections:', error);
                UIUtils.showAlert('Error loading detections');
            }
        },
        
        async fetchDataSources() {
            try {
                this.dataSources = await DataSourcesAPI.fetchDataSources();
            } catch (error) {
                console.error('Error fetching data sources:', error);
            }
        },
        
        async fetchMitreTechniques() {
            try {
                this.mitreTechniques = await MitreAPI.fetchTechniques();
            } catch (error) {
                console.error('Error fetching MITRE techniques:', error);
            }
        },
        
        applyFilters() {
            this.filteredDetections = this.detections.filter(detection => {
                const matchesStatus = this.statusFilter === 'all' || detection.status === this.statusFilter;
                const matchesSeverity = this.severityFilter === 'all' || detection.severity === this.severityFilter;
                const matchesSearch = this.searchTerm === '' || 
                    detection.name.toLowerCase().includes(this.searchTerm.toLowerCase()) || 
                    (detection.description && detection.description.toLowerCase().includes(this.searchTerm.toLowerCase()));
                return matchesStatus && matchesSeverity && matchesSearch;
            });
        },
        
        async createDetection() {
            try {
                await DetectionsAPI.createDetection(this.newDetection);
                this.showForm = false;
                this.resetNewDetection();
                await this.fetchDetections();
                UIUtils.showAlert('Detection created successfully', 'success');
            } catch (error) {
                console.error('Error creating detection:', error);
                UIUtils.showAlert('Error creating detection');
            }
        },
        
        async updateDetection() {
            try {
                await DetectionsAPI.updateDetection(this.selectedDetection.id, this.selectedDetection);
                this.selectedDetection = null;
                await this.fetchDetections();
                UIUtils.showAlert('Detection updated successfully', 'success');
            } catch (error) {
                console.error('Error updating detection:', error);
                UIUtils.showAlert('Error updating detection');
            }
        },
        
        async deleteDetection(id) {
            if (!confirm('Are you sure you want to delete this detection?')) {
                return;
            }
            
            try {
                await DetectionsAPI.deleteDetection(id);
                this.selectedDetection = null;
                await this.fetchDetections();
                UIUtils.showAlert('Detection deleted successfully', 'success');
            } catch (error) {
                console.error('Error deleting detection:', error);
                UIUtils.showAlert('Error deleting detection');
            }
        },
        
        async selectDetection(id) {
            try {
                this.selectedDetection = await DetectionsAPI.fetchDetection(id);
            } catch (error) {
                console.error('Error fetching detection details:', error);
                UIUtils.showAlert('Error loading detection details');
            }
        },
        
        async addDataSource(detectionId, dataSourceId) {
            if (!dataSourceId) return;
            
            try {
                await DetectionsAPI.addDataSource(detectionId, dataSourceId);
                await this.selectDetection(detectionId);
                this.selectedDataSource = null;
                UIUtils.showAlert('Data source added successfully', 'success');
            } catch (error) {
                console.error('Error adding data source:', error);
                UIUtils.showAlert('Error adding data source');
            }
        },
        
        async removeDataSource(detectionId, dataSourceId) {
            if (!confirm('Are you sure you want to remove this data source?')) {
                return;
            }
            
            try {
                await DetectionsAPI.removeDataSource(detectionId, dataSourceId);
                await this.selectDetection(detectionId);
                UIUtils.showAlert('Data source removed successfully', 'success');
            } catch (error) {
                console.error('Error removing data source:', error);
                UIUtils.showAlert('Error removing data source');
            }
        },
        
        async addMitreTechnique(detectionId, techniqueId) {
            if (!techniqueId) return;
            
            try {
                await DetectionsAPI.addMitreTechnique(detectionId, techniqueId);
                await this.selectDetection(detectionId);
                this.selectedMitreTechnique = null;
                UIUtils.showAlert('MITRE technique added successfully', 'success');
            } catch (error) {
                console.error('Error adding MITRE technique:', error);
                UIUtils.showAlert('Error adding MITRE technique');
            }
        },
        
        async removeMitreTechnique(detectionId, techniqueId) {
            if (!confirm('Are you sure you want to remove this MITRE technique?')) {
                return;
            }
            
            try {
                await DetectionsAPI.removeMitreTechnique(detectionId, techniqueId);
                await this.selectDetection(detectionId);
                UIUtils.showAlert('MITRE technique removed successfully', 'success');
            } catch (error) {
                console.error('Error removing MITRE technique:', error);
                UIUtils.showAlert('Error removing MITRE technique');
            }
        },
        
        resetNewDetection() {
            this.newDetection = {
                name: '',
                description: '',
                status: 'idea',
                severity: 'medium',
                risk_points: 50,
                playbook_link: '',
                owner: ''
            };
        },
        
        cancelForm() {
            this.showForm = false;
            this.resetNewDetection();
        },
        
        cancelEdit() {
            this.selectedDetection = null;
            this.selectedDataSource = null;
            this.selectedMitreTechnique = null;
        },
        
        getStatusClass(status) {
            return `status status-${status}`;
        },
        
        getSeverityClass(severity) {
            const severityMap = {
                'low': 'badge-secondary',
                'medium': 'badge-warning', 
                'high': 'badge-danger',
                'critical': 'badge-danger'
            };
            return `badge ${severityMap[severity] || 'badge-secondary'}`;
        },
        
        viewDetection(detectionId) {
            UIUtils.navigateTo(`detections-detail.html?id=${detectionId}`);
        },
        
        initWatchers() {
            this.$nextTick(() => {
                this.$watch('statusFilter', () => this.applyFilters());
                this.$watch('severityFilter', () => this.applyFilters());
                this.$watch('searchTerm', () => this.applyFilters());
            });
        }
    };
}