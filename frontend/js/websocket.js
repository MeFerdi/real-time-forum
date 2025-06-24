// WebSocket client for real-time communication
class WebSocketClient {
    constructor() {
        this.ws = null;
        this.isConnected = false;
        this.reconnectAttempts = 0;
        this.maxReconnectAttempts = 5;
        this.reconnectDelay = 1000; // Start with 1 second
        this.messageHandlers = new Map();
        this.onlineUsers = new Set();
        this.currentConversation = null;
        
        // Bind methods to preserve 'this' context
        this.connect = this.connect.bind(this);
        this.disconnect = this.disconnect.bind(this);
        this.sendMessage = this.sendMessage.bind(this);
        this.handleMessage = this.handleMessage.bind(this);
        this.handleOpen = this.handleOpen.bind(this);
        this.handleClose = this.handleClose.bind(this);
        this.handleError = this.handleError.bind(this);
        
        // Register default message handlers
        this.registerHandler('private_message', this.handlePrivateMessage.bind(this));
        this.registerHandler('user_status', this.handleUserStatus.bind(this));
        this.registerHandler('online_users', this.handleOnlineUsers.bind(this));
        this.registerHandler('typing', this.handleTyping.bind(this));
        this.registerHandler('error', this.handleServerError.bind(this));
    }

    connect() {
        if (this.ws && (this.ws.readyState === WebSocket.CONNECTING || this.ws.readyState === WebSocket.OPEN)) {
            return;
        }

        try {
            const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
            const wsUrl = `${protocol}//${window.location.host}/ws`;
            
            this.ws = new WebSocket(wsUrl);
            this.ws.onopen = this.handleOpen;
            this.ws.onmessage = this.handleMessage;
            this.ws.onclose = this.handleClose;
            this.ws.onerror = this.handleError;
            
            console.log('Attempting to connect to WebSocket...');
        } catch (error) {
            console.error('Failed to create WebSocket connection:', error);
            this.scheduleReconnect();
        }
    }

    disconnect() {
        if (this.ws) {
            this.ws.close();
            this.ws = null;
        }
        this.isConnected = false;
        this.onlineUsers.clear();
    }

    handleOpen(event) {
        console.log('WebSocket connected');
        this.isConnected = true;
        this.reconnectAttempts = 0;
        this.reconnectDelay = 1000;
        
        // Notify UI about connection status
        this.notifyConnectionStatus(true);
    }

    handleMessage(event) {
        try {
            const message = JSON.parse(event.data);
            console.log('Received WebSocket message:', message);
            
            const handler = this.messageHandlers.get(message.type);
            if (handler) {
                handler(message.data, message.timestamp);
            } else {
                console.warn('No handler for message type:', message.type);
            }
        } catch (error) {
            console.error('Error parsing WebSocket message:', error);
        }
    }

    handleClose(event) {
        console.log('WebSocket disconnected:', event.code, event.reason);
        this.isConnected = false;
        this.onlineUsers.clear();
        
        // Notify UI about connection status
        this.notifyConnectionStatus(false);
        
        // Attempt to reconnect if not a normal closure
        if (event.code !== 1000 && this.reconnectAttempts < this.maxReconnectAttempts) {
            this.scheduleReconnect();
        }
    }

    handleError(error) {
        console.error('WebSocket error:', error);
    }

    scheduleReconnect() {
        if (this.reconnectAttempts >= this.maxReconnectAttempts) {
            console.error('Max reconnection attempts reached');
            return;
        }

        this.reconnectAttempts++;
        const delay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1); // Exponential backoff
        
        console.log(`Reconnecting in ${delay}ms (attempt ${this.reconnectAttempts}/${this.maxReconnectAttempts})`);
        
