// Events Add JavaScript functionality
class EventsAddAPI {
    static async fetchDetections() {
        return await APIUtils.fetchAPI('/api/detections');
    }
    
    static async fetchRiskObjects() {
        return await APIUtils.fetchAPI('/api/risk/objects');
    }
    
    static async createEvent(eventData) {
        return await APIUtils.postAPI('/api/events', eventData);
    }
    
    static async getRiskObjectByEntity(entityType, entityValue) {
        const params = new URLSearchParams({
            type: entityType,
            value: entityValue
        });
        return await APIUtils.fetchAPI(`/api/risk/objects/entity?${params}`);
    }
}

// Alpine.js event add data function
function eventAddData() {
    return {
        formData: {
            detection_id: '',
            entity_type: '',
            entity_value: '',
            risk_points: 10,
            timestamp: '',
            raw_data: '',
            context: '',
            is_false_positive: false,
            false_positive_reason: ''
        },
        
        detections: [],
        riskObjects: [],
        filteredEntities: [],
        showEntitySuggestions: false,
        submitting: false,
        successMessage: '',
        errorMessage: '',
        selectedEntityId: null,
        
        async init() {
            await this.loadData();
        },
        
        async loadData() {
            try {
                // Load detections and risk objects in parallel
                const [detections, riskObjects] = await Promise.all([
                    EventsAddAPI.fetchDetections(),
                    EventsAddAPI.fetchRiskObjects()
                ]);
                
                this.detections = detections || [];
                this.riskObjects = riskObjects || [];
            } catch (error) {
                console.error('Error loading data:', error);
                this.errorMessage = 'Error loading form data';
            }
        },
        
        getEntityPlaceholder() {
            switch (this.formData.entity_type) {
                case 'user':
                    return 'e.g., john.doe, admin';
                case 'host':
                    return 'e.g., SERVER-DC01, workstation-123';
                case 'ip':
                    return 'e.g., 192.168.1.100, 10.0.0.5';
                default:
                    return 'Select entity type first';
            }
        },
        
        clearEntitySelection() {
            this.formData.entity_value = '';
            this.selectedEntityId = null;
            this.filteredEntities = [];
            this.showEntitySuggestions = false;
        },
        
        searchEntities() {
            if (!this.formData.entity_type || !this.formData.entity_value) {
                this.filteredEntities = [];
                this.showEntitySuggestions = false;
                return;
            }
            
            const searchTerm = this.formData.entity_value.toLowerCase();
            this.filteredEntities = this.riskObjects.filter(obj => 
                obj.entity_type === this.formData.entity_type &&
                obj.entity_value.toLowerCase().includes(searchTerm)
            ).slice(0, 5); // Limit to 5 suggestions
            
            this.showEntitySuggestions = this.filteredEntities.length > 0;
        },
        
        selectEntity(entity) {
            this.formData.entity_value = entity.entity_value;
            this.selectedEntityId = entity.id;
            this.showEntitySuggestions = false;
        },
        
        validateJSON(jsonString) {
            if (!jsonString) return true; // Empty is valid (optional field)
            try {
                JSON.parse(jsonString);
                return true;
            } catch (e) {
                return false;
            }
        },
        
        async submitEvent() {
            // Clear previous messages
            this.successMessage = '';
            this.errorMessage = '';
            
            // Validate required fields
            if (!this.formData.detection_id) {
                this.errorMessage = 'Please select a detection';
                return;
            }
            
            if (!this.formData.entity_type || !this.formData.entity_value) {
                this.errorMessage = 'Please specify the entity type and value';
                return;
            }
            
            // Validate JSON fields
            if (this.formData.raw_data && !this.validateJSON(this.formData.raw_data)) {
                this.errorMessage = 'Raw Data must be valid JSON format';
                return;
            }
            
            if (this.formData.context && !this.validateJSON(this.formData.context)) {
                this.errorMessage = 'Context must be valid JSON format';
                return;
            }
            
            this.submitting = true;
            
            try {
                // First, get or create the risk object
                let entityId = this.selectedEntityId;
                
                if (!entityId) {
                    // Check if entity exists
                    try {
                        const existingEntity = await EventsAddAPI.getRiskObjectByEntity(
                            this.formData.entity_type,
                            this.formData.entity_value
                        );
                        entityId = existingEntity.id;
                    } catch (error) {
                        // Entity doesn't exist, it will be created when the event is processed
                        // The backend will handle this
                    }
                }
                
                // Prepare event data
                const eventData = {
                    detection_id: parseInt(this.formData.detection_id),
                    entity_type: this.formData.entity_type,
                    entity_value: this.formData.entity_value,
                    risk_points: parseInt(this.formData.risk_points),
                    is_false_positive: this.formData.is_false_positive
                };
                
                // Add optional fields if provided
                if (this.formData.timestamp) {
                    eventData.timestamp = new Date(this.formData.timestamp).toISOString();
                }
                
                if (this.formData.raw_data) {
                    eventData.raw_data = this.formData.raw_data;
                }
                
                if (this.formData.context) {
                    eventData.context = this.formData.context;
                }
                
                if (entityId) {
                    eventData.entity_id = entityId;
                }
                
                // Submit the event
                const result = await EventsAddAPI.createEvent(eventData);
                
                this.successMessage = `Event created successfully! Event ID: ${result.id}`;
                
                // Reset form after successful submission
                setTimeout(() => {
                    this.resetForm();
                    // Optionally redirect to events list
                    if (confirm('Event created successfully! View all events?')) {
                        window.location.href = 'events.html';
                    }
                }, 1500);
                
            } catch (error) {
                console.error('Error creating event:', error);
                this.errorMessage = error.message || 'Error creating event. Please try again.';
            } finally {
                this.submitting = false;
            }
        },
        
        resetForm() {
            this.formData = {
                detection_id: '',
                entity_type: '',
                entity_value: '',
                risk_points: 10,
                timestamp: '',
                raw_data: '',
                context: '',
                is_false_positive: false,
                false_positive_reason: ''
            };
            this.selectedEntityId = null;
            this.filteredEntities = [];
            this.showEntitySuggestions = false;
            this.successMessage = '';
            this.errorMessage = '';
        }
    };
}