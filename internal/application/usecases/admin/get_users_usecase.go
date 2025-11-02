package admin

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/22smeargle/winkr-backend/internal/domain/entities"
	"github.com/22smeargle/winkr-backend/internal/domain/repositories"
	"github.com/22smeargle/winkr-backend/pkg/logger"
)

// GetUsersUseCase handles retrieving users with filtering and pagination
type GetUsersUseCase struct {
	userRepo repositories.UserRepository
}

// NewGetUsersUseCase creates a new GetUsersUseCase
func NewGetUsersUseCase(userRepo repositories.UserRepository) *GetUsersUseCase {
	return &GetUsersUseCase{
		userRepo: userRepo,
	}
}

// UserFilters represents filtering options for user queries
type UserFilters struct {
	Status    string `json:"status" validate:"omitempty,oneof=active inactive banned verified unverified premium basic"`
	Verified  string `json:"verified" validate:"omitempty,oneof=true false"`
	Premium   string `json:"premium" validate:"omitempty,oneof=true false"`
	Search    string `json:"search"`
	SortBy    string `json:"sort_by" validate:"omitempty,oneof=created_at updated_at last_active email first_name"`
	SortOrder string `json:"sort_order" validate:"omitempty,oneof=asc desc"`
	Limit     int    `json:"limit" validate:"min=1,max=100"`
	Offset    int    `json:"offset" validate:"min=0"`
}

// GetUsersRequest represents the request to get users
type GetUsersRequest struct {
	AdminID uuid.UUID   `json:"admin_id" validate:"required"`
	Filters  UserFilters `json:"filters"`
}

// GetUsersResponse represents the response from getting users
type GetUsersResponse struct {
	Users []*entities.User `json:"users"`
	Total int64           `json:"total"`
}

// Execute retrieves users with filtering and pagination
func (uc *GetUsersUseCase) Execute(ctx context.Context, adminID uuid.UUID, filters UserFilters) (*GetUsersResponse, error) {
	logger.Info("GetUsers use case executed", "admin_id", adminID, "filters", filters)

	// Build query parameters
	query := buildUserQuery(filters)

	// Get users from repository
	users, err := uc.userRepo.GetAllUsers(ctx, filters.Limit, filters.Offset)
	if err != nil {
		logger.Error("Failed to get users from repository", err, "admin_id", adminID)
		return nil, fmt.Errorf("failed to get users: %w", err)
	}

	// Get total count for pagination
	total, err := uc.userRepo.GetActiveUsersCount(ctx)
	if err != nil {
		logger.Error("Failed to get total users count", err, "admin_id", adminID)
		return nil, fmt.Errorf("failed to get total users count: %w", err)
	}

	// Apply additional filtering if needed
	filteredUsers := applyUserFilters(users, filters)

	logger.Info("GetUsers use case completed successfully", "admin_id", adminID, "count", len(filteredUsers))
	return &GetUsersResponse{
		Users: filteredUsers,
		Total: total,
	}, nil
}

// buildUserQuery builds a query string based on filters
func buildUserQuery(filters UserFilters) string {
	// This would typically build a SQL query or use an ORM query builder
	// For now, we'll apply filters in memory
	return ""
}

// applyUserFilters applies additional filters to the user list
func applyUserFilters(users []*entities.User, filters UserFilters) []*entities.User {
	var filtered []*entities.User

	for _, user := range users {
		// Status filter
		if filters.Status != "" {
			switch filters.Status {
			case "active":
				if !user.IsActive || user.IsBanned {
					continue
				}
			case "inactive":
				if user.IsActive {
					continue
				}
			case "banned":
				if !user.IsBanned {
					continue
				}
			case "verified":
				if !user.IsVerified {
					continue
				}
			case "unverified":
				if user.IsVerified {
					continue
				}
			case "premium":
				if !user.IsPremium {
					continue
				}
			case "basic":
				if user.IsPremium {
					continue
				}
			}
		}

		// Verified filter
		if filters.Verified != "" {
			if filters.Verified == "true" && !user.IsVerified {
				continue
			}
			if filters.Verified == "false" && user.IsVerified {
				continue
			}
		}

		// Premium filter
		if filters.Premium != "" {
			if filters.Premium == "true" && !user.IsPremium {
				continue
			}
			if filters.Premium == "false" && user.IsPremium {
				continue
			}
		}

		// Search filter
		if filters.Search != "" {
			searchLower := fmt.Sprintf("%s", filters.Search)
			if !containsSearchTerm(user, searchLower) {
				continue
			}
		}

		filtered = append(filtered, user)
	}

	// Apply sorting
	return sortUsers(filtered, filters.SortBy, filters.SortOrder)
}

// containsSearchTerm checks if user matches search term
func containsSearchTerm(user *entities.User, searchTerm string) bool {
	searchLower := searchTerm
	return contains(user.FirstName, searchLower) ||
		contains(user.LastName, searchLower) ||
		contains(user.Email, searchLower)
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && 
		(s[:len(substr)] == substr || 
		 s[len(s)-len(substr):] == substr ||
		 findSubstring(s, substr))
}

