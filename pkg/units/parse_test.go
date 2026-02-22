package units

import "testing"

func TestParseMemory(t *testing.T) {
	tests := []struct {
		input string
		want  int64
	}{
		{"100m", 104857600},
		{"1g", 1073741824},
		{"512k", 524288},
		{"50M", 52428800},
		{"2G", 2147483648},
		{"1024", 1024},
	}

	for _, tt := range tests {
		got, err := ParseMemory(tt.input)
		if err != nil {
			t.Errorf("ParseMemory(%q) error: %v", tt.input, err)
			continue
		}
		if got != tt.want {
			t.Errorf("ParseMemory(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestParseMemoryInvalid(t *testing.T) {
	invalid := []string{"", "abc", "100x", "m"}

	for _, s := range invalid {
		_, err := ParseMemory(s)
		if err == nil {
			t.Errorf("ParseMemory(%q) expected error, got nil", s)
		}
	}
}
