// API client for interacting with the backend
const API = {
    baseUrl: '/api',

    async request(endpoint, options = {}) {
        const url = `${this.baseUrl}${endpoint}`;
        options.headers = {
            'Content-Type': 'application/json',
            ...options.headers
        };
        options.credentials = 'include'; // Send cookies for authentication

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

    // Auth endpoints
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
        return await this.request('/logout', { method: 'POST' });
    },

    async getProfile() {
        return await this.request('/profile');
    },

    // Post endpoints
    async getPosts(category = '') {
        const endpoint = category ? `/posts?category=${category}` : '/posts';
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
        return await this.request('/comments', {
            method: 'POST',
            body: JSON.stringify(commentData)
        });
    },

    async likePost(postId) {
        return await this.request('/posts/like', {
            method: 'POST',
            body: JSON.stringify({ post_id: postId })
        });
    },

    async likeComment(commentId) {
        return await this.request('/comments/like', {
            method: 'POST',
            body: JSON.stringify({ comment_id: commentId })
        });
    },

    async getCategories() {
        return await this.request('/categories');
    }
};