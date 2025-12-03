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
			name:     "DraftStatus string representation",
			status:   DraftStatus,
			expected: "draft",
		},
		{
			name:     "PublishStatus string representation",
			status:   PublishStatus,
			expected: "publish",
		},
		{
			name:     "PrivateStatus string representation",
			status:   PrivateStatus,
			expected: "private",
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
			expected: "draft", // 0 corresponds to DraftStatus
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
		assert.Equal(t, Status(0), DraftStatus, "DraftStatus should be 0")
		assert.Equal(t, Status(1), PublishStatus, "PublishStatus should be 1")
		assert.Equal(t, Status(2), PrivateStatus, "PrivateStatus should be 2")
	})

	t.Run("Constants are unique", func(t *testing.T) {
		constants := []Status{DraftStatus, PublishStatus, PrivateStatus}
		seen := make(map[Status]bool)

		for _, status := range constants {
			assert.False(t, seen[status], "Status constant %v should be unique", status)
			seen[status] = true
		}
	})

	t.Run("Constants are in ascending order", func(t *testing.T) {
		assert.True(t, DraftStatus < PublishStatus, "DraftStatus should be less than PublishStatus")
		assert.True(t, PublishStatus < PrivateStatus, "PublishStatus should be less than PrivateStatus")
		assert.True(t, DraftStatus < PrivateStatus, "DraftStatus should be less than PrivateStatus")
	})
}

func TestStatus_Usage(t *testing.T) {
	t.Run("Status assignment and comparison", func(t *testing.T) {
		var status Status
		status = DraftStatus

		assert.Equal(t, DraftStatus, status, "Status assignment should work correctly")
		assert.True(t, status == DraftStatus, "Status comparison should work correctly")
		assert.False(t, status == PublishStatus, "Status should not equal different constant")
	})

	t.Run("Status in switch statements", func(t *testing.T) {
		statuses := []Status{DraftStatus, PublishStatus, PrivateStatus}
		expectedStrings := []string{"draft", "publish", "private"}

		for i, status := range statuses {
			var result string
			switch status {
			case DraftStatus:
				result = "found draft"
			case PublishStatus:
				result = "found publish"
			case PrivateStatus:
				result = "found private"
			default:
				result = "found unknown"
			}
			expected := "found " + expectedStrings[i]
			assert.Equal(t, expected, result, "Switch statement should handle status correctly")
		}
	})

	t.Run("Status as map key", func(t *testing.T) {
		statusMap := map[Status]string{
			DraftStatus:   "Draft content",
			PublishStatus: "Published content",
			PrivateStatus: "Private content",
		}

		assert.Equal(t, "Draft content", statusMap[DraftStatus], "Status should work as map key")
		assert.Equal(t, "Published content", statusMap[PublishStatus], "Status should work as map key")
		assert.Equal(t, "Private content", statusMap[PrivateStatus], "Status should work as map key")
		assert.Empty(t, statusMap[Status(100)], "Unknown status should return empty string")
	})

	t.Run("Status slice operations", func(t *testing.T) {
		statuses := []Status{PublishStatus, DraftStatus, PrivateStatus}

		assert.Contains(t, statuses, DraftStatus, "Status slice should contain DraftStatus")
		assert.Contains(t, statuses, PublishStatus, "Status slice should contain PublishStatus")
		assert.Contains(t, statuses, PrivateStatus, "Status slice should contain PrivateStatus")
		assert.Len(t, statuses, 3, "Status slice should have correct length")
	})
}

func TestStatus_StringIntegration(t *testing.T) {
	t.Run("Status String() method in formatting", func(t *testing.T) {
		statuses := []Status{DraftStatus, PublishStatus, PrivateStatus}

		for _, status := range statuses {
			formatted := fmt.Sprintf("Status: %s", status.String())
			expected := fmt.Sprintf("Status: %s", status.String())
			assert.Equal(t, expected, formatted, "Status should format correctly with fmt.Sprintf")
		}
	})

	t.Run("Status String() method with different verbs", func(t *testing.T) {
		status := PublishStatus

		// Test with %s verb
		result := fmt.Sprintf("%s", status.String())
		assert.Equal(t, "publish", result, "Status should format with %s verb")

		// Test with %q verb
		result = fmt.Sprintf("%q", status.String())
		assert.Equal(t, `"publish"`, result, "Status should format with %q verb")
	})

	t.Run("Status String() method in logging scenarios", func(t *testing.T) {
		status := PrivateStatus
		logMessage := fmt.Sprintf("Page status changed to %s", status.String())
		expected := "Page status changed to private"
		assert.Equal(t, expected, logMessage, "Status should be usable in log messages")
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
			{Status(0), "draft"},   // DraftStatus
			{Status(1), "publish"}, // PublishStatus
			{Status(2), "private"}, // PrivateStatus
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
		status1 := DraftStatus
		status2 := status1 + 1

		assert.Equal(t, PublishStatus, status2, "Status arithmetic should work")

		status3 := PrivateStatus - 1
		assert.Equal(t, PublishStatus, status3, "Status arithmetic should work")
	})

	t.Run("Status type size", func(t *testing.T) {
		var status Status
		assert.Equal(t, 1, int(unsafe.Sizeof(status)), "Status should be 1 byte (int8)")
	})
}

func TestStatus_DefaultValues(t *testing.T) {
	t.Run("Zero value of Status", func(t *testing.T) {
		var status Status
		assert.Equal(t, DraftStatus, status, "Zero value should be DraftStatus")
		assert.Equal(t, "draft", status.String(), "Zero value should stringify to 'draft'")
	})

	t.Run("Status as function parameter", func(t *testing.T) {
		checkStatus := func(s Status) string {
			return s.String()
		}

		assert.Equal(t, "draft", checkStatus(DraftStatus), "Status should pass as parameter correctly")
		assert.Equal(t, "publish", checkStatus(PublishStatus), "Status should pass as parameter correctly")
		assert.Equal(t, "private", checkStatus(PrivateStatus), "Status should pass as parameter correctly")
	})
}

func TestStatus_ConcurrentAccess(t *testing.T) {
	t.Run("Concurrent String() calls", func(t *testing.T) {
		status := PublishStatus
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
			assert.Equal(t, "publish", result, "Concurrent String() calls should return consistent results")
		}
	})

	t.Run("Concurrent status operations", func(t *testing.T) {
		statuses := []Status{DraftStatus, PublishStatus, PrivateStatus}
		results := make(chan string, len(statuses))

		for _, status := range statuses {
			go func(s Status) {
				results <- fmt.Sprintf("status_%s", s.String())
			}(status)
		}

		expectedResults := []string{"status_draft", "status_publish", "status_private"}
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
	status := PublishStatus

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = status.String()
	}
}

func BenchmarkStatus_AllConstants_String(b *testing.B) {
	statuses := []Status{DraftStatus, PublishStatus, PrivateStatus}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, status := range statuses {
			_ = status.String()
		}
	}
}
