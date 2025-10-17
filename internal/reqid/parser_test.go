// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

// CANARY: REQ=CBIN-139; FEATURE="ParserTests"; ASPECT=Engine; STATUS=STUB; TEST=TestCBIN139_AspectIDParser; UPDATED=2025-10-16
package reqid

import (
	"testing"
)

func TestCBIN139_AspectIDParser(t *testing.T) {
	tests := []struct {
		name    string
		reqID   string
		want    *RequirementID
		wantErr bool
	}{
		{
			name:  "new format CLI",
			reqID: "CBIN-CLI-001",
			want:  &RequirementID{Key: "CBIN", Aspect: "CLI", ID: "001", Format: "v2"},
		},
		{
			name:  "new format Engine with higher ID",
			reqID: "CBIN-Engine-042",
			want:  &RequirementID{Key: "CBIN", Aspect: "Engine", ID: "042", Format: "v2"},
		},
		{
			name:  "old format backward compat",
			reqID: "CBIN-001",
			want:  &RequirementID{Key: "CBIN", ID: "001", Format: "v1"},
		},
		{
			name:  "old format with higher ID",
			reqID: "CBIN-138",
			want:  &RequirementID{Key: "CBIN", ID: "138", Format: "v1"},
		},
		{
			name:    "invalid aspect",
			reqID:   "CBIN-InvalidAspect-001",
			wantErr: true,
		},
		{
			name:    "missing ID segment",
			reqID:   "CBIN-CLI",
			wantErr: true,
		},
		{
			name:    "invalid format entirely",
			reqID:   "NOTVALID",
			wantErr: true,
		},
		{
			name:  "case insensitive aspect (cli)",
			reqID: "CBIN-cli-001",
			want:  &RequirementID{Key: "CBIN", Aspect: "cli", ID: "001", Format: "v2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseRequirementID(tt.reqID)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseRequirementID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				if got.Key != tt.want.Key {
					t.Errorf("ParseRequirementID() Key = %v, want %v", got.Key, tt.want.Key)
				}
				if got.Aspect != tt.want.Aspect {
					t.Errorf("ParseRequirementID() Aspect = %v, want %v", got.Aspect, tt.want.Aspect)
				}
				if got.ID != tt.want.ID {
					t.Errorf("ParseRequirementID() ID = %v, want %v", got.ID, tt.want.ID)
				}
				if got.Format != tt.want.Format {
					t.Errorf("ParseRequirementID() Format = %v, want %v", got.Format, tt.want.Format)
				}
			}
		})
	}
}

func TestCBIN139_RequirementID_String(t *testing.T) {
	tests := []struct {
		name string
		req  *RequirementID
		want string
	}{
		{
			name: "v2 format with aspect",
			req:  &RequirementID{Key: "CBIN", Aspect: "CLI", ID: "001", Format: "v2"},
			want: "CBIN-CLI-001",
		},
		{
			name: "v1 format without aspect",
			req:  &RequirementID{Key: "CBIN", ID: "138", Format: "v1"},
			want: "CBIN-138",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.req.String(); got != tt.want {
				t.Errorf("RequirementID.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
