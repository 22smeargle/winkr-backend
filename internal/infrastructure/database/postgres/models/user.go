package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User represents a user entity in the database
type User struct {
	ID             uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Email          string     `gorm:"uniqueIndex;not null" json:"email"`
	PasswordHash   string     `gorm:"not null" json:"-"`
	FirstName      string     `gorm:"not null" json:"first_name"`
	LastName       string     `gorm:"not null" json:"last_name"`
	DateOfBirth    time.Time  `gorm:"not null" json:"date_of_birth"`
	Gender         string     `gorm:"not null;check:gender IN ('male', 'female', 'other')" json:"gender"`
	InterestedIn   []string   `gorm:"type:text[];not null" json:"interested_in"`
	Bio            *string    `gorm:"type:text" json:"bio"`
	LocationLat    *float64   `gorm:"type:decimal(10,8)" json:"location_lat"`
	LocationLng    *float64   `gorm:"type:decimal(11,8)" json:"location_lng"`
	LocationCity   *string    `gorm:"size:100" json:"location_city"`
	LocationCountry *string    `gorm:"size:100" json:"location_country"`
	IsVerified     bool       `gorm:"default:false" json:"is_verified"`
	VerificationLevel int       `gorm:"default:0;check:verification_level IN (0, 1, 2);index" json:"verification_level"`
	IsPremium      bool       `gorm:"default:false" json:"is_premium"`
	IsActive       bool       `gorm:"default:true;index" json:"is_active"`
	IsBanned       bool       `gorm:"default:false;index" json:"is_banned"`
	LastActive     *time.Time `gorm:"index" json:"last_active"`
	CreatedAt      time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt      *time.Time `gorm:"index" json:"-"`

	// Relationships
	Photos         []*Photo         `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"photos,omitempty"`
	Preferences    *UserPreferences `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"preferences,omitempty"`
	PrimaryPhoto   *Photo           `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE" json:"primary_photo,omitempty"`
	
	// Swipe relationships
	SwipesMade    []*Swipe `gorm:"foreignKey:SwiperID;constraint:OnDelete:CASCADE" json:"-"`
	SwipesReceived []*Swipe `gorm:"foreignKey:SwipedID;constraint:OnDelete:CASCADE" json:"-"`
	
	// Match relationships
	MatchesAsUser1 []*Match `gorm:"foreignKey:User1ID;constraint:OnDelete:CASCADE" json:"-"`
	MatchesAsUser2 []*Match `gorm:"foreignKey:User2ID;constraint:OnDelete:CASCADE" json:"-"`
	
	// Message relationships
	SentMessages     []*Message `gorm:"foreignKey:SenderID;constraint:OnDelete:CASCADE" json:"-"`
	
	// Report relationships
	ReportsMade     []*Report `gorm:"foreignKey:ReporterID;constraint:OnDelete:CASCADE" json:"-"`
	ReportsReceived []*Report `gorm:"foreignKey:ReportedUserID;constraint:OnDelete:CASCADE" json:"-"`
	
	// Subscription relationship
	Subscription *Subscription `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"subscription,omitempty"`
	
	// Verification relationships
	Verifications     []*Verification     `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
	VerificationAttempts []*VerificationAttempt `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
	VerificationBadges []*VerificationBadge `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
	
	// Moderation relationships
	Bans              []*Ban              `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
	Appeals           []*Appeal           `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
	ModerationActions []*ModerationAction `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
	Reputation        *UserReputation     `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"reputation,omitempty"`
	QueueItems        []*ModerationQueue  `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
	BlocksMade        []*Block            `gorm:"foreignKey:BlockerID;constraint:OnDelete:CASCADE" json:"-"`
	BlocksReceived    []*Block            `gorm:"foreignKey:BlockedID;constraint:OnDelete:CASCADE" json:"-"`
	ContentAnalyses   []*ContentAnalysis  `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
}

// TableName returns the table name for the User model
func (User) TableName() string {
	return "users"
}

// BeforeCreate GORM hook
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
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