class Router {
    constructor() {
        this.routes = {
            '/': 'feed-section',
            '/login': 'login-section',
            '/register': 'register-section',
            '/post': 'post-section',
            '/chat': 'chat-section'
        };

        this.params = {};
        window.addEventListener('popstate', () => this.handleRoute());
        this.isAuthenticated = false;
    }

    init() {
        this.handleRoute();
    }

    setAuthenticated(status) {
        this.isAuthenticated = status;
        document.body.classList.toggle('authenticated', status);
        this.handleRoute();
    }

    navigate(path, params = {}) {
        this.params = params;
        window.history.pushState({}, '', path);
        this.handleRoute();
    }

    handleRoute() {
        const path = window.location.pathname;
        
        // Parse URL parameters
        const urlParams = new URLSearchParams(window.location.search);
        this.params = Object.fromEntries(urlParams.entries());
        
        // Redirect to login if not authenticated
        if (!this.isAuthenticated && path !== '/login' && path !== '/register') {
            this.navigate('/login');
            return;
        }

        // Hide all sections
        document.querySelectorAll('.section').forEach(section => {
            section.classList.remove('active');
        });

        // Show the appropriate section
        const sectionId = this.routes[path] || this.routes['/'];
        const section = document.getElementById(sectionId);
        if (section) {
            section.classList.add('active');
            
            // Handle specific section initialization
            if (path === '/chat' && this.params.userId) {
                this.initializeChat(this.params.userId);
            } else if (path === '/post') {
                this.resetCommentForm();
            }

            // Dispatch route change event
            window.dispatchEvent(new CustomEvent('routeChanged', {
                detail: { path, params: this.params }
            }));
        }
    }

    initializeChat(userId) {
        const chatSection = document.getElementById('chat-section');
        if (chatSection) {
            chatSection.setAttribute('data-user-id', userId);
            window.dispatchEvent(new CustomEvent('chatInitialized', {
                detail: { userId }
            }));
        }
    }

    resetCommentForm() {
        const commentForm = document.getElementById('comment-form');
        if (commentForm) {
            commentForm.reset();
        }
    }

    getParams() {
        return this.params;
    }
}
export default new Router();