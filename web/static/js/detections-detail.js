// Detections Detail JavaScript functionality
class DetectionDetailAPI {
    static async fetchDetection(id) {
        return await APIUtils.fetchAPI(`/api/detections/${id}`);
    }

    static async updateDetection(id, detection) {
        return await APIUtils.postAPI(`/api/detections/${id}`, detection);
    }

    static async deleteDetection(id) {
        return await APIUtils.deleteAPI(`/api/detections/${id}`);
    }

    static async fetchFalsePositiveRate(id) {
        return await APIUtils.fetchAPI(`/api/detections/${id}/fp-rate`);
    }

    static async fetchEventCount(id) {
        return await APIUtils.fetchAPI(`/api/detections/${id}/events/count/30days`);
    }

    static async fetchFalsePositiveCount(id) {
        return await APIUtils.fetchAPI(`/api/detections/${id}/false-positives/count/30days`);
    }
}

// Alpine.js detection detail data function
function detectionDetailData() {
    return {
        detection: null,
        loading: true,
        editing: false,
        detectionId: null,
        stats: {
            falsePositiveRate: 0,
            eventCount: 0,
            falsePositiveCount: 0
        },
        
        async init() {
            const urlParams = new URLSearchParams(window.location.search);
            this.detectionId = urlParams.get('id');
            if (this.detectionId) {
                await this.fetchDetectionDetails();
            } else {
                UIUtils.showAlert('No detection ID provided');
                UIUtils.navigateTo('detections-list.html');
            }
        },
        
        async fetchDetectionDetails() {
            this.loading = true;
            try {
                this.detection = await DetectionDetailAPI.fetchDetection(this.detectionId);
                await this.fetchStats();
            } catch (error) {
                console.error('Error fetching detection details:', error);
                UIUtils.showAlert('Error loading detection details');
                UIUtils.navigateTo('detections-list.html');
            }
            this.loading = false;
        },
        
        async fetchStats() {
            try {
                const [fpRate, eventCount, fpCount] = await Promise.all([
                    DetectionDetailAPI.fetchFalsePositiveRate(this.detectionId),
                    DetectionDetailAPI.fetchEventCount(this.detectionId),
                    DetectionDetailAPI.fetchFalsePositiveCount(this.detectionId)
                ]);
                
                this.stats = {
                    falsePositiveRate: fpRate.false_positive_rate || 0,
                    eventCount: eventCount.count || 0,
                    falsePositiveCount: fpCount.count || 0
                };
            } catch (error) {
                console.error('Error fetching stats:', error);
            }
        },
        
        async updateDetection() {
            try {
                await DetectionDetailAPI.updateDetection(this.detectionId, this.detection);
                this.editing = false;
                UIUtils.showAlert('Detection updated successfully');
            } catch (error) {
                console.error('Error updating detection:', error);
                UIUtils.showAlert('Error updating detection');
            }
        },
        
        async deleteDetection() {
            if (!UIUtils.confirmAction('Are you sure you want to delete this detection?')) {
                return;
            }
            
            try {
                await DetectionDetailAPI.deleteDetection(this.detectionId);
                UIUtils.navigateTo('detections-list.html');
            } catch (error) {
                console.error('Error deleting detection:', error);
                UIUtils.showAlert('Error deleting detection');
            }
        },
        
        startEditing() {
            this.editing = true;
        },
        
        cancelEditing() {
            this.editing = false;
            // Reload detection data to cancel changes
            this.fetchDetectionDetails();
        },
        
        getStatusClass(status) {
            return UIUtils.getStatusClass(status);
        },
        
        getSeverityClass(severity) {
            return UIUtils.getSeverityClass(severity);
        },
        
        formatTimestamp(timestamp) {
            return UIUtils.formatDate(timestamp);
        }
    };
}