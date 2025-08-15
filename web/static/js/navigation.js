// Enhanced navigation with command palette and keyboard shortcuts
class NavigationEnhancer {
    constructor() {
        this.commandPalette = null;
        this.shortcuts = new Map();
        this.commands = [];
        this.init();
    }

    init() {
        this.setupCommands();
        this.setupKeyboardShortcuts();
        this.createCommandPalette();
        this.setupBreadcrumbs();
    }

    setupCommands() {
        // Define available commands
        this.commands = [
            { 
                id: 'go-dashboard',
                name: 'Go to Dashboard',
                shortcut: 'g d',
                action: () => window.location.href = '/index.html',
                icon: '🏠'
            },
            {
                id: 'go-detections',
                name: 'Go to Detections',
                shortcut: 'g t',
                action: () => window.location.href = '/detections-list.html',
                icon: '🔍'
            },
            {
                id: 'go-mitre',
                name: 'Go to MITRE ATT&CK',
                shortcut: 'g m',
                action: () => window.location.href = '/mitre.html',
                icon: '🎯'
            },
            {
                id: 'go-events',
                name: 'Go to Events',
                shortcut: 'g e',
                action: () => window.location.href = '/events.html',
                icon: '📊'
            },
            {
                id: 'go-alerts',
                name: 'Go to Risk Alerts',
                shortcut: 'g a',
                action: () => window.location.href = '/risk-alerts.html',
                icon: '🚨'
            },
            {
                id: 'new-detection',
                name: 'Create New Detection',
                shortcut: 'c d',
                action: () => window.location.href = '/detections-add.html',
                icon: '➕'
            },
            {
                id: 'search',
                name: 'Search',
                shortcut: '/',
                action: () => this.focusSearch(),
                icon: '🔎'
            },
            {
                id: 'refresh',
                name: 'Refresh Data',
                shortcut: 'r',
                action: () => window.location.reload(),
                icon: '🔄'
            },
            {
                id: 'help',
                name: 'Show Keyboard Shortcuts',
                shortcut: '?',
                action: () => this.showHelp(),
                icon: '❓'
            },
            {
                id: 'toggle-theme',
                name: 'Toggle Dark Mode',
                shortcut: 't',
                action: () => this.toggleTheme(),
                icon: '🌓'
            }
        ];

        // Add data-specific commands based on current page
        const currentPage = window.location.pathname.split('/').pop();
        if (currentPage.includes('detection')) {
            this.commands.push({
                id: 'edit-detection',
                name: 'Edit Current Detection',
                shortcut: 'e',
                action: () => this.editCurrentItem(),
                icon: '✏️'
            });
        }
    }

