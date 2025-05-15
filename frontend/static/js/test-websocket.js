import WebSocket from 'ws';

const ws = new WebSocket('ws://localhost:8080/ws/messages');

ws.on('open', () => {
    console.log('Connected to WebSocket server');

    // Test new_message
    ws.send(JSON.stringify({
        type: 'new_message',
        userId: '123',
        sender: 'TestUser',
        content: 'Hello, world!',
        timestamp: new Date().toISOString(),
        id: 'msg1'
    }));

    // Test messages_history after 1 second
    setTimeout(() => {
        ws.send(JSON.stringify({
            type: 'messages_history',
            history: [
                { sender: 'User1', content: 'Hi!', timestamp: new Date().toISOString() },
                { sender: 'User2', content: 'Hey!', timestamp: new Date().toISOString() }
            ]
        }));
    }, 1000);

    // Test notification after 2 seconds
    setTimeout(() => {
        ws.send(JSON.stringify({
            type: 'notification',
            content: 'You have a new friend request!'
        }));
    }, 2000);
});

ws.on('message', (data) => {
    console.log('Received:', data.toString());
});

ws.on('close', () => {
    console.log('Disconnected from WebSocket server');
});

ws.on('error', (error) => {
    console.error('WebSocket error:', error);
});