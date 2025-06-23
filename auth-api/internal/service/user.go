package service

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/ecoci/auth-api/internal/auth"
	"github.com/ecoci/auth-api/internal/db"
	"github.com/google/uuid"
)

// UserService handles user-related business logic
type UserService struct {
	db *gorm.DB
}

// NewUserService creates a new user service
func NewUserService(database *gorm.DB) *UserService {
	return &UserService{
		db: database,
	}
}

// CreateOrUpdateUserFromGitHub creates or updates a user from GitHub OAuth data
func (s *UserService) CreateOrUpdateUserFromGitHub(githubUser *auth.GitHubUser) (*db.User, error) {
	var user db.User

	// Try to find existing user by GitHub ID
	err := s.db.Where("github_id = ?", githubUser.ID).First(&user).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to query user: %w", err)
	}

	// If user doesn't exist, create new one
	if err == gorm.ErrRecordNotFound {
		user = db.User{
			GitHubID:       githubUser.ID,
			GitHubUsername: githubUser.Login,
			GitHubEmail:    githubUser.Email,
			AvatarURL:      &githubUser.AvatarURL,
			Name:           githubUser.Name,
		}

		if err := s.db.Create(&user).Error; err != nil {
			return nil, fmt.Errorf("failed to create user: %w", err)
		}
	} else {
		// Update existing user with latest info from GitHub
		user.GitHubUsername = githubUser.Login
		user.GitHubEmail = githubUser.Email
		user.AvatarURL = &githubUser.AvatarURL
		user.Name = githubUser.Name

		if err := s.db.Save(&user).Error; err != nil {
			return nil, fmt.Errorf("failed to update user: %w", err)
		}
	}

	return &user, nil
}

// GetUserByID retrieves a user by their UUID
func (s *UserService) GetUserByID(userID uuid.UUID) (*db.User, error) {
	var user db.User
	err := s.db.Where("id = ?", userID).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// GetUserByGitHubID retrieves a user by their GitHub ID
func (s *UserService) GetUserByGitHubID(githubID int64) (*db.User, error) {
	var user db.User
	err := s.db.Where("github_id = ?", githubID).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// GetUserByGitHubUsername retrieves a user by their GitHub username
func (s *UserService) GetUserByGitHubUsername(username string) (*db.User, error) {
	var user db.User
	err := s.db.Where("github_username = ?", username).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// ListUsers retrieves a paginated list of users
func (s *UserService) ListUsers(limit, offset int) ([]db.User, int64, error) {
	var users []db.User
	var total int64

	// Get total count
	if err := s.db.Model(&db.User{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	// Get paginated results
	if err := s.db.Limit(limit).Offset(offset).Order("created_at DESC").Find(&users).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}

	return users, total, nil
}

// DeleteUser deletes a user and all related data
func (s *UserService) DeleteUser(userID uuid.UUID) error {
	// Using transaction to ensure data consistency
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Delete user's runs first (due to foreign key constraints)
		if err := tx.Where("user_id = ?", userID).Delete(&db.Run{}).Error; err != nil {
			return fmt.Errorf("failed to delete user runs: %w", err)
		}

		// Delete user's repositories
		if err := tx.Where("owner_id = ?", userID).Delete(&db.Repository{}).Error; err != nil {
			return fmt.Errorf("failed to delete user repositories: %w", err)
		}

		// Delete user
		if err := tx.Where("id = ?", userID).Delete(&db.User{}).Error; err != nil {
			return fmt.Errorf("failed to delete user: %w", err)
		}

		return nil
	})
}