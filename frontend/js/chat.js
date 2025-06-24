// Chat UI management
class ChatUI {
    constructor() {
        this.currentConversation = null;
        this.conversations = new Map();
        this.onlineUsers = new Map();
        this.messageHistory = new Map();
        this.typingTimeout = null;
        this.loadingMore = false;
        this.scrollTimeout = null;
        this.searchTimeout = null;

        this.bindEvents();
        this.initializeChat();
    }

    bindEvents() {
        // Chat form submission
        const chatForm = document.getElementById('chat-form');
        if (chatForm) {
            chatForm.addEventListener('submit', (e) => this.handleSendMessage(e));
        }

        // Message input typing indicator
        const messageInput = document.getElementById('message-input');
        if (messageInput) {
            messageInput.addEventListener('input', () => this.handleTyping());
        }

        // User search (with debouncing)
        const userSearch = document.getElementById('user-search');
        if (userSearch) {
            userSearch.addEventListener('input', (e) => this.debouncedHandleUserSearch(e.target.value));
        }

        // New chat button
        const newChatBtn = document.getElementById('new-chat-btn');
        if (newChatBtn) {
            newChatBtn.addEventListener('click', () => this.showUserList());
        }

        // Messages container scroll for pagination (with throttling)
        const messagesContainer = document.getElementById('messages-container');
        if (messagesContainer) {
            messagesContainer.addEventListener('scroll', () => this.throttledHandleScroll());
        }

        // WebSocket status listener
        window.addEventListener('websocket-status', (e) => {
            this.handleConnectionStatus(e.detail.connected);
        });

        // Mobile back button
        const backBtn = document.getElementById('back-to-conversations');
        if (backBtn) {
            backBtn.addEventListener('click', () => this.showConversationsList());
        }
    }

    async initializeChat() {
        // Only initialize if user is authenticated
        if (!window.views || !window.views.currentUser) {
            console.log('User not authenticated, skipping chat initialization');
            return;
        }

        try {
            console.log('Initializing chat for user:', window.views.currentUser.username);

            // Load conversations
            await this.loadConversations();

            // Load all users for new conversations
            await this.loadAllUsers();

        } catch (error) {
            console.error('Failed to initialize chat:', error);
        }
    }

    async loadConversations() {
        try {
            const result = await API.request('/messages/conversations');
            if (result.success && result.data) {
                this.conversations.clear();
                result.data.forEach(conv => {
                    this.conversations.set(conv.user_id, conv);
                });
                this.renderConversations();
            }
        } catch (error) {
            console.error('Failed to load conversations:', error);
            // Don't redirect on auth errors here, let the main app handle it
        }
    }

    async loadAllUsers() {
        try {
            const result = await API.request('/messages/users');
            if (result.success && result.data) {
                result.data.forEach(user => {
                    if (!this.conversations.has(user.id)) {
                        // Add users without conversations
                        this.conversations.set(user.id, {
                            user_id: user.id,
                            username: user.username,
                            first_name: user.first_name,
                            last_name: user.last_name,
                            last_message: null,
                            unread_count: 0,
                            is_online: false
                        });
                    }
                });
                this.renderConversations();
            }
        } catch (error) {
            console.error('Failed to load users:', error);
            // Don't redirect on auth errors here, let the main app handle it
        }
    }

