// Risk Objects Detail JavaScript functionality
class RiskObjectDetailAPI {
    static async fetchRiskObject(id) {
        return await APIUtils.fetchAPI(`/api/risk/objects/${id}`);
    }

    static async fetchEventsByEntity(entityId) {
        return await APIUtils.fetchAPI(`/api/events/entity/${entityId}`);
    }

    static async fetchDetections() {
        return await APIUtils.fetchAPI('/api/detections');
    }
}

// Alpine.js risk object detail data function
function riskObjectDetailData() {
    return {
        riskObject: null,
        events: [],
        detections: {},
        loading: true,
        objectId: null,
        
        async init() {
            const urlParams = new URLSearchParams(window.location.search);
            this.objectId = urlParams.get('id');
            if (this.objectId) {
                await this.fetchRiskObjectDetails();
            } else {
                UIUtils.showAlert('No risk object ID provided');
                UIUtils.navigateTo('risk-objects.html');
            }
        },
        
        async fetchRiskObjectDetails() {
            this.loading = true;
            try {
                this.riskObject = await RiskObjectDetailAPI.fetchRiskObject(this.objectId);
                await Promise.all([
                    this.fetchEvents(),
                    this.fetchDetections()
                ]);
            } catch (error) {
                console.error('Error fetching risk object details:', error);
                UIUtils.showAlert('Error loading risk object details');
                UIUtils.navigateTo('risk-objects.html');
            }
            this.loading = false;
        },
        
        async fetchEvents() {
            try {
                this.events = await RiskObjectDetailAPI.fetchEventsByEntity(this.objectId);
            } catch (error) {
                console.error('Error fetching events:', error);
                this.events = [];
            }
        },
        
        async fetchDetections() {
            try {
                const detectionsList = await RiskObjectDetailAPI.fetchDetections();
                this.detections = {};
                detectionsList.forEach(d => {
                    this.detections[d.id] = d;
                });
            } catch (error) {
                console.error('Error fetching detections:', error);
            }
        },
        
        getDetectionName(detectionId) {
            return this.detections[detectionId]?.name || `Detection ${detectionId}`;
        },
        
        formatTimestamp(timestamp) {
            return UIUtils.formatDate(timestamp);
        },
        
        getRiskScoreClass(score) {
            if (score >= 80) return 'risk-critical';
            if (score >= 60) return 'risk-high';
            if (score >= 40) return 'risk-medium';
            if (score >= 20) return 'risk-low';
            return 'risk-minimal';
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