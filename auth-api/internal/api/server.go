package api

import (
	"fmt"
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"golang.org/x/time/rate"
	"gorm.io/gorm"

	"github.com/ecoci/auth-api/internal/auth"
	"github.com/ecoci/auth-api/internal/config"
	"github.com/ecoci/auth-api/internal/middleware"
	"github.com/ecoci/auth-api/internal/service"
)

// Server represents the API server
type Server struct {
	cfg          *config.Config
	db           *gorm.DB
	router       *gin.Engine
	jwtManager   *auth.JWTManager
	oauthManager *auth.OAuthManager
	userService  *service.UserService
	runService   *service.RunService
	repoService  *service.RepositoryService
}

// NewServer creates a new API server instance
func NewServer(cfg *config.Config, db *gorm.DB) (*Server, error) {
	// Initialize authentication managers
	jwtManager := auth.NewJWTManager(cfg.JWTSecret, cfg.JWTExpiration)
	oauthManager := auth.NewOAuthManager(cfg.GitHubClientID, cfg.GitHubClientSecret, cfg.GitHubRedirectURL)

	// Initialize services
	userService := service.NewUserService(db)
	runService := service.NewRunService(db)
	repoService := service.NewRepositoryService(db)

	// Set Gin mode based on environment
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create router
	router := gin.New()

	server := &Server{
		cfg:          cfg,
		db:           db,
		router:       router,
		jwtManager:   jwtManager,
		oauthManager: oauthManager,
		userService:  userService,
		runService:   runService,
		repoService:  repoService,
	}

	// Setup middleware and routes
	server.setupMiddleware()
	server.setupRoutes()

	return server, nil
}

// setupMiddleware configures middleware for the server
func (s *Server) setupMiddleware() {
	// Recovery and logging middleware
	s.router.Use(gin.Recovery())
	s.router.Use(gin.Logger())

	// CORS middleware
	corsConfig := cors.Config{
		AllowOrigins:     s.cfg.AllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
		AllowCredentials: true,
		MaxAge:           300, // 5 minutes
	}
	s.router.Use(cors.New(corsConfig))

	// Rate limiting middleware
	limiter := rate.NewLimiter(rate.Limit(s.cfg.RateLimitRPS), s.cfg.RateLimitBurst)
	s.router.Use(middleware.RateLimiter(limiter))

	// Security headers middleware
	s.router.Use(middleware.SecurityHeaders())

	// Set trusted proxies
	if err := s.router.SetTrustedProxies(s.cfg.TrustedProxies); err != nil {
		log.Printf("Warning: failed to set trusted proxies: %v", err)
	}
}

// setupRoutes configures API routes
func (s *Server) setupRoutes() {
	// Health check endpoint
	s.router.GET("/health", s.handleHealth)

	// Swagger documentation (only in development)
	if s.cfg.IsDevelopment() {
		s.router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	// Authentication routes
	authGroup := s.router.Group("/auth")
	{
		authGroup.GET("/github", s.handleGitHubAuth)
		authGroup.GET("/github/callback", s.handleGitHubCallback)
		authGroup.POST("/logout", middleware.JWTAuth(s.jwtManager), s.handleLogout)
		authGroup.GET("/me", middleware.JWTAuth(s.jwtManager), s.handleGetMe)
	}

	// API routes (authenticated)
	apiGroup := s.router.Group("/")
	apiGroup.Use(middleware.JWTAuth(s.jwtManager))
	{
		// Runs endpoints
		apiGroup.POST("/runs", s.handleCreateRun)

		// Repositories endpoints
		apiGroup.GET("/repos", s.handleListRepositories)
		apiGroup.GET("/repos/:repo_id/runs", s.handleGetRepositoryRuns)
	}
}

// Start starts the server on the given address
func (s *Server) Start(addr string) error {
	log.Printf("Starting server on %s", addr)
	return s.router.Run(addr)
}

// GetRouter returns the Gin router (useful for testing)
func (s *Server) GetRouter() *gin.Engine {
	return s.router
}