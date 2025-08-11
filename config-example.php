<?php

/**
 * WhatsApp Multi-Session Configuration Example
 * 
 * Add this to your config/services.php file in Laravel
 */

return [
    // ... your other services ...

    'whatsapp' => [
        // Required: Your Go WhatsApp API endpoint
        'api_endpoint' => env('WHATSAPP_API_ENDPOINT', 'http://localhost:8080'),
        
        // Optional: Pre-configured JWT token (get from /api/auth/login)
        'token' => env('WHATSAPP_TOKEN', null),
        
        // Optional: Default session ID to use
        'session_id' => env('WHATSAPP_SESSION_ID', '9760640454'),
        
        // Optional: Admin credentials for authentication
        'username' => env('WHATSAPP_USERNAME', 'admin'),
        'password' => env('WHATSAPP_PASSWORD', 'admin123'),
    ],
];

/**
 * Add these to your .env file:
 * 
 * WHATSAPP_API_ENDPOINT=http://localhost:8080
 * WHATSAPP_TOKEN=your-jwt-token-here
 * WHATSAPP_SESSION_ID=9760640454
 * WHATSAPP_USERNAME=admin
 * WHATSAPP_PASSWORD=admin123
 */