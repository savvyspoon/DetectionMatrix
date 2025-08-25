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
        
        // False positive dialog state
        showFPDialog: false,
        selectedEventId: null,
        fpAnalyst: '',
        fpReasonType: '',
        fpCustomReason: '',
        
        async init() {
            await this.fetchEvents();
            this.initWatchers();
        },
        
        async fetchEvents(page = 1) {
            this.loading = true;
            try {
                const data = await EventsAPI.fetchEvents(page, this.limit);
                // Standard list envelope: { items, page, page_size, total }
                this.events = data.items || [];
                this.currentPage = data.page || 1;
                const pageSize = data.page_size || this.limit;
                this.totalCount = data.total || (this.events ? this.events.length : 0);
                this.totalPages = Math.max(1, Math.ceil(this.totalCount / pageSize));
                this.hasNext = (this.currentPage * pageSize) < this.totalCount;
                this.hasPrev = this.currentPage > 1;
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
                const detectionsResp = await EventsAPI.fetchDetections();
                this.detections = {};
                (detectionsResp.items || detectionsResp || []).forEach(d => {
                    this.detections[d.id] = d;
                });
                
                // Fetch risk objects for lookup
                const riskObjectsResp = await EventsAPI.fetchRiskObjects();
                this.riskObjects = {};
                (riskObjectsResp.items || riskObjectsResp || []).forEach(ro => {
                    this.riskObjects[ro.id] = ro;
                });
            } catch (error) {
                console.error('Error fetching related data:', error);
            }
        },
        
        applyFilters() {
            this.filteredEvents = this.events.filter(event => {
                const matchesFalsePositive = this.falsePositiveFilter === 'all' || 
                    (this.falsePositiveFilter === 'false' && !event.is_false_positive) ||
                    (this.falsePositiveFilter === 'true' && event.is_false_positive);
                
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
        
        showFalsePositiveDialog(eventId) {
            this.selectedEventId = eventId;
            this.fpAnalyst = '';
            this.fpReasonType = '';
            this.fpCustomReason = '';
            this.showFPDialog = true;
        },
        
        closeFPDialog() {
            this.showFPDialog = false;
            this.selectedEventId = null;
            this.fpAnalyst = '';
            this.fpReasonType = '';
            this.fpCustomReason = '';
        },
        
        async confirmMarkAsFalsePositive() {
            if (!this.fpAnalyst || !this.fpReasonType) {
                UIUtils.showAlert('Please fill in all required fields', 'error');
                return;
            }
            
            // Determine the final reason text
            let reason = this.fpReasonType;
            if (this.fpReasonType === 'Other' && this.fpCustomReason) {
                reason = this.fpCustomReason;
            }
            
            try {
                const response = await fetch(`/api/events/${this.selectedEventId}/false-positive`, {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    body: JSON.stringify({
                        reason: reason,
                        analyst_name: this.fpAnalyst
                    })
                });
                
                if (response.ok) {
                    // Find the event and update its status
                    const event = this.events.find(e => e.id === this.selectedEventId);
                    if (event) {
                        event.is_false_positive = true;
                    }
                    this.applyFilters();
                    this.closeFPDialog();
                    UIUtils.showAlert('Event marked as false positive', 'success');
                } else {
                    throw new Error(`Failed to mark event as false positive: ${response.status}`);
                }
            } catch (error) {
                console.error('Error marking event as false positive:', error);
                UIUtils.showAlert('Failed to mark event as false positive', 'error');
            }
        },
        
        getEntityValue(entityId) {
            // Alias for getRiskObjectValue to maintain compatibility
            return this.getRiskObjectValue(entityId);
        },
        
        getPageNumbers() {
            const pages = [];
            const maxVisible = 5;
            let start = Math.max(1, this.currentPage - Math.floor(maxVisible / 2));
            let end = Math.min(this.totalPages, start + maxVisible - 1);
            
            // Adjust start if we're near the end
            if (end - start < maxVisible - 1) {
                start = Math.max(1, end - maxVisible + 1);
            }
            
            for (let i = start; i <= end; i++) {
                pages.push(i);
            }
            
            return pages;
        },
        
        onFilterChange() {
            // Reset to first page when filters change
            this.currentPage = 1;
            this.applyFilters();
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
