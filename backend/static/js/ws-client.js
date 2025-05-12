class MessageWebSocket {
    constructor(userId) {
        this.userId = userId;
        this.connection = null;
        this.reconnectInterval = 5000;
        this.connect();
    }

    connect() {
        this.connection = new WebSocket(`ws://${window.location.host}/ws/messages`);

        this.connection.onopen = () => {
            console.log('WebSocket connected');
        };

        this.connection.onmessage = (event) => {
            const message = JSON.parse(event.data);
            this.handleMessage(message);
        };

        this.connection.onclose = () => {
            console.log('WebSocket disconnected, attempting to reconnect');
            setTimeout(() => this.connect(), this.reconnectInterval);
        };
    }

    handleMessage(message) {
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