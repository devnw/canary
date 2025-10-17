// Copyright (c) 2024 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

// CANARY: REQ=CBIN-140; FEATURE="GapService"; ASPECT=Engine; STATUS=IMPL; UPDATED=2025-10-17
package gap

import (
	"fmt"
	"strings"
	"time"

	"go.spyder.org/canary/internal/storage"
)

// Service provides gap analysis business logic
type Service struct {
	repo *storage.GapRepository
}

// NewService creates a new gap analysis service
func NewService(repo *storage.GapRepository) *Service {
	return &Service{repo: repo}
}

// MarkGap records a new gap analysis entry
func (s *Service) MarkGap(reqID, feature, aspect, category, description, correctiveAction, createdBy string) (string, error) {
	// Validate inputs
	if reqID == "" {
		return "", fmt.Errorf("req_id is required")
	}
	if feature == "" {
		return "", fmt.Errorf("feature is required")
	}
	if category == "" {
		return "", fmt.Errorf("category is required")
	}
	if description == "" {
		return "", fmt.Errorf("description is required")
	}

	// Validate category exists
	categories, err := s.repo.GetCategories()
	if err != nil {
		return "", fmt.Errorf("get categories: %w", err)
	}

	validCategory := false
	for _, cat := range categories {
		if cat.Name == category {
			validCategory = true
			break
		}
	}
	if !validCategory {
		return "", fmt.Errorf("invalid category: %s", category)
	}

	// Generate gap ID
	gapID, err := s.generateGapID(reqID)
	if err != nil {
		return "", fmt.Errorf("generate gap ID: %w", err)
	}

	// Create entry
	entry := &storage.GapEntry{
		GapID:            gapID,
		ReqID:            reqID,
		Feature:          feature,
		Aspect:           aspect,
		Category:         category,
		Description:      description,
		CorrectiveAction: correctiveAction,
		CreatedAt:        time.Now(),
		CreatedBy:        createdBy,
	}

	if err := s.repo.CreateEntry(entry); err != nil {
		return "", fmt.Errorf("create entry: %w", err)
	}

	return gapID, nil
}

// QueryGaps queries gap entries with filters
func (s *Service) QueryGaps(reqID, feature, aspect, category string, limit int) ([]*storage.GapEntry, error) {
	filter := storage.GapQueryFilter{
		ReqID:    reqID,
		Feature:  feature,
		Aspect:   aspect,
		Category: category,
		Limit:    limit,
	}

	entries, err := s.repo.QueryEntries(filter)
	if err != nil {
		return nil, fmt.Errorf("query entries: %w", err)
	}

	return entries, nil
}

// GenerateReport generates a gap analysis report for a requirement
func (s *Service) GenerateReport(reqID string) (string, error) {
	if reqID == "" {
		return "", fmt.Errorf("req_id is required")
	}

	report, err := s.repo.GenerateGapReport(reqID)
	if err != nil {
		return "", fmt.Errorf("generate report: %w", err)
	}

	return report, nil
}

// MarkHelpful marks a gap entry as helpful
func (s *Service) MarkHelpful(gapID string) error {
	if gapID == "" {
		return fmt.Errorf("gap_id is required")
	}

	// Verify gap exists
	_, err := s.repo.GetEntryByGapID(gapID)
	if err != nil {
		return fmt.Errorf("gap not found: %s", gapID)
	}

	if err := s.repo.MarkHelpful(gapID); err != nil {
		return fmt.Errorf("mark helpful: %w", err)
	}

	return nil
}

// MarkUnhelpful marks a gap entry as unhelpful
func (s *Service) MarkUnhelpful(gapID string) error {
	if gapID == "" {
		return fmt.Errorf("gap_id is required")
	}

	// Verify gap exists
	_, err := s.repo.GetEntryByGapID(gapID)
	if err != nil {
		return fmt.Errorf("gap not found: %s", gapID)
	}

	if err := s.repo.MarkUnhelpful(gapID); err != nil {
		return fmt.Errorf("mark unhelpful: %w", err)
	}

	return nil
}

// GetCategories retrieves all available gap categories
func (s *Service) GetCategories() ([]*storage.GapCategory, error) {
	categories, err := s.repo.GetCategories()
	if err != nil {
		return nil, fmt.Errorf("get categories: %w", err)
	}
	return categories, nil
}

