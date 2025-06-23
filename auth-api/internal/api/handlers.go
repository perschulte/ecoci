package api

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/ecoci/auth-api/internal/service"
)

// Health check handler
// @Summary Health check
// @Description Get the health status of the API
// @Tags health
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /health [get]
func (s *Server) handleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"version":   "1.0.0",
	})
}

// GitHub OAuth initiation handler
// @Summary Initiate GitHub OAuth
// @Description Redirect to GitHub OAuth authorization
// @Tags auth
// @Param redirect_uri query string false "Redirect URI after auth"
// @Success 302 "Redirect to GitHub"
// @Failure 400 {object} map[string]interface{}
// @Router /auth/github [get]
func (s *Server) handleGitHubAuth(c *gin.Context) {
	// Generate state parameter for CSRF protection
	state := uuid.New().String()
	
	// Store state in session (simplified - in production use secure session store)
	c.SetCookie("oauth_state", state, 300, "/", s.cfg.CookieDomain, s.cfg.CookieSecure, true)
	
	// Store redirect URI if provided
	if redirectURI := c.Query("redirect_uri"); redirectURI != "" {
		c.SetCookie("redirect_after_auth", redirectURI, 300, "/", s.cfg.CookieDomain, s.cfg.CookieSecure, true)
	}

	// Redirect to GitHub OAuth
	authURL := s.oauthManager.GetAuthURL(state)
	c.Redirect(http.StatusFound, authURL)
}

// GitHub OAuth callback handler
// @Summary GitHub OAuth callback
// @Description Handle GitHub OAuth callback and create session
// @Tags auth
// @Param code query string true "Authorization code"
// @Param state query string false "State parameter"
// @Success 302 "Redirect to application"
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /auth/github/callback [get]
func (s *Server) handleGitHubCallback(c *gin.Context) {
	// Verify state parameter
	state := c.Query("state")
	storedState, err := c.Cookie("oauth_state")
	if err != nil || state != storedState {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":     "Invalid state parameter",
			"code":      "INVALID_STATE",
			"timestamp": time.Now().UTC(),
		})
		return
	}

	// Clear state cookie
	c.SetCookie("oauth_state", "", -1, "/", s.cfg.CookieDomain, s.cfg.CookieSecure, true)

	// Get authorization code
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":     "Missing authorization code",
			"code":      "MISSING_CODE",
			"timestamp": time.Now().UTC(),
		})
		return
	}

	// Exchange code for token
	token, err := s.oauthManager.ExchangeCodeForToken(c.Request.Context(), code)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":     "Failed to exchange code for token",
			"code":      "TOKEN_EXCHANGE_FAILED",
			"timestamp": time.Now().UTC(),
		})
		return
	}

	// Get user info from GitHub
	githubUser, err := s.oauthManager.GetUserInfo(c.Request.Context(), token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":     "Failed to get user info from GitHub",
			"code":      "USER_INFO_FAILED",
			"timestamp": time.Now().UTC(),
		})
		return
	}

	// Create or update user in database
	user, err := s.userService.CreateOrUpdateUserFromGitHub(githubUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":     "Failed to create user",
			"code":      "USER_CREATION_FAILED",
			"timestamp": time.Now().UTC(),
		})
		return
	}

	// Generate JWT token
	jwtToken, err := s.jwtManager.GenerateToken(user.ID, user.GitHubUsername)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":     "Failed to generate auth token",
			"code":      "TOKEN_GENERATION_FAILED",
			"timestamp": time.Now().UTC(),
		})
		return
	}

	// Set JWT cookie
	maxAge := int(s.cfg.JWTExpiration.Seconds())
	c.SetCookie("ecoci_token", jwtToken, maxAge, "/", s.cfg.CookieDomain, s.cfg.CookieSecure, true)

	// Get redirect URI and clear cookie
	redirectURI := "/"
	if storedRedirect, err := c.Cookie("redirect_after_auth"); err == nil {
		redirectURI = storedRedirect
		c.SetCookie("redirect_after_auth", "", -1, "/", s.cfg.CookieDomain, s.cfg.CookieSecure, true)
	}

	c.Redirect(http.StatusFound, redirectURI)
}

// Logout handler
// @Summary Logout user
// @Description Clear authentication session
// @Tags auth
// @Security CookieAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /auth/logout [post]
func (s *Server) handleLogout(c *gin.Context) {
	// Clear JWT cookie
	c.SetCookie("ecoci_token", "", -1, "/", s.cfg.CookieDomain, s.cfg.CookieSecure, true)
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Successfully logged out",
	})
}

// Get current user handler
// @Summary Get current user
// @Description Get information about the authenticated user
// @Tags auth
// @Security CookieAuth
// @Produce json
// @Success 200 {object} db.User
// @Failure 401 {object} map[string]interface{}
// @Router /auth/me [get]
func (s *Server) handleGetMe(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":     "User ID not found in context",
			"code":      "MISSING_USER_ID",
			"timestamp": time.Now().UTC(),
		})
		return
	}

	user, err := s.userService.GetUserByID(userID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":     "Failed to get user information",
			"code":      "USER_FETCH_FAILED",
			"timestamp": time.Now().UTC(),
		})
		return
	}

	c.JSON(http.StatusOK, user)
}

