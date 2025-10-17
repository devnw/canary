// Copyright (c) 2024 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.


// CANARY: REQ=CBIN-140; FEATURE="GapRepository"; ASPECT=Storage; STATUS=IMPL; UPDATED=2025-10-17
package storage

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// GapEntry represents a gap analysis entry
type GapEntry struct {
	ID               int
	GapID            string
	ReqID            string
	Feature          string
	Aspect           string
	Category         string
	Description      string
	CorrectiveAction string
	CreatedAt        time.Time
	CreatedBy        string
	HelpfulCount     int
	UnhelpfulCount   int
}

// GapCategory represents a gap category
type GapCategory struct {
	ID          int
	Name        string
	Description string
	CreatedAt   time.Time
}

// GapConfig represents gap analysis configuration
type GapConfig struct {
	ID                  int
	MaxGapInjection     int
	MinHelpfulThreshold int
	RankingStrategy     string
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// GapQueryFilter represents query filters for gap entries
type GapQueryFilter struct {
	ReqID    string
	Feature  string
	Aspect   string
	Category string
	Limit    int
}

// GapRepository handles gap analysis database operations
type GapRepository struct {
	db *DB
}

// NewGapRepository creates a new gap repository
func NewGapRepository(db *DB) *GapRepository {
	return &GapRepository{db: db}
}

// CreateEntry creates a new gap analysis entry
func (r *GapRepository) CreateEntry(entry *GapEntry) error {
	// Get category ID
	var categoryID int
	err := r.db.conn.Get(&categoryID, "SELECT id FROM gap_categories WHERE name = ?", entry.Category)
	if err != nil {
		return fmt.Errorf("get category ID: %w", err)
	}

	query := `
		INSERT INTO gap_entries (
			gap_id, req_id, feature, aspect, category_id,
			description, corrective_action, created_at, created_by,
			helpful_count, unhelpful_count
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	createdAt := entry.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now()
	}

	createdBy := entry.CreatedBy
	if createdBy == "" {
		createdBy = "unknown"
	}

	_, err = r.db.conn.Exec(query,
		entry.GapID, entry.ReqID, entry.Feature, entry.Aspect, categoryID,
		entry.Description, entry.CorrectiveAction, createdAt, createdBy,
		entry.HelpfulCount, entry.UnhelpfulCount,
	)
	if err != nil {
		return fmt.Errorf("insert gap entry: %w", err)
	}

	return nil
}

// GetEntryByGapID retrieves a gap entry by its gap ID
func (r *GapRepository) GetEntryByGapID(gapID string) (*GapEntry, error) {
	query := `
		SELECT
			e.id, e.gap_id, e.req_id, e.feature, e.aspect,
			c.name as category, e.description, e.corrective_action,
			e.created_at, e.created_by, e.helpful_count, e.unhelpful_count
		FROM gap_entries e
		JOIN gap_categories c ON e.category_id = c.id
		WHERE e.gap_id = ?
	`

	entry := &GapEntry{}
	err := r.db.conn.QueryRow(query, gapID).Scan(
		&entry.ID, &entry.GapID, &entry.ReqID, &entry.Feature, &entry.Aspect,
		&entry.Category, &entry.Description, &entry.CorrectiveAction,
		&entry.CreatedAt, &entry.CreatedBy, &entry.HelpfulCount, &entry.UnhelpfulCount,
	)
	if err != nil {
		return nil, fmt.Errorf("get gap entry: %w", err)
	}

	return entry, nil
}

// GetEntriesByReqID retrieves all gap entries for a requirement
func (r *GapRepository) GetEntriesByReqID(reqID string) ([]*GapEntry, error) {
	query := `
		SELECT
			e.id, e.gap_id, e.req_id, e.feature, e.aspect,
			c.name as category, e.description, e.corrective_action,
			e.created_at, e.created_by, e.helpful_count, e.unhelpful_count
		FROM gap_entries e
		JOIN gap_categories c ON e.category_id = c.id
		WHERE e.req_id = ?
		ORDER BY e.helpful_count DESC, e.created_at DESC
	`

	rows, err := r.db.conn.Query(query, reqID)
	if err != nil {
		return nil, fmt.Errorf("query gap entries: %w", err)
	}
	defer rows.Close()

	return r.scanGapEntries(rows)
}

// MarkHelpful increments the helpful count for a gap entry
func (r *GapRepository) MarkHelpful(gapID string) error {
	query := `UPDATE gap_entries SET helpful_count = helpful_count + 1 WHERE gap_id = ?`
	_, err := r.db.conn.Exec(query, gapID)
	if err != nil {
		return fmt.Errorf("mark helpful: %w", err)
	}
	return nil
}

// MarkUnhelpful increments the unhelpful count for a gap entry
func (r *GapRepository) MarkUnhelpful(gapID string) error {
	query := `UPDATE gap_entries SET unhelpful_count = unhelpful_count + 1 WHERE gap_id = ?`
	_, err := r.db.conn.Exec(query, gapID)
	if err != nil {
		return fmt.Errorf("mark unhelpful: %w", err)
	}
	return nil
}

// QueryEntries queries gap entries with filters
func (r *GapRepository) QueryEntries(filter GapQueryFilter) ([]*GapEntry, error) {
	query := `
		SELECT
			e.id, e.gap_id, e.req_id, e.feature, e.aspect,
			c.name as category, e.description, e.corrective_action,
			e.created_at, e.created_by, e.helpful_count, e.unhelpful_count
		FROM gap_entries e
		JOIN gap_categories c ON e.category_id = c.id
		WHERE 1=1
	`
	args := []interface{}{}

	if filter.ReqID != "" {
		query += " AND e.req_id = ?"
		args = append(args, filter.ReqID)
	}
	if filter.Feature != "" {
		query += " AND e.feature = ?"
		args = append(args, filter.Feature)
	}
	if filter.Aspect != "" {
		query += " AND e.aspect = ?"
		args = append(args, filter.Aspect)
	}
	if filter.Category != "" {
		query += " AND c.name = ?"
		args = append(args, filter.Category)
	}

	query += " ORDER BY e.helpful_count DESC, e.created_at DESC"

	if filter.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filter.Limit)
	}

	rows, err := r.db.conn.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("query entries: %w", err)
	}
	defer rows.Close()

	return r.scanGapEntries(rows)
}

// GetTopGaps retrieves top gaps for a requirement based on configuration
func (r *GapRepository) GetTopGaps(reqID string, config *GapConfig) ([]*GapEntry, error) {
	query := `
		SELECT
			e.id, e.gap_id, e.req_id, e.feature, e.aspect,
			c.name as category, e.description, e.corrective_action,
			e.created_at, e.created_by, e.helpful_count, e.unhelpful_count
		FROM gap_entries e
		JOIN gap_categories c ON e.category_id = c.id
		WHERE e.req_id = ?
		AND e.helpful_count >= ?
	`
	args := []interface{}{reqID, config.MinHelpfulThreshold}

	// Apply ranking strategy
	switch config.RankingStrategy {
	case "helpful_desc":
		query += " ORDER BY e.helpful_count DESC, e.created_at DESC"
	case "recency_desc":
		query += " ORDER BY e.created_at DESC"
	case "weighted":
		// Weighted: (helpful_count * 2) - unhelpful_count, then recency
		query += " ORDER BY (e.helpful_count * 2 - e.unhelpful_count) DESC, e.created_at DESC"
	default:
		query += " ORDER BY e.helpful_count DESC, e.created_at DESC"
	}

	query += " LIMIT ?"
	args = append(args, config.MaxGapInjection)

	rows, err := r.db.conn.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("get top gaps: %w", err)
	}
	defer rows.Close()

	return r.scanGapEntries(rows)
}

// GetCategories retrieves all gap categories
func (r *GapRepository) GetCategories() ([]*GapCategory, error) {
	query := `SELECT id, name, description, created_at FROM gap_categories ORDER BY name ASC`

	rows, err := r.db.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("get categories: %w", err)
	}
	defer rows.Close()

	var categories []*GapCategory
	for rows.Next() {
		cat := &GapCategory{}
		err := rows.Scan(&cat.ID, &cat.Name, &cat.Description, &cat.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan category: %w", err)
		}
		categories = append(categories, cat)
	}

	return categories, rows.Err()
}

// GetConfig retrieves the gap analysis configuration
func (r *GapRepository) GetConfig() (*GapConfig, error) {
	query := `
		SELECT id, max_gap_injection, min_helpful_threshold, ranking_strategy,
			created_at, updated_at
		FROM gap_config
		WHERE id = 1
	`

	config := &GapConfig{}
	err := r.db.conn.QueryRow(query).Scan(
		&config.ID, &config.MaxGapInjection, &config.MinHelpfulThreshold,
		&config.RankingStrategy, &config.CreatedAt, &config.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get config: %w", err)
	}

	return config, nil
}

// UpdateConfig updates the gap analysis configuration
func (r *GapRepository) UpdateConfig(config *GapConfig) error {
	query := `
		UPDATE gap_config
		SET max_gap_injection = ?,
			min_helpful_threshold = ?,
			ranking_strategy = ?,
			updated_at = ?
		WHERE id = 1
	`

	updatedAt := config.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = time.Now()
	}

	_, err := r.db.conn.Exec(query,
		config.MaxGapInjection,
		config.MinHelpfulThreshold,
		config.RankingStrategy,
		updatedAt,
	)
	if err != nil {
		return fmt.Errorf("update config: %w", err)
	}

	return nil
}

// GenerateGapReport generates a formatted gap analysis report
func (r *GapRepository) GenerateGapReport(reqID string) (string, error) {
	entries, err := r.GetEntriesByReqID(reqID)
	if err != nil {
		return "", fmt.Errorf("get entries: %w", err)
	}

	if len(entries) == 0 {
		return fmt.Sprintf("No gap analysis entries found for %s\n", reqID), nil
	}

	var report strings.Builder
	report.WriteString(fmt.Sprintf("# Gap Analysis Report for %s\n\n", reqID))
	report.WriteString(fmt.Sprintf("Total Gaps: %d\n\n", len(entries)))

	// Group by category
	categoryGroups := make(map[string][]*GapEntry)
	for _, entry := range entries {
		categoryGroups[entry.Category] = append(categoryGroups[entry.Category], entry)
	}

	for category, catEntries := range categoryGroups {
		report.WriteString(fmt.Sprintf("## Category: %s (%d)\n\n", category, len(catEntries)))

		for _, entry := range catEntries {
			report.WriteString(fmt.Sprintf("### %s - %s\n", entry.GapID, entry.Feature))
			report.WriteString(fmt.Sprintf("**Description:** %s\n\n", entry.Description))
			if entry.CorrectiveAction != "" {
				report.WriteString(fmt.Sprintf("**Corrective Action:** %s\n\n", entry.CorrectiveAction))
			}
			report.WriteString(fmt.Sprintf("**Helpful:** %d | **Unhelpful:** %d | **Created:** %s\n\n",
				entry.HelpfulCount, entry.UnhelpfulCount, entry.CreatedAt.Format("2006-01-02")))
			report.WriteString("---\n\n")
		}
	}

	return report.String(), nil
}

// scanGapEntries scans gap entries from SQL rows
func (r *GapRepository) scanGapEntries(rows *sql.Rows) ([]*GapEntry, error) {
	var entries []*GapEntry
	for rows.Next() {
		entry := &GapEntry{}
		err := rows.Scan(
			&entry.ID, &entry.GapID, &entry.ReqID, &entry.Feature, &entry.Aspect,
			&entry.Category, &entry.Description, &entry.CorrectiveAction,
			&entry.CreatedAt, &entry.CreatedBy, &entry.HelpfulCount, &entry.UnhelpfulCount,
		)
		if err != nil {
			return nil, fmt.Errorf("scan gap entry: %w", err)
		}
		entries = append(entries, entry)
	}
	return entries, rows.Err()
}