    setupKeyboardShortcuts() {
        let keyBuffer = '';
        let bufferTimeout;

        document.addEventListener('keydown', (e) => {
            // Skip if user is typing in an input
            if (e.target.matches('input, textarea, select')) {
                return;
            }

            // Command palette shortcut (Cmd/Ctrl + K)
            if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
                e.preventDefault();
                this.toggleCommandPalette();
                return;
            }

            // Escape closes command palette
            if (e.key === 'Escape') {
                this.hideCommandPalette();
                return;
            }

            // Build key buffer for multi-key shortcuts
            clearTimeout(bufferTimeout);
            keyBuffer += e.key.toLowerCase();

            // Check for matching shortcuts
            const matchingCommand = this.commands.find(cmd => 
                cmd.shortcut === keyBuffer || 
                cmd.shortcut.startsWith(keyBuffer)
            );

            if (matchingCommand && matchingCommand.shortcut === keyBuffer) {
                e.preventDefault();
                matchingCommand.action();
                keyBuffer = '';
            } else if (!this.commands.some(cmd => cmd.shortcut.startsWith(keyBuffer))) {
                keyBuffer = '';
            }

            // Clear buffer after timeout
            bufferTimeout = setTimeout(() => {
                keyBuffer = '';
            }, 1000);
        });
    }

    createCommandPalette() {
        // Create command palette container
        const palette = document.createElement('div');
        palette.id = 'command-palette';
        palette.style.cssText = `
            position: fixed;
            top: 50%;
            left: 50%;
            transform: translate(-50%, -50%) scale(0.95);
            width: 90%;
            max-width: 600px;
            background: white;
            border-radius: 12px;
            box-shadow: 0 25px 50px rgba(0, 0, 0, 0.2);
            z-index: 10000;
            display: none;
            opacity: 0;
            transition: all 0.2s ease;
        `;

        // Create search input
        const searchContainer = document.createElement('div');
        searchContainer.style.cssText = `
            padding: 20px;
            border-bottom: 1px solid #e5e7eb;
        `;

        const searchInput = document.createElement('input');
        searchInput.type = 'text';
        searchInput.placeholder = 'Type a command or search...';
        searchInput.style.cssText = `
            width: 100%;
            padding: 12px;
            font-size: 16px;
            border: 1px solid #e5e7eb;
            border-radius: 8px;
            outline: none;
        `;
        searchInput.addEventListener('input', (e) => this.filterCommands(e.target.value));

        searchContainer.appendChild(searchInput);
        palette.appendChild(searchContainer);

        // Create commands list
        const commandsList = document.createElement('div');
        commandsList.id = 'commands-list';
        commandsList.style.cssText = `
            max-height: 400px;
            overflow-y: auto;
            padding: 10px;
        `;

        palette.appendChild(commandsList);
        document.body.appendChild(palette);

        this.commandPalette = palette;
        this.searchInput = searchInput;
        this.commandsList = commandsList;

        // Click outside to close
        document.addEventListener('click', (e) => {
            if (!palette.contains(e.target)) {
                this.hideCommandPalette();
            }
        });
    }

    filterCommands(query) {
        const filtered = query ? 
            this.commands.filter(cmd => 
                cmd.name.toLowerCase().includes(query.toLowerCase())
            ) : this.commands;

        this.renderCommands(filtered);
    }

    renderCommands(commands) {
        this.commandsList.innerHTML = '';

        commands.forEach((cmd, index) => {
            const item = document.createElement('div');
            item.style.cssText = `
                display: flex;
                align-items: center;
                padding: 12px;
                border-radius: 8px;
                cursor: pointer;
                transition: background 0.2s;
            `;
            
            item.innerHTML = `
                <span style="font-size: 20px; margin-right: 12px;">${cmd.icon}</span>
                <div style="flex: 1;">
                    <div style="font-weight: 500;">${cmd.name}</div>
                    ${cmd.shortcut ? `<div style="font-size: 12px; color: #6b7280; margin-top: 2px;">Shortcut: ${cmd.shortcut}</div>` : ''}
                </div>
            `;

            item.addEventListener('mouseenter', () => {
                item.style.background = '#f3f4f6';
            });

            item.addEventListener('mouseleave', () => {
                item.style.background = 'transparent';
            });

            item.addEventListener('click', () => {
                cmd.action();
                this.hideCommandPalette();
            });

            // Keyboard navigation
            if (index === 0) {
                item.dataset.selected = 'true';
                item.style.background = '#f3f4f6';
            }

            this.commandsList.appendChild(item);
        });
    }

    toggleCommandPalette() {
        if (this.commandPalette.style.display === 'none') {
            this.showCommandPalette();
        } else {
            this.hideCommandPalette();
        }
    }

    showCommandPalette() {
        this.commandPalette.style.display = 'block';
        setTimeout(() => {
            this.commandPalette.style.opacity = '1';
            this.commandPalette.style.transform = 'translate(-50%, -50%) scale(1)';
        }, 10);
        this.searchInput.value = '';
        this.searchInput.focus();
        this.renderCommands(this.commands);
    }

    hideCommandPalette() {
        this.commandPalette.style.opacity = '0';
        this.commandPalette.style.transform = 'translate(-50%, -50%) scale(0.95)';
        setTimeout(() => {
            this.commandPalette.style.display = 'none';
        }, 200);
    }

    setupBreadcrumbs() {
        // Create breadcrumb container if it doesn't exist
        const main = document.querySelector('main');
        if (!main) return;

        const breadcrumbContainer = document.createElement('nav');
        breadcrumbContainer.className = 'breadcrumbs';
        breadcrumbContainer.style.cssText = `
            padding: 10px 0;
            margin-bottom: 20px;
            font-size: 14px;
        `;

        const path = window.location.pathname.split('/').filter(p => p);
        const breadcrumbs = [
            { name: 'Home', url: '/' }
        ];

        // Add current page
        const currentPage = document.title.split(' - ')[0] || 'Page';
        if (currentPage !== 'RiskMatrix') {
            breadcrumbs.push({ name: currentPage, url: window.location.pathname });
        }

        breadcrumbContainer.innerHTML = breadcrumbs.map((crumb, index) => {
            const isLast = index === breadcrumbs.length - 1;
            return `
                <span>
                    ${index > 0 ? '<span style="margin: 0 8px; color: #9ca3af;">›</span>' : ''}
                    ${isLast ? 
                        `<span style="color: #6b7280;">${crumb.name}</span>` :
                        `<a href="${crumb.url}" style="color: #3498db; text-decoration: none;">${crumb.name}</a>`
                    }
                </span>
            `;
        }).join('');

        main.insertBefore(breadcrumbContainer, main.firstChild);
    }

    focusSearch() {
        const searchInput = document.querySelector('input[type="search"], input[placeholder*="Search"]');
        if (searchInput) {
            searchInput.focus();
            searchInput.select();
        } else {
            this.showCommandPalette();
        }
    }

    showHelp() {
        const helpModal = document.createElement('div');
        helpModal.style.cssText = `
            position: fixed;
            top: 0;
            left: 0;
            right: 0;
            bottom: 0;
            background: rgba(0, 0, 0, 0.5);
            display: flex;
            align-items: center;
            justify-content: center;
            z-index: 10001;
        `;

        const helpContent = document.createElement('div');
        helpContent.style.cssText = `
            background: white;
            border-radius: 12px;
            padding: 30px;
            max-width: 500px;
            max-height: 80vh;
            overflow-y: auto;
        `;

        helpContent.innerHTML = `
            <h2 style="margin-bottom: 20px;">Keyboard Shortcuts</h2>
            <div style="font-size: 14px; line-height: 1.8;">
                ${this.commands.map(cmd => `
                    <div style="display: flex; justify-content: space-between; padding: 8px 0; border-bottom: 1px solid #f3f4f6;">
                        <span>${cmd.icon} ${cmd.name}</span>
                        <kbd style="background: #f3f4f6; padding: 4px 8px; border-radius: 4px; font-family: monospace;">${cmd.shortcut}</kbd>
                    </div>
                `).join('')}
                <div style="margin-top: 20px; padding-top: 20px; border-top: 1px solid #e5e7eb;">
                    <strong>Command Palette:</strong> 
                    <kbd style="background: #f3f4f6; padding: 4px 8px; border-radius: 4px; font-family: monospace;">Cmd/Ctrl + K</kbd>
                </div>
            </div>
            <button onclick="this.parentElement.parentElement.remove()" style="
                margin-top: 20px;
                padding: 10px 20px;
                background: #3498db;
                color: white;
                border: none;
                border-radius: 6px;
                cursor: pointer;
                width: 100%;
            ">Close</button>
        `;

        helpModal.appendChild(helpContent);
        document.body.appendChild(helpModal);

        helpModal.addEventListener('click', (e) => {
            if (e.target === helpModal) {
                helpModal.remove();
            }
        });
    }

    toggleTheme() {
        const isDark = document.body.classList.toggle('dark-mode');
        localStorage.setItem('theme', isDark ? 'dark' : 'light');
        
        // Add dark mode styles if not present
        if (!document.getElementById('dark-mode-styles')) {
            const style = document.createElement('style');
            style.id = 'dark-mode-styles';
            style.textContent = `
                body.dark-mode {
                    --surface-primary: #1a1a1a;
                    --surface-secondary: #2a2a2a;
                    --text-color: #e0e0e0;
                    --light-bg: #121212;
                    --border-color: #333;
                }
                
                body.dark-mode .card {
                    background: linear-gradient(135deg, #1a1a1a 0%, #2a2a2a 100%);
                    border-color: rgba(255, 255, 255, 0.1);
                    color: #e0e0e0;
                }
                
                body.dark-mode header {
                    background: linear-gradient(135deg, #1a1a1a 0%, #2a2a2a 100%);
                }
                
                body.dark-mode input,
                body.dark-mode select,
                body.dark-mode textarea {
                    background: #2a2a2a;
                    color: #e0e0e0;
                    border-color: #444;
                }
                
                body.dark-mode table {
                    background: #1a1a1a;
                    color: #e0e0e0;
                }
                
                body.dark-mode th {
                    background: #2a2a2a;
                    color: #e0e0e0;
                }
                
                body.dark-mode tbody tr:hover {
                    background: rgba(52, 152, 219, 0.1);
                }
            `;
            document.head.appendChild(style);
        }

        window.notify?.info(`Dark mode ${isDark ? 'enabled' : 'disabled'}`);
    }

    editCurrentItem() {
        // Find edit button on current page
        const editBtn = document.querySelector('[href*="edit"], [data-action="edit"], .btn-edit');
        if (editBtn) {
            editBtn.click();
        }
    }
}

// Initialize navigation enhancements
document.addEventListener('DOMContentLoaded', () => {
    window.navigationEnhancer = new NavigationEnhancer();
    
    // Restore theme preference
    const savedTheme = localStorage.getItem('theme');
    if (savedTheme === 'dark') {
        document.body.classList.add('dark-mode');
    }
});