// Create run handler
// @Summary Create CO2 measurement run
// @Description Store a new CO2 measurement run
// @Tags runs
// @Security CookieAuth
// @Accept json
// @Produce json
// @Param run body service.RunCreateRequest true "Run data"
// @Success 201 {object} db.Run
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 422 {object} map[string]interface{}
// @Router /runs [post]
func (s *Server) handleCreateRun(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":     "User ID not found in context",
			"code":      "MISSING_USER_ID",
			"timestamp": time.Now().UTC(),
		})
		return
	}

	var req service.RunCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":     "Invalid request body",
			"code":      "INVALID_REQUEST_BODY",
			"timestamp": time.Now().UTC(),
			"details":   err.Error(),
		})
		return
	}

	// Validate required fields
	if req.EnergyKWh < 0 || req.CO2Kg < 0 || req.DurationS < 0 {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error":     "Energy, CO2, and duration values must be non-negative",
			"code":      "VALIDATION_FAILED",
			"timestamp": time.Now().UTC(),
		})
		return
	}

	// Create the run
	run, err := s.runService.CreateRun(userID.(uuid.UUID), &req, s.repoService)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":     "Failed to create run",
			"code":      "RUN_CREATION_FAILED",
			"timestamp": time.Now().UTC(),
			"details":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, run)
}

// List repositories handler
// @Summary List repositories with CO2 statistics
// @Description Get paginated list of repositories with aggregated CO2 data
// @Tags repositories
// @Security CookieAuth
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param sort query string false "Sort field" Enums(total_co2,avg_co2,run_count,last_run) default(total_co2)
// @Param order query string false "Sort order" Enums(asc,desc) default(desc)
// @Param owner query string false "Filter by owner username"
// @Param name query string false "Filter by repository name"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /repos [get]
func (s *Server) handleListRepositories(c *gin.Context) {
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	// Parse sorting parameters
	sortBy := c.DefaultQuery("sort", "total_co2")
	order := c.DefaultQuery("order", "desc")
	if order != "asc" && order != "desc" {
		order = "desc"
	}

	// Parse filters
	filters := make(map[string]interface{})
	if owner := c.Query("owner"); owner != "" {
		filters["owner"] = owner
	}
	if name := c.Query("name"); name != "" {
		filters["name"] = name
	}

	// Get repositories with stats
	repos, total, err := s.repoService.ListRepositoriesWithStats(limit, offset, sortBy, order, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":     "Failed to list repositories",
			"code":      "REPOSITORIES_FETCH_FAILED",
			"timestamp": time.Now().UTC(),
		})
		return
	}

	// Calculate pagination info
	totalPages := (total + int64(limit) - 1) / int64(limit)
	
	c.JSON(http.StatusOK, gin.H{
		"repositories": repos,
		"pagination": gin.H{
			"page":     page,
			"limit":    limit,
			"total":    total,
			"pages":    totalPages,
			"has_next": int64(page) < totalPages,
			"has_prev": page > 1,
		},
	})
}

// Get repository runs handler
// @Summary Get runs for a repository
// @Description Get paginated list of runs for a specific repository
// @Tags repositories
// @Security CookieAuth
// @Produce json
// @Param repo_id path string true "Repository UUID"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param from_date query string false "Filter from date (ISO 8601)"
// @Param to_date query string false "Filter to date (ISO 8601)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /repos/{repo_id}/runs [get]
func (s *Server) handleGetRepositoryRuns(c *gin.Context) {
	// Parse repository ID
	repoIDStr := c.Param("repo_id")
	repoID, err := uuid.Parse(repoIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":     "Invalid repository ID",
			"code":      "INVALID_REPO_ID",
			"timestamp": time.Now().UTC(),
		})
		return
	}

	// Check if repository exists
	_, err = s.repoService.GetRepositoryByID(repoID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":     "Repository not found",
			"code":      "REPOSITORY_NOT_FOUND",
			"timestamp": time.Now().UTC(),
		})
		return
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	// Parse date filters
	filters := make(map[string]interface{})
	if fromDate := c.Query("from_date"); fromDate != "" {
		if parsedDate, err := time.Parse(time.RFC3339, fromDate); err == nil {
			filters["from_date"] = parsedDate
		}
	}
	if toDate := c.Query("to_date"); toDate != "" {
		if parsedDate, err := time.Parse(time.RFC3339, toDate); err == nil {
			filters["to_date"] = parsedDate
		}
	}

	// Get runs
	runs, total, err := s.repoService.GetRepositoryRuns(repoID, limit, offset, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":     "Failed to get repository runs",
			"code":      "RUNS_FETCH_FAILED",
			"timestamp": time.Now().UTC(),
		})
		return
	}

	// Calculate pagination info
	totalPages := (total + int64(limit) - 1) / int64(limit)
	
	c.JSON(http.StatusOK, gin.H{
		"runs": runs,
		"pagination": gin.H{
			"page":     page,
			"limit":    limit,
			"total":    total,
			"pages":    totalPages,
			"has_next": int64(page) < totalPages,
			"has_prev": page > 1,
		},
	})
}