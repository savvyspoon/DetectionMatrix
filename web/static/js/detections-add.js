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
            class_id: null,
            query: ''
        },
        detectionClasses: [],
        dataSources: [],
        selectedDataSources: [],
        selectedDataSourceToAdd: '',
        mitreTechniques: [],
        selectedMitreTechniques: [],
        selectedMitreTechniqueToAdd: '',
        
        get availableDataSources() {
            return this.dataSources.filter(ds => !this.selectedDataSources.includes(ds.id));
        },
        
        get availableMitreTechniques() {
            return this.mitreTechniques.filter(mt => !this.selectedMitreTechniques.includes(mt.id));
        },
        
        async init() {
            await Promise.all([
                this.fetchDetectionClasses(),
                this.fetchDataSources(),
                this.fetchMitreTechniques()
            ]);
        },
        
        async fetchDetectionClasses() {
            try {
                this.detectionClasses = await DetectionAddAPI.fetchDetectionClasses();
            } catch (error) {
                console.error('Error fetching detection classes:', error);
                this.detectionClasses = [];
            }
        },
        
        async fetchDataSources() {
            try {
                const response = await fetch('/api/datasources');
                if (response.ok) {
                    const data = await response.json();
                    // Handle ListResponse structure from API
                    if (data && typeof data === 'object' && 'items' in data) {
                        this.dataSources = Array.isArray(data.items) ? data.items : [];
                    } else if (Array.isArray(data)) {
                        this.dataSources = data;
                    } else {
                        this.dataSources = [];
                    }
                }
            } catch (error) {
                console.error('Error fetching data sources:', error);
                this.dataSources = [];
            }
        },
        
        async fetchMitreTechniques() {
            try {
                const response = await fetch('/api/mitre/techniques');
                if (response.ok) {
                    const data = await response.json();
                    // Handle ListResponse structure from API
                    if (data && typeof data === 'object' && 'items' in data) {
                        this.mitreTechniques = Array.isArray(data.items) ? data.items : [];
                    } else if (Array.isArray(data)) {
                        this.mitreTechniques = data;
                    } else {
                        this.mitreTechniques = [];
                    }
                }
            } catch (error) {
                console.error('Error fetching MITRE techniques:', error);
                this.mitreTechniques = [];
            }
        },
        
        addDataSource() {
            if (this.selectedDataSourceToAdd && !this.selectedDataSources.includes(parseInt(this.selectedDataSourceToAdd))) {
                this.selectedDataSources.push(parseInt(this.selectedDataSourceToAdd));
                this.selectedDataSourceToAdd = '';
            }
        },
        
        removeDataSource(dataSourceId) {
            const index = this.selectedDataSources.indexOf(dataSourceId);
            if (index > -1) {
                this.selectedDataSources.splice(index, 1);
            }
        },
        
        getDataSourceName(dataSourceId) {
            const ds = this.dataSources.find(d => d.id === dataSourceId);
            return ds ? ds.name : 'Unknown';
        },
        
        addMitreTechnique() {
            if (this.selectedMitreTechniqueToAdd && !this.selectedMitreTechniques.includes(this.selectedMitreTechniqueToAdd)) {
                this.selectedMitreTechniques.push(this.selectedMitreTechniqueToAdd);
                this.selectedMitreTechniqueToAdd = '';
            }
        },
        
        removeMitreTechnique(techniqueId) {
            const index = this.selectedMitreTechniques.indexOf(techniqueId);
            if (index > -1) {
                this.selectedMitreTechniques.splice(index, 1);
            }
        },
        
        getMitreTechniqueName(techniqueId) {
            const technique = this.mitreTechniques.find(t => t.id === techniqueId);
            return technique ? `${technique.id} - ${technique.name}` : 'Unknown';
        },
        
        async createDetection() {
            try {
                // Create the detection without the relationship fields
                const detectionData = { ...this.newDetection };
                // Remove fields that aren't part of the detection model
                delete detectionData.data_source_ids;
                delete detectionData.mitre_technique_ids;
                
                // Clean up empty fields that should be omitted
                if (detectionData.class_id === null) {
                    delete detectionData.class_id;
                }
                if (!detectionData.risk_object || detectionData.risk_object === '') {
                    delete detectionData.risk_object;
                }
                if (!detectionData.query || detectionData.query === '') {
                    delete detectionData.query;
                }
                if (!detectionData.playbook_link || detectionData.playbook_link === '') {
                    delete detectionData.playbook_link;
                }
                if (!detectionData.owner || detectionData.owner === '') {
                    delete detectionData.owner;
                }
                if (!detectionData.testing_description || detectionData.testing_description === '') {
                    delete detectionData.testing_description;
                }
                
                console.log('Sending detection data:', detectionData);
                
                // Create the detection
                const createdDetection = await DetectionAddAPI.createDetection(detectionData);
                
                // Add MITRE techniques
                for (const techniqueId of this.selectedMitreTechniques) {
                    try {
                        await fetch(`/api/detections/${createdDetection.id}/mitre/${techniqueId}`, {
                            method: 'POST'
                        });
                    } catch (err) {
                        console.error(`Error adding MITRE technique ${techniqueId}:`, err);
                    }
                }
                
                // Add data sources
                for (const dataSourceId of this.selectedDataSources) {
                    try {
                        await fetch(`/api/detections/${createdDetection.id}/datasource/${dataSourceId}`, {
                            method: 'POST'
                        });
                    } catch (err) {
                        console.error(`Error adding data source ${dataSourceId}:`, err);
                    }
                }
                
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
                class_id: null,
                query: ''
            };
            this.selectedDataSources = [];
            this.selectedMitreTechniques = [];
            this.selectedDataSourceToAdd = '';
            this.selectedMitreTechniqueToAdd = '';
        }
    };
}
