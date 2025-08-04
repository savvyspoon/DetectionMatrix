// Events List JavaScript functionality
class EventsAPI {
    static async fetchEvents(page = 1, limit = 20) {
        return await APIUtils.fetchAPI(`/api/events?page=${page}&limit=${limit}`);
    }

    static async fetchDetections() {
        return await APIUtils.fetchAPI('/api/detections');
    }

    static async fetchRiskObjects() {
        return await APIUtils.fetchAPI('/api/risk/objects');
    }
}

// Alpine.js events list data function
function eventsListData() {
    return {
        events: [],
        filteredEvents: [],
        detections: {},
        riskObjects: {},
        falsePositiveFilter: 'all',
        searchTerm: '',
        currentPage: 1,
        limit: 20,
        totalPages: 1,
        totalCount: 0,
        hasNext: false,
        hasPrev: false,
        loading: false,
        
        async init() {
            await this.fetchEvents();
            this.initWatchers();
        },
        
        async fetchEvents(page = 1) {
            this.loading = true;
            try {
                const data = await EventsAPI.fetchEvents(page, this.limit);
                this.events = data.events || [];
                this.currentPage = data.pagination.page;
                this.totalPages = data.pagination.total_pages;
                this.totalCount = data.pagination.total_count;
                this.hasNext = data.pagination.has_next;
                this.hasPrev = data.pagination.has_prev;
                await this.fetchRelatedData();
                this.applyFilters();
            } catch (error) {
                console.error('Error fetching events:', error);
                UIUtils.showAlert('Error loading events');
            }
            this.loading = false;
        },
        
        async fetchRelatedData() {
            try {
                // Fetch detections for lookup
                const detectionsList = await EventsAPI.fetchDetections();
                this.detections = {};
                detectionsList.forEach(d => {
                    this.detections[d.id] = d;
                });
                
                // Fetch risk objects for lookup
                const riskObjectsList = await EventsAPI.fetchRiskObjects();
                this.riskObjects = {};
                riskObjectsList.forEach(ro => {
                    this.riskObjects[ro.id] = ro;
                });
            } catch (error) {
                console.error('Error fetching related data:', error);
            }
        },
        
        applyFilters() {
            this.filteredEvents = this.events.filter(event => {
                const matchesFalsePositive = this.falsePositiveFilter === 'all' || 
                    (this.falsePositiveFilter === 'valid' && !event.is_false_positive) ||
                    (this.falsePositiveFilter === 'false_positive' && event.is_false_positive);
                
                const matchesSearch = this.searchTerm === '' || 
                    event.raw_data?.toLowerCase().includes(this.searchTerm.toLowerCase()) ||
                    this.getDetectionName(event.detection_id).toLowerCase().includes(this.searchTerm.toLowerCase()) ||
                    this.getRiskObjectValue(event.entity_id).toLowerCase().includes(this.searchTerm.toLowerCase());
                
                return matchesFalsePositive && matchesSearch;
            });
        },
        
        async nextPage() {
            if (this.hasNext) {
                await this.fetchEvents(this.currentPage + 1);
            }
        },
        
        async prevPage() {
            if (this.hasPrev) {
                await this.fetchEvents(this.currentPage - 1);
            }
        },
        
        async goToPage(page) {
            if (page >= 1 && page <= this.totalPages) {
                await this.fetchEvents(page);
            }
        },
        
        getDetectionName(detectionId) {
            return this.detections[detectionId]?.name || `Detection ${detectionId}`;
        },
        
        getRiskObjectValue(entityId) {
            const riskObject = this.riskObjects[entityId];
            return riskObject ? `${riskObject.entity_value} (${riskObject.entity_type})` : `Entity ${entityId}`;
        },
        
        formatTimestamp(timestamp) {
            return UIUtils.formatDate(timestamp);
        },
        
        getFalsePositiveClass(isFalsePositive) {
            return isFalsePositive ? 'badge badge-warning' : 'badge badge-success';
        },
        
        getFalsePositiveText(isFalsePositive) {
            return isFalsePositive ? 'False Positive' : 'Valid';
        },
        
        viewEvent(eventId) {
            UIUtils.navigateTo(`events-detail.html?id=${eventId}`);
        },
        
        // Watchers for filter changes
        initWatchers() {
            this.$nextTick(() => {
                this.$watch('falsePositiveFilter', () => {
                    this.applyFilters();
                });
                
                this.$watch('searchTerm', () => {
                    this.applyFilters();
                });
            });
        }
    };
}