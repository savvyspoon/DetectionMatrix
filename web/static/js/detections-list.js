// Detections List JavaScript functionality
class DetectionsAPI {
    static async fetchDetections() {
        return await APIUtils.fetchAPI('/api/detections');
    }

    static async deleteDetection(id) {
        return await APIUtils.deleteAPI(`/api/detections/${id}`);
    }
}

// Alpine.js detections list data function
function detectionsListData() {
    return {
        detections: [],
        filteredDetections: [],
        statusFilter: 'all',
        severityFilter: 'all',
        searchTerm: '',
        
        async init() {
            await this.fetchDetections();
            this.initWatchers();
        },
        
        async fetchDetections() {
            try {
                this.detections = await DetectionsAPI.fetchDetections();
                this.applyFilters();
            } catch (error) {
                console.error('Error fetching detections:', error);
                UIUtils.showAlert('Error loading detections');
            }
        },
        
        applyFilters() {
            const filters = {
                status: this.statusFilter,
                severity: this.severityFilter,
                search: this.searchTerm
            };
            
            this.filteredDetections = FilterUtils.filterItems(this.detections, filters);
        },
        
        async deleteDetection(id) {
            if (!UIUtils.confirmAction('Are you sure you want to delete this detection?')) {
                return;
            }
            
            try {
                await DetectionsAPI.deleteDetection(id);
                await this.fetchDetections();
            } catch (error) {
                console.error('Error deleting detection:', error);
                UIUtils.showAlert('Error deleting detection');
            }
        },
        
        getStatusClass(status) {
            return UIUtils.getStatusClass(status);
        },
        
        getSeverityClass(severity) {
            return UIUtils.getSeverityClass(severity);
        },
        
        // Watchers to trigger filtering when values change
        watchStatusFilter() {
            this.$watch('statusFilter', () => {
                this.applyFilters();
            });
        },

        watchSeverityFilter() {
            this.$watch('severityFilter', () => {
                this.applyFilters();
            });
        },

        watchSearchTerm() {
            this.$watch('searchTerm', () => {
                this.applyFilters();
            });
        },

        initWatchers() {
            this.$nextTick(() => {
                this.watchStatusFilter();
                this.watchSeverityFilter();  
                this.watchSearchTerm();
            });
        }
    };
}