// GetTopGapsForPlan retrieves top gaps for a requirement to inject into planning
func (s *Service) GetTopGapsForPlan(reqID string) ([]*storage.GapEntry, error) {
	if reqID == "" {
		return nil, fmt.Errorf("req_id is required")
	}

	// Get configuration
	config, err := s.repo.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("get config: %w", err)
	}

	// Get top gaps
	gaps, err := s.repo.GetTopGaps(reqID, config)
	if err != nil {
		return nil, fmt.Errorf("get top gaps: %w", err)
	}

	return gaps, nil
}

// FormatGapsForInjection formats gaps for injection into plan command prompt
func (s *Service) FormatGapsForInjection(reqID string) (string, error) {
	gaps, err := s.GetTopGapsForPlan(reqID)
	if err != nil {
		return "", fmt.Errorf("get gaps for plan: %w", err)
	}

	if len(gaps) == 0 {
		return "", nil // No gaps to inject
	}

	var output strings.Builder
	output.WriteString("\n\n## Past Implementation Gaps\n\n")
	output.WriteString(fmt.Sprintf("The following gaps were identified in previous implementations of %s:\n\n", reqID))

	for i, gap := range gaps {
		output.WriteString(fmt.Sprintf("%d. **%s** (%s)\n", i+1, gap.Feature, gap.Category))
		output.WriteString(fmt.Sprintf("   - **Problem:** %s\n", gap.Description))
		if gap.CorrectiveAction != "" {
			output.WriteString(fmt.Sprintf("   - **Solution:** %s\n", gap.CorrectiveAction))
		}
		output.WriteString(fmt.Sprintf("   - **Helpfulness:** %d helpful, %d unhelpful\n", gap.HelpfulCount, gap.UnhelpfulCount))
		output.WriteString("\n")
	}

	output.WriteString("**Action:** Review these gaps and ensure your implementation avoids similar mistakes.\n\n")

	return output.String(), nil
}

// UpdateConfig updates gap analysis configuration
func (s *Service) UpdateConfig(maxGapInjection, minHelpfulThreshold int, rankingStrategy string) error {
	// Validate ranking strategy
	validStrategies := map[string]bool{
		"helpful_desc": true,
		"recency_desc": true,
		"weighted":     true,
	}

	if !validStrategies[rankingStrategy] {
		return fmt.Errorf("invalid ranking strategy: %s (must be helpful_desc, recency_desc, or weighted)", rankingStrategy)
	}

	if maxGapInjection < 0 {
		return fmt.Errorf("max_gap_injection must be >= 0")
	}

	if minHelpfulThreshold < 0 {
		return fmt.Errorf("min_helpful_threshold must be >= 0")
	}

	config := &storage.GapConfig{
		MaxGapInjection:     maxGapInjection,
		MinHelpfulThreshold: minHelpfulThreshold,
		RankingStrategy:     rankingStrategy,
		UpdatedAt:           time.Now(),
	}

	if err := s.repo.UpdateConfig(config); err != nil {
		return fmt.Errorf("update config: %w", err)
	}

	return nil
}

// GetConfig retrieves current gap analysis configuration
func (s *Service) GetConfig() (*storage.GapConfig, error) {
	config, err := s.repo.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("get config: %w", err)
	}
	return config, nil
}

// generateGapID generates a unique gap ID for a requirement
func (s *Service) generateGapID(reqID string) (string, error) {
	// Get existing entries for this requirement
	entries, err := s.repo.GetEntriesByReqID(reqID)
	if err != nil {
		return "", err
	}

	// Find next available number
	nextNum := 1
	for _, entry := range entries {
		// Extract number from GAP-CBIN-XXX-NNN
		parts := strings.Split(entry.GapID, "-")
		if len(parts) >= 4 {
			var num int
			fmt.Sscanf(parts[3], "%d", &num)
			if num >= nextNum {
				nextNum = num + 1
			}
		}
	}

	// Generate gap ID: GAP-{REQ_ID}-{NUMBER}
	gapID := fmt.Sprintf("GAP-%s-%03d", reqID, nextNum)
	return gapID, nil
}
