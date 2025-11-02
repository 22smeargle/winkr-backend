package valueobjects

import (
	"fmt"
	"strings"
)

// Gender represents a gender value object
type Gender string

// Valid gender constants
const (
	GenderMale   Gender = "male"
	GenderFemale Gender = "female"
	GenderOther  Gender = "other"
)

// NewGender creates a new Gender value object
func NewGender(gender string) (Gender, error) {
	normalized := strings.ToLower(strings.TrimSpace(gender))
	
	switch Gender(normalized) {
	case GenderMale, GenderFemale, GenderOther:
		return Gender(normalized), nil
	default:
		return "", fmt.Errorf("invalid gender: %s", gender)
	}
}

// IsValid checks if the gender is valid
func (g Gender) IsValid() bool {
	return g == GenderMale || g == GenderFemale || g == GenderOther
}

// String returns the string representation of gender
func (g Gender) String() string {
	return string(g)
}

// IsMale returns true if gender is male
func (g Gender) IsMale() bool {
	return g == GenderMale
}

// IsFemale returns true if gender is female
func (g Gender) IsFemale() bool {
	return g == GenderFemale
}

// IsOther returns true if gender is other
func (g Gender) IsOther() bool {
	return g == GenderOther
}

// GetDisplayName returns the display name for the gender
func (g Gender) GetDisplayName() string {
	switch g {
	case GenderMale:
		return "Male"
	case GenderFemale:
		return "Female"
	case GenderOther:
		return "Other"
	default:
		return "Unknown"
	}
}

// GetAllGenders returns all valid genders
func GetAllGenders() []Gender {
	return []Gender{GenderMale, GenderFemale, GenderOther}
}

// GetGenderDisplayNames returns all gender display names
func GetGenderDisplayNames() []string {
	genders := GetAllGenders()
	displayNames := make([]string, len(genders))
	for i, gender := range genders {
		displayNames[i] = gender.GetDisplayName()
	}
	return displayNames
}

// ParseGender parses a string into a Gender value object
func ParseGender(gender string) (Gender, error) {
	return NewGender(gender)
}

// GenderFromString creates a Gender from string (panics on invalid input)
func GenderFromString(gender string) Gender {
	g, err := NewGender(gender)
	if err != nil {
		panic(fmt.Sprintf("invalid gender: %s", gender))
	}
	return g
}

// InterestedIn represents a slice of genders a user is interested in
type InterestedIn []Gender

// NewInterestedIn creates a new InterestedIn value object
func NewInterestedIn(genders []string) (InterestedIn, error) {
	var result InterestedIn
	
	for _, gender := range genders {
		g, err := NewGender(gender)
		if err != nil {
			return nil, fmt.Errorf("invalid gender in interested_in: %s", gender)
		}
		result = append(result, g)
	}
	
	if len(result) == 0 {
		return nil, fmt.Errorf("at least one gender must be specified in interested_in")
	}
	
	return result, nil
}

// IsValid checks if the interested in list is valid
func (ii InterestedIn) IsValid() bool {
	if len(ii) == 0 {
		return false
	}
	
	for _, gender := range ii {
		if !gender.IsValid() {
			return false
		}
	}
	
	return true
}

// Contains checks if the interested in list contains a specific gender
func (ii InterestedIn) Contains(gender Gender) bool {
	for _, g := range ii {
		if g == gender {
			return true
		}
	}
	return false
}

// ToStringSlice converts InterestedIn to string slice
func (ii InterestedIn) ToStringSlice() []string {
	result := make([]string, len(ii))
	for i, gender := range ii {
		result[i] = gender.String()
	}
	return result
}

// ToDisplayNames converts InterestedIn to display names
func (ii InterestedIn) ToDisplayNames() []string {
	result := make([]string, len(ii))
	for i, gender := range ii {
		result[i] = gender.GetDisplayName()
	}
	return result
}

// Equals checks if two InterestedIn slices are equal
func (ii InterestedIn) Equals(other InterestedIn) bool {
	if len(ii) != len(other) {
		return false
	}
	
	for _, gender := range ii {
		if !other.Contains(gender) {
			return false
		}
	}
	
	return true
}

// Clone creates a copy of the InterestedIn slice
func (ii InterestedIn) Clone() InterestedIn {
	result := make(InterestedIn, len(ii))
	copy(result, ii)
	return result
}