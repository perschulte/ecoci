package service

import (
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/ecoci/auth-api/internal/db"
)

// RepositoryService handles repository-related business logic
type RepositoryService struct {
	db *gorm.DB
}

// NewRepositoryService creates a new repository service
func NewRepositoryService(database *gorm.DB) *RepositoryService {
	return &RepositoryService{
		db: database,
	}
}

// RepositoryCreateRequest represents the data needed to create/update a repository
type RepositoryCreateRequest struct {
	Name        string  `json:"name"`
	FullName    string  `json:"full_name"`
	Description *string `json:"description"`
	Private     bool    `json:"private"`
	HTMLURL     string  `json:"html_url"`
}

// CreateOrUpdateRepository creates or updates a repository
func (s *RepositoryService) CreateOrUpdateRepository(ownerID uuid.UUID, req *RepositoryCreateRequest) (*db.Repository, error) {
	var repo db.Repository

	// Try to find existing repository by full name and owner
	err := s.db.Where("full_name = ? AND owner_id = ?", req.FullName, ownerID).First(&repo).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to query repository: %w", err)
	}

	// If repository doesn't exist, create new one
	if err == gorm.ErrRecordNotFound {
		repo = db.Repository{
			OwnerID:     ownerID,
			Name:        req.Name,
			FullName:    req.FullName,
			Description: req.Description,
			Private:     req.Private,
			HTMLURL:     req.HTMLURL,
		}

		if err := s.db.Create(&repo).Error; err != nil {
			return nil, fmt.Errorf("failed to create repository: %w", err)
		}
	} else {
		// Update existing repository
		repo.Name = req.Name
		repo.Description = req.Description
		repo.Private = req.Private
		repo.HTMLURL = req.HTMLURL

		if err := s.db.Save(&repo).Error; err != nil {
			return nil, fmt.Errorf("failed to update repository: %w", err)
		}
	}

	return &repo, nil
}

// GetRepositoryByID retrieves a repository by ID
func (s *RepositoryService) GetRepositoryByID(repoID uuid.UUID) (*db.Repository, error) {
	var repo db.Repository
	err := s.db.Preload("Owner").Where("id = ?", repoID).First(&repo).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("repository not found")
		}
		return nil, fmt.Errorf("failed to get repository: %w", err)
	}

	return &repo, nil
}

// ListRepositoriesWithStats retrieves repositories with CO2 statistics
func (s *RepositoryService) ListRepositoriesWithStats(limit, offset int, sortBy, order string, filters map[string]interface{}) ([]db.RepositoryStats, int64, error) {
	// Build base query with joins and aggregations
	query := s.db.Table("repositories r").
		Select(`
			r.id, r.owner_id, r.github_repo_id, r.name, r.full_name, r.description, 
			r.private, r.html_url, r.created_at, r.updated_at,
			u.id as "owner.id", u.github_username as "owner.github_username", 
			u.github_email as "owner.github_email", u.avatar_url as "owner.avatar_url",
			u.name as "owner.name", u.created_at as "owner.created_at",
			COALESCE(SUM(runs.co2_kg), 0) as total_co2_kg,
			COALESCE(AVG(runs.co2_kg), 0) as avg_co2_kg,
			COALESCE(SUM(runs.energy_kwh), 0) as total_energy_kwh,
			COALESCE(AVG(runs.energy_kwh), 0) as avg_energy_kwh,
			COALESCE(COUNT(runs.id), 0) as run_count,
			COALESCE(MAX(runs.created_at), r.created_at) as last_run_at
		`).
		Joins("LEFT JOIN users u ON r.owner_id = u.id").
		Joins("LEFT JOIN runs ON r.id = runs.repository_id").
		Group("r.id, u.id").
		Having("COUNT(runs.id) > 0") // Only include repos with runs

	// Apply filters
	if owner, ok := filters["owner"]; ok {
		query = query.Where("u.github_username = ?", owner)
	}
	if name, ok := filters["name"]; ok {
		query = query.Where("r.name ILIKE ?", "%"+name.(string)+"%")
	}

	// Count total results
	var total int64
	countQuery := s.db.Table("(?) as counted", query).Count(&total)
	if countQuery.Error != nil {
		return nil, 0, fmt.Errorf("failed to count repositories: %w", countQuery.Error)
	}

	// Apply sorting
	switch sortBy {
	case "total_co2":
		query = query.Order("total_co2_kg " + order)
	case "avg_co2":
		query = query.Order("avg_co2_kg " + order)
	case "run_count":
		query = query.Order("run_count " + order)
	case "last_run":
		query = query.Order("last_run_at " + order)
	default:
		query = query.Order("total_co2_kg DESC")
	}

	// Apply pagination
	query = query.Limit(limit).Offset(offset)

	// Execute query
	rows, err := query.Rows()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to execute repository stats query: %w", err)
	}
	defer rows.Close()

	var results []db.RepositoryStats
	for rows.Next() {
		var stat db.RepositoryStats
		var owner db.User

		err := rows.Scan(
			&stat.ID, &stat.OwnerID, &stat.GitHubRepoID, &stat.Name, &stat.FullName,
			&stat.Description, &stat.Private, &stat.HTMLURL, &stat.CreatedAt, &stat.UpdatedAt,
			&owner.ID, &owner.GitHubUsername, &owner.GitHubEmail, &owner.AvatarURL,
			&owner.Name, &owner.CreatedAt,
			&stat.Stats.TotalCO2Kg, &stat.Stats.AvgCO2Kg,
			&stat.Stats.TotalEnergyKWh, &stat.Stats.AvgEnergyKWh,
			&stat.Stats.RunCount, &stat.Stats.LastRunAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan repository stats: %w", err)
		}

		stat.Owner = &owner
		results = append(results, stat)
	}

	return results, total, nil
}

