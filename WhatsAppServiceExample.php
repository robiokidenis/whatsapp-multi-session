<?php

/**
 * Example usage of the updated WhatsAppService with session_id and token
 * 
 * TWO WAYS TO USE:
 * 1. Configuration-based (recommended): Set values in config/services.php
 * 2. Parameter-based: Pass values directly to methods
 */

require_once 'WhatsAppService.php';

use App\Services\WhatsAppService;

// METHOD 1: Configuration-based usage
// Add to config/services.php:
/*
'whatsapp' => [
    'api_endpoint' => env('WHATSAPP_API_ENDPOINT', 'http://localhost:8080'),
    'token' => env('WHATSAPP_TOKEN', 'your-jwt-token'),
    'session_id' => env('WHATSAPP_SESSION_ID', '9760640454'),
    'username' => env('WHATSAPP_USERNAME', 'admin'),
    'password' => env('WHATSAPP_PASSWORD', 'admin123'),
],
*/

// METHOD 2: Parameter-based usage (for multiple sessions/tokens)
$sessionId = '9760640454';
$token = 'your-jwt-token-here';
$testPhoneNumber = '6281381393739';

class WhatsAppExample
{
    private ?string $sessionId;
    private ?string $token;

    public function __construct(?string $sessionId = null, ?string $token = null)
    {
        $this->sessionId = $sessionId;
        $this->token = $token;
    }

    /**
     * Example: Authenticate and get a new token
     */
    public function authenticate(): void
    {
        echo "🔐 Authenticating with WhatsApp API...\n";
        
        $result = WhatsAppService::authenticate();
        
        if ($result['success']) {
            echo "✅ Authentication successful!\n";
            echo "Token: " . substr($result['token'], 0, 20) . "...\n";
            echo "User: " . ($result['user']['username'] ?? 'N/A') . "\n";
            
            // Update token for further operations
            $this->token = $result['token'];
        } else {
            echo "❌ Authentication failed: " . $result['message'] . "\n";
        }
        echo "\n";
    }

    /**
     * Example: Send a regular message
     */
    public function sendMessage(string $to, string $message): void
    {
        echo "📱 Sending message to $to...\n";
        
        // Use either config-based or parameter-based approach
        if ($this->sessionId && $this->token) {
            // Parameter-based
            $result = WhatsAppService::sendMessage($to, $message, $this->sessionId, $this->token);
        } else {
            // Config-based
            $result = WhatsAppService::sendMessage($to, $message);
        }
        
        if ($result['code'] === 'SUCCESS') {
            echo "✅ Message sent successfully!\n";
            echo "Message ID: " . ($result['results']['message_id'] ?? 'N/A') . "\n";
        } else {
            echo "❌ Failed to send message: " . $result['message'] . "\n";
        }
        echo "\n";
    }

    /**
     * Example: Send typing indicator
     */
    public function sendTypingIndicator(string $to): void
    {
        echo "⌨️  Sending typing indicator to $to...\n";
        
        // First set session online for better reliability
        if ($this->sessionId && $this->token) {
            $onlineResult = WhatsAppService::setOnline($this->sessionId, $this->token);
            $typingResult = WhatsAppService::sendTyping($to, $this->sessionId, $this->token);
        } else {
            $onlineResult = WhatsAppService::setOnline();
            $typingResult = WhatsAppService::sendTyping($to);
        }

        if ($onlineResult['code'] === 'SUCCESS') {
            echo "🟢 Session set to online\n";
        }

        if ($typingResult['code'] === 'SUCCESS') {
            echo "✅ Typing indicator sent!\n";
            echo "💡 The recipient should see 'typing...' in WhatsApp\n";
            
            // Wait 5 seconds then stop typing
            echo "⏳ Waiting 5 seconds...\n";
            sleep(5);
            
            if ($this->sessionId && $this->token) {
                $stopResult = WhatsAppService::stopTyping($to, $this->sessionId, $this->token);
            } else {
                $stopResult = WhatsAppService::stopTyping($to);
            }
            
            if ($stopResult['code'] === 'SUCCESS') {
                echo "⏹️  Typing indicator stopped\n";
            }
        } else {
            echo "❌ Failed to send typing indicator: " . $typingResult['message'] . "\n";
        }
        echo "\n";
    }

    /**
     * Example: Check if a phone number exists on WhatsApp
     */
    public function checkPhoneNumber(string $phoneNumber): void
    {
        echo "🔍 Checking if $phoneNumber exists on WhatsApp...\n";
        
        $result = WhatsAppService::getUserInfo($phoneNumber, $this->sessionId, $this->token);
        
        if ($result) {
            echo "✅ Phone number exists on WhatsApp!\n";
            print_r($result);
        } else {
            echo "❌ Phone number not found or not on WhatsApp\n";
        }
        echo "\n";
    }

