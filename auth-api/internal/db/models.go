package db

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User represents a GitHub OAuth authenticated user
type User struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	GitHubID        int64     `gorm:"uniqueIndex;not null" json:"github_id"`
	GitHubUsername  string    `gorm:"index;not null" json:"github_username"`
	GitHubEmail     *string   `json:"github_email"`
	AvatarURL       *string   `json:"avatar_url"`
	Name            *string   `json:"name"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`

	// Relationships
	Repositories []Repository `gorm:"foreignKey:OwnerID" json:"repositories,omitempty"`
	Runs         []Run        `gorm:"foreignKey:UserID" json:"runs,omitempty"`
}

// Repository represents a GitHub repository
type Repository struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OwnerID      uuid.UUID `gorm:"type:uuid;not null;index" json:"owner_id"`
	GitHubRepoID int64     `gorm:"uniqueIndex;not null" json:"github_repo_id"`
	Name         string    `gorm:"not null" json:"name"`
	FullName     string    `gorm:"index;not null" json:"full_name"`
	Description  *string   `json:"description"`
	Private      bool      `gorm:"not null;default:false" json:"private"`
	HTMLURL      string    `gorm:"not null" json:"html_url"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// Relationships
	Owner *User `gorm:"foreignKey:OwnerID" json:"owner,omitempty"`
	Runs  []Run `gorm:"foreignKey:RepositoryID" json:"runs,omitempty"`
}

// Run represents a CO2 measurement run
type Run struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID       uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	RepositoryID uuid.UUID `gorm:"type:uuid;not null;index" json:"repository_id"`

	// CO2 measurement data
	EnergyKWh  float64 `gorm:"type:decimal(12,6);not null;check:energy_kwh >= 0" json:"energy_kwh"`
	CO2Kg      float64 `gorm:"type:decimal(12,6);not null;check:co2_kg >= 0" json:"co2_kg"`
	DurationS  float64 `gorm:"type:decimal(10,3);not null;check:duration_s >= 0" json:"duration_s"`

	// Additional metadata
	RunMetadata   JSONB   `gorm:"type:jsonb" json:"run_metadata,omitempty"`
	GitCommitSHA  *string `gorm:"size:40" json:"git_commit_sha,omitempty"`
	BranchName    *string `json:"branch_name,omitempty"`
	WorkflowName  *string `json:"workflow_name,omitempty"`

	CreatedAt time.Time `gorm:"index:idx_runs_created_at" json:"created_at"`

	// Relationships
	User       *User       `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Repository *Repository `gorm:"foreignKey:RepositoryID" json:"repository,omitempty"`
}

// JSONB represents a JSONB field for PostgreSQL
type JSONB map[string]interface{}

// Value implements the driver.Valuer interface for JSONB
func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan implements the sql.Scanner interface for JSONB
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal JSONB value: %v", value)
	}

	return json.Unmarshal(bytes, j)
}

// RepositoryStats represents aggregated statistics for a repository
type RepositoryStats struct {
	Repository
	Stats struct {
		TotalCO2Kg      float64   `json:"total_co2_kg"`
		AvgCO2Kg        float64   `json:"avg_co2_kg"`
		TotalEnergyKWh  float64   `json:"total_energy_kwh"`
		AvgEnergyKWh    float64   `json:"avg_energy_kwh"`
		RunCount        int64     `json:"run_count"`
		LastRunAt       time.Time `json:"last_run_at"`
	} `json:"stats"`
}

// BeforeCreate sets the ID if not already set for User
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

// BeforeCreate sets the ID if not already set for Repository
func (r *Repository) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}

// BeforeCreate sets the ID if not already set for Run
func (r *Run) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}

// TableName returns the table name for User
func (User) TableName() string {
	return "users"
}

// TableName returns the table name for Repository
func (Repository) TableName() string {
	return "repositories"
}

// TableName returns the table name for Run
func (Run) TableName() string {
	return "runs"
}