// GetRepositoryRuns retrieves runs for a specific repository
func (s *RepositoryService) GetRepositoryRuns(repoID uuid.UUID, limit, offset int, filters map[string]interface{}) ([]db.Run, int64, error) {
	query := s.db.Where("repository_id = ?", repoID)

	// Apply date filters
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
		return nil, 0, fmt.Errorf("failed to get repository runs: %w", err)
	}

	return runs, total, nil
}

// GetRepositoryStats retrieves aggregated statistics for a repository
func (s *RepositoryService) GetRepositoryStats(repoID uuid.UUID) (*db.RepositoryStats, error) {
	var stat db.RepositoryStats

	// Get repository info
	if err := s.db.Preload("Owner").Where("id = ?", repoID).First(&stat.Repository).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("repository not found")
		}
		return nil, fmt.Errorf("failed to get repository: %w", err)
	}

	// Get aggregated stats
	row := s.db.Table("runs").
		Select(`
			COALESCE(SUM(co2_kg), 0) as total_co2_kg,
			COALESCE(AVG(co2_kg), 0) as avg_co2_kg,
			COALESCE(SUM(energy_kwh), 0) as total_energy_kwh,
			COALESCE(AVG(energy_kwh), 0) as avg_energy_kwh,
			COALESCE(COUNT(id), 0) as run_count,
			COALESCE(MAX(created_at), NOW()) as last_run_at
		`).
		Where("repository_id = ?", repoID).
		Row()

	err := row.Scan(
		&stat.Stats.TotalCO2Kg,
		&stat.Stats.AvgCO2Kg,
		&stat.Stats.TotalEnergyKWh,
		&stat.Stats.AvgEnergyKWh,
		&stat.Stats.RunCount,
		&stat.Stats.LastRunAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository stats: %w", err)
	}

	return &stat, nil
}

// DeleteRepository deletes a repository and all related runs
func (s *RepositoryService) DeleteRepository(repoID uuid.UUID) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Delete all runs for this repository
		if err := tx.Where("repository_id = ?", repoID).Delete(&db.Run{}).Error; err != nil {
			return fmt.Errorf("failed to delete repository runs: %w", err)
		}

		// Delete the repository
		if err := tx.Where("id = ?", repoID).Delete(&db.Repository{}).Error; err != nil {
			return fmt.Errorf("failed to delete repository: %w", err)
		}

		return nil
	})
}