# WhatsApp Multi-Session PHP Integration

Updated WhatsAppService.php to work with Go WhatsApp Multi-Session API with full typing indicator support.

## 🚀 Features

- ✅ **Typing Indicators** - Send and stop typing indicators
- ✅ **Token Authentication** - JWT Bearer token support
- ✅ **Configuration-based** - Use Laravel config or direct parameters
- ✅ **Session Management** - Multiple session support
- ✅ **Auto Authentication** - Built-in login functionality
- ✅ **Full API Coverage** - Messages, files, groups, presence

## ⚙️ Configuration

### 1. Update `config/services.php`:

```php
'whatsapp' => [
    'api_endpoint' => env('WHATSAPP_API_ENDPOINT', 'http://localhost:8080'),
    'token' => env('WHATSAPP_TOKEN', null),
    'session_id' => env('WHATSAPP_SESSION_ID', '9760640454'),
    'username' => env('WHATSAPP_USERNAME', 'admin'),
    'password' => env('WHATSAPP_PASSWORD', 'admin123'),
],
```

### 2. Update `.env`:

```env
WHATSAPP_API_ENDPOINT=http://localhost:8080
WHATSAPP_TOKEN=your-jwt-token-here
WHATSAPP_SESSION_ID=9760640454
WHATSAPP_USERNAME=admin
WHATSAPP_PASSWORD=admin123
```

## 📱 Usage Examples

### Method 1: Configuration-based (Recommended)

```php
use App\Services\WhatsAppService;

// Uses values from config/services.php
WhatsAppService::sendTyping('6281381393739');
WhatsAppService::sendMessage('6281381393739', 'Hello!');
WhatsAppService::stopTyping('6281381393739');
```

### Method 2: Parameter-based (Multiple sessions)

```php
use App\Services\WhatsAppService;

$sessionId = '9760640454';
$token = 'your-jwt-token';

WhatsAppService::sendTyping('6281381393739', $sessionId, $token);
WhatsAppService::sendMessage('6281381393739', 'Hello!', $sessionId, $token);
WhatsAppService::stopTyping('6281381393739', $sessionId, $token);
```

### Authentication

```php
// Get new token
$auth = WhatsAppService::authenticate();
if ($auth['success']) {
    $token = $auth['token'];
    // Use token for subsequent calls
}

// Or use config credentials
$auth = WhatsAppService::authenticate('admin', 'password123');
```

## 🔥 Typing Indicator Workflow

```php
use App\Services\WhatsAppService;

// Complete typing workflow
WhatsAppService::setOnline();                           // Set session online
WhatsAppService::sendTyping('6281381393739');          // Start typing
sleep(3);                                              // Wait 3 seconds
WhatsAppService::stopTyping('6281381393739');          // Stop typing  
WhatsAppService::sendMessage('6281381393739', 'Hello!'); // Send message
```

## 📋 Available Methods

| Method | Description | Config Support |
|--------|-------------|----------------|
| `authenticate()` | Get JWT token | ✅ |
| `sendMessage()` | Send text message | ✅ |
| `sendTyping()` | Start typing indicator | ✅ |
| `stopTyping()` | Stop typing indicator | ✅ |
| `setOnline()` | Set session online | ✅ |
| `sendGroupMessage()` | Send group message | ✅ |
| `sendFile()` | Send file attachment | ✅ |
| `listGroups()` | Get WhatsApp groups | ✅ |
| `getUserInfo()` | Check phone number | ✅ |

## 🎯 Key Features

### Flexible Usage
- **Config-based**: Set once, use everywhere
- **Parameter-based**: Perfect for multiple sessions/tokens

### Enhanced Typing
- **Proper phone formatting**: `6281381393739` → `6281381393739@s.whatsapp.net`
- **Online presence**: Automatically sets session online
- **Error handling**: Comprehensive error responses

### Response Format
```php
[
    'code' => 'SUCCESS|ERROR',
    'message' => 'Human readable message',
    'results' => [...] // API response data
]
```

## 🔧 Files

- `WhatsAppService.php` - Main service class
- `config-example.php` - Laravel configuration example  
- `WhatsAppServiceExample.php` - Usage examples
- `PHP-README.md` - This documentation

## ⚡ Quick Start

1. Copy `WhatsAppService.php` to your Laravel project
2. Update `config/services.php` with WhatsApp settings
3. Use the service in your controllers:

```php
// In your controller
use App\Services\WhatsAppService;

public function sendWithTyping($phoneNumber, $message) 
{
    WhatsAppService::sendTyping($phoneNumber);
    sleep(2);
    WhatsAppService::stopTyping($phoneNumber);
    return WhatsAppService::sendMessage($phoneNumber, $message);
}
```

🎉 **Ready to use!** Your PHP application now supports WhatsApp typing indicators!