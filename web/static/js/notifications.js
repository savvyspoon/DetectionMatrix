// Toast notification system
class NotificationSystem {
    constructor() {
        this.container = null;
        this.init();
    }

    init() {
        // Create container if it doesn't exist
        if (!document.getElementById('notification-container')) {
            this.container = document.createElement('div');
            this.container.id = 'notification-container';
            this.container.style.cssText = `
                position: fixed;
                top: 20px;
                right: 20px;
                z-index: 9999;
                display: flex;
                flex-direction: column;
                gap: 10px;
                pointer-events: none;
            `;
            document.body.appendChild(this.container);
        } else {
            this.container = document.getElementById('notification-container');
        }
    }

    show(message, type = 'info', duration = 5000) {
        const notification = document.createElement('div');
        notification.className = `notification notification-${type}`;
        
        const icons = {
            success: '✓',
            error: '✕',
            warning: '⚠',
            info: 'ℹ'
        };

        const colors = {
            success: 'linear-gradient(135deg, #10b981 0%, #059669 100%)',
            error: 'linear-gradient(135deg, #ef4444 0%, #dc2626 100%)',
            warning: 'linear-gradient(135deg, #f59e0b 0%, #d97706 100%)',
            info: 'linear-gradient(135deg, #3498db 0%, #2980b9 100%)'
        };

        notification.style.cssText = `
            display: flex;
            align-items: center;
            gap: 12px;
            padding: 16px 20px;
            background: white;
            border-radius: 8px;
            box-shadow: 0 10px 25px rgba(0, 0, 0, 0.1), 0 4px 10px rgba(0, 0, 0, 0.06);
            min-width: 300px;
            max-width: 500px;
            pointer-events: auto;
            cursor: pointer;
            animation: slideIn 0.3s ease;
            position: relative;
            overflow: hidden;
        `;

        // Add gradient bar
        const gradientBar = document.createElement('div');
        gradientBar.style.cssText = `
            position: absolute;
            top: 0;
            left: 0;
            right: 0;
            height: 3px;
            background: ${colors[type]};
        `;
        notification.appendChild(gradientBar);

        // Add icon
        const icon = document.createElement('div');
        icon.style.cssText = `
            width: 24px;
            height: 24px;
            border-radius: 50%;
            background: ${colors[type]};
            color: white;
            display: flex;
            align-items: center;
            justify-content: center;
            font-weight: bold;
            flex-shrink: 0;
        `;
        icon.textContent = icons[type];
        notification.appendChild(icon);

        // Add message
        const messageEl = document.createElement('div');
        messageEl.style.cssText = `
            flex: 1;
            color: #333;
            font-size: 14px;
            line-height: 1.5;
        `;
        messageEl.textContent = message;
        notification.appendChild(messageEl);

        // Add close button
        const closeBtn = document.createElement('button');
        closeBtn.style.cssText = `
            background: none;
            border: none;
            color: #999;
            cursor: pointer;
            padding: 4px;
            font-size: 18px;
            line-height: 1;
            transition: color 0.2s;
        `;
        closeBtn.innerHTML = '&times;';
        closeBtn.onmouseover = () => closeBtn.style.color = '#333';
        closeBtn.onmouseout = () => closeBtn.style.color = '#999';
        notification.appendChild(closeBtn);

        // Add progress bar for auto-dismiss
        const progressBar = document.createElement('div');
        progressBar.style.cssText = `
            position: absolute;
            bottom: 0;
            left: 0;
            height: 2px;
            background: ${colors[type]};
            opacity: 0.3;
            animation: progress ${duration}ms linear;
        `;
        notification.appendChild(progressBar);

        // Click handlers
        const dismiss = () => {
            notification.style.animation = 'slideOut 0.3s ease';
            setTimeout(() => notification.remove(), 300);
        };

        closeBtn.onclick = dismiss;
        notification.onclick = dismiss;

        // Auto dismiss
        if (duration > 0) {
            setTimeout(dismiss, duration);
        }

        this.container.appendChild(notification);

        // Add animations
        if (!document.getElementById('notification-styles')) {
            const style = document.createElement('style');
            style.id = 'notification-styles';
            style.textContent = `
                @keyframes slideIn {
                    from {
                        transform: translateX(100%);
                        opacity: 0;
                    }
                    to {
                        transform: translateX(0);
                        opacity: 1;
                    }
                }
                @keyframes slideOut {
                    from {
                        transform: translateX(0);
                        opacity: 1;
                    }
                    to {
                        transform: translateX(100%);
                        opacity: 0;
                    }
                }
                @keyframes progress {
                    from { width: 100%; }
                    to { width: 0%; }
                }
                @media (max-width: 768px) {
                    #notification-container {
                        left: 10px !important;
                        right: 10px !important;
                        top: 10px !important;
                    }
                    #notification-container .notification {
                        min-width: auto !important;
                        max-width: none !important;
                    }
                }
            `;
            document.head.appendChild(style);
        }

        return notification;
    }

    success(message, duration) {
        return this.show(message, 'success', duration);
    }

    error(message, duration) {
        return this.show(message, 'error', duration);
    }

    warning(message, duration) {
        return this.show(message, 'warning', duration);
    }

    info(message, duration) {
        return this.show(message, 'info', duration);
    }
}

// Initialize global notification system
window.notify = new NotificationSystem();

// Add to HTMX requests
document.addEventListener('DOMContentLoaded', () => {
    // Success responses
    document.body.addEventListener('htmx:afterOnLoad', (event) => {
        if (event.detail.xhr.status >= 200 && event.detail.xhr.status < 300) {
            const message = event.detail.xhr.getResponseHeader('X-Success-Message');
            if (message) {
                window.notify.success(message);
            }
        }
    });

    // Error responses
    document.body.addEventListener('htmx:responseError', (event) => {
        const message = event.detail.xhr.getResponseHeader('X-Error-Message') || 
                       'An error occurred. Please try again.';
        window.notify.error(message);
    });

    // Network errors
    document.body.addEventListener('htmx:sendError', (event) => {
        window.notify.error('Network error. Please check your connection.');
    });

    // Loading states
    document.body.addEventListener('htmx:beforeRequest', (event) => {
        const target = event.detail.target;
        if (target && target.dataset.loadingMessage) {
            target._loadingNotification = window.notify.info(target.dataset.loadingMessage, 0);
        }
    });

    document.body.addEventListener('htmx:afterRequest', (event) => {
        const target = event.detail.target;
        if (target && target._loadingNotification) {
            target._loadingNotification.click();
            delete target._loadingNotification;
        }
    });
});