<?php

namespace App\Services;

use App\Events\WhatsAppApiCall;
use Illuminate\Http\Client\Response;
use Illuminate\Support\Facades\Config;
use Illuminate\Support\Facades\Http;

class WhatsAppService
{
    private const MAX_MESSAGE_LENGTH = 65536;

    private static function getHttpClient(?string $token = null)
    {
        // Use provided token or get from config
        $authToken = $token ?: Config::get('services.whatsapp.token');
        
        return Http::withToken($authToken)
            ->baseUrl(Config::get('services.whatsapp.api_endpoint'));
    }

    /**
     * Authenticate and get a new token
     * 
     * @param string|null $username
     * @param string|null $password
     * @return array
     */
    public static function authenticate(?string $username = null, ?string $password = null): array
    {
        $username = $username ?: Config::get('services.whatsapp.username', 'admin');
        $password = $password ?: Config::get('services.whatsapp.password', 'admin123');

        try {
            $response = Http::baseUrl(Config::get('services.whatsapp.api_endpoint'))
                ->post('/api/auth/login', [
                    'username' => $username,
                    'password' => $password
                ]);

            if ($response->successful()) {
                $data = $response->json();
                if (isset($data['success']) && $data['success'] === true) {
                    return [
                        'success' => true,
                        'token' => $data['data']['token'],
                        'expires_at' => $data['data']['expires_at'] ?? null,
                        'user' => $data['data']['user'] ?? null
                    ];
                }
            }

            return [
                'success' => false,
                'token' => null,
                'message' => $response->json()['message'] ?? 'Authentication failed'
            ];
        } catch (\Exception $e) {
            return [
                'success' => false,
                'token' => null,
                'message' => 'Authentication error: ' . $e->getMessage()
            ];
        }
    }

    /**
     * Get session ID from config or use default
     */
    private static function getSessionId(?string $sessionId = null): string
    {
        return $sessionId ?: Config::get('services.whatsapp.session_id');
    }

    public static function getUserInfo(string $phoneNumber, ?string $sessionId = null, ?string $token = null): ?array
    {
        $response = self::getRequest("/api/sessions/$sessionId/check-number", $token, [
            'to' => $phoneNumber
        ]);

        if (isset($response['success']) && $response['success'] === true) {
            return $response['data'];
        }

        return null;
    }

    public static function sendMessage(string $to, string $message, ?string $sessionId = null, ?string $token = null): array
    {
        if (empty($to)) {
            return [
                'code' => 'ERROR',
                'message' => 'Recipient phone number is required',
                'results' => null,
            ];
        }

        if (empty($message)) {
            return [
                'code' => 'ERROR',
                'message' => 'Message content is required',
                'results' => null,
            ];
        }

        $sessionId = self::getSessionId($sessionId);
        $messageChunks = self::splitMessage($message);
        $lastResponse = null;

        foreach ($messageChunks as $chunk) {
            $response = self::sendChunkedMessage($to, $chunk, $sessionId, $token);
            $success = $response->successful();

            event(new WhatsAppApiCall('send', $to, $chunk, $response, true));

            if (! $success) {
                return [
                    'code' => 'ERROR',
                    'message' => 'Failed to send message',
                    'results' => null,
                ];
            }

            $lastResponse = $response;
        }

        $responseBody = $lastResponse->json();

        return [
            'code' => $responseBody['success'] ? 'SUCCESS' : 'ERROR',
            'message' => $responseBody['message'] ?? 'Message sent successfully',
            'results' => $responseBody['data'] ?? null,
        ];
    }

    public static function sendGroupMessage(string $to, string $message, ?string $sessionId = null, ?string $token = null): array
    {
        $sessionId = self::getSessionId($sessionId);
        $messageChunks = self::splitMessage($message);
        $lastResponse = null;

        foreach ($messageChunks as $chunk) {
            $response = self::sendChunkedMessage($to, $chunk, $sessionId, $token);
            $success = $response->successful();

            event(new WhatsAppApiCall('send', $to, $chunk, $response, $success));

            if (! $success) {
                return [
                    'code' => 'ERROR',
                    'message' => 'Failed to send group message',
                    'results' => null,
                ];
            }

            $lastResponse = $response;
        }

        $responseBody = $lastResponse->json();

        return [
            'code' => $responseBody['success'] ? 'SUCCESS' : 'ERROR',
            'message' => $responseBody['message'] ?? 'Group message sent successfully',
            'results' => $responseBody['data'] ?? null,
        ];
    }

