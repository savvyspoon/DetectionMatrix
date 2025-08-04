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
        
        viewEvent(eventId) {
            UIUtils.navigateTo(`events-detail.html?id=${eventId}`);
        },
        
        viewDetection(detectionId) {
            UIUtils.navigateTo(`detections-detail.html?id=${detectionId}`);
        }
    };
}