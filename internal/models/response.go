package models

// StandardResponse represents the standard API response structure
type StandardResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Code    string      `json:"code,omitempty"` // Error code for client handling
}

// SuccessResponse creates a success response
func SuccessResponse(message string, data interface{}) StandardResponse {
	return StandardResponse{
		Success: true,
		Message: message,
		Data:    data,
	}
}

// ErrorResponse creates an error response
func ErrorResponse(message string, code string) StandardResponse {
	return StandardResponse{
		Success: false,
		Error:   message,
		Code:    code,
	}
}

// Error codes for standardized error handling
const (
	ErrCodeNotFound            = "NOT_FOUND"
	ErrCodeUnauthorized        = "UNAUTHORIZED"
	ErrCodeBadRequest          = "BAD_REQUEST"
	ErrCodeServiceUnavailable  = "SERVICE_UNAVAILABLE"
	ErrCodeInternalServer      = "INTERNAL_SERVER_ERROR"
	ErrCodeSessionNotConnected = "SESSION_NOT_CONNECTED"
	ErrCodeSessionNotLoggedIn  = "SESSION_NOT_LOGGED_IN"
	ErrCodeInvalidInput        = "INVALID_INPUT"
	ErrCodeAlreadyExists       = "ALREADY_EXISTS"
	ErrCodeRateLimited         = "RATE_LIMITED"
	ErrCodeForbidden           = "FORBIDDEN"
)