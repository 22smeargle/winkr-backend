package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/domain/entities"
	"github.com/22smeargle/winkr-backend/internal/domain/repositories"
)

// DiscoveryUser represents a user in discovery results
type DiscoveryUser struct {
	ID               uuid.UUID  `json:"id"`
	FirstName         string     `json:"first_name"`
	Age              int        `json:"age"`
	Bio              *string    `json:"bio"`
	Location         *Location  `json:"location,omitempty"`
	Distance         float64    `json:"distance"` // in kilometers
	IsVerified       bool        `json:"is_verified"`
	VerificationLevel int         `json:"verification_level"`
	IsPremium        bool        `json:"is_premium"`
	Photos           []*Photo    `json:"photos"`
	LastActive       *time.Time  `json:"last_active"`
	CreatedAt        time.Time   `json:"created_at"`
}

// Location represents location information
type Location struct {
	Lat     float64 `json:"lat"`
	Lng     float64 `json:"lng"`
	City    *string `json:"city,omitempty"`
	Country *string `json:"country,omitempty"`
}

// Photo represents a photo in discovery results
type Photo struct {
	ID                uuid.UUID `json:"id"`
	URL               string    `json:"url"`
	IsPrimary         bool      `json:"is_primary"`
	VerificationStatus string    `json:"verification_status"`
}

// Match represents a match between two users
type Match struct {
	ID        uuid.UUID `json:"id"`
	User1ID   uuid.UUID `json:"user1_id"`
	User2ID   uuid.UUID `json:"user2_id"`
	MatchedAt  time.Time `json:"matched_at"`
	IsActive   bool      `json:"is_active"`
	User       *User     `json:"user"` // The other user in the match
}

// MatchWithDetails represents a match with additional details
type MatchWithDetails struct {
	*Match
	LastMessage    *Message   `json:"last_message,omitempty"`
	UnreadCount    int        `json:"unread_count"`
	HasConversation bool       `json:"has_conversation"`
}

// Message represents a message in match details
type Message struct {
	ID           uuid.UUID `json:"id"`
	Content      string    `json:"content"`
	MessageType  string    `json:"message_type"`
	IsRead       bool      `json:"is_read"`
	SenderID     uuid.UUID `json:"sender_id"`
	CreatedAt    time.Time `json:"created_at"`
}

// User represents a user in match results
type User struct {
	ID               uuid.UUID  `json:"id"`
	FirstName         string     `json:"first_name"`
	Age              int        `json:"age"`
	Bio              *string    `json:"bio"`
	IsVerified       bool        `json:"is_verified"`
	VerificationLevel int         `json:"verification_level"`
	IsPremium        bool        `json:"is_premium"`
	Photos           []*Photo    `json:"photos"`
}

// DiscoveryStats represents discovery statistics for a user
type DiscoveryStats struct {
	UserID           uuid.UUID `json:"user_id"`
	TotalSwipes       int64     `json:"total_swipes"`
	TotalLikes        int64     `json:"total_likes"`
	TotalPasses       int64     `json:"total_passes"`
	TotalMatches      int64     `json:"total_matches"`
	ActiveMatches     int64     `json:"active_matches"`
	LikeRate          float64   `json:"like_rate"`          // percentage
	MatchRate         float64   `json:"match_rate"`         // percentage
	SwipesToday       int64     `json:"swipes_today"`
	SwipesThisWeek   int64     `json:"swipes_this_week"`
	SwipesThisMonth  int64     `json:"swipes_this_month"`
	MatchesToday      int64     `json:"matches_today"`
	MatchesThisWeek   int64     `json:"matches_this_week"`
	MatchesThisMonth  int64     `json:"matches_this_month"`
	ProfileViews       int64     `json:"profile_views"`
	PhotosCount       int64     `json:"photos_count"`
	LastActiveDays    int       `json:"last_active_days"`
	AccountAgeDays    int       `json:"account_age_days"`
	GeneratedAt       time.Time `json:"generated_at"`
}

