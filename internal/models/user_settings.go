package models

import (
	"time"
)

// UserSettings represents user preferences and configuration
type UserSettings struct {
	ID                  int        `json:"id"`
	UserID              int        `json:"user_id"`
	Timezone            string     `json:"timezone"`
	DateFormat          string     `json:"date_format"`
	TimeFormat          string     `json:"time_format"`
	Language            string     `json:"language"`
	EmailNotifications  bool       `json:"email_notifications"`
	PushNotifications   bool       `json:"push_notifications"`
	SMSNotifications    bool       `json:"sms_notifications"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           *time.Time `json:"updated_at,omitempty"`
}

// CreateUserSettingsRequest represents user settings creation request
type CreateUserSettingsRequest struct {
	Timezone            string `json:"timezone" validate:"required"`
	DateFormat          string `json:"date_format" validate:"required"`
	TimeFormat          string `json:"time_format" validate:"required"`
	Language            string `json:"language" validate:"required"`
	EmailNotifications  bool   `json:"email_notifications"`
	PushNotifications   bool   `json:"push_notifications"`
	SMSNotifications    bool   `json:"sms_notifications"`
}

// UpdateUserSettingsRequest represents user settings update request
type UpdateUserSettingsRequest struct {
	Timezone            *string `json:"timezone,omitempty"`
	DateFormat          *string `json:"date_format,omitempty"`
	TimeFormat          *string `json:"time_format,omitempty"`
	Language            *string `json:"language,omitempty"`
	EmailNotifications  *bool   `json:"email_notifications,omitempty"`
	PushNotifications   *bool   `json:"push_notifications,omitempty"`
	SMSNotifications    *bool   `json:"sms_notifications,omitempty"`
}

// UserSettingsResponse represents the API response for user settings
type UserSettingsResponse struct {
	Timezone     string                 `json:"timezone"`
	DateFormat   string                 `json:"date_format"`
	TimeFormat   string                 `json:"time_format"`
	Language     string                 `json:"language"`
	Notifications NotificationSettings `json:"notifications"`
}

// NotificationSettings represents notification preferences
type NotificationSettings struct {
	Email bool `json:"email"`
	Push  bool `json:"push"`
	SMS   bool `json:"sms"`
}

// ToResponse converts UserSettings to UserSettingsResponse
func (us *UserSettings) ToResponse() UserSettingsResponse {
	return UserSettingsResponse{
		Timezone:   us.Timezone,
		DateFormat: us.DateFormat,
		TimeFormat: us.TimeFormat,
		Language:   us.Language,
		Notifications: NotificationSettings{
			Email: us.EmailNotifications,
			Push:  us.PushNotifications,
			SMS:   us.SMSNotifications,
		},
	}
}

// GetDefaultSettings returns default user settings
func GetDefaultSettings() UserSettings {
	now := time.Now()
	return UserSettings{
		Timezone:            "UTC",
		DateFormat:          "YYYY-MM-DD",
		TimeFormat:          "24h",
		Language:            "en",
		EmailNotifications:  true,
		PushNotifications:   true,
		SMSNotifications:    false,
		CreatedAt:           now,
	}
}