// Common utility functions used across the application

class APIUtils {
    /**
     * Generic API fetch with error handling
     */
    static async fetchAPI(url, options = {}) {
        try {
            const response = await fetch(url, options);
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }
            return await response.json();
        } catch (error) {
            console.error(`API Error (${url}):`, error);
            throw error;
        }
    }

    /**
     * DELETE request helper
     */
    static async deleteAPI(url) {
        const response = await fetch(url, { method: 'DELETE' });
        if (!response.ok) {
            throw new Error(`Delete failed! status: ${response.status}`);
        }
        return response;
    }

    /**
     * POST request helper
     */
    static async postAPI(url, data) {
        const response = await fetch(url, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(data)
        });
        
        if (!response.ok) {
            const errorText = await response.text();
            console.error('Server error:', errorText);
            throw new Error(`Post failed! status: ${response.status}`);
        }
        
        return await response.json();
    }
}

class UIUtils {
    /**
     * Status CSS class mappings
     */
    static getStatusClass(status) {
        return `status status-${status}`;
    }

    /**
     * Severity badge CSS class mappings
     */
    static getSeverityClass(severity) {
        switch(severity) {
            case 'low': return 'badge badge-secondary';
            case 'medium': return 'badge badge-warning';
            case 'high': return 'badge badge-danger';
            case 'critical': return 'badge badge-danger';
            default: return 'badge';
        }
    }

    /**
     * Format dates consistently
     */
    static formatDate(dateString) {
        return new Date(dateString).toLocaleString();
    }

    /**
     * Confirmation dialog helper
     */
    static confirmAction(message) {
        return confirm(message);
    }

    /**
     * Show alert message with better UI
     */
    static showAlert(message, type = 'info') {
        // Create notification element
        const notification = document.createElement('div');
        notification.className = `notification notification-${type}`;
        notification.style.cssText = `
            position: fixed;
            top: 20px;
            right: 20px;
            z-index: 9999;
            padding: 12px 16px;
            border-radius: 4px;
            color: white;
            max-width: 400px;
            font-size: 14px;
            box-shadow: 0 2px 8px rgba(0,0,0,0.2);
            transition: opacity 0.3s ease;
        `;
        
        // Set background color based on type
        switch(type) {
            case 'error':
                notification.style.backgroundColor = '#dc3545';
                break;
            case 'warning':
                notification.style.backgroundColor = '#ffc107';
                notification.style.color = '#212529';
                break;
            case 'success':
                notification.style.backgroundColor = '#28a745';
                break;
            default:
                notification.style.backgroundColor = '#17a2b8';
        }
        
        notification.textContent = message;
        document.body.appendChild(notification);
        
        // Auto-remove after 5 seconds
        setTimeout(() => {
            notification.style.opacity = '0';
            setTimeout(() => {
                if (notification.parentNode) {
                    notification.parentNode.removeChild(notification);
                }
            }, 300);
        }, 5000);
    }

    /**
     * Navigate to URL
     */
    static navigateTo(url) {
        window.location.href = url;
    }
}

class FilterUtils {
    /**
     * Generic array filtering function
     */
    static filterItems(items, filters) {
        return items.filter(item => {
            return Object.entries(filters).every(([key, filterValue]) => {
                if (filterValue === 'all' || filterValue === '') {
                    return true;
                }
                
                if (key === 'search') {
                    const searchTerm = filterValue.toLowerCase();
                    return item.name?.toLowerCase().includes(searchTerm) ||
                           item.description?.toLowerCase().includes(searchTerm);
                }
                
                return item[key] === filterValue;
            });
        });
    }

    /**
     * Extract URL parameter
     */
    static getURLParameter(name) {
        const urlParams = new URLSearchParams(window.location.search);
        return urlParams.get(name);
    }
}

// Export for use in other modules if needed
if (typeof module !== 'undefined' && module.exports) {
    module.exports = { APIUtils, UIUtils, FilterUtils };
}