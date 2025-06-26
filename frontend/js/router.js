class Router {
    constructor() {
        this.routes = {
            '/': 'feed-section',
            '/login': 'login-section',
            '/register': 'register-section',
            '/post': 'post-section',
            '/chat': 'chat-section'
        };

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

    navigate(path) {
        window.history.pushState({}, '', path);
        this.handleRoute();
    }

    handleRoute() {
        const path = window.location.pathname;
        
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

            

            // Refresh chat when navigating to chat section
           if (path === '/chat' && window.Chat) {
            if (!window.Chat.isInitialized) {
                window.Chat.initializeChat();
            }
            if (window.chatUI) {
                window.chatUI.refreshChatUI();
            }
        }
            

            // Reset comment form if navigating away from post view
            if (path !== '/post') {
                const commentForm = document.getElementById('comment-form');
                if (commentForm) {
                    commentForm.reset();
                }
            }
        }
    }
}

const router = new Router();