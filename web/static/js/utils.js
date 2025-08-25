// Common utility functions used across the application

class APIUtils {
    /**
     * Generic API fetch with standardized JSON error parsing
     */
    static async fetchAPI(url, options = {}) {
        try {
            const response = await fetch(url, options);
            if (!response.ok) {
                const err = await APIUtils.parseError(response);
                UIUtils.showAlert(err.message || `Request failed (${response.status})`, 'error');
                throw new Error(err.message || `HTTP ${response.status}`);
            }
            // Handle empty responses (204, or no body)
            const contentType = response.headers.get('content-type') || '';
            if (!contentType.includes('application/json')) return null;
            return await response.json();
        } catch (error) {
            console.error(`API Error (${url}):`, error);
            throw error;
        }
    }

    /**
     * DELETE request helper with error parsing
     */
    static async deleteAPI(url) {
        const response = await fetch(url, { method: 'DELETE' });
        if (!response.ok) {
            const err = await APIUtils.parseError(response);
            UIUtils.showAlert(err.message || `Delete failed (${response.status})`, 'error');
            throw new Error(err.message || `HTTP ${response.status}`);
        }
        return true;
    }

    /**
     * POST request helper with error parsing
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
            const err = await APIUtils.parseError(response);
            UIUtils.showAlert(err.message || `Post failed (${response.status})`, 'error');
            throw new Error(err.message || `HTTP ${response.status}`);
        }
        return await response.json();
    }

    /**
     * PUT request helper with error parsing
     */
    static async putAPI(url, data) {
        const response = await fetch(url, {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(data)
        });
        if (!response.ok) {
            const err = await APIUtils.parseError(response);
            UIUtils.showAlert(err.message || `Put failed (${response.status})`, 'error');
            throw new Error(err.message || `HTTP ${response.status}`);
        }
        return await response.json();
    }

    /**
     * Parse standardized JSON error { error, code, request_id }
     */
    static async parseError(response) {
        try {
            const contentType = response.headers.get('content-type') || '';
            if (contentType.includes('application/json')) {
                const body = await response.json();
                return { message: body.error, code: body.code, requestId: body.request_id };
            }
            const text = await response.text();
            return { message: text || response.statusText, code: response.status };
        } catch (e) {
            return { message: response.statusText, code: response.status };
        }
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
