package validator

import (
	"regexp"
	"strings"

	"github.com/22smeargle/winkr-backend/internal/application/dto"
	"github.com/22smeargle/winkr-backend/pkg/errors"
)

// ProfileValidator handles profile validation
type ProfileValidator struct {
	validator *Validator
}

// NewProfileValidator creates a new ProfileValidator instance
func NewProfileValidator() *ProfileValidator {
	return &ProfileValidator{
		validator: NewValidator(),
	}
}

// ValidateUpdateProfileRequest validates update profile request
func (v *ProfileValidator) ValidateUpdateProfileRequest(req *dto.UpdateProfileRequestDTO) error {
	// Validate first name if provided
	if req.FirstName != nil {
		if err := v.validateFirstName(*req.FirstName); err != nil {
			return err
		}
	}

	// Validate last name if provided
	if req.LastName != nil {
		if err := v.validateLastName(*req.LastName); err != nil {
			return err
		}
	}

	// Validate bio if provided
	if req.Bio != nil {
		if err := v.validateBio(*req.Bio); err != nil {
			return err
		}
	}

	// Validate interested in if provided
	if len(req.InterestedIn) > 0 {
		if err := v.validateInterestedIn(req.InterestedIn); err != nil {
			return err
		}
	}

	// Validate preferences if provided
	if req.Preferences != nil {
		if err := v.validatePreferences(req.Preferences); err != nil {
			return err
		}
	}

	return nil
}

// ValidateUpdateLocationRequest validates update location request
func (v *ProfileValidator) ValidateUpdateLocationRequest(req *dto.UpdateLocationRequestDTO) error {
	// Validate latitude
	if req.Latitude < -90 || req.Latitude > 90 {
		return errors.ErrInvalidLatitude
	}

	// Validate longitude
	if req.Longitude < -180 || req.Longitude > 180 {
		return errors.ErrInvalidLongitude
	}

	// Validate city if provided
	if req.City != nil {
		if err := v.validateCity(*req.City); err != nil {
			return err
		}
	}

	// Validate country if provided
	if req.Country != nil {
		if err := v.validateCountry(*req.Country); err != nil {
			return err
		}
	}

	return nil
}

// ValidateDeleteAccountRequest validates delete account request
func (v *ProfileValidator) ValidateDeleteAccountRequest(req *dto.DeleteAccountRequestDTO) error {
	// Validate confirmation
	if !req.Confirm {
		return errors.ErrAccountDeletionNotConfirmed
	}

	// Validate reason if provided
	if req.Reason != "" {
		if err := v.validateReason(req.Reason); err != nil {
			return err
		}
	}

	return nil
}

// validateFirstName validates first name
func (v *ProfileValidator) validateFirstName(firstName string) error {
	firstName = strings.TrimSpace(firstName)

	if len(firstName) < 2 {
		return errors.ErrFirstNameTooShort
	}

	if len(firstName) > 100 {
		return errors.ErrFirstNameTooLong
	}

	// Check for invalid characters
	if matched, _ := regexp.MatchString(`^[a-zA-Z\s\-']+$`, firstName); !matched {
		return errors.ErrInvalidFirstName
	}

	// Check for inappropriate content
	if v.containsInappropriateContent(firstName) {
		return errors.ErrInappropriateFirstName
	}

	return nil
}

// validateLastName validates last name
func (v *ProfileValidator) validateLastName(lastName string) error {
	lastName = strings.TrimSpace(lastName)

	if len(lastName) < 2 {
		return errors.ErrLastNameTooShort
	}

	if len(lastName) > 100 {
		return errors.ErrLastNameTooLong
	}

	// Check for invalid characters
	if matched, _ := regexp.MatchString(`^[a-zA-Z\s\-']+$`, lastName); !matched {
		return errors.ErrInvalidLastName
	}

	// Check for inappropriate content
	if v.containsInappropriateContent(lastName) {
		return errors.ErrInappropriateLastName
	}

	return nil
}

// validateBio validates bio
func (v *ProfileValidator) validateBio(bio string) error {
	bio = strings.TrimSpace(bio)

	if len(bio) > 500 {
		return errors.ErrBioTooLong
	}

	// Check for inappropriate content
	if v.containsInappropriateContent(bio) {
		return errors.ErrInappropriateBio
	}

	// Check for spam patterns
	if v.containsSpamPatterns(bio) {
		return errors.ErrSpamBio
	}

	return nil
}

// validateInterestedIn validates interested in preferences
func (v *ProfileValidator) validateInterestedIn(interestedIn []string) error {
	if len(interestedIn) == 0 {
		return errors.ErrInterestedInRequired
	}

	validGenders := map[string]bool{
		"male":   true,
		"female": true,
		"other":  true,
	}

	for _, gender := range interestedIn {
		gender = strings.ToLower(strings.TrimSpace(gender))
		if !validGenders[gender] {
			return errors.ErrInvalidGender
		}
	}

	return nil
}

