// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

package utils

import "testing"

// CANARY: REQ=CBIN-205; FEATURE="ContextCaps"; ASPECT=CLI; STATUS=TESTED; TEST=TestCANARY_CBIN_205_EffectiveLimit; UPDATED=2026-08-28
func TestCANARY_CBIN_205_EffectiveLimit(t *testing.T) {
	tests := []struct {
		name string
		flag int
		def  int
		want int
	}{
		{"unset (zero) maps to default", 0, 20, 20},
		{"explicit -1 maps to unlimited (0)", -1, 20, 0},
		{"any other negative maps to unlimited (0)", -5, 20, 0},
		{"positive value passes through", 5, 20, 5},
		{"positive value larger than default passes through", 100, 20, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EffectiveLimit(tt.flag, tt.def)
			if got != tt.want {
				t.Errorf("EffectiveLimit(%d, %d) = %d, want %d", tt.flag, tt.def, got, tt.want)
			}
		})
	}
}
