package pages

import (
	"fmt"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

func TestStatus_String(t *testing.T) {
	tests := []struct {
		name     string
		status   Status
		expected string
	}{
		{
			name:     "Draft string representation",
			status:   Draft,
			expected: "draft",
		},
		{
			name:     "Published string representation",
			status:   Published,
			expected: "published",
		},
		{
			name:     "Unknown status (negative value)",
			status:   Status(-1),
			expected: "unknown",
		},
		{
			name:     "Unknown status (high positive value)",
			status:   Status(127), // Maximum int8 value
			expected: "unknown",
		},
		{
			name:     "Unknown status (zero value)",
			status:   Status(0),
			expected: "draft", // 0 corresponds to Draft
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.status.String()
			assert.Equal(t, tt.expected, result, "Status.String() should return correct string representation")
		})
	}
}

func TestStatus_Constants(t *testing.T) {
	t.Run("Constant values are correct", func(t *testing.T) {
		assert.Equal(t, Status(0), Draft, "Draft should be 0")
		assert.Equal(t, Status(1), Published, "Published should be 1")
	})

	t.Run("Constants are unique", func(t *testing.T) {
		constants := []Status{Draft, Published}
		seen := make(map[Status]bool)

		for _, status := range constants {
			assert.False(t, seen[status], "Status constant %v should be unique", status)
			seen[status] = true
		}
	})

	t.Run("Constants are in ascending order", func(t *testing.T) {
		assert.True(t, Draft < Published, "Draft should be less than Published")
	})
}

func TestStatus_Usage(t *testing.T) {
	t.Run("Status in switch statements", func(t *testing.T) {
		statuses := []Status{Draft, Published}

		for _, status := range statuses {
			var result string
			switch status {
			case Draft:
				result = "found draft"
			case Published:
				result = "found published"
			default:
				result = "found unknown"
			}
			expected := "found " + status.String()
			assert.Equal(t, expected, result, "Switch statement should handle status correctly")
		}
	})

	t.Run("Status as map key", func(t *testing.T) {
		statusMap := map[Status]string{
			Draft:     "Draft content",
			Published: "Published content",
		}

		assert.Equal(t, "Draft content", statusMap[Draft], "Status should work as map key")
		assert.Equal(t, "Published content", statusMap[Published], "Status should work as map key")
		assert.Empty(t, statusMap[Status(100)], "Unknown status should return empty string")
	})

	t.Run("Status slice operations", func(t *testing.T) {
		statuses := []Status{Published, Draft}

		assert.Contains(t, statuses, Draft, "Status slice should contain Draft")
		assert.Contains(t, statuses, Published, "Status slice should contain Published")
		assert.Len(t, statuses, 2, "Status slice should have correct length")
	})
}

func TestStatus_StringIntegration(t *testing.T) {
	t.Run("Status String() method in formatting", func(t *testing.T) {
		statuses := []Status{Draft, Published}

		for _, status := range statuses {
			formatted := fmt.Sprintf("Status: %s", status)
			expected := fmt.Sprintf("Status: %s", status.String())
			assert.Equal(t, expected, formatted, "Status should format correctly with fmt.Sprintf")
		}
	})
}

func TestStatus_EdgeCases(t *testing.T) {
	t.Run("Status with maximum int8 value", func(t *testing.T) {
		status := Status(127) // Maximum int8 value
		result := status.String()
		assert.Equal(t, "unknown", result, "Maximum int8 value should return 'unknown'")
	})

	t.Run("Status with minimum int8 value", func(t *testing.T) {
		status := Status(-128) // Minimum int8 value
		result := status.String()
		assert.Equal(t, "unknown", result, "Minimum int8 value should return 'unknown'")
	})

	t.Run("Status with edge values around constants", func(t *testing.T) {
		testCases := []struct {
			status   Status
			expected string
		}{
			{Status(-1), "unknown"},
			{Status(0), "draft"},     // Draft
			{Status(1), "published"}, // Published
			{Status(2), "unknown"},   // No more PrivateStatus
			{Status(3), "unknown"},
		}

		for _, tc := range testCases {
			t.Run(fmt.Sprintf("Status_%d", tc.status), func(t *testing.T) {
				result := tc.status.String()
				assert.Equal(t, tc.expected, result, "Status %d should return '%s'", tc.status, tc.expected)
			})
		}
	})
}

func TestStatus_TypeProperties(t *testing.T) {
	t.Run("Status is int8 type", func(t *testing.T) {
		var status Status
		assert.Equal(t, int8(0), int8(status), "Status should be convertible to int8")
	})

	t.Run("Status arithmetic operations", func(t *testing.T) {
		status1 := Draft
		status2 := status1 + 1

		assert.Equal(t, Published, status2, "Status arithmetic should work")
	})

	t.Run("Status type size", func(t *testing.T) {
		var status Status
		assert.Equal(t, 1, int(unsafe.Sizeof(status)), "Status should be 1 byte (int8)")
	})
}

func TestStatus_DefaultValues(t *testing.T) {
	t.Run("Zero value of Status", func(t *testing.T) {
		var status Status
		assert.Equal(t, Draft, status, "Zero value should be Draft")
		assert.Equal(t, "draft", status.String(), "Zero value should stringify to 'draft'")
	})

	t.Run("Status as function parameter", func(t *testing.T) {
		checkStatus := func(s Status) string {
			return s.String()
		}

		assert.Equal(t, "draft", checkStatus(Draft), "Status should pass as parameter correctly")
		assert.Equal(t, "published", checkStatus(Published), "Status should pass as parameter correctly")
	})
}

func TestStatus_ConcurrentAccess(t *testing.T) {
	t.Run("Concurrent String() calls", func(t *testing.T) {
		status := Published
		results := make(chan string, 100)

		// Launch multiple goroutines calling String() concurrently
		for i := 0; i < 100; i++ {
			go func() {
				results <- status.String()
			}()
		}

		// Collect results
		for i := 0; i < 100; i++ {
			result := <-results
			assert.Equal(t, "published", result, "Concurrent String() calls should return consistent results")
		}
	})

	t.Run("Concurrent status operations", func(t *testing.T) {
		statuses := []Status{Draft, Published}
		results := make(chan string, len(statuses))

		for _, status := range statuses {
			go func(s Status) {
				results <- fmt.Sprintf("status_%s", s.String())
			}(status)
		}

		expectedResults := []string{"status_draft", "status_published"}
		actualResults := make([]string, 0, len(statuses))

		for i := 0; i < len(statuses); i++ {
			actualResults = append(actualResults, <-results)
		}

		// Sort both slices for comparison (since goroutine execution order is not guaranteed)
		assert.ElementsMatch(t, expectedResults, actualResults, "Concurrent status operations should produce correct results")
	})
}

// Benchmark tests for performance validation
func BenchmarkStatus_String(b *testing.B) {
	status := Published

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = status.String()
	}
}

func BenchmarkStatus_AllConstants_String(b *testing.B) {
	statuses := []Status{Draft, Published}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, status := range statuses {
			_ = status.String()
		}
	}
}