        setTimeout(() => {
            this.connect();
        }, delay);
    }

    sendMessage(type, data) {
        if (!this.isConnected || !this.ws) {
            console.error('WebSocket not connected');
            return false;
        }

        const message = {
            type: type,
            data: data,
            timestamp: new Date().toISOString()
        };

        try {
            this.ws.send(JSON.stringify(message));
            return true;
        } catch (error) {
            console.error('Error sending WebSocket message:', error);
            return false;
        }
    }

    sendPrivateMessage(receiverID, content) {
        return this.sendMessage('private_message', {
            receiver_id: receiverID,
            content: content
        });
    }

    sendTypingIndicator(receiverID) {
        return this.sendMessage('typing', {
            receiver_id: receiverID
        });
    }

    registerHandler(messageType, handler) {
        this.messageHandlers.set(messageType, handler);
    }

    // Default message handlers
    handlePrivateMessage(data, timestamp) {
        console.log('Received private message:', data);

        // Always update UI with new message
        if (window.chatUI) {
            window.chatUI.addMessage(data);
        }

        // Show notification if not in current conversation and message is from another user
        const currentUserId = window.views && window.views.currentUser ? window.views.currentUser.id : 0;
        const isFromOtherUser = data.sender_id !== currentUserId;
        const isCurrentConversation = this.currentConversation === data.sender_id;

        if (isFromOtherUser && !isCurrentConversation) {
            this.showMessageNotification(data);
        }
    }

    handleUserStatus(data, timestamp) {
        console.log('User status update:', data);
        
        if (data.status === 'online') {
            this.onlineUsers.add(data.user_id);
        } else {
            this.onlineUsers.delete(data.user_id);
        }
        
        // Update UI
        if (window.chatUI) {
            window.chatUI.updateUserStatus(data.user_id, data.status);
        }
    }

    handleOnlineUsers(data, timestamp) {
        console.log('Online users:', data);
        
        this.onlineUsers.clear();
        data.forEach(user => {
            this.onlineUsers.add(user.user_id);
        });
        
        // Update UI
        if (window.chatUI) {
            window.chatUI.updateOnlineUsers(data);
        }
    }

    handleTyping(data, timestamp) {
        console.log('Typing indicator:', data);
        
        // Update UI with typing indicator
        if (window.chatUI) {
            window.chatUI.showTypingIndicator(data.sender_id, data.username);
        }
    }

    handleServerError(data, timestamp) {
        console.error('Server error:', data);
        
        // Show error to user
        if (data.error) {
            this.showErrorNotification(data.error);
        }
    }

    // Utility methods
    notifyConnectionStatus(connected) {
        // Update connection status indicator in UI
        const statusIndicator = document.getElementById('connection-status');
        if (statusIndicator) {
            statusIndicator.className = connected ? 'connected' : 'disconnected';
            statusIndicator.textContent = connected ? 'Connected' : 'Disconnected';
        }
        
        // Dispatch custom event for other components
        window.dispatchEvent(new CustomEvent('websocket-status', {
            detail: { connected }
        }));
    }

    showMessageNotification(messageData) {
        // Create a simple notification
        if ('Notification' in window && Notification.permission === 'granted') {
            new Notification(`New message from ${messageData.sender.username}`, {
                body: messageData.content.substring(0, 100),
                icon: '/favicon.ico'
            });
        }
    }

    showErrorNotification(error) {
        // Show error in UI
        console.error('WebSocket error:', error);
        
        // You can implement a toast notification system here
        alert('Error: ' + error);
    }

    isUserOnline(userID) {
        return this.onlineUsers.has(userID);
    }

    setCurrentConversation(userID) {
        this.currentConversation = userID;
    }

    getCurrentConversation() {
        return this.currentConversation;
    }
}

// Create global WebSocket client instance
window.wsClient = new WebSocketClient();

// Request notification permission on first user interaction
function requestNotificationPermission() {
    if ('Notification' in window && Notification.permission === 'default') {
        Notification.requestPermission();
    }
}

// Add event listener for first user interaction
document.addEventListener('click', requestNotificationPermission, { once: true });
