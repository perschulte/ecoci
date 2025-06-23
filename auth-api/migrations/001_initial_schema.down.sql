-- Migration rollback: Drop initial schema

-- Drop triggers first
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
DROP TRIGGER IF EXISTS update_repositories_updated_at ON repositories;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop tables in reverse dependency order
DROP TABLE IF EXISTS runs;
DROP TABLE IF EXISTS repositories;
DROP TABLE IF EXISTS users;