    public static function listGroups(?string $sessionId = null, ?string $token = null): array
    {
        $sessionId = self::getSessionId($sessionId);
        return self::getRequest("/api/sessions/$sessionId/groups", $token);
    }

    public static function sendFile(string $to, string $caption, string $filePath, ?string $sessionId = null, ?string $token = null): bool
    {
        $sessionId = self::getSessionId($sessionId);
        $response = self::postRequest("/api/sessions/$sessionId/send-attachment", [
            'to' => $to,
            'caption' => $caption,
            'file' => base64_encode(file_get_contents($filePath)),
            'filename' => basename($filePath)
        ], $token);

        return $response->successful();
    }

    public static function sendTyping(string $to, ?string $sessionId = null, ?string $token = null): array
    {
        $sessionId = self::getSessionId($sessionId);
        $response = self::postRequest("/api/sessions/$sessionId/typing", [
            'to' => $to
        ], $token);

        $responseBody = $response->json();

        return [
            'code' => $response->successful() ? 'SUCCESS' : 'ERROR',
            'message' => $responseBody['message'] ?? ($response->successful() ? 'Typing indicator sent' : 'Failed to send typing indicator'),
            'results' => $responseBody
        ];
    }

    public static function stopTyping(string $to, ?string $sessionId = null, ?string $token = null): array
    {
        $sessionId = self::getSessionId($sessionId);
        $response = self::postRequest("/api/sessions/$sessionId/stop-typing", [
            'to' => $to
        ], $token);

        $responseBody = $response->json();

        return [
            'code' => $response->successful() ? 'SUCCESS' : 'ERROR',
            'message' => $responseBody['message'] ?? ($response->successful() ? 'Typing indicator stopped' : 'Failed to stop typing indicator'),
            'results' => $responseBody
        ];
    }

    public static function setOnline(?string $sessionId = null, ?string $token = null): array
    {
        $sessionId = self::getSessionId($sessionId);
        $response = self::postRequest("/api/sessions/$sessionId/set-online", [], $token);

        $responseBody = $response->json();

        return [
            'code' => $response->successful() ? 'SUCCESS' : 'ERROR',
            'message' => $responseBody['message'] ?? ($response->successful() ? 'Session set to online' : 'Failed to set session online'),
            'results' => $responseBody
        ];
    }

    private static function splitMessage(string $message): array
    {
        return strlen($message) > self::MAX_MESSAGE_LENGTH
            ? str_split($message, self::MAX_MESSAGE_LENGTH)
            : [$message];
    }

    private static function sendChunkedMessage(string $to, string $message, ?string $sessionId, ?string $token): Response
    {
        // :TODO filter bad world
        return self::postRequest("/api/sessions/$sessionId/send", [
            'to' => $to,
            'message' => $message,
        ], $token);
    }

    private static function getRequest(string $endpoint, ?string $token, array $data = []): array
    {
        try {
            $client = self::getHttpClient($token);
            $response = empty($data) ? $client->get($endpoint) : $client->post($endpoint, $data);
            
            if ($response->successful()) {
                return $response->json();
            }

            return [];
        } catch (\Exception $e) {
            return [];
        }
    }

    private static function postRequest(string $endpoint, array $data, ?string $token, bool $isMultipart = false): Response
    {
        try {
            $client = self::getHttpClient($token);
            $client = $isMultipart ? $client->asMultipart() : $client;

            return $client->post($endpoint, $data);
        } catch (\Exception $e) {
            throw $e;
        }
    }

    private static function putRequest(string $endpoint, array $data, ?string $token): Response
    {
        try {
            return self::getHttpClient($token)->put($endpoint, $data);
        } catch (\Exception $e) {
            throw $e;
        }
    }
}
