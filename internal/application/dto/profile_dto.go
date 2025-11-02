package dto

import (
	"github.com/go-playground/validator/v10"
)

// ProfileDTOs contain all profile related data transfer objects

// UpdateProfileRequestDTO represents update profile request DTO
type UpdateProfileRequestDTO struct {
	FirstName    *string       `json:"first_name" validate:"omitempty,min=2,max=100"`
	LastName     *string       `json:"last_name" validate:"omitempty,min=2,max=100"`
	Bio          *string       `json:"bio" validate:"omitempty,max=500"`
	InterestedIn []string      `json:"interested_in" validate:"omitempty,min=1,dive,oneof=male female other"`
	Preferences  *PreferencesDTO `json:"preferences"`
}

// UpdateLocationRequestDTO represents update location request DTO
type UpdateLocationRequestDTO struct {
	Latitude  float64  `json:"latitude" validate:"required,min=-90,max=90"`
	Longitude float64  `json:"longitude" validate:"required,min=-180,max=180"`
	City      *string  `json:"city" validate:"omitempty,min=1,max=100"`
	Country   *string  `json:"country" validate:"omitempty,min=1,max=100"`
}

// DeleteAccountRequestDTO represents delete account request DTO
type DeleteAccountRequestDTO struct {
	Password string `json:"password"`
	Reason   string `json:"reason" validate:"omitempty,max=500"`
	Confirm  bool   `json:"confirm" validate:"required"`
}

// PreferencesDTO represents user preferences DTO
type PreferencesDTO struct {
	AgeMin      int `json:"age_min" validate:"omitempty,min=18,max=100"`
	AgeMax      int `json:"age_max" validate:"omitempty,min=18,max=100"`
	MaxDistance int `json:"max_distance" validate:"omitempty,min=1,max=500"`
	ShowMe      bool `json:"show_me"`
}

// ProfileResponseDTO represents profile response DTO
type ProfileResponseDTO struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorDTO   `json:"error,omitempty"`
}

// UserProfileResponseDTO represents user profile response DTO
type UserProfileResponseDTO struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorDTO   `json:"error,omitempty"`
}

// MatchesResponseDTO represents matches response DTO
type MatchesResponseDTO struct {
	Success    bool           `json:"success"`
	Data       interface{}     `json:"data,omitempty"`
	Pagination *PaginationDTO  `json:"pagination,omitempty"`
	Error      *ErrorDTO      `json:"error,omitempty"`
}

// MessageResponseDTO represents a simple message response DTO
type MessageResponseDTO struct {
	Success bool     `json:"success"`
	Message string    `json:"message"`
	Error   *ErrorDTO `json:"error,omitempty"`
}

// PaginationDTO represents pagination information DTO
type PaginationDTO struct {
	Total  int `json:"total"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

// ProfileStatsDTO represents profile statistics DTO
type ProfileStatsDTO struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorDTO   `json:"error,omitempty"`
}

// ProfileCompletionDTO represents profile completion DTO
type ProfileCompletionDTO struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorDTO   `json:"error,omitempty"`
}

// BlockUserRequestDTO represents block user request DTO
type BlockUserRequestDTO struct {
	UserID string `json:"user_id" validate:"required,uuid"`
}

// UnblockUserRequestDTO represents unblock user request DTO
type UnblockUserRequestDTO struct {
	UserID string `json:"user_id" validate:"required,uuid"`
}

// ReportUserRequestDTO represents report user request DTO
type ReportUserRequestDTO struct {
	UserID      string `json:"user_id" validate:"required,uuid"`
	Reason      string `json:"reason" validate:"required,oneof=inappropriate_behavior fake_profile spam harassment other"`
	Description string `json:"description" validate:"omitempty,max=1000"`
}

// BlockedUsersResponseDTO represents blocked users response DTO
type BlockedUsersResponseDTO struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorDTO   `json:"error,omitempty"`
}

// Validate validates DTO using validator
func (dto *UpdateProfileRequestDTO) Validate() error {
	validate := validator.New()
	return validate.Struct(dto)
}

func (dto *UpdateLocationRequestDTO) Validate() error {
	validate := validator.New()
	return validate.Struct(dto)
}

func (dto *DeleteAccountRequestDTO) Validate() error {
	validate := validator.New()
	return validate.Struct(dto)
}

func (dto *PreferencesDTO) Validate() error {
	validate := validator.New()
	return validate.Struct(dto)
}

func (dto *BlockUserRequestDTO) Validate() error {
	validate := validator.New()
	return validate.Struct(dto)
}

func (dto *UnblockUserRequestDTO) Validate() error {
	validate := validator.New()
	return validate.Struct(dto)
}

func (dto *ReportUserRequestDTO) Validate() error {
	validate := validator.New()
	return validate.Struct(dto)
}

// NewProfileResponseDTO creates a new profile response DTO
func NewProfileResponseDTO(data interface{}) *ProfileResponseDTO {
	return &ProfileResponseDTO{
		Success: true,
		Data:    data,
	}
}

// NewUserProfileResponseDTO creates a new user profile response DTO
func NewUserProfileResponseDTO(data interface{}) *UserProfileResponseDTO {
	return &UserProfileResponseDTO{
		Success: true,
		Data:    data,
	}
}

// NewMatchesResponseDTO creates a new matches response DTO
func NewMatchesResponseDTO(data interface{}, total, limit, offset int) *MatchesResponseDTO {
	return &MatchesResponseDTO{
		Success: true,
		Data:    data,
		Pagination: &PaginationDTO{
			Total:  total,
			Limit:  limit,
			Offset: offset,
		},
	}
}

// NewMessageResponseDTO creates a new message response DTO
func NewMessageResponseDTO(message string) *MessageResponseDTO {
	return &MessageResponseDTO{
		Success: true,
		Message: message,
	}
}

// NewProfileStatsResponseDTO creates a new profile stats response DTO
func NewProfileStatsResponseDTO(data interface{}) *ProfileStatsDTO {
	return &ProfileStatsDTO{
		Success: true,
		Data:    data,
	}
}

// NewProfileCompletionResponseDTO creates a new profile completion response DTO
func NewProfileCompletionResponseDTO(data interface{}) *ProfileCompletionDTO {
	return &ProfileCompletionDTO{
		Success: true,
		Data:    data,
	}
}

// NewBlockedUsersResponseDTO creates a new blocked users response DTO
func NewBlockedUsersResponseDTO(data interface{}) *BlockedUsersResponseDTO {
	return &BlockedUsersResponseDTO{
		Success: true,
		Data:    data,
	}
}

// NewProfileErrorResponseDTO creates a new profile error response DTO
func NewProfileErrorResponseDTO(code, message, details string) *ProfileResponseDTO {
	return &ProfileResponseDTO{
		Success: false,
		Error: &ErrorDTO{
			Code:    code,
			Message: message,
			Details: details,
		},
	}
}

// NewMatchesErrorResponseDTO creates a new matches error response DTO
func NewMatchesErrorResponseDTO(code, message, details string) *MatchesResponseDTO {
	return &MatchesResponseDTO{
		Success: false,
		Error: &ErrorDTO{
			Code:    code,
			Message: message,
			Details: details,
		},
	}
}

// NewMessageErrorResponseDTO creates a new message error response DTO
func NewMessageErrorResponseDTO(code, message, details string) *MessageResponseDTO {
	return &MessageResponseDTO{
		Success: false,
		Error: &ErrorDTO{
			Code:    code,
			Message: message,
			Details: details,
		},
	}
}