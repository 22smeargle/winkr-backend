package entities

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user entity in the system
type User struct {
	ID             uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Email          string     `json:"email" gorm:"uniqueIndex;not null"`
	PasswordHash   string     `json:"-" gorm:"not null"`
	FirstName      string     `json:"first_name" gorm:"not null"`
	LastName       string     `json:"last_name" gorm:"not null"`
	DateOfBirth    time.Time  `json:"date_of_birth" gorm:"not null"`
	Gender         string     `json:"gender" gorm:"not null;check:gender IN ('male', 'female', 'other')"`
	InterestedIn   []string   `json:"interested_in" gorm:"type:text[];not null"`
	Bio            *string    `json:"bio"`
	LocationLat    *float64   `json:"location_lat"`
	LocationLng    *float64   `json:"location_lng"`
	LocationCity   *string    `json:"location_city"`
	LocationCountry *string    `json:"location_country"`
	IsVerified     bool             `json:"is_verified" gorm:"default:false"`
	VerificationLevel VerificationLevel `json:"verification_level" gorm:"default:0;check:verification_level IN (0, 1, 2)"`
	IsPremium      bool       `json:"is_premium" gorm:"default:false"`
	IsActive       bool       `json:"is_active" gorm:"default:true"`
	IsBanned       bool       `json:"is_banned" gorm:"default:false"`
	LastActive     *time.Time `json:"last_active"`
	CreatedAt      time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName returns the table name for the User entity
func (User) TableName() string {
	return "users"
}

// GetAge returns the user's age based on date of birth
func (u *User) GetAge() int {
	now := time.Now()
	age := now.Year() - u.DateOfBirth.Year()
	if now.YearDay() < u.DateOfBirth.YearDay() {
		age--
	}
	return age
}

// HasLocation returns true if the user has location data
func (u *User) HasLocation() bool {
	return u.LocationLat != nil && u.LocationLng != nil
}

// GetLocation returns the user's location as a tuple
func (u *User) GetLocation() (lat, lng float64, hasLocation bool) {
	if !u.HasLocation() {
		return 0, 0, false
	}
	return *u.LocationLat, *u.LocationLng, true
}

// IsComplete returns true if the user profile is complete
func (u *User) IsComplete() bool {
	return u.FirstName != "" && 
		   u.LastName != "" && 
		   !u.DateOfBirth.IsZero() && 
		   u.Gender != "" && 
		   len(u.InterestedIn) > 0 &&
		   u.HasLocation()
}

// GetVerificationLevel returns user's verification level
func (u *User) GetVerificationLevel() VerificationLevel {
	return u.VerificationLevel
}

// SetVerificationLevel sets user's verification level
func (u *User) SetVerificationLevel(level VerificationLevel) {
	u.VerificationLevel = level
}

// IsVerificationLevelAtLeast returns true if user's verification level is at least the specified level
func (u *User) IsVerificationLevelAtLeast(level VerificationLevel) bool {
	return u.VerificationLevel >= level
}

// IsSelfieVerified returns true if user has selfie verification
func (u *User) IsSelfieVerified() bool {
	return u.VerificationLevel >= VerificationLevelSelfie
}

// IsDocumentVerified returns true if user has document verification
func (u *User) IsDocumentVerified() bool {
	return u.VerificationLevel >= VerificationLevelDocument
}