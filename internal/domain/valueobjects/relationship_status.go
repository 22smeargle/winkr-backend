package valueobjects

import (
	"fmt"
	"strings"
)

// RelationshipStatus represents a relationship status value object
type RelationshipStatus string

// Valid relationship status constants
const (
	RelationshipSingle   RelationshipStatus = "single"
	RelationshipDating   RelationshipStatus = "dating"
	RelationshipEngaged RelationshipStatus = "engaged"
	RelationshipMarried  RelationshipStatus = "married"
	RelationshipDivorced RelationshipStatus = "divorced"
	RelationshipWidowed  RelationshipStatus = "widowed"
	RelationshipSeparated RelationshipStatus = "separated"
)

// NewRelationshipStatus creates a new RelationshipStatus value object
func NewRelationshipStatus(status string) (RelationshipStatus, error) {
	normalized := strings.ToLower(strings.TrimSpace(status))
	
	switch RelationshipStatus(normalized) {
	case RelationshipSingle, RelationshipDating, RelationshipEngaged, 
		 RelationshipMarried, RelationshipDivorced, RelationshipWidowed, RelationshipSeparated:
		return RelationshipStatus(normalized), nil
	default:
		return "", fmt.Errorf("invalid relationship status: %s", status)
	}
}

// IsValid checks if the relationship status is valid
func (rs RelationshipStatus) IsValid() bool {
	validStatuses := []RelationshipStatus{
		RelationshipSingle, RelationshipDating, RelationshipEngaged,
		RelationshipMarried, RelationshipDivorced, RelationshipWidowed, RelationshipSeparated,
	}
	
	for _, validStatus := range validStatuses {
		if rs == validStatus {
			return true
		}
	}
	
	return false
}

// String returns the string representation of relationship status
func (rs RelationshipStatus) String() string {
	return string(rs)
}

// IsSingle returns true if relationship status is single
func (rs RelationshipStatus) IsSingle() bool {
	return rs == RelationshipSingle
}

// IsDating returns true if relationship status is dating
func (rs RelationshipStatus) IsDating() bool {
	return rs == RelationshipDating
}

// IsEngaged returns true if relationship status is engaged
func (rs RelationshipStatus) IsEngaged() bool {
	return rs == RelationshipEngaged
}

// IsMarried returns true if relationship status is married
func (rs RelationshipStatus) IsMarried() bool {
	return rs == RelationshipMarried
}

// IsDivorced returns true if relationship status is divorced
func (rs RelationshipStatus) IsDivorced() bool {
	return rs == RelationshipDivorced
}

// IsWidowed returns true if relationship status is widowed
func (rs RelationshipStatus) IsWidowed() bool {
	return rs == RelationshipWidowed
}

// IsSeparated returns true if relationship status is separated
func (rs RelationshipStatus) IsSeparated() bool {
	return rs == RelationshipSeparated
}

// IsAvailable returns true if person is available for dating
func (rs RelationshipStatus) IsAvailable() bool {
	return rs == RelationshipSingle || rs == RelationshipDivorced || rs == RelationshipWidowed || rs == RelationshipSeparated
}

// IsInRelationship returns true if person is in a relationship
func (rs RelationshipStatus) IsInRelationship() bool {
	return rs == RelationshipDating || rs == RelationshipEngaged || rs == RelationshipMarried
}

// GetDisplayName returns the display name for relationship status
func (rs RelationshipStatus) GetDisplayName() string {
	switch rs {
	case RelationshipSingle:
		return "Single"
	case RelationshipDating:
		return "Dating"
	case RelationshipEngaged:
		return "Engaged"
	case RelationshipMarried:
		return "Married"
	case RelationshipDivorced:
		return "Divorced"
	case RelationshipWidowed:
		return "Widowed"
	case RelationshipSeparated:
		return "Separated"
	default:
		return "Unknown"
	}
}

// GetAllRelationshipStatuses returns all valid relationship statuses
func GetAllRelationshipStatuses() []RelationshipStatus {
	return []RelationshipStatus{
		RelationshipSingle, RelationshipDating, RelationshipEngaged,
		RelationshipMarried, RelationshipDivorced, RelationshipWidowed, RelationshipSeparated,
	}
}

