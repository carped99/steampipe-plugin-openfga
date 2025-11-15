package openfga

import (
	"testing"
)

func TestSplitObject(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantType string
		wantID   string
	}{
		{"normal case", "doc:123", "doc", "123"},
		{"empty string", "", "", ""},
		{"no colon", "document", "document", ""},
		{"trailing colon", "user:", "user", ""},
		{"leading colon", ":42", "", "42"},
		{"multiple colons", "a:b:c", "a", "b:c"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotType, gotID := splitObject(tt.input)
			if gotType != tt.wantType || gotID != tt.wantID {
				t.Fatalf("splitObject(%q) = (%q, %q), want (%q, %q)", tt.input, gotType, gotID, tt.wantType, tt.wantID)
			}
		})
	}
}
