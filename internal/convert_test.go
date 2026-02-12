package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBytesToString(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		want  string
	}{
		{
			name:  "empty byte slice",
			input: []byte{},
			want:  "",
		},
		{
			name:  "nil byte slice",
			input: nil,
			want:  "",
		},
		{
			name:  "simple ASCII string",
			input: []byte("hello world"),
			want:  "hello world",
		},
		{
			name:  "string with special characters",
			input: []byte("hello!@#$%^&*()"),
			want:  "hello!@#$%^&*()",
		},
		{
			name:  "string with numbers",
			input: []byte("1234567890"),
			want:  "1234567890",
		},
		{
			name:  "string with mixed content",
			input: []byte("Test123!@#"),
			want:  "Test123!@#",
		},
		{
			name:  "string with spaces and tabs",
			input: []byte("hello\tworld\nfoo bar"),
			want:  "hello\tworld\nfoo bar",
		},
		{
			name:  "single character",
			input: []byte("a"),
			want:  "a",
		},
		{
			name:  "string with unicode characters",
			input: []byte("Hello 世界"),
			want:  "Hello 世界",
		},
		{
			name:  "string with emoji",
			input: []byte("Hello 😀🎉"),
			want:  "Hello 😀🎉",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BytesToString(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBytesToStringZeroCopy(t *testing.T) {
	t.Run("zero copy conversion", func(t *testing.T) {
		b := []byte("test string")
		s := BytesToString(b)

		assert.Equal(t, "test string", s)

		b[0] = 'T'
		assert.Equal(t, "Test string", s)
	})
}

func BenchmarkBytesToString(b *testing.B) {
	data := []byte("test string for benchmark")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = BytesToString(data)
	}
}
