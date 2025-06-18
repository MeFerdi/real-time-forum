import API from './api.js';
import router from './router.js';
import views from './views.js';
import WebSocketClient from './ws.js';

// Create Forum namespace
const Forum = {
    api: API,
    router: router,
    views: views,
    ws: null,
    
    async init() {
        // Initialize core components
        this.router.init();
        await this.views.init();
        
        // Setup WebSocket if authenticated
        if (this.router.isAuthenticated) {
            this.ws = new WebSocketClient();
            this.setupWebSocketHandlers();
        }
        
        await this.loadCategories();
        this.setupEventListeners();
    },

    async loadCategories() {
        const result = await this.api.getCategories();
        if (result.success) {
            const categoryFilter = document.getElementById('categoryFilter');
            if (categoryFilter) {
                categoryFilter.innerHTML = '<option value="">All Categories</option>';
                result.data.forEach(category => {
                    const option = document.createElement('option');
                    option.value = category.id;
                    option.textContent = category.name;
                    categoryFilter.appendChild(option);
                });
            }
        }
    },

    setupWebSocketHandlers() {
        if (!this.ws || !this.ws.socket) {
            console.error('WebSocket not initialized');
            return;
        }
    
        this.ws.socket.addEventListener('message', event => {
            try {
                const message = JSON.parse(event.data);
                
                switch (message.type) {
                    case 'chat':
                        this.handleChatMessage(message.data);
                        break;
                    case 'status':
                        this.updateUserStatus(message.data);
                        break;
                    case 'users':
                        this.updateOnlineUsers(message.data);
                        break;
                    default:
                        console.warn('Unknown message type:', message.type);
                }
            } catch (error) {
                console.error('WebSocket message error:', error);
            }
        });
    
        // Add connection status handlers
        this.ws.socket.addEventListener('open', () => {
            console.log('WebSocket Connected');
        });
    
        this.ws.socket.addEventListener('close', () => {
            console.log('WebSocket Disconnected');
            setTimeout(() => this.initWebSocket(), 5000);
        });
    
        this.ws.socket.addEventListener('error', (error) => {
            console.error('WebSocket Error:', error);
        });
    },
    setupEventListeners() {
        // Chat form handler
        const chatForm = document.getElementById('chat-form');
        if (chatForm) {
            chatForm.addEventListener('submit', this.handleChatSubmit.bind(this));
        }

        // Category filter handler
        const categoryFilter = document.getElementById('categoryFilter');
        if (categoryFilter) {
            categoryFilter.addEventListener('change', (e) => {
                this.views.loadPosts(e.target.value);
            });
        }

        // Post form handler
        const postForm = document.getElementById('quick-post-form');
        if (postForm) {
            postForm.addEventListener('submit', this.handlePostSubmit.bind(this));
        }
    },

    handleChatSubmit(e) {
        e.preventDefault();
        const form = e.target;
        const content = form.message.value.trim();
        const receiverId = form.getAttribute('data-receiver-id');
        
        if (content && receiverId) {
            this.ws.sendMessage({
                type: 'chat',
                data: {
                    receiver_id: parseInt(receiverId),
                    content: content
                }
            });
            form.reset();
        }
    },

    handlePostSubmit(e) {
        e.preventDefault();
        const form = e.target;
        const postData = {
            title: form.title.value.trim(),
            content: form.content.value.trim(),
            category_id: parseInt(form.category.value)
        };

        if (postData.title && postData.content && postData.category_id) {
            this.api.createPost(postData).then(result => {
                if (result.success) {
                    form.reset();
                    this.views.loadPosts();
                }
            });
        }
    },

    handleChatMessage(message) {
        const chatContainer = document.getElementById('chat-messages');
        if (chatContainer) {
            const messageElement = this.createMessageElement(message);
            chatContainer.appendChild(messageElement);
            chatContainer.scrollTop = chatContainer.scrollHeight;
        }
    },

    createMessageElement(message) {
        const div = document.createElement('div');
        div.className = `message ${message.sender_id === this.views.getCurrentUser()?.id ? 'sent' : 'received'}`;
        div.innerHTML = `
            <div class="message-content">${message.content}</div>
            <div class="message-meta">
                <span class="message-time">${new Date(message.created_at).toLocaleTimeString()}</span>
            </div>
        `;
        return div;
    },

    updateUserStatus(data) {
        const userElement = document.querySelector(`[data-user-id="${data.user_id}"]`);
        if (userElement) {
            userElement.classList.toggle('online', data.online);
        }
    },

    updateOnlineUsers(users) {
        const usersList = document.getElementById('users-list');
        if (usersList) {
            usersList.innerHTML = users.map(user => this.createUserElement(user)).join('');
        }
    },

    createUserElement(user) {
        return `
            <div class="user-item ${user.online ? 'online' : 'offline'}" 
                 data-user-id="${user.id}">
                <div class="user-avatar">
                    <img src="https://ui-avatars.com/api/?name=${user.username}" alt="${user.username}">
                    <span class="status-indicator"></span>
                </div>
                <div class="user-info">
                    <span class="user-name">${user.username}</span>
                    <span class="last-message">${user.last_message || ''}</span>
                </div>
            </div>
        `;
    }
};

// Initialize when DOM is ready
document.addEventListener('DOMContentLoaded', () => {
    window.Forum = Forum;
    Forum.init();
});

export default Forum;