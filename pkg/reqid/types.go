// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

// CANARY: REQ=REQ-151; FEATURE="FixedPrefixes"; ASPECT=Engine; STATUS=IMPL; UPDATED=2025-11-01
package reqid

import (
	"fmt"
	"regexp"
	"strconv"
)

// IDType represents the type of identifier
type IDType string

const (
	// TypeREQ represents a requirement specification
	TypeREQ IDType = "REQ"

	// TypeTASK represents a task within a requirement plan
	TypeTASK IDType = "TASK"

	// TypeBUG represents a bug related to a requirement or task
	TypeBUG IDType = "BUG"
)

// ID represents a parsed identifier (REQ, TASK, or BUG)
type ID struct {
	Type   IDType // REQ, TASK, or BUG
	Number int    // Numeric identifier (e.g., 123)

	// Relationships
	ParentID string // For TASK/BUG: the REQ or TASK this relates to (optional)
}

var (
	// Pattern for all ID formats: REQ-123, TASK-456, BUG-789
	idPattern = regexp.MustCompile(`^(REQ|TASK|BUG)-(\d+)$`)
)

// Parse parses an ID string into its components
func Parse(id string) (*ID, error) {
	matches := idPattern.FindStringSubmatch(id)
	if matches == nil {
		return nil, fmt.Errorf("invalid ID format: %q (expected REQ-###, TASK-###, or BUG-###)", id)
	}

	typ := IDType(matches[1])
	number, err := strconv.Atoi(matches[2])
	if err != nil {
		return nil, fmt.Errorf("invalid number in ID %q: %w", id, err)
	}

	return &ID{
		Type:   typ,
		Number: number,
	}, nil
}

// String returns the string representation of the ID
func (id *ID) String() string {
	return fmt.Sprintf("%s-%d", id.Type, id.Number)
}

// IsREQ returns true if this is a REQ identifier
func (id *ID) IsREQ() bool {
	return id.Type == TypeREQ
}

// IsTASK returns true if this is a TASK identifier
func (id *ID) IsTASK() bool {
	return id.Type == TypeTASK
}

// IsBUG returns true if this is a BUG identifier
func (id *ID) IsBUG() bool {
	return id.Type == TypeBUG
}

// Validate checks if the ID is valid
func (id *ID) Validate() error {
	if id.Type != TypeREQ && id.Type != TypeTASK && id.Type != TypeBUG {
		return fmt.Errorf("invalid ID type: %q", id.Type)
	}

	if id.Number <= 0 {
		return fmt.Errorf("ID number must be positive, got: %d", id.Number)
	}

	return nil
}

// New creates a new ID
func New(typ IDType, number int) *ID {
	return &ID{
		Type:   typ,
		Number: number,
	}
}

// NewREQ creates a new REQ identifier
func NewREQ(number int) *ID {
	return &ID{
		Type:   TypeREQ,
		Number: number,
	}
}

// NewTASK creates a new TASK identifier
func NewTASK(number int) *ID {
	return &ID{
		Type:   TypeTASK,
		Number: number,
	}
}

// NewBUG creates a new BUG identifier
func NewBUG(number int) *ID {
	return &ID{
		Type:   TypeBUG,
		Number: number,
	}
}
