// Events Detail JavaScript functionality
class EventAPI {
    static async fetchEvent(id) {
        const response = await fetch(`/api/events/${id}`);
        if (response.ok) {
            return await response.json();
        }
        throw new Error('Event not found');
    }

    static async fetchDetection(detectionId) {
        const response = await fetch(`/api/detections/${detectionId}`);
        if (response.ok) {
            return await response.json();
        }
        return null;
    }

    static async fetchRiskObject(entityId) {
        const response = await fetch(`/api/risk/objects/${entityId}`);
        if (response.ok) {
            return await response.json();
        }
        return null;
    }

    static async markAsFalsePositive(eventId, reason, analystName) {
        const response = await fetch(`/api/events/${eventId}/false-positive`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                reason: reason,
                analyst_name: analystName
            })
        });
        
        if (!response.ok) {
            throw new Error('Failed to mark event as false positive');
        }
    }

    static async unmarkAsFalsePositive(eventId) {
        const response = await fetch(`/api/events/${eventId}/false-positive`, {
            method: 'DELETE'
        });
        
        if (!response.ok) {
            throw new Error('Failed to unmark event as false positive');
        }
    }
}

class EventUtils {
    static formatTimestamp(timestamp) {
        return new Date(timestamp).toLocaleString();
    }
    
    static getFalsePositiveClass(isFalsePositive) {
        return isFalsePositive ? 'badge badge-warning' : 'badge badge-success';
    }
    
    static getFalsePositiveText(isFalsePositive) {
        return isFalsePositive ? 'False Positive' : 'Valid';
    }
    
    static formatContext(context) {
        if (!context) return 'No context available';
        try {
            const parsed = JSON.parse(context);
            return JSON.stringify(parsed, null, 2);
        } catch (e) {
            return context;
        }
    }
}

// Alpine.js event detail data function
function eventDetailData() {
    return {
        event: null,
        detection: null,
        riskObject: null,
        dataSources: [],
        loading: true,
        
        async init() {
            await this.loadEvent();
        },
        
        async loadEvent() {
            const urlParams = new URLSearchParams(window.location.search);
            const id = urlParams.get('id');
            
            if (!id) {
                alert('No event ID provided');
                window.location.href = 'events.html';
                return;
            }
            
            try {
                // Fetch event details
                this.event = await EventAPI.fetchEvent(id);
                
                // Fetch related data
                await Promise.all([
                    this.fetchDetection(this.event.detection_id),
                    this.fetchRiskObject(this.event.entity_id)
                ]);
                
                this.loading = false;
            } catch (error) {
                console.error('Error loading event:', error);
                alert('Event not found');
                window.location.href = 'events.html';
            }
        },
        
        async fetchDetection(detectionId) {
            try {
                this.detection = await EventAPI.fetchDetection(detectionId);
                this.dataSources = this.detection?.data_sources || [];
            } catch (error) {
                console.error('Error fetching detection:', error);
            }
        },
        
        async fetchRiskObject(entityId) {
            try {
                this.riskObject = await EventAPI.fetchRiskObject(entityId);
            } catch (error) {
                console.error('Error fetching risk object:', error);
            }
        },
        
        formatTimestamp(timestamp) {
            return EventUtils.formatTimestamp(timestamp);
        },
        
        getFalsePositiveClass(isFalsePositive) {
            return EventUtils.getFalsePositiveClass(isFalsePositive);
        },
        
        getFalsePositiveText(isFalsePositive) {
            return EventUtils.getFalsePositiveText(isFalsePositive);
        },
        
        formatContext(context) {
            return EventUtils.formatContext(context);
        },
        
        async markAsFalsePositive() {
            const reason = prompt('Please provide a reason for marking this as a false positive:');
            if (!reason) return;
            
            const analystName = prompt('Please enter your name:');
            if (!analystName) return;
            
            try {
                await EventAPI.markAsFalsePositive(this.event.id, reason, analystName);
                await this.loadEvent(); // Refresh event data
            } catch (error) {
                console.error('Error marking false positive:', error);
                alert('Error marking event as false positive');
            }
        },
        
        async unmarkAsFalsePositive() {
            if (!confirm('Are you sure you want to unmark this event as a false positive? This will restore the risk points.')) {
                return;
            }
            
            try {
                await EventAPI.unmarkAsFalsePositive(this.event.id);
                await this.loadEvent(); // Refresh event data
            } catch (error) {
                console.error('Error unmarking false positive:', error);
                alert('Error unmarking event as false positive');
            }
        }
    };
}