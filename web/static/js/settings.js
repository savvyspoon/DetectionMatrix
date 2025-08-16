// Settings page functionality for Detection Classes management

class SettingsAPI {
    static async fetchDetectionClasses() {
        try {
            const response = await fetch('/api/detection-classes');
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }
            return await response.json();
        } catch (error) {
            console.error('Error fetching detection classes:', error);
            throw error;
        }
    }
    
    static async createDetectionClass(classData) {
        try {
            const response = await fetch('/api/detection-classes', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(classData)
            });
            
            if (!response.ok) {
                const errorText = await response.text();
                throw new Error(errorText || `HTTP error! status: ${response.status}`);
            }
            return await response.json();
        } catch (error) {
            console.error('Error creating detection class:', error);
            throw error;
        }
    }
    
    static async updateDetectionClass(id, classData) {
        try {
            const response = await fetch(`/api/detection-classes/${id}`, {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(classData)
            });
            
            if (!response.ok) {
                const errorText = await response.text();
                throw new Error(errorText || `HTTP error! status: ${response.status}`);
            }
            return response.status === 204 ? {} : await response.json();
        } catch (error) {
            console.error('Error updating detection class:', error);
            throw error;
        }
    }
    
    static async deleteDetectionClass(id) {
        try {
            const response = await fetch(`/api/detection-classes/${id}`, {
                method: 'DELETE'
            });
            
            if (!response.ok) {
                const errorText = await response.text();
                throw new Error(errorText || `HTTP error! status: ${response.status}`);
            }
            return true;
        } catch (error) {
            console.error('Error deleting detection class:', error);
            throw error;
        }
    }
}

function settingsData() {
    return {
        classes: [],
        showAddForm: false,
        editingClass: null,
        loading: true,
        error: null,
        newClass: {
            name: '',
            description: '',
            color: '#6B7280',
            icon: '',
            display_order: 999
        },
        
        async init() {
            await this.loadClasses();
        },
        
        async loadClasses() {
            this.loading = true;
            this.error = null;
            try {
                this.classes = await SettingsAPI.fetchDetectionClasses();
            } catch (error) {
                this.error = 'Failed to load detection classes: ' + error.message;
                this.classes = [];
            } finally {
                this.loading = false;
            }
        },
        
        get sortedClasses() {
            return [...this.classes].sort((a, b) => {
                // First sort by display_order
                if (a.display_order !== b.display_order) {
                    return a.display_order - b.display_order;
                }
                // Then by name if display_order is the same
                return a.name.localeCompare(b.name);
            });
        },
        
        async saveClass() {
            if (!this.newClass.name) {
                this.showAlert('Class name is required', 'error');
                return;
            }
            
            try {
                if (this.editingClass) {
                    // Update existing class
                    await SettingsAPI.updateDetectionClass(this.editingClass.id, this.newClass);
                    this.showAlert('Class updated successfully', 'success');
                } else {
                    // Create new class
                    await SettingsAPI.createDetectionClass(this.newClass);
                    this.showAlert('Class added successfully', 'success');
                }
                
                await this.loadClasses();
                this.cancelAdd();
            } catch (error) {
                this.showAlert('Failed to save class: ' + error.message, 'error');
            }
        },
        
        editClass(classItem) {
            if (classItem.is_system) {
                this.showAlert('System classes cannot be modified', 'warning');
                return;
            }
            
            this.editingClass = classItem;
            this.newClass = {
                name: classItem.name,
                description: classItem.description || '',
                color: classItem.color || '#6B7280',
                icon: classItem.icon || '',
                display_order: classItem.display_order || 999
            };
            this.showAddForm = true;
        },
        
        async deleteClass(id) {
            const classItem = this.classes.find(c => c.id === id);
            if (!classItem) return;
            
            if (classItem.is_system) {
                this.showAlert('System classes cannot be deleted', 'warning');
                return;
            }
            
            if (!confirm(`Are you sure you want to delete the class "${classItem.name}"? Detections using this class will have their class unset.`)) {
                return;
            }
            
            try {
                await SettingsAPI.deleteDetectionClass(id);
                this.showAlert('Class deleted successfully', 'success');
                await this.loadClasses();
            } catch (error) {
                this.showAlert('Failed to delete class: ' + error.message, 'error');
            }
        },
        
        cancelAdd() {
            this.showAddForm = false;
            this.editingClass = null;
            this.newClass = {
                name: '',
                description: '',
                color: '#6B7280',
                icon: '',
                display_order: 999
            };
        },
        
        showAlert(message, type = 'info') {
            // Simple alert implementation - you can enhance this with a toast notification
            const alertClass = type === 'error' ? 'alert-danger' : 
                             type === 'success' ? 'alert-success' : 
                             type === 'warning' ? 'alert-warning' : 'alert-info';
            
            // Create alert element
            const alertDiv = document.createElement('div');
            alertDiv.className = `alert ${alertClass}`;
            alertDiv.style.cssText = 'position: fixed; top: 20px; right: 20px; padding: 1rem; background: white; border-radius: 4px; box-shadow: 0 2px 8px rgba(0,0,0,0.2); z-index: 1000; min-width: 250px;';
            alertDiv.textContent = message;
            
            // Add appropriate background color
            if (type === 'error') {
                alertDiv.style.backgroundColor = '#fee';
                alertDiv.style.borderLeft = '4px solid #dc3545';
            } else if (type === 'success') {
                alertDiv.style.backgroundColor = '#efe';
                alertDiv.style.borderLeft = '4px solid #28a745';
            } else if (type === 'warning') {
                alertDiv.style.backgroundColor = '#ffeaa7';
                alertDiv.style.borderLeft = '4px solid #ffc107';
            }
            
            document.body.appendChild(alertDiv);
            
            // Auto-remove after 3 seconds
            setTimeout(() => {
                alertDiv.remove();
            }, 3000);
        }
    };
}