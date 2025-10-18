// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

// CANARY: REQ=CBIN-146; FEATURE="ContextManagement"; ASPECT=Engine; STATUS=IMPL; UPDATED=2025-10-18
package storage

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ContextManager manages the current project context
type ContextManager struct {
	manager  *DatabaseManager
	registry *ProjectRegistry
}

// NewContextManager creates a new context manager
func NewContextManager(manager *DatabaseManager) *ContextManager {
	return &ContextManager{
		manager:  manager,
		registry: NewProjectRegistry(manager),
	}
}

// DetectProject attempts to detect the current project from the working directory
func (cm *ContextManager) DetectProject() (*Project, error) {
	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("get working directory: %w", err)
	}

	// List all registered projects
	projects, err := cm.registry.List()
	if err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}

	if len(projects) == 0 {
		return nil, errors.New("no projects registered")
	}

	// Find project that matches current path or is a parent of current path
	var matchedProject *Project
	maxMatchLength := 0

	for _, project := range projects {
		// Normalize paths for comparison
		projectPath := filepath.Clean(project.Path)
		currentPath := filepath.Clean(cwd)

		// Check if current path is under project path
		if currentPath == projectPath || strings.HasPrefix(currentPath, projectPath+string(filepath.Separator)) {
			// Use longest match (most specific project)
			if len(projectPath) > maxMatchLength {
				matchedProject = project
				maxMatchLength = len(projectPath)
			}
		}
	}

	if matchedProject == nil {
		return nil, fmt.Errorf("no project found for path: %s", cwd)
	}

	return matchedProject, nil
}

// SwitchTo switches the current project context to the specified project ID
func (cm *ContextManager) SwitchTo(projectID string) error {
	// Verify project exists
	project, err := cm.registry.GetByID(projectID)
	if err != nil {
		return fmt.Errorf("get project: %w", err)
	}

	// Deactivate all projects first
	if err := cm.deactivateAll(); err != nil {
		return fmt.Errorf("deactivate projects: %w", err)
	}

	// Activate the target project
	if err := cm.setActive(project.ID, true); err != nil {
		return fmt.Errorf("activate project: %w", err)
	}

	return nil
}

// GetCurrent returns the currently active project
func (cm *ContextManager) GetCurrent() (*Project, error) {
	// Query for active project
	query := `
		SELECT id, name, path, active, created_at, COALESCE(metadata, '') as metadata
		FROM projects
		WHERE active = 1
		LIMIT 1
	`

	p := &Project{}
	err := cm.manager.conn.QueryRow(query).Scan(
		&p.ID, &p.Name, &p.Path, &p.Active, &p.CreatedAt, &p.Metadata,
	)

	if err != nil {
		return nil, errors.New("no active project context set")
	}

	return p, nil
}

// deactivateAll sets all projects to inactive
func (cm *ContextManager) deactivateAll() error {
	query := `UPDATE projects SET active = 0`

	_, err := cm.manager.conn.Exec(query)
	if err != nil {
		return fmt.Errorf("update projects: %w", err)
	}

	return nil
}

// setActive sets the active flag for a project
func (cm *ContextManager) setActive(projectID string, active bool) error {
	query := `UPDATE projects SET active = ? WHERE id = ?`

	result, err := cm.manager.conn.Exec(query, active, projectID)
	if err != nil {
		return fmt.Errorf("update project: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("project %s not found", projectID)
	}

	return nil
}
