// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

// CANARY: REQ=CBIN-146; FEATURE="ProjectRegistry"; ASPECT=Storage; STATUS=IMPL; UPDATED=2025-10-18
package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// Project represents a registered project in the canary system
type Project struct {
	ID        string
	Name      string
	Path      string
	Active    bool
	CreatedAt string
	Metadata  string // JSON metadata
}

// ProjectRegistry manages project registration and queries
type ProjectRegistry struct {
	manager *DatabaseManager
}

// NewProjectRegistry creates a new project registry
func NewProjectRegistry(manager *DatabaseManager) *ProjectRegistry {
	return &ProjectRegistry{
		manager: manager,
	}
}

// Register adds a new project to the registry
func (pr *ProjectRegistry) Register(project *Project) error {
	if project == nil {
		return errors.New("project cannot be nil")
	}

	if project.Name == "" {
		return errors.New("project name is required")
	}

	if project.Path == "" {
		return errors.New("project path is required")
	}

	// Generate slug from project name
	baseSlug := generateSlug(project.Name)

	// Check for slug collisions and generate unique ID
	slug, err := pr.generateUniqueSlug(baseSlug)
	if err != nil {
		return fmt.Errorf("generate unique slug: %w", err)
	}

	project.ID = slug
	project.CreatedAt = time.Now().UTC().Format(time.RFC3339)

	// Ensure projects table exists
	if err := pr.ensureProjectsTable(); err != nil {
		return fmt.Errorf("ensure projects table: %w", err)
	}

	// Insert project
	query := `
		INSERT INTO projects (id, name, path, active, created_at, metadata)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err = pr.manager.conn.Exec(query,
		project.ID,
		project.Name,
		project.Path,
		project.Active,
		project.CreatedAt,
		project.Metadata,
	)

	if err != nil {
		// Check for unique constraint violation on path
		if strings.Contains(err.Error(), "UNIQUE") || strings.Contains(err.Error(), "unique") {
			return fmt.Errorf("project with path %s already exists", project.Path)
		}
		return fmt.Errorf("insert project: %w", err)
	}

	return nil
}

// List returns all registered projects
func (pr *ProjectRegistry) List() ([]*Project, error) {
	if err := pr.ensureProjectsTable(); err != nil {
		return nil, fmt.Errorf("ensure projects table: %w", err)
	}

	query := `
		SELECT id, name, path, active, created_at, COALESCE(metadata, '') as metadata
		FROM projects
		ORDER BY created_at DESC
	`

	rows, err := pr.manager.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("query projects: %w", err)
	}
	defer rows.Close()

	var projects []*Project
	for rows.Next() {
		p := &Project{}
		err := rows.Scan(&p.ID, &p.Name, &p.Path, &p.Active, &p.CreatedAt, &p.Metadata)
		if err != nil {
			return nil, fmt.Errorf("scan project: %w", err)
		}
		projects = append(projects, p)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate projects: %w", err)
	}

	return projects, nil
}

// Remove deletes a project from the registry
func (pr *ProjectRegistry) Remove(id string) error {
	if err := pr.ensureProjectsTable(); err != nil {
		return fmt.Errorf("ensure projects table: %w", err)
	}

	query := `DELETE FROM projects WHERE id = ?`

	result, err := pr.manager.conn.Exec(query, id)
	if err != nil {
		return fmt.Errorf("delete project: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("project with id %s not found", id)
	}

	return nil
}

// GetByID retrieves a project by its ID
func (pr *ProjectRegistry) GetByID(id string) (*Project, error) {
	if err := pr.ensureProjectsTable(); err != nil {
		return nil, fmt.Errorf("ensure projects table: %w", err)
	}

	query := `
		SELECT id, name, path, active, created_at, COALESCE(metadata, '') as metadata
		FROM projects
		WHERE id = ?
	`

	p := &Project{}
	err := pr.manager.conn.QueryRow(query, id).Scan(
		&p.ID, &p.Name, &p.Path, &p.Active, &p.CreatedAt, &p.Metadata,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("project with id %s not found", id)
	}

	if err != nil {
		return nil, fmt.Errorf("query project: %w", err)
	}

	return p, nil
}

// GetByPath retrieves a project by its path
func (pr *ProjectRegistry) GetByPath(path string) (*Project, error) {
	if err := pr.ensureProjectsTable(); err != nil {
		return nil, fmt.Errorf("ensure projects table: %w", err)
	}

	query := `
		SELECT id, name, path, active, created_at, COALESCE(metadata, '') as metadata
		FROM projects
		WHERE path = ?
	`

	p := &Project{}
	err := pr.manager.conn.QueryRow(query, path).Scan(
		&p.ID, &p.Name, &p.Path, &p.Active, &p.CreatedAt, &p.Metadata,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("project with path %s not found", path)
	}

	if err != nil {
		return nil, fmt.Errorf("query project: %w", err)
	}

	return p, nil
}

// generateSlug creates a URL-friendly slug from a project name
func generateSlug(name string) string {
	// Convert to lowercase
	slug := strings.ToLower(name)

	// Replace spaces and underscores with hyphens
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, "_", "-")

	// Remove special characters, keep only alphanumeric and hyphens
	reg := regexp.MustCompile("[^a-z0-9-]+")
	slug = reg.ReplaceAllString(slug, "-")

	// Remove leading/trailing hyphens
	slug = strings.Trim(slug, "-")

	// Replace multiple consecutive hyphens with single hyphen
	reg = regexp.MustCompile("-+")
	slug = reg.ReplaceAllString(slug, "-")

	return slug
}

// generateUniqueSlug ensures the slug is unique, appending a counter if needed
func (pr *ProjectRegistry) generateUniqueSlug(baseSlug string) (string, error) {
	slug := baseSlug
	counter := 2

	for {
		// Check if slug exists
		exists, err := pr.slugExists(slug)
		if err != nil {
			return "", err
		}

		if !exists {
			return slug, nil
		}

		// Slug exists, try with counter
		slug = fmt.Sprintf("%s-%d", baseSlug, counter)
		counter++

		// Safety limit to prevent infinite loops
		if counter > 1000 {
			return "", errors.New("unable to generate unique slug after 1000 attempts")
		}
	}
}

// slugExists checks if a slug is already in use
func (pr *ProjectRegistry) slugExists(slug string) (bool, error) {
	if err := pr.ensureProjectsTable(); err != nil {
		return false, err
	}

	query := `SELECT COUNT(*) FROM projects WHERE id = ?`

	var count int
	err := pr.manager.conn.QueryRow(query, slug).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("check slug existence: %w", err)
	}

	return count > 0, nil
}

// ensureProjectsTable creates the projects table if it doesn't exist
func (pr *ProjectRegistry) ensureProjectsTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS projects (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			path TEXT NOT NULL UNIQUE,
			active BOOLEAN DEFAULT FALSE,
			created_at TEXT NOT NULL,
			metadata TEXT
		)
	`

	_, err := pr.manager.conn.Exec(query)
	if err != nil {
		return fmt.Errorf("create projects table: %w", err)
	}

	// Create index on path
	indexQuery := `CREATE INDEX IF NOT EXISTS idx_projects_path ON projects(path)`
	_, err = pr.manager.conn.Exec(indexQuery)
	if err != nil {
		return fmt.Errorf("create path index: %w", err)
	}

	return nil
}