    renderConversations() {
        const conversationsList = document.getElementById('conversations-list');
        if (!conversationsList) return;

        // Sort conversations by last message time, then alphabetically
        const sortedConversations = Array.from(this.conversations.values()).sort((a, b) => {
            if (a.last_message && b.last_message) {
                return new Date(b.last_message.created_at) - new Date(a.last_message.created_at);
            } else if (a.last_message) {
                return -1;
            } else if (b.last_message) {
                return 1;
            } else {
                return a.username.localeCompare(b.username);
            }
        });

        conversationsList.innerHTML = sortedConversations.map(conv => {
            const isOnline = this.onlineUsers.has(conv.user_id);
            const lastMessageTime = conv.last_message ? 
                this.formatRelativeTime(conv.last_message.created_at) : '';
            const lastMessagePreview = conv.last_message ? 
                conv.last_message.content.substring(0, 50) + (conv.last_message.content.length > 50 ? '...' : '') : 
                'No messages yet';

            return `
                <div class="conversation-item ${this.currentConversation === conv.user_id ? 'active' : ''}" 
                     data-user-id="${conv.user_id}" 
                     onclick="window.chatUI.selectConversation(${conv.user_id})">
                    <div class="conversation-avatar">
                        <img src="https://ui-avatars.com/api/?name=${conv.username}&background=random" 
                             alt="${conv.username}'s avatar" />
                        <div class="status-indicator ${isOnline ? 'online' : 'offline'}"></div>
                    </div>
                    <div class="conversation-info">
                        <div class="conversation-header">
                            <span class="username">${conv.username}</span>
                            <span class="timestamp">${lastMessageTime}</span>
                        </div>
                        <div class="conversation-preview">
                            <span class="last-message">${lastMessagePreview}</span>
                            ${conv.unread_count > 0 ? `<span class="unread-badge">${conv.unread_count}</span>` : ''}
                        </div>
                    </div>
                </div>
            `;
        }).join('');
    }

    async selectConversation(userID) {
        this.currentConversation = userID;

        // Update WebSocket client
        if (window.wsClient) {
            window.wsClient.setCurrentConversation(userID);
        }

        // Update UI
        this.updateChatHeader(userID);
        this.enableMessageInput();

        // Mobile: Show chat main and hide sidebar
        if (window.innerWidth < 768) {
            const chatSidebar = document.querySelector('.chat-sidebar');
            const chatMain = document.querySelector('.chat-main');
            const backBtn = document.getElementById('back-to-conversations');

            if (chatSidebar && chatMain) {
                chatSidebar.style.display = 'none';
                chatMain.classList.add('active');
                chatMain.style.display = 'flex';
                if (backBtn) backBtn.style.display = 'flex';
            }
        }

        // Load message history
        await this.loadMessageHistory(userID);

        // Mark messages as read
        await this.markMessagesAsRead(userID);

        // Update conversations list
        this.renderConversations();
    }

    updateChatHeader(userID) {
        const conversation = this.conversations.get(userID);
        if (!conversation) return;

        const chatHeader = document.getElementById('chat-header');
        const isOnline = this.onlineUsers.has(userID);
        
        chatHeader.innerHTML = `
            <div class="chat-user-info">
                <div class="avatar">
                    <img src="https://ui-avatars.com/api/?name=${conversation.username}&background=random" 
                         alt="${conversation.username}'s avatar" />
                    <div class="status-indicator ${isOnline ? 'online' : 'offline'}"></div>
                </div>
                <div class="user-details">
                    <span class="username">${conversation.username}</span>
                    <span class="status">${isOnline ? 'Online' : 'Offline'}</span>
                </div>
            </div>
        `;
    }

    enableMessageInput() {
        const messageInput = document.getElementById('message-input');
        const sendBtn = document.getElementById('send-btn');
        
        if (messageInput) {
            messageInput.disabled = false;
            messageInput.placeholder = 'Type a message...';
        }
        if (sendBtn) {
            sendBtn.disabled = false;
        }
    }

    async loadMessageHistory(userID, offset = 0) {
        try {
            const result = await API.request(`/messages/history?user_id=${userID}&limit=10&offset=${offset}`);
            if (result.success && result.data) {
                const messages = Array.isArray(result.data) ? result.data : [];

                if (offset === 0) {
                    // Clear existing messages for new conversation
                    this.messageHistory.set(userID, messages);
                    this.renderMessages(messages);
                } else {
                    // Prepend older messages
                    const existingMessages = this.messageHistory.get(userID) || [];
                    const allMessages = [...messages, ...existingMessages];
                    this.messageHistory.set(userID, allMessages);
                    this.prependMessages(messages);
                }
            } else {
                // Handle case where no messages exist
                if (offset === 0) {
                    this.messageHistory.set(userID, []);
                    this.renderMessages([]);
                }
            }
        } catch (error) {
            console.error('Failed to load message history:', error);
            // Handle authentication errors
            if (error.message.includes('401')) {
                console.log('User not authenticated, redirecting to login');
                router.navigate('/login');
            }
        }
    }

