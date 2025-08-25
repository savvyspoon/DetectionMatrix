// Risk Alerts JavaScript functionality
class RiskAlertsAPI {
    static async fetchAlerts(page = 1, pageSize = 20, statusFilter = '') {
        const params = new URLSearchParams();
        params.append('page', page.toString());
        params.append('limit', pageSize.toString());
        
        if (statusFilter) {
            params.append('status', statusFilter);
        }
        
        const url = `/api/risk/alerts?${params.toString()}`;
        return await APIUtils.fetchAPI(url);
    }

    static async fetchRiskObjects() {
        return await APIUtils.fetchAPI('/api/risk/objects');
    }

    static async updateAlertStatus(alertId, status) {
        return await APIUtils.postAPI(`/api/risk/alerts/${alertId}`, { status });
    }
}

// Alpine.js risk alerts data function
function riskAlertsData() {
    return {
        alerts: [],
        filteredAlerts: [],
        riskObjects: {},
        searchTerm: '',
        statusFilter: '',
        sortBy: 'triggered_at',
        sortOrder: 'desc',
        hideClosed: false,
        
        // Pagination state
        currentPage: 1,
        pageSize: 20,
        totalCount: 0,
        totalPages: 0,
        isPaginated: false,
        
        async init() {
            await this.fetchAlerts();
            this.initWatchers();
        },
        
        async fetchAlerts() {
            try {
                const data = await RiskAlertsAPI.fetchAlerts(this.currentPage, this.pageSize, this.statusFilter);
                // Standard list envelope
                this.alerts = data.items || [];
                this.totalCount = data.total || (this.alerts ? this.alerts.length : 0);
                this.currentPage = data.page || this.currentPage;
                const pageSize = data.page_size || this.pageSize;
                this.totalPages = Math.max(1, Math.ceil(this.totalCount / pageSize));
                this.isPaginated = true;
                
                await this.fetchRiskObjects();
                this.applySortingAndFiltering();
            } catch (error) {
                console.error('Error fetching alerts:', error);
                UIUtils.showAlert('Error loading risk alerts');
            }
        },
        
        async fetchRiskObjects() {
            try {
                const riskObjectsResp = await RiskAlertsAPI.fetchRiskObjects();
                this.riskObjects = {};
                (riskObjectsResp.items || riskObjectsResp || []).forEach(ro => {
                    this.riskObjects[ro.id] = ro;
                });
            } catch (error) {
                console.error('Error fetching risk objects:', error);
            }
        },
        
        applySortingAndFiltering() {
            let filtered = [...this.alerts];
            
            // Apply search filter
            if (this.searchTerm) {
                const searchLower = this.searchTerm.toLowerCase();
                filtered = filtered.filter(alert => {
                    const riskObject = this.riskObjects[alert.entity_id];
                    const entityValue = riskObject ? riskObject.entity_value : alert.entity_id.toString();
                    return entityValue.toLowerCase().includes(searchLower) ||
                           alert.status.toLowerCase().includes(searchLower);
                });
            }
            
            // Hide closed alerts if checkbox is checked
            if (this.hideClosed) {
                filtered = filtered.filter(alert => alert.status.toLowerCase() !== 'closed');
            }
            
            // Apply sorting with closed alerts always at bottom (unless hidden)
            filtered.sort((a, b) => {
                // First, check if either alert is closed (only if not hiding closed)
                if (!this.hideClosed) {
                    const aIsClosed = a.status.toLowerCase() === 'closed';
                    const bIsClosed = b.status.toLowerCase() === 'closed';
                    
                    // If one is closed and the other isn't, closed goes to bottom
                    if (aIsClosed && !bIsClosed) return 1;
                    if (!aIsClosed && bIsClosed) return -1;
                }
                
                // If both are closed or both are not closed, apply normal sorting
                let aVal = a[this.sortBy];
                let bVal = b[this.sortBy];
                
                // Handle timestamps
                if (this.sortBy === 'triggered_at' || this.sortBy === 'resolved_at') {
                    aVal = new Date(aVal);
                    bVal = new Date(bVal);
                }
                
                if (this.sortOrder === 'asc') {
                    return aVal < bVal ? -1 : aVal > bVal ? 1 : 0;
                } else {
                    return aVal > bVal ? -1 : aVal < bVal ? 1 : 0;
                }
            });
            
            this.filteredAlerts = filtered;
        },
        
        async updateStatus(alert, newStatus) {
            try {
                await RiskAlertsAPI.updateAlertStatus(alert.id, newStatus);
                alert.status = newStatus;
                if (newStatus === 'resolved') {
                    alert.resolved_at = new Date().toISOString();
                }
                this.applySortingAndFiltering();
            } catch (error) {
                console.error('Error updating alert status:', error);
                UIUtils.showAlert('Error updating alert status');
            }
        },
        
        async changePage(newPage) {
            if (newPage >= 1 && newPage <= this.totalPages) {
                this.currentPage = newPage;
                await this.fetchAlerts();
            }
        },
        
        async nextPage() {
            if (this.currentPage < this.totalPages) {
                await this.changePage(this.currentPage + 1);
            }
        },
        
        async previousPage() {
            if (this.currentPage > 1) {
                await this.changePage(this.currentPage - 1);
            }
        },
        
        async goToPage(page) {
            await this.changePage(page);
        },
        
        async changePageSize() {
            this.currentPage = 1;
            await this.fetchAlerts();
        },
        
        getPageNumbers() {
            const pages = [];
            const maxPagesToShow = 5;
            const startPage = Math.max(1, this.currentPage - Math.floor(maxPagesToShow / 2));
            const endPage = Math.min(this.totalPages, startPage + maxPagesToShow - 1);
            
            for (let i = startPage; i <= endPage; i++) {
                pages.push(i);
            }
            return pages;
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
        
        getRiskObjectValue(entityId) {
            const riskObject = this.riskObjects[entityId];
            return riskObject ? `${riskObject.entity_value} (${riskObject.entity_type})` : `Entity ${entityId}`;
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
        
        formatTimestamp(timestamp) {
            return timestamp ? UIUtils.formatDate(timestamp) : 'N/A';
        },
        
        getStatusClass(status) {
            const statusLower = status.toLowerCase().replace(/\s+/g, '-');
            switch (statusLower) {
                case 'new': return 'status-new';
                case 'triage': return 'status-triage';
                case 'investigation': return 'status-investigation';
                case 'on-hold': return 'status-on-hold';
                case 'incident': return 'status-incident';
                case 'closed': return 'status-closed';
                default: return 'status-unknown';
            }
        },
        
        getSortIcon(field) {
            if (this.sortBy !== field) return '↕️';
            return this.sortOrder === 'asc' ? '↑' : '↓';
        },
        
        viewAlert(alertId) {
            UIUtils.navigateTo(`risk-alerts-detail.html?id=${alertId}`);
        },
        
        // Watchers for filter changes
        initWatchers() {
            this.$nextTick(() => {
                this.$watch('searchTerm', () => {
                    this.applySortingAndFiltering();
                });
                
                this.$watch('statusFilter', async () => {
                    this.currentPage = 1;
                    await this.fetchAlerts();
                });
                
                this.$watch('hideClosed', () => {
                    this.applySortingAndFiltering();
                });
            });
        }
    };
}