// NewDiscoveryUser creates a new DiscoveryUser from entities
func NewDiscoveryUser(user *entities.User, photos []*entities.Photo, distance float64) *DiscoveryUser {
	discoveryPhotos := make([]*Photo, 0, len(photos))
	for _, photo := range photos {
		discoveryPhotos = append(discoveryPhotos, &Photo{
			ID:               photo.ID,
			URL:               photo.FileURL,
			IsPrimary:         photo.IsPrimary,
			VerificationStatus: string(photo.VerificationStatus),
		})
	}

	var location *Location
	if user.HasLocation() {
		lat, lng, _ := user.GetLocation()
		location = &Location{
			Lat:     lat,
			Lng:     lng,
			City:    user.LocationCity,
			Country: user.LocationCountry,
		}
	}

	return &DiscoveryUser{
		ID:               user.ID,
		FirstName:         user.FirstName,
		Age:              user.GetAge(),
		Bio:              user.Bio,
		Location:         location,
		Distance:         distance,
		IsVerified:       user.IsVerified,
		VerificationLevel: int(user.VerificationLevel),
		IsPremium:        user.IsPremium,
		Photos:           discoveryPhotos,
		LastActive:       user.LastActive,
		CreatedAt:        user.CreatedAt,
	}
}

// NewMatch creates a new Match from entities
func NewMatch(match *entities.Match, user1, user2 *entities.User) *Match {
	// Determine which user is the "other" user (this would depend on context)
	// For now, we'll use user2 as the other user
	otherUser := user2

	userPhotos := make([]*Photo, 0) // Would need to fetch photos
	user := &User{
		ID:               otherUser.ID,
		FirstName:         otherUser.FirstName,
		Age:              otherUser.GetAge(),
		Bio:              otherUser.Bio,
		IsVerified:       otherUser.IsVerified,
		VerificationLevel: int(otherUser.VerificationLevel),
		IsPremium:        otherUser.IsPremium,
		Photos:           userPhotos,
	}

	return &Match{
		ID:       match.ID,
		User1ID:  match.User1ID,
		User2ID:  match.User2ID,
		MatchedAt: match.MatchedAt,
		IsActive:  match.IsActive,
		User:      user,
	}
}

// NewMatchWithDetails creates a new MatchWithDetails from entities
func NewMatchWithDetails(matchWithDetails *repositories.MatchWithDetails, photos []*entities.Photo) *MatchWithDetails {
	// Convert base match
	match := &Match{
		ID:       matchWithDetails.ID,
		User1ID:  matchWithDetails.User1ID,
		User2ID:  matchWithDetails.User2ID,
		MatchedAt: matchWithDetails.MatchedAt,
		IsActive:  matchWithDetails.IsActive,
	}

	// Convert other user
	otherUser := matchWithDetails.OtherUser
	userPhotos := make([]*Photo, 0, len(photos))
	for _, photo := range photos {
		userPhotos = append(userPhotos, &Photo{
			ID:               photo.ID,
			URL:               photo.FileURL,
			IsPrimary:         photo.IsPrimary,
			VerificationStatus: string(photo.VerificationStatus),
		})
	}

	user := &User{
		ID:               otherUser.ID,
		FirstName:         otherUser.FirstName,
		Age:              otherUser.GetAge(),
		Bio:              otherUser.Bio,
		IsVerified:       otherUser.IsVerified,
		VerificationLevel: int(otherUser.VerificationLevel),
		IsPremium:        otherUser.IsPremium,
		Photos:           userPhotos,
	}

	// Convert last message if exists
	var lastMessage *Message
	if matchWithDetails.LastMessage != nil {
		lastMessage = &Message{
			ID:          matchWithDetails.LastMessage.ID,
			Content:     matchWithDetails.LastMessage.Content,
			MessageType: string(matchWithDetails.LastMessage.MessageType),
			IsRead:      matchWithDetails.LastMessage.IsRead,
			SenderID:    matchWithDetails.LastMessage.SenderID,
			CreatedAt:   matchWithDetails.LastMessage.CreatedAt,
		}
	}

	return &MatchWithDetails{
		Match:            match,
		User:             user,
		LastMessage:       lastMessage,
		UnreadCount:      matchWithDetails.UnreadCount,
		HasConversation:   matchWithDetails.HasConversation,
	}
}