    renderMessages(messages) {
        const messagesContainer = document.getElementById('messages-container');
        if (!messagesContainer) return;

        if (messages.length === 0) {
            messagesContainer.innerHTML = `
                <div class="no-messages">
                    <i class="fas fa-comments"></i>
                    <p>No messages yet. Start the conversation!</p>
                </div>
            `;
            return;
        }

        messagesContainer.innerHTML = messages.map(msg => this.renderMessage(msg)).join('');
        this.scrollToBottom();
    }

    prependMessages(messages) {
        const messagesContainer = document.getElementById('messages-container');
        if (!messagesContainer || messages.length === 0) return;

        const oldScrollHeight = messagesContainer.scrollHeight;
        const messagesHTML = messages.map(msg => this.renderMessage(msg)).join('');
        messagesContainer.insertAdjacentHTML('afterbegin', messagesHTML);
        
        // Maintain scroll position
        const newScrollHeight = messagesContainer.scrollHeight;
        messagesContainer.scrollTop = newScrollHeight - oldScrollHeight;
    }

    renderMessage(message) {
        const currentUserId = window.views && window.views.currentUser ? window.views.currentUser.id : 0;
        const isOwn = message.sender_id === currentUserId;
        const timestamp = this.formatMessageTime(message.created_at);

        return `
            <div class="message ${isOwn ? 'own' : 'other'}">
                <div class="message-content">
                    <div class="message-text">${this.escapeHtml(message.content)}</div>
                    <div class="message-meta">
                        <span class="sender">${message.sender ? message.sender.username : 'Unknown'}</span>
                        <span class="timestamp">${timestamp}</span>
                    </div>
                </div>
            </div>
        `;
    }

    addMessage(messageData) {
        const currentUserId = window.views && window.views.currentUser ? window.views.currentUser.id : 0;

        // Determine the conversation partner
        const conversationPartnerId = messageData.sender_id === currentUserId ?
                                    messageData.receiver_id : messageData.sender_id;

        // Add to message history
        const messages = this.messageHistory.get(conversationPartnerId) || [];
        messages.push(messageData);
        this.messageHistory.set(conversationPartnerId, messages);

        // Update conversation
        let conversation = this.conversations.get(conversationPartnerId);
        if (conversation) {
            conversation.last_message = messageData;
            // Only increment unread count for received messages
            if (messageData.sender_id !== currentUserId) {
                conversation.unread_count = (conversation.unread_count || 0) + 1;
            }
        } else {
            // Create new conversation if it doesn't exist
            conversation = {
                user_id: conversationPartnerId,
                username: messageData.sender ? messageData.sender.username : 'Unknown',
                first_name: messageData.sender ? messageData.sender.first_name : '',
                last_name: messageData.sender ? messageData.sender.last_name : '',
                last_message: messageData,
                unread_count: messageData.sender_id !== currentUserId ? 1 : 0,
                is_online: this.onlineUsers.has(conversationPartnerId)
            };
            this.conversations.set(conversationPartnerId, conversation);
        }

        // Update UI if this is the current conversation
        if (this.currentConversation === conversationPartnerId) {
            const messagesContainer = document.getElementById('messages-container');
            if (messagesContainer) {
                // Remove "no messages" placeholder if it exists
                const noMessages = messagesContainer.querySelector('.no-messages');
                if (noMessages) {
                    noMessages.remove();
                }

                const messageHTML = this.renderMessage(messageData);
                messagesContainer.insertAdjacentHTML('beforeend', messageHTML);
                this.scrollToBottom();

                // Auto-mark as read if it's the current conversation and message is from other user
                if (messageData.sender_id !== currentUserId) {
                    this.markMessagesAsRead(messageData.sender_id);
                    conversation.unread_count = 0;
                }
            }
        }

        // Update conversations list to reflect new message
        this.renderConversations();

        console.log('Message added:', messageData);
    }

