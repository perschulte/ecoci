package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/ecoci/auth-api/internal/auth"
	"github.com/ecoci/auth-api/internal/config"
	"github.com/ecoci/auth-api/internal/db"
	"github.com/ecoci/auth-api/internal/service"
)

func setupTestServer(t *testing.T) (*Server, func()) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)
	
	// Create in-memory database
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto-migrate tables
	err = database.AutoMigrate(&db.User{}, &db.Repository{}, &db.Run{})
	require.NoError(t, err)

	// Create test config
	cfg := &config.Config{
		JWTSecret:      "test-secret",
		JWTExpiration:  time.Hour,
		CookieDomain:   "localhost",
		CookieSecure:   false,
		AllowedOrigins: []string{"http://localhost:3000"},
		RateLimitRPS:   100,
		RateLimitBurst: 200,
		TrustedProxies: []string{"127.0.0.1"},
		Environment:    "test",
	}

	// Create server
	server, err := NewServer(cfg, database)
	require.NoError(t, err)

	cleanup := func() {
		sqlDB, _ := database.DB()
		sqlDB.Close()
	}

	return server, cleanup
}

func createTestUser(t *testing.T, db *gorm.DB) *db.User {
	user := &db.User{
		GitHubID:       12345,
		GitHubUsername: "testuser",
		GitHubEmail:    stringPtr("test@example.com"),
		Name:           stringPtr("Test User"),
	}
	require.NoError(t, db.Create(user).Error)
	return user
}

func createTestRepository(t *testing.T, database *gorm.DB, ownerID uuid.UUID) *db.Repository {
	repo := &db.Repository{
		OwnerID:      ownerID,
		GitHubRepoID: 67890,
		Name:         "testrepo",
		FullName:     "testuser/testrepo",
		HTMLURL:      "https://github.com/testuser/testrepo",
		Description:  stringPtr("Test repository"),
		Private:      false,
	}
	require.NoError(t, database.Create(repo).Error)
	return repo
}

func createTestRun(t *testing.T, database *gorm.DB, userID, repoID uuid.UUID) *db.Run {
	run := &db.Run{
		UserID:       userID,
		RepositoryID: repoID,
		EnergyKWh:    0.5,
		CO2Kg:        0.3,
		DurationS:    120.0,
	}
	require.NoError(t, database.Create(run).Error)
	return run
}

func generateTestJWT(t *testing.T, server *Server, userID uuid.UUID, username string) string {
	token, err := server.jwtManager.GenerateToken(userID, username)
	require.NoError(t, err)
	return token
}

func TestHandleHealth(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	server.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, "healthy", response["status"])
	assert.Contains(t, response, "timestamp")
	assert.Equal(t, "1.0.0", response["version"])
}

