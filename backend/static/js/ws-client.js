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
        }
    }

    handleNewMessage(message) {
        // Update UI with new message
        // Mark message as read
    }
}