// GetAvailableStatuses returns relationship statuses that indicate availability
func GetAvailableStatuses() []RelationshipStatus {
	return []RelationshipStatus{
		RelationshipSingle, RelationshipDivorced, RelationshipWidowed, RelationshipSeparated,
	}
}

// GetInRelationshipStatuses returns relationship statuses that indicate being in a relationship
func GetInRelationshipStatuses() []RelationshipStatus {
	return []RelationshipStatus{
		RelationshipDating, RelationshipEngaged, RelationshipMarried,
	}
}

// GetRelationshipStatusDisplayNames returns all relationship status display names
func GetRelationshipStatusDisplayNames() []string {
	statuses := GetAllRelationshipStatuses()
	displayNames := make([]string, len(statuses))
	for i, status := range statuses {
		displayNames[i] = status.GetDisplayName()
	}
	return displayNames
}

// ParseRelationshipStatus parses a string into a RelationshipStatus value object
func ParseRelationshipStatus(status string) (RelationshipStatus, error) {
	return NewRelationshipStatus(status)
}

// RelationshipStatusFromString creates a RelationshipStatus from string (panics on invalid input)
func RelationshipStatusFromString(status string) RelationshipStatus {
	rs, err := NewRelationshipStatus(status)
	if err != nil {
		panic(fmt.Sprintf("invalid relationship status: %s", status))
	}
	return rs
}

// RelationshipStatusFilter represents a filter for relationship statuses
type RelationshipStatusFilter struct {
	Statuses []RelationshipStatus `json:"statuses"`
	IncludeAvailable bool           `json:"include_available"`
	IncludeInRelationship bool      `json:"include_in_relationship"`
}

// NewRelationshipStatusFilter creates a new relationship status filter
func NewRelationshipStatusFilter() *RelationshipStatusFilter {
	return &RelationshipStatusFilter{
		Statuses:              []RelationshipStatus{},
		IncludeAvailable:       true,
		IncludeInRelationship:  true,
	}
}

// AddStatus adds a relationship status to the filter
func (rsf *RelationshipStatusFilter) AddStatus(status RelationshipStatus) {
	rsf.Statuses = append(rsf.Statuses, status)
}

// RemoveStatus removes a relationship status from the filter
func (rsf *RelationshipStatusFilter) RemoveStatus(status RelationshipStatus) {
	for i, s := range rsf.Statuses {
		if s == status {
			rsf.Statuses = append(rsf.Statuses[:i], rsf.Statuses[i+1:]...)
			break
		}
	}
}

// Clear clears all statuses from the filter
func (rsf *RelationshipStatusFilter) Clear() {
	rsf.Statuses = []RelationshipStatus{}
}

// Contains checks if the filter contains a specific relationship status
func (rsf *RelationshipStatusFilter) Contains(status RelationshipStatus) bool {
	for _, s := range rsf.Statuses {
		if s == status {
			return true
		}
	}
	return false
}

// GetFilteredStatuses returns the filtered relationship statuses based on the filter settings
func (rsf *RelationshipStatusFilter) GetFilteredStatuses() []RelationshipStatus {
	var result []RelationshipStatus
	
	if len(rsf.Statuses) > 0 {
		result = rsf.Statuses
	} else {
		result = GetAllRelationshipStatuses()
	}
	
	if !rsf.IncludeAvailable {
		result = filterStatuses(result, GetAvailableStatuses(), true)
	}
	
	if !rsf.IncludeInRelationship {
		result = filterStatuses(result, GetInRelationshipStatuses(), true)
	}
	
	return result
}

// filterStatuses filters out statuses that are in the exclude list
func filterStatuses(statuses, exclude []RelationshipStatus, exclude bool) []RelationshipStatus {
	var result []RelationshipStatus
	
	for _, status := range statuses {
		shouldExclude := false
		for _, excl := range exclude {
			if status == excl {
				shouldExclude = true
				break
			}
		}
		
		if exclude == shouldExclude {
			result = append(result, status)
		}
	}
	
	return result
}