// Detections List JavaScript functionality
class DetectionsAPI {
    static async fetchDetections() {
        return await APIUtils.fetchAPI('/api/detections');
    }

    static async fetchDetectionClasses() {
        return await APIUtils.fetchAPI('/api/detection-classes');
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
        detectionClasses: [],
        statusFilter: 'all',
        severityFilter: 'all',
        classFilter: 'all',
        searchTerm: '',
        
        async init() {
            await Promise.all([
                this.fetchDetections(),
                this.fetchDetectionClasses()
            ]);
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
        
        async fetchDetectionClasses() {
            try {
                this.detectionClasses = await DetectionsAPI.fetchDetectionClasses();
            } catch (error) {
                console.error('Error fetching detection classes:', error);
                // Continue without classes if there's an error
                this.detectionClasses = [];
            }
        },
        
        applyFilters() {
            let filtered = [...this.detections];
            
            // Apply status filter
            if (this.statusFilter !== 'all') {
                filtered = filtered.filter(d => d.status === this.statusFilter);
            }
            
            // Apply severity filter
            if (this.severityFilter !== 'all') {
                filtered = filtered.filter(d => d.severity === this.severityFilter);
            }
            
            // Apply class filter
            if (this.classFilter !== 'all') {
                const classId = parseInt(this.classFilter);
                filtered = filtered.filter(d => d.class_id === classId);
            }
            
            // Apply search filter
            if (this.searchTerm) {
                const search = this.searchTerm.toLowerCase();
                filtered = filtered.filter(d => 
                    d.name?.toLowerCase().includes(search) ||
                    d.description?.toLowerCase().includes(search) ||
                    d.owner?.toLowerCase().includes(search) ||
                    d.class?.name?.toLowerCase().includes(search)
                );
            }
            
            this.filteredDetections = filtered;
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
        
        watchClassFilter() {
            this.$watch('classFilter', () => {
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
                this.watchClassFilter();
                this.watchSearchTerm();
            });
        }
    };
}