    async handleSendMessage(e) {
        e.preventDefault();

        if (!this.currentConversation) return;

        const messageInput = document.getElementById('message-input');
        const content = messageInput.value.trim();

        if (!content) return;

        const currentUserId = window.views && window.views.currentUser ? window.views.currentUser.id : 0;

        // Create optimistic message for immediate UI update
        const optimisticMessage = {
            id: Date.now(), // Temporary ID
            sender_id: currentUserId,
            receiver_id: this.currentConversation,
            content: content,
            is_read: false,
            created_at: new Date().toISOString(),
            sender: window.views.currentUser
        };

        // Clear input immediately for better UX
        messageInput.value = '';

        // Send via WebSocket if connected, otherwise use HTTP API
        let success = false;
        if (window.wsClient && window.wsClient.isConnected) {
            success = window.wsClient.sendPrivateMessage(this.currentConversation, content);
            if (success) {
                // Add optimistic message immediately for WebSocket
                this.addMessage(optimisticMessage);
            }
        } else {
            // Fallback to HTTP API
            try {
                const result = await API.request('/messages/send', {
                    method: 'POST',
                    body: JSON.stringify({
                        receiver_id: this.currentConversation,
                        content: content
                    })
                });
                success = result.success;
                if (success) {
                    this.addMessage(result.data);
                }
            } catch (error) {
                console.error('Failed to send message:', error);
                // Restore input value on error
                messageInput.value = content;
            }
        }

        if (!success && window.wsClient && window.wsClient.isConnected) {
            // Restore input value on WebSocket error
            messageInput.value = content;
        }
    }

    handleTyping() {
        if (!this.currentConversation || !window.wsClient || !window.wsClient.isConnected) return;

        // Clear existing timeout
        if (this.typingTimeout) {
            clearTimeout(this.typingTimeout);
        }

        // Send typing indicator
        window.wsClient.sendTypingIndicator(this.currentConversation);

        // Set timeout to stop typing indicator
        this.typingTimeout = setTimeout(() => {
            // Could send "stop typing" indicator here if implemented
        }, 3000);
    }

    showTypingIndicator(senderID, username) {
        if (this.currentConversation !== senderID) return;

        const typingIndicator = document.getElementById('typing-indicator');
        if (typingIndicator) {
            typingIndicator.querySelector('span').textContent = username;
            typingIndicator.style.display = 'block';

            // Hide after 3 seconds
            setTimeout(() => {
                typingIndicator.style.display = 'none';
            }, 3000);
        }
    }

    updateUserStatus(userID, status) {
        if (status === 'online') {
            this.onlineUsers.set(userID, true);
        } else {
            this.onlineUsers.delete(userID);
        }

        // Update conversation if it exists
        const conversation = this.conversations.get(userID);
        if (conversation) {
            conversation.is_online = status === 'online';
        }

        // Update UI
        this.renderConversations();
        if (this.currentConversation === userID) {
            this.updateChatHeader(userID);
        }
    }

    updateOnlineUsers(users) {
        this.onlineUsers.clear();
        users.forEach(user => {
            this.onlineUsers.set(user.user_id, true);
            
            // Update conversation status
            const conversation = this.conversations.get(user.user_id);
            if (conversation) {
                conversation.is_online = true;
            }
        });

        this.renderConversations();
        if (this.currentConversation) {
            this.updateChatHeader(this.currentConversation);
        }
    }

