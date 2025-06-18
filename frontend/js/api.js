const API = {
    baseUrl: '/api',
    ws: null,
    messageCallbacks: [],
    reconnectTimeout: null,
    isConnecting: false,

    // WebSocket Management
    initWebSocket() {
        if (this.isConnecting) return;
        this.isConnecting = true;

        this.ws = new WebSocket(`ws://${window.location.host}/ws`);
        
        this.ws.onopen = () => {
            this.isConnecting = false;
            console.log('WebSocket Connected');
        };

        this.ws.onmessage = (event) => {
            const message = JSON.parse(event.data);
            this.messageCallbacks.forEach(callback => callback(message));
        };

        this.ws.onclose = () => {
            this.isConnecting = false;
            if (this.reconnectTimeout) clearTimeout(this.reconnectTimeout);
            this.reconnectTimeout = setTimeout(() => this.initWebSocket(), 1000);
        };

        this.ws.onerror = (error) => {
            console.error('WebSocket Error:', error);
            this.ws.close();
        };
    },

    // Basic HTTP Request Handler
    async request(endpoint, options = {}) {
        const url = `${this.baseUrl}${endpoint}`;
        options.headers = {
            'Content-Type': 'application/json',
            ...options.headers
        };
        options.credentials = 'include';

        try {
            const response = await fetch(url, options);
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }
            const data = await response.json();
            return { success: true, data };
        } catch (error) {
            console.error('API Error:', error);
            return { success: false, error: error.message };
        }
    },

    // Message Handling
    onMessage(callback) {
        this.messageCallbacks.push(callback);
        return () => {
            this.messageCallbacks = this.messageCallbacks.filter(cb => cb !== callback);
        };
    },

    sendMessage(receiverId, content) {
        if (!this.ws || this.ws.readyState !== WebSocket.OPEN) return;
        this.ws.send(JSON.stringify({
            type: 'chat',
            data: {
                receiver_id: receiverId,
                content: content
            }
        }));
    },

    // Authentication Endpoints
    async register(userData) {
        return await this.request('/register', {
            method: 'POST',
            body: JSON.stringify(userData)
        });
    },

    async login(credentials) {
        return await this.request('/login', {
            method: 'POST',
            body: JSON.stringify(credentials)
        });
    },

    async logout() {
        try {
            const response = await fetch(`${this.baseUrl}/logout`, {
                method: 'POST',
                credentials: 'include'
            });
            this.ws?.close();
            return { success: response.ok };
        } catch (error) {
            console.error('Logout Error:', error);
            return { success: false, error: error.message };
        }
    },

    // User Endpoints
    async getProfile() {
        return await this.request('/profile');
    },

    async getUsers() {
        console.log('Fetching users from:', 'api/api/users');
        return await this.request('/api/users');
    },

    async getOnlineUsers() {
        return await this.request('/users/online');
    },

    async getUserById(userId) {
        return await this.request(`api/users/${userId}`);
    },

    async updateUserStatus(status) {
        return await this.request('api/users/status', {
            method: 'POST',
            body: JSON.stringify({ status })
        });
    },
    // Chat Endpoints
    async getChatHistory(userId, offset = 0) {
        return await this.request(`/messages?user_id=${userId}&offset=${offset}`);
    },

    async markMessageAsRead(messageId) {
        return await this.request(`/messages/${messageId}/read`, {
            method: 'POST'
        });
    },

    // Post Endpoints
    async getPosts(categoryId = '') {
        const endpoint = categoryId ? `/posts?category_id=${categoryId}` : '/posts';
        return await this.request(endpoint);
    },

    async getPost(postId) {
        return await this.request(`/posts/get?id=${postId}`);
    },

    async createPost(postData) {
        return await this.request('/posts/create', {
            method: 'POST',
            body: JSON.stringify(postData)
        });
    },

    async createComment(commentData) {
        return await this.request(`/posts/${commentData.post_id}/comments`, {
            method: 'POST',
            body: JSON.stringify({ content: commentData.content })
        });
    },

    async likePost(postId) {
        return await this.request(`/posts/like?post_id=${postId}`, {
            method: 'POST'
        });
    },

    async likeComment(commentId) {
        return await this.request(`/comments/like?comment_id=${commentId}`, {
            method: 'POST'
        });
    },

    // Category Endpoints
    async getCategories() {
        return await this.request('/categories');
    }
};

// Initialize WebSocket connection
if (document.readyState === 'complete') {
    API.initWebSocket();
} else {
    window.addEventListener('load', () => API.initWebSocket());
}
export default API;