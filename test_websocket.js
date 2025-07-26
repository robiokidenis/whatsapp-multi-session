const WebSocket = require('ws');

// Test WebSocket connection for QR code
const ws = new WebSocket('ws://localhost:8080/api/ws/9049935660');

ws.on('open', function open() {
    console.log('WebSocket connected successfully');
});

ws.on('message', function message(data) {
    const message = JSON.parse(data.toString());
    console.log('Received message:', message);
    
    if (message.type === 'qr') {
        console.log('QR Code received, length:', message.data.qr.length);
        console.log('QR timeout:', message.data.timeout);
    } else if (message.type === 'error') {
        console.log('Error:', message.error);
    }
});

ws.on('error', function error(err) {
    console.log('WebSocket error:', err.message);
});

ws.on('close', function close(code, reason) {
    console.log('WebSocket closed:', code, reason.toString());
});

// Keep alive for 30 seconds
setTimeout(() => {
    console.log('Closing connection...');
    ws.close();
}, 30000);