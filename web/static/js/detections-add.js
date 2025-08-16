// Detections Add JavaScript functionality
class DetectionAddAPI {
    static async createDetection(detection) {
        return await APIUtils.postAPI('/api/detections', detection);
    }
    
    static async fetchDetectionClasses() {
        return await APIUtils.fetchAPI('/api/detection-classes');
    }
}

// Alpine.js detection add data function
function detectionAddData() {
    return {
        newDetection: {
            name: '',
            description: '',
            status: 'idea',
            severity: 'medium',
            risk_points: 50,
            playbook_link: '',
            owner: '',
            risk_object: '',
            testing_description: '',
            class_id: null
        },
        detectionClasses: [],
        
        async init() {
            await this.fetchDetectionClasses();
        },
        
        async fetchDetectionClasses() {
            try {
                this.detectionClasses = await DetectionAddAPI.fetchDetectionClasses();
            } catch (error) {
                console.error('Error fetching detection classes:', error);
                this.detectionClasses = [];
            }
        },
        
        async createDetection() {
            try {
                await DetectionAddAPI.createDetection(this.newDetection);
                UIUtils.navigateTo('detections-list.html');
            } catch (error) {
                console.error('Error creating detection:', error);
                UIUtils.showAlert('Error creating detection. Please try again.');
            }
        },

        resetForm() {
            this.newDetection = {
                name: '',
                description: '',
                status: 'idea',
                severity: 'medium',
                risk_points: 50,
                playbook_link: '',
                owner: '',
                risk_object: '',
                testing_description: '',
                class_id: null
            };
        }
    };
}