// Frontend performance optimizations

// Lazy loading for images and heavy components
class LazyLoader {
    constructor() {
        this.imageObserver = null;
        this.scriptObserver = null;
        this.init();
    }

    init() {
        // Set up Intersection Observer for images
        if ('IntersectionObserver' in window) {
            this.imageObserver = new IntersectionObserver((entries) => {
                entries.forEach(entry => {
                    if (entry.isIntersecting) {
                        this.loadImage(entry.target);
                        this.imageObserver.unobserve(entry.target);
                    }
                });
            }, {
                rootMargin: '50px 0px',
                threshold: 0.01
            });

            // Observe all lazy images
            document.querySelectorAll('img[data-src]').forEach(img => {
                this.imageObserver.observe(img);
            });
        } else {
            // Fallback for older browsers
            this.loadAllImages();
        }

        // Set up lazy loading for scripts
        this.setupScriptLazyLoading();
    }

    loadImage(img) {
        const src = img.dataset.src;
        if (!src) return;

        // Create a new image to preload
        const newImg = new Image();
        newImg.onload = () => {
            img.src = src;
            img.classList.add('loaded');
            delete img.dataset.src;
        };
        newImg.src = src;
    }

    loadAllImages() {
        document.querySelectorAll('img[data-src]').forEach(img => {
            this.loadImage(img);
        });
    }

    setupScriptLazyLoading() {
        // Scripts are now loaded directly in HTML to avoid conflicts
        // Chart.js and Alpine.js are included in the HTML head
        
        // Check if Chart.js is loaded and initialize charts if needed
        if (window.Chart && window.chartManager) {
            window.chartManager.initializeCharts();
        }
        
        // Alpine.js is auto-initialized when loaded with defer attribute
    }

    loadScript(src, callback) {
        const script = document.createElement('script');
        script.src = src;
        script.async = true;
        script.onload = callback;
        document.head.appendChild(script);
    }
}

// Debounce utility for search and filter inputs
function debounce(func, wait) {
    let timeout;
    return function executedFunction(...args) {
        const later = () => {
            clearTimeout(timeout);
            func(...args);
        };
        clearTimeout(timeout);
        timeout = setTimeout(later, wait);
    };
}

// Virtual scrolling for large lists
class VirtualScroller {
    constructor(container, items, itemHeight = 50) {
        this.container = container;
        this.items = items;
        this.itemHeight = itemHeight;
        this.visibleItems = [];
        this.scrollTop = 0;
        this.containerHeight = 0;
        this.totalHeight = 0;
        this.startIndex = 0;
        this.endIndex = 0;
        
        this.init();
    }

    init() {
        this.containerHeight = this.container.clientHeight;
        this.totalHeight = this.items.length * this.itemHeight;
        
        // Create viewport
        this.viewport = document.createElement('div');
        this.viewport.style.height = `${this.totalHeight}px`;
        this.viewport.style.position = 'relative';
        
        // Create content container
        this.content = document.createElement('div');
        this.content.style.position = 'absolute';
        this.content.style.top = '0';
        this.content.style.left = '0';
        this.content.style.right = '0';
        
        this.viewport.appendChild(this.content);
        this.container.appendChild(this.viewport);
        
        // Add scroll listener
        this.container.addEventListener('scroll', () => {
            this.handleScroll();
        });
        
        // Initial render
        this.render();
    }

    handleScroll() {
        this.scrollTop = this.container.scrollTop;
        this.render();
    }

    render() {
        // Calculate visible range
        this.startIndex = Math.floor(this.scrollTop / this.itemHeight);
        this.endIndex = Math.min(
            this.items.length - 1,
            Math.floor((this.scrollTop + this.containerHeight) / this.itemHeight) + 1
        );
        
        // Update content position
        this.content.style.transform = `translateY(${this.startIndex * this.itemHeight}px)`;
        
        // Render visible items
        const fragment = document.createDocumentFragment();
        for (let i = this.startIndex; i <= this.endIndex; i++) {
            const item = this.createItemElement(this.items[i], i);
            fragment.appendChild(item);
        }
        
        this.content.innerHTML = '';
        this.content.appendChild(fragment);
    }

    createItemElement(item, index) {
        const div = document.createElement('div');
        div.style.height = `${this.itemHeight}px`;
        div.innerHTML = item.html || `Item ${index}`;
        return div;
    }
}

// Request batching for API calls
class RequestBatcher {
    constructor(batchSize = 10, delay = 100) {
        this.batchSize = batchSize;
        this.delay = delay;
        this.queue = [];
        this.timeout = null;
    }