    /**
     * Example: Send a file
     */
    public function sendFile(string $to, string $filePath, string $caption = ''): void
    {
        echo "📎 Sending file to $to...\n";
        
        if (!file_exists($filePath)) {
            echo "❌ File not found: $filePath\n";
            return;
        }

        $result = WhatsAppService::sendFile($to, $caption, $filePath, $this->sessionId, $this->token);
        
        if ($result) {
            echo "✅ File sent successfully!\n";
        } else {
            echo "❌ Failed to send file\n";
        }
        echo "\n";
    }

    /**
     * Example: Get WhatsApp groups
     */
    public function listGroups(): void
    {
        echo "👥 Fetching WhatsApp groups...\n";
        
        $result = WhatsAppService::listGroups($this->sessionId, $this->token);
        
        if (!empty($result)) {
            echo "✅ Found groups:\n";
            foreach ($result as $group) {
                echo "  - " . ($group['name'] ?? 'Unnamed Group') . "\n";
            }
        } else {
            echo "❌ No groups found or failed to fetch\n";
        }
        echo "\n";
    }

    /**
     * Example: Complete workflow - typing then message
     */
    public function sendMessageWithTyping(string $to, string $message): void
    {
        echo "🚀 Complete workflow: typing indicator + message\n";
        
        // Step 1: Set online
        WhatsAppService::setOnline($this->sessionId, $this->token);
        echo "1. ✅ Set session online\n";
        
        // Step 2: Send typing
        $typingResult = WhatsAppService::sendTyping($to, $this->sessionId, $this->token);
        if ($typingResult['code'] === 'SUCCESS') {
            echo "2. ⌨️  Typing indicator sent\n";
        }
        
        // Step 3: Wait a bit (simulating typing)
        sleep(3);
        
        // Step 4: Stop typing and send message
        WhatsAppService::stopTyping($to, $this->sessionId, $this->token);
        echo "3. ⏹️  Stopped typing\n";
        
        // Step 5: Send actual message
        $messageResult = WhatsAppService::sendMessage($to, $message, $this->sessionId, $this->token);
        if ($messageResult['code'] === 'SUCCESS') {
            echo "4. 📨 Message sent successfully!\n";
        } else {
            echo "4. ❌ Failed to send message: " . $messageResult['message'] . "\n";
        }
        echo "\n";
    }
}

// Example usage
if (php_sapi_name() === 'cli') {
    echo "🤖 WhatsApp Multi-Session PHP Service Example\n";
    echo "============================================\n\n";

    $testPhoneNumber = '6281381393739';  // Replace with actual phone number

    // EXAMPLE 1: Using configuration from config/services.php
    echo "📋 METHOD 1: Configuration-based usage\n";
    echo "====================================\n";
    $whatsappConfig = new WhatsAppExample(); // Uses config values
    
    // Authenticate if needed
    $whatsappConfig->authenticate();
    
    // Send typing indicator (using config)
    echo "Testing typing indicator with config...\n";
    WhatsAppService::sendTyping($testPhoneNumber);
    sleep(3);
    WhatsAppService::stopTyping($testPhoneNumber);
    
    // Send message (using config)
    WhatsAppService::sendMessage($testPhoneNumber, 'Hello from config-based PHP! 🚀');
    
    echo "\n";

    // EXAMPLE 2: Using direct parameters (for multiple sessions)
    echo "📋 METHOD 2: Parameter-based usage\n";
    echo "=================================\n";
    $sessionId = '9760640454';
    $token = 'your-jwt-token-here';
    
    $whatsappParam = new WhatsAppExample($sessionId, $token);

    // Test various functions with parameters
    $whatsappParam->checkPhoneNumber($testPhoneNumber);
    $whatsappParam->sendTypingIndicator($testPhoneNumber);
    $whatsappParam->sendMessage($testPhoneNumber, 'Hello from parameter-based PHP! 🚀');
    $whatsappParam->sendMessageWithTyping($testPhoneNumber, 'Typing indicator message! ⌨️');
    $whatsappParam->listGroups();

    echo "\n📝 USAGE SUMMARY:\n";
    echo "================\n";
    echo "✅ Config-based: WhatsAppService::sendTyping(\$to)\n";
    echo "✅ Parameter-based: WhatsAppService::sendTyping(\$to, \$sessionId, \$token)\n";
    echo "✅ Authentication: WhatsAppService::authenticate()\n";
    echo "✅ All methods support both approaches!\n\n";
    
    echo "✅ All tests completed!\n";
}