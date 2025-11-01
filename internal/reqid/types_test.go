// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package reqid

import (
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		input   string
		want    *ID
		wantErr bool
	}{
		{"REQ-1", &ID{Type: TypeREQ, Number: 1}, false},
		{"REQ-123", &ID{Type: TypeREQ, Number: 123}, false},
		{"TASK-1", &ID{Type: TypeTASK, Number: 1}, false},
		{"TASK-456", &ID{Type: TypeTASK, Number: 456}, false},
		{"BUG-1", &ID{Type: TypeBUG, Number: 1}, false},
		{"BUG-789", &ID{Type: TypeBUG, Number: 789}, false},
		
		// Invalid formats
		{"CBIN-001", nil, true},           // Old format
		{"REQ-ABC", nil, true},            // Non-numeric
		{"INVALID-123", nil, true},        // Wrong prefix
		{"REQ", nil, true},                // Missing number
		{"123", nil, true},                // Missing type
		{"REQ-", nil, true},               // Missing number
		{"", nil, true},                   // Empty
	}
	
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := Parse(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.Type != tt.want.Type || got.Number != tt.want.Number {
					t.Errorf("Parse(%q) = %v, want %v", tt.input, got, tt.want)
				}
				if got.String() != tt.input {
					t.Errorf("Parse(%q).String() = %q, want %q", tt.input, got.String(), tt.input)
				}
			}
		})
	}
}

func TestIDTypeChecks(t *testing.T) {
	req := NewREQ(1)
	if !req.IsREQ() {
		t.Error("NewREQ(1).IsREQ() = false, want true")
	}
	if req.IsTASK() || req.IsBUG() {
		t.Error("REQ should not be TASK or BUG")
	}
	
	task := NewTASK(1)
	if !task.IsTASK() {
		t.Error("NewTASK(1).IsTASK() = false, want true")
	}
	if task.IsREQ() || task.IsBUG() {
		t.Error("TASK should not be REQ or BUG")
	}
	
	bug := NewBUG(1)
	if !bug.IsBUG() {
		t.Error("NewBUG(1).IsBUG() = false, want true")
	}
	if bug.IsREQ() || bug.IsTASK() {
		t.Error("BUG should not be REQ or TASK")
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		id      *ID
		wantErr bool
	}{
		{"valid REQ", NewREQ(1), false},
		{"valid TASK", NewTASK(123), false},
		{"valid BUG", NewBUG(456), false},
		{"zero number", &ID{Type: TypeREQ, Number: 0}, true},
		{"negative number", &ID{Type: TypeREQ, Number: -1}, true},
		{"invalid type", &ID{Type: "INVALID", Number: 1}, true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.id.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("ID.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestString(t *testing.T) {
	tests := []struct {
		id   *ID
		want string
	}{
		{NewREQ(1), "REQ-1"},
		{NewREQ(123), "REQ-123"},
		{NewTASK(456), "TASK-456"},
		{NewBUG(789), "BUG-789"},
	}
	
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.id.String(); got != tt.want {
				t.Errorf("ID.String() = %q, want %q", got, tt.want)
			}
		})
	}
}
