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
                // Handle authentication errors specifically
                if (response.status === 401) {
                    console.warn('Authentication required for:', endpoint);
                    return { success: false, error: 'Authentication required', status: 401 };
                }
                throw new Error(`HTTP error! status: ${response.status}`);
            }
            const data = await response.json();
            console.log(`API ${endpoint} response:`, data); // Debug logging
            return { success: true, data };
        } catch (error) {
            console.error('API Error for', endpoint, ':', error);
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
        try {
            const response = await fetch(`${this.baseUrl}/logout`, {
                method: 'POST',
                credentials: 'include'
            });

            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            // Try to parse JSON response, fall back to success if response is empty
            try {
                const data = await response.json();
                return { success: data.success };
            } catch (e) {
                // If response is empty or invalid JSON, check if request was successful
                return { success: response.ok };
            }
        } catch (error) {
            console.error('Logout Error:', error);
            return { success: false, error: error.message };
        }
    },

    async getProfile() {
        return await this.request('/profile');
    },

    // Post endpoints
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
            body: JSON.stringify({ content: commentData.content }),
            credentials: 'include'
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

    async getCategories() {
        return await this.request('/categories');
    },

    // Message endpoints
    async getConversations() {
        return await this.request('/messages/conversations');
    },

    async getConversationHistory(userID, limit = 10, offset = 0) {
        return await this.request(`/messages/history?user_id=${userID}&limit=${limit}&offset=${offset}`);
    },

    async sendMessage(receiverID, content) {
        return await this.request('/messages/send', {
            method: 'POST',
            body: JSON.stringify({
                receiver_id: receiverID,
                content: content
            })
        });
    },

    async markMessagesAsRead(senderID) {
        return await this.request('/messages/mark-read', {
            method: 'POST',
            body: JSON.stringify({
                sender_id: senderID
            })
        });
    },

    async getAllUsers() {
        return await this.request('/messages/users');
    }
};