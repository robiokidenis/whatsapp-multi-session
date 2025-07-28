package handlers

import (
	"encoding/json"
	"net/http"
	
	"whatsapp-multi-session/internal/models"
)

// HandleError writes appropriate error response based on error type
func HandleError(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")
	
	var response models.StandardResponse
	
	// Check error type and set appropriate status code and error code
	switch err.(type) {
	case models.NotFoundError:
		w.WriteHeader(http.StatusNotFound)
		response = models.ErrorResponse(err.Error(), models.ErrCodeNotFound)
	case models.UnauthorizedError:
		w.WriteHeader(http.StatusUnauthorized)
		response = models.ErrorResponse(err.Error(), models.ErrCodeUnauthorized)
	case models.BadRequestError:
		w.WriteHeader(http.StatusBadRequest)
		response = models.ErrorResponse(err.Error(), models.ErrCodeBadRequest)
	case models.ServiceUnavailableError:
		w.WriteHeader(http.StatusServiceUnavailable)
		response = models.ErrorResponse(err.Error(), models.ErrCodeServiceUnavailable)
	default:
		// For any other errors, return 500
		w.WriteHeader(http.StatusInternalServerError)
		response = models.ErrorResponse("Internal server error", models.ErrCodeInternalServer)
	}
	
	json.NewEncoder(w).Encode(response)
}

// HandleErrorWithMessage writes error response with custom message and code
func HandleErrorWithMessage(w http.ResponseWriter, statusCode int, message string, code string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	response := models.ErrorResponse(message, code)
	json.NewEncoder(w).Encode(response)
}

// WriteSuccessResponse writes a standard success response
func WriteSuccessResponse(w http.ResponseWriter, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	
	response := models.SuccessResponse(message, data)
	json.NewEncoder(w).Encode(response)
}