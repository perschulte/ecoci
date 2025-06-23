package service

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/ecoci/auth-api/internal/auth"
	"github.com/ecoci/auth-api/internal/db"
)

func setupTestDB(t *testing.T) (*gorm.DB, func()) {
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto-migrate tables
	err = database.AutoMigrate(&db.User{}, &db.Repository{}, &db.Run{})
	require.NoError(t, err)

	cleanup := func() {
		sqlDB, _ := database.DB()
		sqlDB.Close()
	}

	return database, cleanup
}

func TestUserService_CreateOrUpdateUserFromGitHub(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	service := NewUserService(database)

	githubUser := &auth.GitHubUser{
		ID:        12345,
		Login:     "testuser",
		Email:     stringPtr("test@example.com"),
		Name:      stringPtr("Test User"),
		AvatarURL: "https://github.com/avatar.jpg",
	}

	t.Run("create new user", func(t *testing.T) {
		user, err := service.CreateOrUpdateUserFromGitHub(githubUser)
		require.NoError(t, err)
		
		assert.NotEqual(t, uuid.Nil, user.ID)
		assert.Equal(t, int64(12345), user.GitHubID)
		assert.Equal(t, "testuser", user.GitHubUsername)
		assert.Equal(t, "test@example.com", *user.GitHubEmail)
		assert.Equal(t, "Test User", *user.Name)
		assert.Equal(t, "https://github.com/avatar.jpg", *user.AvatarURL)
	})

	t.Run("update existing user", func(t *testing.T) {
		// Update GitHub user info
		githubUser.Login = "updateduser"
		githubUser.Email = stringPtr("updated@example.com")
		githubUser.Name = stringPtr("Updated User")

		user, err := service.CreateOrUpdateUserFromGitHub(githubUser)
		require.NoError(t, err)
		
		assert.Equal(t, "updateduser", user.GitHubUsername)
		assert.Equal(t, "updated@example.com", *user.GitHubEmail)
		assert.Equal(t, "Updated User", *user.Name)
		
		// Verify only one user exists in database
		var count int64
		database.Model(&db.User{}).Count(&count)
		assert.Equal(t, int64(1), count)
	})
}

func TestUserService_GetUserByID(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	service := NewUserService(database)
	
	// Create test user
	testUser := &db.User{
		GitHubID:       12345,
		GitHubUsername: "testuser",
		GitHubEmail:    stringPtr("test@example.com"),
	}
	require.NoError(t, database.Create(testUser).Error)

	t.Run("existing user", func(t *testing.T) {
		user, err := service.GetUserByID(testUser.ID)
		require.NoError(t, err)
		
		assert.Equal(t, testUser.ID, user.ID)
		assert.Equal(t, testUser.GitHubUsername, user.GitHubUsername)
		assert.Equal(t, testUser.GitHubEmail, user.GitHubEmail)
	})

	t.Run("non-existing user", func(t *testing.T) {
		nonExistentID := uuid.New()
		_, err := service.GetUserByID(nonExistentID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user not found")
	})
}

func TestUserService_GetUserByGitHubID(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	service := NewUserService(database)
	
	// Create test user
	testUser := &db.User{
		GitHubID:       12345,
		GitHubUsername: "testuser",
		GitHubEmail:    stringPtr("test@example.com"),
	}
	require.NoError(t, database.Create(testUser).Error)

	t.Run("existing user", func(t *testing.T) {
		user, err := service.GetUserByGitHubID(12345)
		require.NoError(t, err)
		
		assert.Equal(t, testUser.ID, user.ID)
		assert.Equal(t, int64(12345), user.GitHubID)
	})

	t.Run("non-existing user", func(t *testing.T) {
		_, err := service.GetUserByGitHubID(99999)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user not found")
	})
}

func TestUserService_GetUserByGitHubUsername(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	service := NewUserService(database)
	
	// Create test user
	testUser := &db.User{
		GitHubID:       12345,
		GitHubUsername: "testuser",
		GitHubEmail:    stringPtr("test@example.com"),
	}
	require.NoError(t, database.Create(testUser).Error)

	t.Run("existing user", func(t *testing.T) {
		user, err := service.GetUserByGitHubUsername("testuser")
		require.NoError(t, err)
		
		assert.Equal(t, testUser.ID, user.ID)
		assert.Equal(t, "testuser", user.GitHubUsername)
	})

	t.Run("non-existing user", func(t *testing.T) {
		_, err := service.GetUserByGitHubUsername("nonexistent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user not found")
	})
}

func TestUserService_ListUsers(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	service := NewUserService(database)
	
	// Create test users
	for i := 0; i < 5; i++ {
		user := &db.User{
			GitHubID:       int64(12345 + i),
			GitHubUsername: fmt.Sprintf("testuser%d", i),
		}
		require.NoError(t, database.Create(user).Error)
	}

	t.Run("list all users", func(t *testing.T) {
		users, total, err := service.ListUsers(10, 0)
		require.NoError(t, err)
		
		assert.Equal(t, int64(5), total)
		assert.Len(t, users, 5)
	})

	t.Run("paginated list", func(t *testing.T) {
		users, total, err := service.ListUsers(2, 0)
		require.NoError(t, err)
		
		assert.Equal(t, int64(5), total)
		assert.Len(t, users, 2)
		
		// Get next page
		users2, total2, err := service.ListUsers(2, 2)
		require.NoError(t, err)
		
		assert.Equal(t, int64(5), total2)
		assert.Len(t, users2, 2)
		
		// Ensure different users
		assert.NotEqual(t, users[0].ID, users2[0].ID)
	})
}

func TestUserService_DeleteUser(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	service := NewUserService(database)
	
	// Create test user
	testUser := &db.User{
		GitHubID:       12345,
		GitHubUsername: "testuser",
	}
	require.NoError(t, database.Create(testUser).Error)

	// Create repository owned by user
	testRepo := &db.Repository{
		OwnerID:      testUser.ID,
		GitHubRepoID: 67890,
		Name:         "testrepo",
		FullName:     "testuser/testrepo",
		HTMLURL:      "https://github.com/testuser/testrepo",
	}
	require.NoError(t, database.Create(testRepo).Error)

	// Create run by user
	testRun := &db.Run{
		UserID:       testUser.ID,
		RepositoryID: testRepo.ID,
		EnergyKWh:    0.5,
		CO2Kg:        0.3,
		DurationS:    120.0,
	}
	require.NoError(t, database.Create(testRun).Error)

	t.Run("delete user with related data", func(t *testing.T) {
		err := service.DeleteUser(testUser.ID)
		require.NoError(t, err)
		
		// Verify user is deleted
		var count int64
		database.Model(&db.User{}).Where("id = ?", testUser.ID).Count(&count)
		assert.Equal(t, int64(0), count)
		
		// Verify related data is deleted
		database.Model(&db.Repository{}).Where("owner_id = ?", testUser.ID).Count(&count)
		assert.Equal(t, int64(0), count)
		
		database.Model(&db.Run{}).Where("user_id = ?", testUser.ID).Count(&count)
		assert.Equal(t, int64(0), count)
	})

	t.Run("delete non-existing user", func(t *testing.T) {
		nonExistentID := uuid.New()
		err := service.DeleteUser(nonExistentID)
		require.NoError(t, err) // GORM delete doesn't error for non-existent records
	})
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}