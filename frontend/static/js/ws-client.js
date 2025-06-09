class MessageWebSocket {
    constructor(userId) {
        this.userId = userId;
        this.connection = null;
        this.reconnectInterval = 5000;
        this.connect();
        console.log('MessageWebSocket constructed with userId:', userId);
    }

    connect() {
        this.connection = new WebSocket(`ws://${window.location.host}/ws/messages`);

        this.connection.onopen = () => {
            console.log('WebSocket connected');
        };

        this.connection.onmessage = (event) => {
            console.log('WebSocket message received:', event.data);
            try {
                const message = JSON.parse(event.data);
                this.handleMessage(message);
            } catch (e) {
                console.error('Failed to parse WebSocket message:', e);
            }
        };

        this.connection.onerror = (error) => {
            console.error('WebSocket error:', error);
        };

        this.connection.onclose = () => {
            console.log('WebSocket disconnected, attempting to reconnect');
            setTimeout(() => this.connect(), this.reconnectInterval);
        };
    }

    handleMessage(message) {
        console.log('Handling message:', message);
        switch (message.type) {
            case 'new_message':
                this.handleNewMessage(message);
                break;
            case 'messages_history':
                this.handleMessageHistory(message);
                break;
            case 'notification':
                this.handleNotification(message);
                break;
            case 'comment_added':
                console.log('Handling comment_added message:', message);
                this.handleCommentAdded(message);
                break;
            case 'comment_updated':
                this.handleCommentUpdated(message);
                break;
            case 'comment_deleted':
                this.handleCommentDeleted(message);
                break;
            default:
                console.log('Unknown message type:', message.type);
        }
    }

    handleNewMessage(message) {
        const messageElement = document.createElement('div');
        messageElement.className = 'message';
        messageElement.innerHTML = `
            <div class="message-sender">${message.sender}</div>
            <div class="message-content">${message.content}</div>
            <div class="message-timestamp">${new Date(message.timestamp).toLocaleString()}</div>
        `;

        const messageContainer = document.getElementById('message-container');
        if (messageContainer) {
            messageContainer.appendChild(messageElement);
            messageContainer.scrollTop = messageContainer.scrollHeight;
        }

        this.sendReadReceipt(message.id);
    }

    handleMessageHistory(message) {
        const messageContainer = document.getElementById('message-container');
        if (messageContainer) {
            messageContainer.innerHTML = '';
            
            message.history.forEach(msg => {
                const messageElement = document.createElement('div');
                messageElement.className = 'message';
                messageElement.innerHTML = `
                    <div class="message-sender">${msg.sender}</div>
                    <div class="message-content">${msg.content}</div>
                    <div class="message-timestamp">${new Date(msg.timestamp).toLocaleString()}</div>
                `;
                messageContainer.appendChild(messageElement);
            });
            
            messageContainer.scrollTop = messageContainer.scrollHeight;
        }
    }

    handleNotification(message) {
        const notification = document.createElement('div');
        notification.className = 'notification';
        notification.textContent = message.content;
        
       
        document.body.appendChild(notification);

        setTimeout(() => {
            notification.remove();
        }, 5000);
    }

    handleCommentAdded(message) {
        if (!message.data || !message.post_id) {
            console.log('Invalid comment_added message:', message);
            return;
        }

        const commentsContainer = document.querySelector(`[data-post-id="${message.post_id}"] .comments-container`);
        if (!commentsContainer) {
            console.log('Comments container not found for post:', message.post_id);
            return;
        }

        console.log('Adding new comment to UI:', message.data);

        const currentUserId = JSON.parse(atob(localStorage.getItem('token').split('.')[1])).user_id;
        const commentElement = document.createElement('div');
        commentElement.className = 'bg-gray-50 p-3 rounded-lg';
        commentElement.dataset.commentId = message.data.id;
        commentElement.innerHTML = `
            <div class="flex items-center justify-between mb-1">
                <div class="flex items-center gap-2">
                    <span class="font-medium text-sm">@${message.data.user.nickname}</span>
                    <span class="text-xs text-gray-500">${new Date(message.data.createdAt).toLocaleString()}</span>
                </div>
                ${message.data.user.id === currentUserId ? `
                    <div class="flex items-center gap-2">
                        <button class="edit-comment-btn text-xs text-blue-600 hover:text-blue-800">Edit</button>
                        <button class="delete-comment-btn text-xs text-red-600 hover:text-red-800">Delete</button>
                    </div>
                ` : ''}
            </div>
            <div class="comment-content">
                <p class="text-sm text-gray-700">${message.data.content}</p>
            </div>
            ${message.data.user.id === currentUserId ? `
                <form class="edit-comment-form hidden mt-2">
                    <div class="flex gap-2">
                        <input type="text" class="flex-grow p-2 border rounded-lg text-sm" value="${message.data.content}">
                        <button type="submit" class="bg-blue-500 text-white px-3 py-1 rounded-lg text-sm hover:bg-blue-600">Save</button>
                        <button type="button" class="cancel-edit-btn bg-gray-300 text-gray-700 px-3 py-1 rounded-lg text-sm hover:bg-gray-400">Cancel</button>
                    </div>
                </form>
            ` : ''}
        `;
        commentsContainer.insertBefore(commentElement, commentsContainer.firstChild);

        // Re-attach event handlers for the new comment's buttons
        if (typeof setupCommentActions === 'function') {
            setupCommentActions();
        }
    }

    handleCommentUpdated(message) {
        if (!message.data || !message.data.id) return;

        const commentElement = document.querySelector(`[data-comment-id="${message.data.id}"]`);
        if (!commentElement) return;

        commentElement.innerHTML = `
            <div class="flex items-center gap-2 mb-1">
                <span class="font-medium text-sm">@${message.data.user.nickname}</span>
                <span class="text-xs text-gray-500">${new Date(message.data.createdAt).toLocaleString()}</span>
            </div>
            <p class="text-sm text-gray-700">${message.data.content}</p>
        `;
    }

    handleCommentDeleted(message) {
        if (!message.data) return;

        const commentElement = document.querySelector(`[data-comment-id="${message.data}"]`);
        if (commentElement) {
            commentElement.remove();
        }
    }

    sendMessage(content) {
        if (this.connection && this.connection.readyState === WebSocket.OPEN) {
            this.connection.send(JSON.stringify({
                type: 'new_message',
                userId: this.userId,
                content
            }));
        }
    }

    sendReadReceipt(messageId) {
        if (this.connection && this.connection.readyState === WebSocket.OPEN) {
            this.connection.send(JSON.stringify({
                type: 'read_receipt',
                userId: this.userId,
                messageId
            }));
        }
    }
}