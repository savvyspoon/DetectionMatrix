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
            throw new Error(`Post failed! status: ${response.status}`);
        }
        
        return response;
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
     * Show alert message
     */
    static showAlert(message) {
        alert(message);
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