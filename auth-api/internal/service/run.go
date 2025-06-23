package service

import (
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/ecoci/auth-api/internal/db"
)

// RunService handles run-related business logic
type RunService struct {
	db *gorm.DB
}

// NewRunService creates a new run service
func NewRunService(database *gorm.DB) *RunService {
	return &RunService{
		db: database,
	}
}

// RunCreateRequest represents the data needed to create a run
type RunCreateRequest struct {
	EnergyKWh     float64                `json:"energy_kwh" validate:"required,min=0"`
	CO2Kg         float64                `json:"co2_kg" validate:"required,min=0"`
	DurationS     float64                `json:"duration_s" validate:"required,min=0"`
	GitCommitSHA  *string                `json:"git_commit_sha,omitempty" validate:"omitempty,len=40"`
	BranchName    *string                `json:"branch_name,omitempty"`
	WorkflowName  *string                `json:"workflow_name,omitempty"`
	Repository    RepositoryCreateRequest `json:"repository" validate:"required"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// CreateRun creates a new CO2 measurement run
func (s *RunService) CreateRun(userID uuid.UUID, req *RunCreateRequest, repoService *RepositoryService) (*db.Run, error) {
	return s.db.Transaction(func(tx *gorm.DB) (*db.Run, error) {
		// Create or update repository first
		repo, err := repoService.CreateOrUpdateRepository(userID, &req.Repository)
		if err != nil {
			return nil, fmt.Errorf("failed to create/update repository: %w", err)
		}

		// Convert metadata to JSONB
		var metadata db.JSONB
		if req.Metadata != nil {
			metadata = db.JSONB(req.Metadata)
		}

		// Create the run
		run := db.Run{
			UserID:       userID,
			RepositoryID: repo.ID,
			EnergyKWh:    req.EnergyKWh,
			CO2Kg:        req.CO2Kg,
			DurationS:    req.DurationS,
			RunMetadata:  metadata,
			GitCommitSHA: req.GitCommitSHA,
			BranchName:   req.BranchName,
			WorkflowName: req.WorkflowName,
		}

		if err := s.db.Create(&run).Error; err != nil {
			return nil, fmt.Errorf("failed to create run: %w", err)
		}

		// Load relationships for response
		if err := s.db.Preload("User").Preload("Repository").First(&run, run.ID).Error; err != nil {
			return nil, fmt.Errorf("failed to load run relationships: %w", err)
		}

		return &run, nil
	})
}

// GetRunByID retrieves a run by ID
func (s *RunService) GetRunByID(runID uuid.UUID) (*db.Run, error) {
	var run db.Run
	err := s.db.Preload("User").Preload("Repository").Where("id = ?", runID).First(&run).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("run not found")
		}
		return nil, fmt.Errorf("failed to get run: %w", err)
	}

	return &run, nil
}

// ListUserRuns retrieves runs for a specific user
func (s *RunService) ListUserRuns(userID uuid.UUID, limit, offset int, filters map[string]interface{}) ([]db.Run, int64, error) {
	query := s.db.Where("user_id = ?", userID)

	// Apply filters
	if repoID, ok := filters["repository_id"]; ok {
		query = query.Where("repository_id = ?", repoID)
	}
	if fromDate, ok := filters["from_date"]; ok {
		query = query.Where("created_at >= ?", fromDate)
	}
	if toDate, ok := filters["to_date"]; ok {
		query = query.Where("created_at <= ?", toDate)
	}

	// Count total
	var total int64
	if err := query.Model(&db.Run{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count runs: %w", err)
	}

	// Get paginated results
	var runs []db.Run
	if err := query.Preload("User").Preload("Repository").
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&runs).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list runs: %w", err)
	}

	return runs, total, nil
}

// GetUserStats retrieves aggregated CO2 statistics for a user
func (s *RunService) GetUserStats(userID uuid.UUID) (*UserStats, error) {
	var stats UserStats

	row := s.db.Table("runs").
		Select(`
			COALESCE(SUM(co2_kg), 0) as total_co2_kg,
			COALESCE(AVG(co2_kg), 0) as avg_co2_kg,
			COALESCE(SUM(energy_kwh), 0) as total_energy_kwh,
			COALESCE(AVG(energy_kwh), 0) as avg_energy_kwh,
			COALESCE(COUNT(id), 0) as run_count,
			COALESCE(COUNT(DISTINCT repository_id), 0) as repository_count,
			COALESCE(MAX(created_at), NOW()) as last_run_at
		`).
		Where("user_id = ?", userID).
		Row()

	err := row.Scan(
		&stats.TotalCO2Kg,
		&stats.AvgCO2Kg,
		&stats.TotalEnergyKWh,
		&stats.AvgEnergyKWh,
		&stats.RunCount,
		&stats.RepositoryCount,
		&stats.LastRunAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get user stats: %w", err)
	}

	return &stats, nil
}

// DeleteRun deletes a run
func (s *RunService) DeleteRun(runID uuid.UUID, userID uuid.UUID) error {
	result := s.db.Where("id = ? AND user_id = ?", runID, userID).Delete(&db.Run{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete run: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("run not found or not owned by user")
	}
	return nil
}

// GetRunsByRepository retrieves runs for a specific repository
func (s *RunService) GetRunsByRepository(repoID uuid.UUID, limit, offset int) ([]db.Run, int64, error) {
	var runs []db.Run
	var total int64

	// Count total
	if err := s.db.Model(&db.Run{}).Where("repository_id = ?", repoID).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count runs: %w", err)
	}

	// Get paginated results
	if err := s.db.Where("repository_id = ?", repoID).
		Preload("User").Preload("Repository").
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&runs).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get runs: %w", err)
	}

	return runs, total, nil
}

// UserStats represents aggregated statistics for a user
type UserStats struct {
	TotalCO2Kg      float64 `json:"total_co2_kg"`
	AvgCO2Kg        float64 `json:"avg_co2_kg"`
	TotalEnergyKWh  float64 `json:"total_energy_kwh"`
	AvgEnergyKWh    float64 `json:"avg_energy_kwh"`
	RunCount        int64   `json:"run_count"`
	RepositoryCount int64   `json:"repository_count"`
	LastRunAt       string  `json:"last_run_at"`
}