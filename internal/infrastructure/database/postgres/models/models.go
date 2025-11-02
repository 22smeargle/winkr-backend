package models

// AllModels returns all database models for migration
func AllModels() []interface{} {
	return []interface{}{
		&User{},
		&Photo{},
		&Message{},
		&Conversation{},
		&Match{},
		&Swipe{},
		&UserPreferences{},
		&Report{},
		&AdminUser{},
		&Subscription{},
		&Verification{},
		&VerificationAttempt{},
		&VerificationBadge{},
		&Ban{},
		&Appeal{},
		&ModerationAction{},
		&UserReputation{},
		&ModerationQueue{},
		&Block{},
		&ContentAnalysis{},
	}
}