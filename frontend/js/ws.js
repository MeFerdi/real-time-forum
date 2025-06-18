class WebSocketClient {
    constructor(url) {
        this.url = url;
        this.ws = null;
        this.messageQueue = [];
        this.eventHandlers = new Map();
        this.reconnectAttempts = 0;
        this.maxReconnectAttempts = 5;
        this.heartbeatInterval = null;
        this.isConnecting = false;
    }

    connect() {
        if (this.isConnecting) return;
        this.isConnecting = true;

        try {
            this.ws = new WebSocket(this.url);
            this.setupEventHandlers();
            this.startHeartbeat();
        } catch (error) {
            console.error('WebSocket connection error:', error);
            this.handleReconnection();
        }
    }

    setupEventHandlers() {
        this.ws.onopen = () => {
            console.log('WebSocket connected');
            this.isConnecting = false;
            this.reconnectAttempts = 0;
            this.processMessageQueue();
            this.emit('connected');
        };

        this.ws.onclose = () => {
            this.isConnecting = false;
            this.stopHeartbeat();
            this.handleReconnection();
            this.emit('disconnected');
        };

        this.ws.onerror = (error) => {
            console.error('WebSocket error:', error);
            this.emit('error', error);
        };

        this.ws.onmessage = (event) => {
            try {
                const message = JSON.parse(event.data);
                this.handleMessage(message);
            } catch (error) {
                console.error('Message parsing error:', error);
            }
        };
    }

    handleMessage(message) {
        switch (message.type) {
            case 'pong':
                // Handle heartbeat response
                break;
            case 'chat':
                this.emit('chat', message.data);
                break;
            case 'status':
                this.emit('status', message.data);
                break;
            default:
                this.emit('message', message);
        }
    }

    send(type, data) {
        const message = JSON.stringify({ type, data });
        
        if (this.ws && this.ws.readyState === WebSocket.OPEN) {
            this.ws.send(message);
        } else {
            this.messageQueue.push(message);
        }
    }

    processMessageQueue() {
        while (this.messageQueue.length > 0 && 
               this.ws && 
               this.ws.readyState === WebSocket.OPEN) {
            const message = this.messageQueue.shift();
            this.ws.send(message);
        }
    }

    startHeartbeat() {
        this.heartbeatInterval = setInterval(() => {
            this.send('ping', {});
        }, 30000);
    }

    stopHeartbeat() {
        if (this.heartbeatInterval) {
            clearInterval(this.heartbeatInterval);
            this.heartbeatInterval = null;
        }
    }

    handleReconnection() {
        if (this.reconnectAttempts < this.maxReconnectAttempts) {
            this.reconnectAttempts++;
            const delay = Math.min(1000 * Math.pow(2, this.reconnectAttempts), 30000);
            setTimeout(() => this.connect(), delay);
        } else {
            this.emit('maxReconnectAttemptsReached');
        }
    }

    on(event, handler) {
        if (!this.eventHandlers.has(event)) {
            this.eventHandlers.set(event, new Set());
        }
        this.eventHandlers.get(event).add(handler);
    }

    off(event, handler) {
        if (this.eventHandlers.has(event)) {
            this.eventHandlers.get(event).delete(handler);
        }
    }

    emit(event, data) {
        if (this.eventHandlers.has(event)) {
            this.eventHandlers.get(event).forEach(handler => handler(data));
        }
    }

    disconnect() {
        this.stopHeartbeat();
        if (this.ws) {
            this.ws.close();
            this.ws = null;
        }
    }
}
export default WebSocketClient;