package entities

import (
	"time"

	"github.com/google/uuid"
)

// Match represents a match between two users
type Match struct {
	ID         uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	User1ID    uuid.UUID  `json:"user1_id" gorm:"type:uuid;not null;index"`
	User2ID    uuid.UUID  `json:"user2_id" gorm:"type:uuid;not null;index"`
	MatchedAt  time.Time  `json:"matched_at" gorm:"autoCreateTime"`
	IsActive   bool       `json:"is_active" gorm:"default:true"`
	CreatedAt  time.Time  `json:"created_at" gorm:"autoCreateTime"`

	// Relationships
	User1 *User         `json:"user1,omitempty" gorm:"foreignKey:User1ID"`
	User2 *User         `json:"user2,omitempty" gorm:"foreignKey:User2ID"`
	Conversation *Conversation `json:"conversation,omitempty" gorm:"foreignKey:MatchID"`
}

// TableName returns the table name for Match entity
func (Match) TableName() string {
	return "matches"
}

// Swipe represents a user's swipe action on another user
type Swipe struct {
	ID        uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	SwiperID  uuid.UUID  `json:"swiper_id" gorm:"type:uuid;not null;index"`
	SwipedID  uuid.UUID  `json:"swiped_id" gorm:"type:uuid;not null;index"`
	IsLike    bool       `json:"is_like" gorm:"not null"`
	CreatedAt time.Time  `json:"created_at" gorm:"autoCreateTime"`

	// Relationships
	Swiper *User `json:"swiper,omitempty" gorm:"foreignKey:SwiperID"`
	Swiped *User `json:"swiped,omitempty" gorm:"foreignKey:SwipedID"`
}

// TableName returns the table name for Swipe entity
func (Swipe) TableName() string {
	return "swipes"
}

// UserPreferences represents user's matching preferences
type UserPreferences struct {
	ID          uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID      uuid.UUID  `json:"user_id" gorm:"type:uuid;not null;uniqueIndex"`
	AgeMin      int        `json:"age_min" gorm:"default:18"`
	AgeMax      int        `json:"age_max" gorm:"default:100"`
	MaxDistance int        `json:"max_distance" gorm:"default:50"` // in kilometers
	ShowMe      bool       `json:"show_me" gorm:"default:true"`
	CreatedAt   time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time  `json:"updated_at" gorm:"autoUpdateTime"`

	// Relationships
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// TableName returns the table name for UserPreferences entity
func (UserPreferences) TableName() string {
	return "user_preferences"
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

// IsLikeAction returns true if swipe is a like
// IsLike returns true if the swipe is a like
func (s *Swipe) IsLikeAction() bool {
	return s.IsLike
}

// IsPass returns true if the swipe is a pass
func (s *Swipe) IsPass() bool {
	return !s.IsLike
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