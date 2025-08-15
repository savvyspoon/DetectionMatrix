// Risk Alerts Detail JavaScript functionality
class RiskAlertDetailAPI {
    static async fetchAlert(alertId) {
        return await APIUtils.fetchAPI(`/api/risk/alerts/${alertId}`);
    }

    static async fetchRiskObject(entityId) {
        return await APIUtils.fetchAPI(`/api/risk/objects/${entityId}`);
    }

    static async fetchEventsForAlert(alertId) {
        return await APIUtils.fetchAPI(`/api/risk/alerts/${alertId}/events`);
    }

    static async fetchDetections() {
        return await APIUtils.fetchAPI('/api/detections');
    }

    static async updateAlertStatus(alertId, status) {
        return await APIUtils.postAPI(`/api/risk/alerts/${alertId}`, { status });
    }
}

// Alpine.js risk alert detail data function
function riskAlertDetailData() {
    return {
        alert: null,
        riskObject: null,
        events: [],
        detections: {},
        loading: true,
        alertId: null,
        
        // Form data properties for two-way binding
        formData: {
            status: 'New',
            owner: '',
            notes: ''
        },
        
        async init() {
            const urlParams = new URLSearchParams(window.location.search);
            this.alertId = urlParams.get('id');
            if (this.alertId) {
                await this.fetchAlertDetails();
            } else {
                UIUtils.showAlert('No alert ID provided');
                UIUtils.navigateTo('risk-alerts.html');
            }
        },
        
        async fetchAlertDetails() {
            this.loading = true;
            try {
                // Fetch alert details
                this.alert = await RiskAlertDetailAPI.fetchAlert(this.alertId);
                
                // Initialize form data with alert values
                this.initializeFormData();
                
                await Promise.all([
                    this.fetchRiskObject(),
                    this.fetchEvents(),
                    this.fetchDetections()
                ]);
            } catch (error) {
                console.error('Error fetching alert details:', error);
                UIUtils.showAlert('Error loading alert details');
                UIUtils.navigateTo('risk-alerts.html');
            }
            this.loading = false;
        },
        
        initializeFormData() {
            if (this.alert) {
                this.formData.status = this.alert.status || 'New';
                this.formData.owner = this.alert.owner || '';
                this.formData.notes = this.alert.notes || '';
            }
        },
        
        async fetchRiskObject() {
            if (!this.alert?.entity_id) return;
            
            try {
                this.riskObject = await RiskAlertDetailAPI.fetchRiskObject(this.alert.entity_id);
            } catch (error) {
                console.error('Error fetching risk object:', error);
            }
        },
        
        async fetchEvents() {
            try {
                this.events = await RiskAlertDetailAPI.fetchEventsForAlert(this.alertId);
            } catch (error) {
                console.error('Error fetching events:', error);
                this.events = [];
            }
        },
        
        async fetchDetections() {
            try {
                const detectionsList = await RiskAlertDetailAPI.fetchDetections();
                this.detections = {};
                detectionsList.forEach(d => {
                    this.detections[d.id] = d;
                });
            } catch (error) {
                console.error('Error fetching detections:', error);
            }
        },
        
        async updateStatus(newStatus) {
            try {
                await RiskAlertDetailAPI.updateAlertStatus(this.alertId, newStatus);
                this.alert.status = newStatus;
                if (newStatus === 'resolved') {
                    this.alert.resolved_at = new Date().toISOString();
                }
            } catch (error) {
                console.error('Error updating alert status:', error);
                UIUtils.showAlert('Error updating alert status');
            }
        },
        
        getDetectionName(detectionId) {
            return this.detections[detectionId]?.name || `Detection ${detectionId}`;
        },
        
        formatTimestamp(timestamp) {
            return timestamp ? UIUtils.formatDate(timestamp) : 'N/A';
        },
        
        getStatusClass(status) {
            switch (status) {
                case 'open': return 'status-open';
                case 'investigating': return 'status-investigating';
                case 'resolved': return 'status-resolved';
                default: return 'status-unknown';
            }
        },
        
        getFalsePositiveClass(isFalsePositive) {
            return isFalsePositive ? 'badge badge-warning' : 'badge badge-success';
        },
        
        getFalsePositiveText(isFalsePositive) {
            return isFalsePositive ? 'False Positive' : 'Valid';
        },
        
        // Risk level utility functions
        getRiskLevelClass(score) {
            if (!score) return 'risk-low';
            if (score >= 100) return 'risk-critical';
            if (score >= 75) return 'risk-high';
            if (score >= 25) return 'risk-medium';
            return 'risk-low';
        },
        
        getRiskLevelText(score) {
            if (!score) return 'Low';
            if (score >= 100) return 'Critical';
            if (score >= 75) return 'High';
            if (score >= 25) return 'Medium';
            return 'Low';
        },
        
        // Parse context JSON string
        parseContext(contextString) {
            if (!contextString) return null;
            try {
                return JSON.parse(contextString);
            } catch (error) {
                console.error('Error parsing context:', error);
                return null;
            }
        },
        
        // Update alert function for form submission
        async updateAlert() {
            if (!this.alert) return;
            
            try {
                const updateData = {
                    status: this.formData.status,
                    owner: this.formData.owner,
                    notes: this.formData.notes
                };
                
                const response = await fetch(`/api/risk/alerts/${this.alertId}`, {
                    method: 'PUT',
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    body: JSON.stringify(updateData)
                });
                
                if (!response.ok) {
                    throw new Error(`Update failed! status: ${response.status}`);
                }
                
                // Update the alert object with new values
                this.alert.status = this.formData.status;
                this.alert.owner = this.formData.owner;
                this.alert.notes = this.formData.notes;
                
                UIUtils.showAlert('Alert updated successfully', 'success');
            } catch (error) {
                console.error('Error updating alert:', error);
                UIUtils.showAlert('Error updating alert', 'error');
            }
        },
        
        viewEvent(eventId) {
            UIUtils.navigateTo(`events-detail.html?id=${eventId}`);
        },
        
        viewDetection(detectionId) {
            UIUtils.navigateTo(`detections-detail.html?id=${detectionId}`);
        }
    };
}