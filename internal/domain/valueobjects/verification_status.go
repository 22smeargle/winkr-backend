package valueobjects

import (
	"fmt"
	"strings"
)

// VerificationStatus represents a verification status value object
type VerificationStatus string

// Valid verification status constants
const (
	VerificationStatusPending  VerificationStatus = "pending"
	VerificationStatusApproved VerificationStatus = "approved"
	VerificationStatusRejected VerificationStatus = "rejected"
)

// NewVerificationStatus creates a new VerificationStatus value object
func NewVerificationStatus(status string) (VerificationStatus, error) {
	normalized := strings.ToLower(strings.TrimSpace(status))
	
	switch VerificationStatus(normalized) {
	case VerificationStatusPending, VerificationStatusApproved, VerificationStatusRejected:
		return VerificationStatus(normalized), nil
	default:
		return "", fmt.Errorf("invalid verification status: %s", status)
	}
}

// IsValid checks if the verification status is valid
func (vs VerificationStatus) IsValid() bool {
	return vs == VerificationStatusPending || vs == VerificationStatusApproved || vs == VerificationStatusRejected
}

// String returns the string representation of verification status
func (vs VerificationStatus) String() string {
	return string(vs)
}

// IsPending returns true if verification status is pending
func (vs VerificationStatus) IsPending() bool {
	return vs == VerificationStatusPending
}

// IsApproved returns true if verification status is approved
func (vs VerificationStatus) IsApproved() bool {
	return vs == VerificationStatusApproved
}

// IsRejected returns true if verification status is rejected
func (vs VerificationStatus) IsRejected() bool {
	return vs == VerificationStatusRejected
}

// IsVerified returns true if verification status is approved
func (vs VerificationStatus) IsVerified() bool {
	return vs.IsApproved()
}

// IsUnverified returns true if verification status is not approved
func (vs VerificationStatus) IsUnverified() bool {
	return !vs.IsApproved()
}

// GetDisplayName returns the display name for verification status
func (vs VerificationStatus) GetDisplayName() string {
	switch vs {
	case VerificationStatusPending:
		return "Pending"
	case VerificationStatusApproved:
		return "Approved"
	case VerificationStatusRejected:
		return "Rejected"
	default:
		return "Unknown"
	}
}

// GetColor returns the color associated with the verification status
func (vs VerificationStatus) GetColor() string {
	switch vs {
	case VerificationStatusPending:
		return "orange"
	case VerificationStatusApproved:
		return "green"
	case VerificationStatusRejected:
		return "red"
	default:
		return "gray"
	}
}

// GetIcon returns the icon associated with the verification status
func (vs VerificationStatus) GetIcon() string {
	switch vs {
	case VerificationStatusPending:
		return "clock"
	case VerificationStatusApproved:
		return "check-circle"
	case VerificationStatusRejected:
		return "x-circle"
	default:
		return "question-circle"
	}
}

// GetAllVerificationStatuses returns all valid verification statuses
func GetAllVerificationStatuses() []VerificationStatus {
	return []VerificationStatus{
		VerificationStatusPending,
		VerificationStatusApproved,
		VerificationStatusRejected,
	}
}

// GetVerificationStatusDisplayNames returns all verification status display names
func GetVerificationStatusDisplayNames() []string {
	statuses := GetAllVerificationStatuses()
	displayNames := make([]string, len(statuses))
	for i, status := range statuses {
		displayNames[i] = status.GetDisplayName()
	}
	return displayNames
}

// ParseVerificationStatus parses a string into a VerificationStatus value object
func ParseVerificationStatus(status string) (VerificationStatus, error) {
	return NewVerificationStatus(status)
}

// VerificationStatusFromString creates a VerificationStatus from string (panics on invalid input)
func VerificationStatusFromString(status string) VerificationStatus {
	vs, err := NewVerificationStatus(status)
	if err != nil {
		panic(fmt.Sprintf("invalid verification status: %s", status))
	}
	return vs
}

