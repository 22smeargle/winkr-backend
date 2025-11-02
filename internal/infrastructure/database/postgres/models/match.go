package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Match represents a match between two users in database
type Match struct {
	ID        uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	User1ID   uuid.UUID  `gorm:"type:uuid;not null;index" json:"user1_id"`
	User2ID   uuid.UUID  `gorm:"type:uuid;not null;index" json:"user2_id"`
	MatchedAt time.Time  `gorm:"autoCreateTime" json:"matched_at"`
	IsActive   bool       `gorm:"default:true;index" json:"is_active"`
	CreatedAt  time.Time  `gorm:"autoCreateTime" json:"created_at"`

	// Relationships
	User1        *User         `gorm:"foreignKey:User1ID;constraint:OnDelete:CASCADE" json:"user1,omitempty"`
	User2        *User         `gorm:"foreignKey:User2ID;constraint:OnDelete:CASCADE" json:"user2,omitempty"`
	Conversation *Conversation `gorm:"foreignKey:MatchID;constraint:OnDelete:CASCADE" json:"conversation,omitempty"`
}

// TableName returns the table name for Match model
func (Match) TableName() string {
	return "matches"
}

// BeforeCreate GORM hook
func (m *Match) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}

// IsUserInMatch checks if a user is part of the match
func (m *Match) IsUserInMatch(userID uuid.UUID) bool {
	return m.User1ID == userID || m.User2ID == userID
}

// GetOtherUserID returns the ID of the other user in the match
func (m *Match) GetOtherUserID(userID uuid.UUID) (uuid.UUID, bool) {
	if m.User1ID == userID {
		return m.User2ID, true
	}
	if m.User2ID == userID {
		return m.User1ID, true
	}
	return uuid.Nil, false
}

// Deactivate deactivates the match
func (m *Match) Deactivate() {
	m.IsActive = false
}

// Activate activates the match
func (m *Match) Activate() {
	m.IsActive = true
}

// Swipe represents a user's swipe action on another user in database
type Swipe struct {
	ID       uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	SwiperID uuid.UUID  `gorm:"type:uuid;not null;index" json:"swiper_id"`
	SwipedID uuid.UUID  `gorm:"type:uuid;not null;index" json:"swiped_id"`
	IsLike   bool       `gorm:"not null" json:"is_like"`
	CreatedAt time.Time  `gorm:"autoCreateTime" json:"created_at"`

	// Relationships
	Swiper *User `gorm:"foreignKey:SwiperID;constraint:OnDelete:CASCADE" json:"swiper,omitempty"`
	Swiped *User `gorm:"foreignKey:SwipedID;constraint:OnDelete:CASCADE" json:"swiped,omitempty"`
}

// TableName returns the table name for Swipe model
func (Swipe) TableName() string {
	return "swipes"
}

// BeforeCreate GORM hook
func (s *Swipe) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

// IsLikeAction returns true if the swipe is a like
func (s *Swipe) IsLikeAction() bool {
	return s.IsLike
}

// IsPass returns true if the swipe is a pass
func (s *Swipe) IsPass() bool {
	return !s.IsLike
}

// UserPreferences represents user's matching preferences in database
type UserPreferences struct {
	ID          uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID      uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex" json:"user_id"`
	AgeMin      int        `gorm:"default:18" json:"age_min"`
	AgeMax      int        `gorm:"default:100" json:"age_max"`
	MaxDistance int        `gorm:"default:50" json:"max_distance"` // in kilometers
	ShowMe      bool       `gorm:"default:true" json:"show_me"`
	CreatedAt   time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime" json:"updated_at"`

	// Relationships
	User *User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
}

// TableName returns the table name for UserPreferences model
func (UserPreferences) TableName() string {
	return "user_preferences"
}

// BeforeCreate GORM hook
func (up *UserPreferences) BeforeCreate(tx *gorm.DB) error {
	if up.ID == uuid.Nil {
		up.ID = uuid.New()
	}
	return nil
}

// IsAgeInRange checks if an age is within the user's preferences
func (up *UserPreferences) IsAgeInRange(age int) bool {
	return age >= up.AgeMin && age <= up.AgeMax
}

// IsWithinDistance checks if a distance is within the user's preferences
func (up *UserPreferences) IsWithinDistance(distance int) bool {
	return distance <= up.MaxDistance
}

// CanBeSeen returns true if the user wants to be seen by others
func (up *UserPreferences) CanBeSeen() bool {
	return up.ShowMe
}

// UpdatePreferences updates the user preferences
func (up *UserPreferences) UpdatePreferences(ageMin, ageMax, maxDistance int, showMe bool) {
	up.AgeMin = ageMin
	up.AgeMax = ageMax
	up.MaxDistance = maxDistance
	up.ShowMe = showMe
}