func TestHandleGetMe(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	// Get database connection
	database := server.db
	
	t.Run("authenticated user", func(t *testing.T) {
		// Create test user
		user := createTestUser(t, database)
		
		// Generate JWT token
		token := generateTestJWT(t, server, user.ID, user.GitHubUsername)
		
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/auth/me", nil)
		req.AddCookie(&http.Cookie{
			Name:  "ecoci_token",
			Value: token,
		})
		
		server.router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response db.User
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.Equal(t, user.ID, response.ID)
		assert.Equal(t, user.GitHubUsername, response.GitHubUsername)
	})

	t.Run("unauthenticated user", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/auth/me", nil)
		
		server.router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestHandleCreateRun(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	database := server.db
	user := createTestUser(t, database)
	token := generateTestJWT(t, server, user.ID, user.GitHubUsername)

	t.Run("valid run creation", func(t *testing.T) {
		runData := service.RunCreateRequest{
			EnergyKWh: 0.5,
			CO2Kg:     0.3,
			DurationS: 120.0,
			Repository: service.RepositoryCreateRequest{
				Name:     "testrepo",
				FullName: "testuser/testrepo",
				HTMLURL:  "https://github.com/testuser/testrepo",
				Private:  false,
			},
			Metadata: map[string]interface{}{
				"cpu_cores": 4,
				"memory_gb": 8,
			},
		}
		
		jsonData, _ := json.Marshal(runData)
		
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/runs", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{
			Name:  "ecoci_token",
			Value: token,
		})
		
		server.router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusCreated, w.Code)
		
		var response db.Run
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.Equal(t, user.ID, response.UserID)
		assert.Equal(t, 0.5, response.EnergyKWh)
		assert.Equal(t, 0.3, response.CO2Kg)
		assert.Equal(t, 120.0, response.DurationS)
		assert.NotNil(t, response.RunMetadata)
	})

	t.Run("invalid run data", func(t *testing.T) {
		runData := service.RunCreateRequest{
			EnergyKWh: -0.5, // Invalid negative value
			CO2Kg:     0.3,
			DurationS: 120.0,
			Repository: service.RepositoryCreateRequest{
				Name:     "testrepo",
				FullName: "testuser/testrepo",
				HTMLURL:  "https://github.com/testuser/testrepo",
			},
		}
		
		jsonData, _ := json.Marshal(runData)
		
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/runs", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{
			Name:  "ecoci_token",
			Value: token,
		})
		
		server.router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	})

	t.Run("unauthenticated request", func(t *testing.T) {
		runData := service.RunCreateRequest{
			EnergyKWh: 0.5,
			CO2Kg:     0.3,
			DurationS: 120.0,
			Repository: service.RepositoryCreateRequest{
				Name:     "testrepo",
				FullName: "testuser/testrepo",
				HTMLURL:  "https://github.com/testuser/testrepo",
			},
		}
		
		jsonData, _ := json.Marshal(runData)
		
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/runs", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		
		server.router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestHandleListRepositories(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	database := server.db
	user := createTestUser(t, database)
	token := generateTestJWT(t, server, user.ID, user.GitHubUsername)
	
	// Create test repository and runs
	repo := createTestRepository(t, database, user.ID)
	createTestRun(t, database, user.ID, repo.ID)
	createTestRun(t, database, user.ID, repo.ID)

	t.Run("list repositories", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/repos", nil)
		req.AddCookie(&http.Cookie{
			Name:  "ecoci_token",
			Value: token,
		})
		
		server.router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.Contains(t, response, "repositories")
		assert.Contains(t, response, "pagination")
		
		repos := response["repositories"].([]interface{})
		assert.Len(t, repos, 1)
		
		pagination := response["pagination"].(map[string]interface{})
		assert.Equal(t, float64(1), pagination["page"])
		assert.Equal(t, float64(1), pagination["total"])
	})

	t.Run("list repositories with pagination", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/repos?page=1&limit=10", nil)
		req.AddCookie(&http.Cookie{
			Name:  "ecoci_token",
			Value: token,
		})
		
		server.router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		pagination := response["pagination"].(map[string]interface{})
		assert.Equal(t, float64(1), pagination["page"])
		assert.Equal(t, float64(10), pagination["limit"])
	})

	t.Run("unauthenticated request", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/repos", nil)
		
		server.router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestHandleGetRepositoryRuns(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	database := server.db
	user := createTestUser(t, database)
	token := generateTestJWT(t, server, user.ID, user.GitHubUsername)
	
	// Create test repository and runs
	repo := createTestRepository(t, database, user.ID)
	createTestRun(t, database, user.ID, repo.ID)
	createTestRun(t, database, user.ID, repo.ID)

	t.Run("get repository runs", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/repos/"+repo.ID.String()+"/runs", nil)
		req.AddCookie(&http.Cookie{
			Name:  "ecoci_token",
			Value: token,
		})
		
		server.router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.Contains(t, response, "runs")
		assert.Contains(t, response, "pagination")
		
		runs := response["runs"].([]interface{})
		assert.Len(t, runs, 2)
		
		pagination := response["pagination"].(map[string]interface{})
		assert.Equal(t, float64(2), pagination["total"])
	})

	t.Run("invalid repository ID", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/repos/invalid-uuid/runs", nil)
		req.AddCookie(&http.Cookie{
			Name:  "ecoci_token",
			Value: token,
		})
		
		server.router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("non-existent repository", func(t *testing.T) {
		nonExistentID := uuid.New()
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/repos/"+nonExistentID.String()+"/runs", nil)
		req.AddCookie(&http.Cookie{
			Name:  "ecoci_token",
			Value: token,
		})
		
		server.router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("unauthenticated request", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/repos/"+repo.ID.String()+"/runs", nil)
		
		server.router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestHandleLogout(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	database := server.db
	user := createTestUser(t, database)
	token := generateTestJWT(t, server, user.ID, user.GitHubUsername)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/auth/logout", nil)
	req.AddCookie(&http.Cookie{
		Name:  "ecoci_token",
		Value: token,
	})
	
	server.router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, "Successfully logged out", response["message"])
	
	// Check that cookie is cleared
	cookies := w.Result().Cookies()
	var tokenCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == "ecoci_token" {
			tokenCookie = cookie
			break
		}
	}
	
	require.NotNil(t, tokenCookie)
	assert.Equal(t, "", tokenCookie.Value)
	assert.Equal(t, -1, tokenCookie.MaxAge)
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}