// validatePreferences validates user preferences
func (v *ProfileValidator) validatePreferences(prefs *dto.PreferencesDTO) error {
	// Validate age range
	if prefs.AgeMin < 18 {
		return errors.ErrAgeMinTooYoung
	}

	if prefs.AgeMax > 100 {
		return errors.ErrAgeMaxTooOld
	}

	if prefs.AgeMin >= prefs.AgeMax {
		return errors.ErrInvalidAgeRange
	}

	// Validate max distance
	if prefs.MaxDistance < 1 || prefs.MaxDistance > 500 {
		return errors.ErrInvalidMaxDistance
	}

	return nil
}

// validateCity validates city name
func (v *ProfileValidator) validateCity(city string) error {
	city = strings.TrimSpace(city)

	if len(city) == 0 {
		return errors.ErrCityRequired
	}

	if len(city) > 100 {
		return errors.ErrCityTooLong
	}

	// Check for invalid characters
	if matched, _ := regexp.MatchString(`^[a-zA-Z\s\-']+$`, city); !matched {
		return errors.ErrInvalidCity
	}

	return nil
}

// validateCountry validates country name
func (v *ProfileValidator) validateCountry(country string) error {
	country = strings.TrimSpace(country)

	if len(country) == 0 {
		return errors.ErrCountryRequired
	}

	if len(country) > 100 {
		return errors.ErrCountryTooLong
	}

	// Check for invalid characters
	if matched, _ := regexp.MatchString(`^[a-zA-Z\s\-']+$`, country); !matched {
		return errors.ErrInvalidCountry
	}

	return nil
}

// validateReason validates deletion reason
func (v *ProfileValidator) validateReason(reason string) error {
	reason = strings.TrimSpace(reason)

	if len(reason) > 500 {
		return errors.ErrReasonTooLong
	}

	// Check for inappropriate content
	if v.containsInappropriateContent(reason) {
		return errors.ErrInappropriateReason
	}

	return nil
}

// containsInappropriateContent checks for inappropriate content in text
func (v *ProfileValidator) containsInappropriateContent(text string) bool {
	// List of inappropriate words (simplified for example)
	inappropriateWords := []string{
		"spam", "scam", "fraud", "illegal", "drugs", "violence",
		"porn", "sex", "nude", "naked", "escort", "prostitute",
		// Add more words as needed
	}

	textLower := strings.ToLower(text)
	for _, word := range inappropriateWords {
		if strings.Contains(textLower, word) {
			return true
		}
	}

	return false
}

// containsSpamPatterns checks for spam patterns in text
func (v *ProfileValidator) containsSpamPatterns(text string) bool {
	// Common spam patterns
	spamPatterns := []string{
		`www\.`,
		`http`,
		`\.com`,
		`click here`,
		`buy now`,
		`limited time`,
		`act fast`,
		`free money`,
		`make money`,
		// Add more patterns as needed
	}

	textLower := strings.ToLower(text)
	for _, pattern := range spamPatterns {
		if matched, _ := regexp.MatchString(pattern, textLower); matched {
			return true
		}
	}

	return false
}

// ValidateProfilePhoto validates profile photo
func (v *ProfileValidator) ValidateProfilePhoto(filename string, fileSize int64, contentType string) error {
	// Validate file type
	validTypes := []string{
		"image/jpeg",
		"image/jpg",
		"image/png",
		"image/webp",
	}

	isValidType := false
	for _, validType := range validTypes {
		if contentType == validType {
			isValidType = true
			break
		}
	}

	if !isValidType {
		return errors.ErrInvalidFileType
	}

	// Validate file size (5MB max)
	if fileSize > 5*1024*1024 {
		return errors.ErrFileTooLarge
	}

	// Validate file extension
	validExtensions := []string{".jpg", ".jpeg", ".png", ".webp"}
	hasValidExtension := false
	for _, ext := range validExtensions {
		if strings.HasSuffix(strings.ToLower(filename), ext) {
			hasValidExtension = true
			break
		}
	}

	if !hasValidExtension {
		return errors.ErrInvalidFileExtension
	}

	return nil
}

// ValidateProfileCompletion validates profile completion requirements
func (v *ProfileValidator) ValidateProfileCompletion(completion int) error {
	if completion < 0 || completion > 100 {
		return errors.ErrInvalidCompletionPercentage
	}

	return nil
}

// ValidateAge validates user age
func (v *ProfileValidator) ValidateAge(age int) error {
	if age < 18 {
		return errors.ErrAgeTooYoung
	}

	if age > 100 {
		return errors.ErrAgeTooOld
	}

	return nil
}

// ValidateDistance validates distance
func (v *ProfileValidator) ValidateDistance(distance int) error {
	if distance < 0 {
		return errors.ErrInvalidDistance
	}

	if distance > 500 {
		return errors.ErrDistanceTooFar
	}

	return nil
}