-- Migration: Initial schema for EcoCI Auth API
-- Creates tables for users, repositories, and CO2 measurement runs

-- Users table for GitHub OAuth authentication
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    github_id BIGINT NOT NULL UNIQUE,
    github_username VARCHAR(255) NOT NULL,
    github_email VARCHAR(255),
    avatar_url TEXT,
    name VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Index for fast GitHub ID lookups during authentication
CREATE INDEX idx_users_github_id ON users(github_id);
CREATE INDEX idx_users_github_username ON users(github_username);

-- Repositories table to track GitHub repositories
CREATE TABLE repositories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    github_repo_id BIGINT NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    full_name VARCHAR(255) NOT NULL, -- owner/repo format
    description TEXT,
    private BOOLEAN NOT NULL DEFAULT false,
    html_url TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Indexes for repository lookups
CREATE INDEX idx_repositories_owner_id ON repositories(owner_id);
CREATE INDEX idx_repositories_github_repo_id ON repositories(github_repo_id);
CREATE INDEX idx_repositories_full_name ON repositories(full_name);

-- CO2 measurement runs table
CREATE TABLE runs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    repository_id UUID NOT NULL REFERENCES repositories(id) ON DELETE CASCADE,
    
    -- CO2 measurement data (from CLI schema)
    energy_kwh DECIMAL(12, 6) NOT NULL CHECK (energy_kwh >= 0),
    co2_kg DECIMAL(12, 6) NOT NULL CHECK (co2_kg >= 0),
    duration_s DECIMAL(10, 3) NOT NULL CHECK (duration_s >= 0),
    
    -- Additional metadata
    run_metadata JSONB, -- Store additional fields from CLI
    git_commit_sha VARCHAR(40), -- Git commit hash if available
    branch_name VARCHAR(255), -- Git branch name if available
    workflow_name VARCHAR(255), -- CI workflow name if available
    
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Indexes for efficient queries
CREATE INDEX idx_runs_user_id ON runs(user_id);
CREATE INDEX idx_runs_repository_id ON runs(repository_id);
CREATE INDEX idx_runs_created_at ON runs(created_at DESC);
CREATE INDEX idx_runs_co2_kg ON runs(co2_kg);

-- Composite index for repository CO2 aggregation queries
CREATE INDEX idx_runs_repo_co2_date ON runs(repository_id, created_at DESC, co2_kg);

-- Trigger to update updated_at timestamps automatically
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_users_updated_at 
    BEFORE UPDATE ON users 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_repositories_updated_at 
    BEFORE UPDATE ON repositories 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Add comments for documentation
COMMENT ON TABLE users IS 'GitHub OAuth authenticated users';
COMMENT ON TABLE repositories IS 'GitHub repositories tracked for CO2 measurements';
COMMENT ON TABLE runs IS 'CO2 measurement runs from CLI executions';

COMMENT ON COLUMN runs.energy_kwh IS 'Energy consumption in kilowatt-hours';
COMMENT ON COLUMN runs.co2_kg IS 'CO2 emissions in kilograms';
COMMENT ON COLUMN runs.duration_s IS 'Execution duration in seconds';
COMMENT ON COLUMN runs.run_metadata IS 'Additional metadata from CLI measurements (JSON)';