    add(request) {
        return new Promise((resolve, reject) => {
            this.queue.push({ request, resolve, reject });
            
            if (this.queue.length >= this.batchSize) {
                this.flush();
            } else {
                this.scheduleFlush();
            }
        });
    }

    scheduleFlush() {
        if (this.timeout) return;
        
        this.timeout = setTimeout(() => {
            this.flush();
        }, this.delay);
    }

    async flush() {
        if (this.timeout) {
            clearTimeout(this.timeout);
            this.timeout = null;
        }

        if (this.queue.length === 0) return;

        const batch = this.queue.splice(0, this.batchSize);
        
        try {
            // Send batch request
            const response = await fetch('/api/batch', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    requests: batch.map(b => b.request)
                })
            });

            const results = await response.json();
            
            // Resolve individual promises
            batch.forEach((item, index) => {
                if (results[index].error) {
                    item.reject(results[index].error);
                } else {
                    item.resolve(results[index].data);
                }
            });
        } catch (error) {
            // Reject all promises in batch
            batch.forEach(item => item.reject(error));
        }
    }
}

// Prefetching for likely navigation targets
class Prefetcher {
    constructor() {
        this.prefetched = new Set();
        this.init();
    }

    init() {
        // Prefetch on hover
        document.addEventListener('mouseover', (e) => {
            const link = e.target.closest('a[href]');
            if (link && this.shouldPrefetch(link)) {
                this.prefetch(link.href);
            }
        });

        // Prefetch visible links
        if ('IntersectionObserver' in window) {
            const observer = new IntersectionObserver((entries) => {
                entries.forEach(entry => {
                    if (entry.isIntersecting) {
                        const links = entry.target.querySelectorAll('a[href]');
                        links.forEach(link => {
                            if (this.shouldPrefetch(link)) {
                                this.prefetch(link.href);
                            }
                        });
                    }
                });
            });

            // Observe navigation areas
            document.querySelectorAll('nav, .navigation').forEach(nav => {
                observer.observe(nav);
            });
        }
    }

    shouldPrefetch(link) {
        // Don't prefetch external links, downloads, or already prefetched
        return !link.href.includes('#') &&
               !link.href.includes('mailto:') &&
               !link.download &&
               link.hostname === window.location.hostname &&
               !this.prefetched.has(link.href);
    }

    prefetch(url) {
        if (this.prefetched.has(url)) return;
        
        this.prefetched.add(url);
        
        // Use link prefetch
        const link = document.createElement('link');
        link.rel = 'prefetch';
        link.href = url;
        document.head.appendChild(link);
    }
}

// Service Worker for offline support and caching
if ('serviceWorker' in navigator) {
    window.addEventListener('load', () => {
        navigator.serviceWorker.register('/sw.js').catch(() => {
            // Service worker registration failed, app will work without offline support
        });
    });
}

// Initialize performance optimizations
document.addEventListener('DOMContentLoaded', () => {
    // Initialize lazy loader
    window.lazyLoader = new LazyLoader();
    
    // Initialize prefetcher
    window.prefetcher = new Prefetcher();
    
    // Initialize request batcher
    window.requestBatcher = new RequestBatcher();
    
    // Add debouncing to search inputs
    document.querySelectorAll('input[type="search"], input.search, input[data-debounce]').forEach(input => {
        const delay = parseInt(input.dataset.debounce) || 300;
        const originalHandler = input.oninput;
        
        if (originalHandler) {
            input.oninput = debounce(originalHandler, delay);
        }
    });
    
    // Initialize virtual scrolling for large lists
    document.querySelectorAll('[data-virtual-scroll]').forEach(container => {
        const items = Array.from(container.children).map(child => ({
            html: child.outerHTML
        }));
        
        if (items.length > 50) {
            container.innerHTML = '';
            new VirtualScroller(container, items);
        }
    });
    
    // Optimize animations with will-change
    document.querySelectorAll('.card, .btn, [data-animate]').forEach(el => {
        el.addEventListener('mouseenter', () => {
            el.style.willChange = 'transform';
        });
        
        el.addEventListener('mouseleave', () => {
            setTimeout(() => {
                el.style.willChange = 'auto';
            }, 300);
        });
    });
    
    // Use passive event listeners for scroll performance
    document.addEventListener('scroll', () => {}, { passive: true });
    document.addEventListener('touchstart', () => {}, { passive: true });
    document.addEventListener('wheel', () => {}, { passive: true });
});