    async markMessagesAsRead(senderID) {
        try {
            await API.request('/messages/mark-read', {
                method: 'POST',
                body: JSON.stringify({ sender_id: senderID })
            });

            // Update conversation unread count
            const conversation = this.conversations.get(senderID);
            if (conversation) {
                conversation.unread_count = 0;
            }
        } catch (error) {
            console.error('Failed to mark messages as read:', error);
        }
    }

    handleConnectionStatus(connected) {
        const messageInput = document.getElementById('message-input');
        const sendBtn = document.getElementById('send-btn');
        
        if (!connected) {
            if (messageInput) messageInput.placeholder = 'Reconnecting...';
            if (sendBtn) sendBtn.disabled = true;
        } else if (this.currentConversation) {
            if (messageInput) messageInput.placeholder = 'Type a message...';
            if (sendBtn) sendBtn.disabled = false;
        }
    }

    throttledHandleScroll() {
        if (this.scrollTimeout) return;

        this.scrollTimeout = setTimeout(() => {
            this.handleScroll();
            this.scrollTimeout = null;
        }, 100); // Throttle to 100ms
    }

    handleScroll() {
        const messagesContainer = document.getElementById('messages-container');
        if (!messagesContainer || this.loadingMore || !this.currentConversation) return;

        // Check if scrolled to top
        if (messagesContainer.scrollTop === 0) {
            this.loadMoreMessages();
        }
    }

    async loadMoreMessages() {
        if (this.loadingMore || !this.currentConversation) return;

        this.loadingMore = true;
        const currentMessages = this.messageHistory.get(this.currentConversation) || [];
        await this.loadMessageHistory(this.currentConversation, currentMessages.length);
        this.loadingMore = false;
    }

    debouncedHandleUserSearch(query) {
        if (this.searchTimeout) {
            clearTimeout(this.searchTimeout);
        }

        this.searchTimeout = setTimeout(() => {
            this.handleUserSearch(query);
        }, 300); // Debounce to 300ms
    }

    handleUserSearch(query) {
        const conversationItems = document.querySelectorAll('.conversation-item');
        conversationItems.forEach(item => {
            const username = item.querySelector('.username').textContent.toLowerCase();
            if (username.includes(query.toLowerCase())) {
                item.style.display = 'block';
            } else {
                item.style.display = 'none';
            }
        });
    }

    showUserList() {
        // This could open a modal with all users to start new conversations
        // For now, just scroll to the conversations list
        const conversationsList = document.getElementById('conversations-list');
        if (conversationsList) {
            conversationsList.scrollIntoView({ behavior: 'smooth' });
        }
    }

    showConversationsList() {
        // Mobile: Show sidebar and hide chat main
        if (window.innerWidth < 768) {
            const chatSidebar = document.querySelector('.chat-sidebar');
            const chatMain = document.querySelector('.chat-main');
            const backBtn = document.getElementById('back-to-conversations');

            if (chatSidebar && chatMain) {
                chatSidebar.style.display = 'flex';
                chatMain.classList.remove('active');
                chatMain.style.display = 'none';
                if (backBtn) backBtn.style.display = 'none';
            }
        }

        // Clear current conversation
        this.currentConversation = null;
        if (window.wsClient) {
            window.wsClient.setCurrentConversation(null);
        }
    }

    scrollToBottom() {
        const messagesContainer = document.getElementById('messages-container');
        if (messagesContainer) {
            messagesContainer.scrollTop = messagesContainer.scrollHeight;
        }
    }

    formatRelativeTime(dateString) {
        const date = new Date(dateString);
        const now = new Date();
        const seconds = Math.floor((now - date) / 1000);
        
        if (seconds < 60) return 'now';
        if (seconds < 3600) return `${Math.floor(seconds / 60)}m`;
        if (seconds < 86400) return `${Math.floor(seconds / 3600)}h`;
        if (seconds < 604800) return `${Math.floor(seconds / 86400)}d`;
        
        return date.toLocaleDateString();
    }

    formatMessageTime(dateString) {
        const date = new Date(dateString);
        return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
    }

    escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }
}

// Create global chat UI instance
window.chatUI = new ChatUI();