// VerificationStatusFilter represents a filter for verification statuses
type VerificationStatusFilter struct {
	Statuses []VerificationStatus `json:"statuses"`
	IncludePending  bool              `json:"include_pending"`
	IncludeApproved bool              `json:"include_approved"`
	IncludeRejected bool              `json:"include_rejected"`
}

// NewVerificationStatusFilter creates a new verification status filter
func NewVerificationStatusFilter() *VerificationStatusFilter {
	return &VerificationStatusFilter{
		Statuses:       []VerificationStatus{},
		IncludePending:  true,
		IncludeApproved: true,
		IncludeRejected: true,
	}
}

// AddStatus adds a verification status to the filter
func (vsf *VerificationStatusFilter) AddStatus(status VerificationStatus) {
	vsf.Statuses = append(vsf.Statuses, status)
}

// RemoveStatus removes a verification status from the filter
func (vsf *VerificationStatusFilter) RemoveStatus(status VerificationStatus) {
	for i, s := range vsf.Statuses {
		if s == status {
			vsf.Statuses = append(vsf.Statuses[:i], vsf.Statuses[i+1:]...)
			break
		}
	}
}

// Clear clears all statuses from the filter
func (vsf *VerificationStatusFilter) Clear() {
	vsf.Statuses = []VerificationStatus{}
}

// Contains checks if the filter contains a specific verification status
func (vsf *VerificationStatusFilter) Contains(status VerificationStatus) bool {
	for _, s := range vsf.Statuses {
		if s == status {
			return true
		}
	}
	return false
}

// GetFilteredStatuses returns the filtered verification statuses based on the filter settings
func (vsf *VerificationStatusFilter) GetFilteredStatuses() []VerificationStatus {
	var result []VerificationStatus
	
	if len(vsf.Statuses) > 0 {
		result = vsf.Statuses
	} else {
		result = GetAllVerificationStatuses()
	}
	
	if !vsf.IncludePending {
		result = filterVerificationStatuses(result, []VerificationStatus{VerificationStatusPending})
	}
	
	if !vsf.IncludeApproved {
		result = filterVerificationStatuses(result, []VerificationStatus{VerificationStatusApproved})
	}
	
	if !vsf.IncludeRejected {
		result = filterVerificationStatuses(result, []VerificationStatus{VerificationStatusRejected})
	}
	
	return result
}

// filterVerificationStatuses filters out statuses that are in the exclude list
func filterVerificationStatuses(statuses []VerificationStatus, exclude []VerificationStatus) []VerificationStatus {
	var result []VerificationStatus
	
	for _, status := range statuses {
		shouldExclude := false
		for _, excl := range exclude {
			if status == excl {
				shouldExclude = true
				break
			}
		}
		
		if !shouldExclude {
			result = append(result, status)
		}
	}
	
	return result
}

// VerificationStatusTransition represents a valid transition between verification statuses
type VerificationStatusTransition struct {
	From VerificationStatus `json:"from"`
	To   VerificationStatus `json:"to"`
}

// GetValidTransitions returns all valid verification status transitions
func GetValidTransitions() []VerificationStatusTransition {
	return []VerificationStatusTransition{
		{From: VerificationStatusPending, To: VerificationStatusApproved},
		{From: VerificationStatusPending, To: VerificationStatusRejected},
		{From: VerificationStatusRejected, To: VerificationStatusPending},
	}
}

// IsValidTransition checks if a transition from one status to another is valid
func IsValidTransition(from, to VerificationStatus) bool {
	transitions := GetValidTransitions()
	for _, transition := range transitions {
		if transition.From == from && transition.To == to {
			return true
		}
	}
	return false
}

// GetNextPossibleStatuses returns the next possible statuses from the current status
func GetNextPossibleStatuses(current VerificationStatus) []VerificationStatus {
	var result []VerificationStatus
	
	transitions := GetValidTransitions()
	for _, transition := range transitions {
		if transition.From == current {
			result = append(result, transition.To)
		}
	}
	
	return result
}