// findSubstring finds a substring in a string (case-insensitive)
func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// sortUsers sorts users based on the specified criteria
func sortUsers(users []*entities.User, sortBy, sortOrder string) []*entities.User {
	if len(users) <= 1 {
		return users
	}

	// Create a copy to avoid modifying the original slice
	sorted := make([]*entities.User, len(users))
	copy(sorted, users)

	// Apply sorting based on sortBy field
	switch sortBy {
	case "created_at":
		if sortOrder == "desc" {
			// Sort by created_at descending
			for i := 0; i < len(sorted)-1; i++ {
				for j := i + 1; j < len(sorted); j++ {
					if sorted[i].CreatedAt.Before(sorted[j].CreatedAt) {
						sorted[i], sorted[j] = sorted[j], sorted[i]
					}
				}
			}
		} else {
			// Sort by created_at ascending
			for i := 0; i < len(sorted)-1; i++ {
				for j := i + 1; j < len(sorted); j++ {
					if sorted[i].CreatedAt.After(sorted[j].CreatedAt) {
						sorted[i], sorted[j] = sorted[j], sorted[i]
					}
				}
			}
		}
	case "updated_at":
		if sortOrder == "desc" {
			// Sort by updated_at descending
			for i := 0; i < len(sorted)-1; i++ {
				for j := i + 1; j < len(sorted); j++ {
					if sorted[i].UpdatedAt.Before(sorted[j].UpdatedAt) {
						sorted[i], sorted[j] = sorted[j], sorted[i]
					}
				}
			}
		} else {
			// Sort by updated_at ascending
			for i := 0; i < len(sorted)-1; i++ {
				for j := i + 1; j < len(sorted); j++ {
					if sorted[i].UpdatedAt.After(sorted[j].UpdatedAt) {
						sorted[i], sorted[j] = sorted[j], sorted[i]
					}
				}
			}
		}
	case "last_active":
		if sortOrder == "desc" {
			// Sort by last_active descending
			for i := 0; i < len(sorted)-1; i++ {
				for j := i + 1; j < len(sorted); j++ {
					if sorted[i].LastActive != nil && sorted[j].LastActive != nil {
						if sorted[i].LastActive.Before(*sorted[j].LastActive) {
							sorted[i], sorted[j] = sorted[j], sorted[i]
						}
					} else if sorted[i].LastActive == nil && sorted[j].LastActive != nil {
						// Put nil last_active at the end
						sorted[i], sorted[j] = sorted[j], sorted[i]
					}
				}
			}
		} else {
			// Sort by last_active ascending
			for i := 0; i < len(sorted)-1; i++ {
				for j := i + 1; j < len(sorted); j++ {
					if sorted[i].LastActive != nil && sorted[j].LastActive != nil {
						if sorted[i].LastActive.After(*sorted[j].LastActive) {
							sorted[i], sorted[j] = sorted[j], sorted[i]
						}
					} else if sorted[i].LastActive != nil && sorted[j].LastActive == nil {
						// Put nil last_active at the end
						sorted[i], sorted[j] = sorted[j], sorted[i]
					}
				}
			}
		}
	case "email":
		if sortOrder == "desc" {
			// Sort by email descending
			for i := 0; i < len(sorted)-1; i++ {
				for j := i + 1; j < len(sorted); j++ {
					if sorted[i].Email < sorted[j].Email {
						sorted[i], sorted[j] = sorted[j], sorted[i]
					}
				}
			}
		} else {
			// Sort by email ascending
			for i := 0; i < len(sorted)-1; i++ {
				for j := i + 1; j < len(sorted); j++ {
					if sorted[i].Email > sorted[j].Email {
						sorted[i], sorted[j] = sorted[j], sorted[i]
					}
				}
			}
		}
	case "first_name":
		if sortOrder == "desc" {
			// Sort by first_name descending
			for i := 0; i < len(sorted)-1; i++ {
				for j := i + 1; j < len(sorted); j++ {
					if sorted[i].FirstName < sorted[j].FirstName {
						sorted[i], sorted[j] = sorted[j], sorted[i]
					}
				}
			}
		} else {
			// Sort by first_name ascending
			for i := 0; i < len(sorted)-1; i++ {
				for j := i + 1; j < len(sorted); j++ {
					if sorted[i].FirstName > sorted[j].FirstName {
						sorted[i], sorted[j] = sorted[j], sorted[i]
					}
				}
			}
		}
	default:
		// Default sort by created_at descending
		for i := 0; i < len(sorted)-1; i++ {
			for j := i + 1; j < len(sorted); j++ {
				if sorted[i].CreatedAt.Before(sorted[j].CreatedAt) {
					sorted[i], sorted[j] = sorted[j], sorted[i]
				}
			}
		}
	}

	return sorted
}