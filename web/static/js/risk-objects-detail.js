// Risk Objects Detail JavaScript functionality
class RiskObjectDetailAPI {
    static async fetchRiskObject(id) {
        return await APIUtils.fetchAPI(`/api/risk/objects/${id}`);
    }

    static async fetchEventsByEntity(entityId) {
        return await APIUtils.fetchAPI(`/api/events/entity/${entityId}`);
    }

    static async fetchDetections() {
        return await APIUtils.fetchAPI('/api/detections');
    }
}

// Alpine.js risk object detail data function
function riskObjectDetailData() {
    return {
        riskObject: null,
        events: [],
        detections: {},
        riskAlerts: [],
        loading: true,
        objectId: null,
        chart: null,
        chartInitialized: false,
        
        async init() {
            const urlParams = new URLSearchParams(window.location.search);
            this.objectId = urlParams.get('id');
            if (this.objectId) {
                await this.fetchRiskObjectDetails();
            } else {
                UIUtils.showAlert('No risk object ID provided');
                UIUtils.navigateTo('risk-objects.html');
            }
        },
        
        async fetchRiskObjectDetails() {
            this.loading = true;
            try {
                this.riskObject = await RiskObjectDetailAPI.fetchRiskObject(this.objectId);
                console.log('Fetched risk object:', this.riskObject);
                
                await Promise.all([
                    this.fetchEvents(),
                    this.fetchDetections(),
                    this.fetchRiskAlerts()
                ]);
                
                console.log('Events loaded:', this.events.length, 'events');
                console.log('Events data:', this.events);
                
                this.loading = false;
                
                // Wait for Alpine.js to update the DOM (hide loading state, show chart container)
                // Use both setTimeout and requestAnimationFrame to ensure the DOM is fully updated
                setTimeout(() => {
                    requestAnimationFrame(() => {
                        console.log('DOM should be updated, checking canvas visibility...');
                        const canvas = document.getElementById('riskTrendChart');
                        const container = canvas?.parentElement;
                        console.log('Canvas element:', canvas);
                        console.log('Container element:', container);
                        console.log('Container computed style:', container ? getComputedStyle(container).display : 'N/A');
                        console.log('Container offsetWidth:', container?.offsetWidth);
                        console.log('Container offsetHeight:', container?.offsetHeight);
                        
                        console.log('Initializing chart...');
                        this.initRiskTrendChart();
                    });
                }, 100);
            } catch (error) {
                console.error('Error fetching risk object details:', error);
                this.loading = false;
                UIUtils.showAlert('Error loading risk object details');
                UIUtils.navigateTo('risk-objects.html');
            }
        },
        
        async fetchEvents() {
            try {
                this.events = await RiskObjectDetailAPI.fetchEventsByEntity(this.objectId);
            } catch (error) {
                console.error('Error fetching events:', error);
                this.events = [];
            }
        },
        
        async fetchDetections() {
            try {
                const detectionsList = await RiskObjectDetailAPI.fetchDetections();
                this.detections = {};
                detectionsList.forEach(d => {
                    this.detections[d.id] = d;
                });
            } catch (error) {
                console.error('Error fetching detections:', error);
            }
        },
        
        async fetchRiskAlerts() {
            try {
                // For now, return empty array - could be implemented to fetch alerts for this specific risk object
                this.riskAlerts = [];
            } catch (error) {
                console.error('Error fetching risk alerts:', error);
                this.riskAlerts = [];
            }
        },
        
        getDetectionName(detectionId) {
            return this.detections[detectionId]?.name || `Detection ${detectionId}`;
        },
        
        formatTimestamp(timestamp) {
            return UIUtils.formatDate(timestamp);
        },
        
        getRiskScoreClass(score) {
            if (score >= 80) return 'risk-critical';
            if (score >= 60) return 'risk-high';
            if (score >= 40) return 'risk-medium';
            if (score >= 20) return 'risk-low';
            return 'risk-minimal';
        },
        
        getFalsePositiveClass(isFalsePositive) {
            return isFalsePositive ? 'badge badge-warning' : 'badge badge-success';
        },
        
        getFalsePositiveText(isFalsePositive) {
            return isFalsePositive ? 'False Positive' : 'Valid';
        },
        
        viewEvent(eventId) {
            UIUtils.navigateTo(`events-detail.html?id=${eventId}`);
        },
        
        viewDetection(detectionId) {
            UIUtils.navigateTo(`detections-detail.html?id=${detectionId}`);
        },
        
        getRiskLevelClass(score) {
            if (score >= 80) return 'risk-critical';
            if (score >= 60) return 'risk-high';
            if (score >= 40) return 'risk-medium';
            if (score >= 20) return 'risk-low';
            return 'risk-minimal';
        },
        
        getRiskLevelText(score) {
            if (score >= 80) return 'Critical';
            if (score >= 60) return 'High';
            if (score >= 40) return 'Medium';
            if (score >= 20) return 'Low';
            return 'Minimal';
        },
        
        getEntityTypeIcon(entityType) {
            switch (entityType) {
                case 'user': return '👤';
                case 'host': return '💻';
                case 'ip': return '🌐';
                default: return '📊';
            }
        },
        
        calculateRiskScoreHistory() {
            // Show current score even if no events
            if (!this.events || this.events.length === 0) {
                if (this.riskObject && this.riskObject.current_score > 0) {
                    return [{
                        timestamp: new Date(this.riskObject.last_seen || new Date()),
                        score: this.riskObject.current_score,
                        change: 0,
                        event: null
                    }];
                }
                return [];
            }
            
            // Sort events by timestamp
            const sortedEvents = [...this.events].sort((a, b) => 
                new Date(a.timestamp) - new Date(b.timestamp)
            );
            
            const history = [];
            let currentScore = 0;
            
            // Calculate cumulative score over time
            sortedEvents.forEach(event => {
                if (!event.is_false_positive) {
                    const previousScore = currentScore;
                    currentScore += event.risk_points || 0;
                    
                    history.push({
                        timestamp: new Date(event.timestamp),
                        score: currentScore,
                        change: currentScore - previousScore,
                        event: event
                    });
                }
            });
            
            // Add current score as the latest entry if we have a risk object
            if (this.riskObject && history.length > 0) {
                const lastHistoryScore = history[history.length - 1].score;
                if (this.riskObject.current_score !== lastHistoryScore) {
                    history.push({
                        timestamp: new Date(),
                        score: this.riskObject.current_score,
                        change: this.riskObject.current_score - lastHistoryScore,
                        event: null
                    });
                }
            }
            
            return history.reverse(); // Show most recent first
        },
        
        initRiskTrendChart() {
            console.log('initRiskTrendChart called');
            console.log('Risk object:', this.riskObject);
            console.log('Events:', this.events);
            
            // Prevent duplicate initialization
            if (this.chartInitialized) {
                console.log('Chart already initialized, skipping');
                return;
            }
            
            // Don't initialize if we have absolutely no data
            if (!this.riskObject) {
                console.log('No risk object, skipping chart init');
                return;
            }
            
            const canvas = document.getElementById('riskTrendChart');
            if (!canvas) {
                console.log('Canvas element not found');
                return;
            }
            console.log('Canvas found:', canvas);
            
            // Check if the canvas container is visible
            const container = canvas.parentElement;
            if (!container || container.offsetWidth === 0 || container.offsetHeight === 0) {
                console.log('Canvas container is not visible or has zero dimensions');
                console.log('Container:', container);
                console.log('Container dimensions:', container?.offsetWidth, 'x', container?.offsetHeight);
                console.log('Container display style:', getComputedStyle(container).display);
                console.log('Skipping chart initialization - container not ready');
                return;
            }
            
            // Destroy existing chart if it exists
            if (this.chart) {
                this.chart.destroy();
                this.chart = null;
            }
            
            let history = this.calculateRiskScoreHistory().reverse(); // Show chronological order for chart
            console.log('History for chart:', history);
            
            // If no history but we have a current score, create a minimal chart with current score
            if (history.length === 0 && this.riskObject.current_score > 0) {
                history = [{
                    timestamp: new Date(this.riskObject.last_seen || new Date()),
                    score: this.riskObject.current_score,
                    change: 0,
                    event: null
                }];
                console.log('Using current score for chart:', history);
            }
            
            if (history.length === 0) {
                console.log('No history data available, skipping chart');
                return;
            }
            
            // Get the canvas context
            const ctx = canvas.getContext('2d');
            console.log('Got canvas context:', ctx);
            
            // Clear any existing dimensions to let Chart.js handle sizing
            canvas.style.width = '';
            canvas.style.height = '';
            
            try {
                this.chart = new Chart(ctx, {
                    type: 'line',
                    data: {
                        labels: history.map(h => this.formatTimestamp(h.timestamp)),
                        datasets: [{
                            label: 'Risk Score',
                            data: history.map(h => h.score),
                            borderColor: '#dc3545',
                            backgroundColor: 'rgba(220, 53, 69, 0.1)',
                            borderWidth: 2,
                            fill: true,
                            tension: 0.4
                        }]
                    },
                    options: {
                        responsive: true,
                        maintainAspectRatio: false,
                        scales: {
                            y: {
                                beginAtZero: true,
                                title: {
                                    display: true,
                                    text: 'Risk Score'
                                }
                            },
                            x: {
                                title: {
                                    display: true,
                                    text: 'Time'
                                }
                            }
                        },
                        plugins: {
                            legend: {
                                display: false
                            },
                            title: {
                                display: true,
                                text: 'Risk Score Over Time'
                            }
                        }
                    }
                });
                
                this.chartInitialized = true;
                console.log('Chart created successfully:', this.chart);
            } catch (error) {
                console.error('Error creating chart:', error);
            }
        },
        
        // Method to reset and reinitialize chart (for debugging)
        resetChart() {
            console.log('Resetting chart...');
            this.chartInitialized = false;
            if (this.chart) {
                this.chart.destroy();
                this.chart = null;
            }
            
            setTimeout(() => {
                this.initRiskTrendChart();
            }, 100);
        }
    };
}