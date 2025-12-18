package pages

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVisibility_String(t *testing.T) {
	tests := []struct {
		name       string
		visibility Visibility
		expected   string
	}{
		{
			name:       "Private visibility",
			visibility: Private,
			expected:   "private",
		},
		{
			name:       "Public visibility",
			visibility: Public,
			expected:   "public",
		},
		{
			name:       "Unknown visibility - invalid value",
			visibility: Visibility(99),
			expected:   "unknown",
		},
		{
			name:       "Zero value visibility",
			visibility: Visibility(0),
			expected:   "private",
		},
		{
			name:       "Negative visibility",
			visibility: Visibility(-1),
			expected:   "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.visibility.String()
			assert.Equal(t, tt.expected, result, "String() should return the expected value")
		})
	}
}

func TestVisibilityFromString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected Visibility
	}{
		{
			name:     "Public string",
			input:    "public",
			expected: Public,
		},
		{
			name:     "Private string - default case",
			input:    "private",
			expected: Private,
		},
		{
			name:     "Empty string",
			input:    "",
			expected: Private,
		},
		{
			name:     "Invalid string",
			input:    "invalid",
			expected: Private,
		},
		{
			name:     "Mixed case public",
			input:    "Public",
			expected: Private,
		},
		{
			name:     "Uppercase public",
			input:    "PUBLIC",
			expected: Private,
		},
		{
			name:     "Whitespace around public",
			input:    " public ",
			expected: Private,
		},
		{
			name:     "Partial match",
			input:    "pub",
			expected: Private,
		},
		{
			name:     "Number string",
			input:    "123",
			expected: Private,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := VisibilityFromString(tt.input)
			assert.Equal(t, tt.expected, result, "VisibilityFromString() should return the expected visibility")
		})
	}
}

func TestVisibility_Constants(t *testing.T) {
	tests := []struct {
		name      string
		constant  Visibility
		expected  int
		stringVal string
	}{
		{
			name:      "Private constant",
			constant:  Private,
			expected:  0,
			stringVal: "private",
		},
		{
			name:      "Public constant",
			constant:  Public,
			expected:  1,
			stringVal: "public",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, int(tt.constant), "Constant should have expected integer value")
			assert.Equal(t, tt.stringVal, tt.constant.String(), "Constant should have expected string representation")
		})
	}
}

func TestVisibility_RoundTrip(t *testing.T) {
	tests := []struct {
		name     string
		original Visibility
	}{
		{
			name:     "Private round trip",
			original: Private,
		},
		{
			name:     "Public round trip",
			original: Public,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert to string and back
			str := tt.original.String()
			result := VisibilityFromString(str)

			assert.Equal(t, tt.original, result, "Round trip conversion should preserve the original value")
		})
	}
}

func TestVisibility_ValuesCoverage(t *testing.T) {
	// Ensure we have tested all possible valid visibility values
	validVisibilities := []Visibility{Private, Public}

	for _, v := range validVisibilities {
		str := v.String()
		result := VisibilityFromString(str)

		assert.Equal(t, v, result, "Visibility %d should round trip correctly", v)
		require.NotEqual(t, "unknown", v.String(), "Valid visibility should not return 'unknown'")
	}
}

func BenchmarkVisibility_String(b *testing.B) {
	visibilities := []Visibility{Private, Public}

	for _, v := range visibilities {
		b.Run(v.String(), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = v.String()
			}
		})
	}
}

func BenchmarkVisibilityFromString(b *testing.B) {
	inputs := []string{"public", "private", "invalid"}

	for _, input := range inputs {
		b.Run(input, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = VisibilityFromString(input)
			}
		})
	}
}
