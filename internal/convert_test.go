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

func TestToTitle(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "single path segment",
			input: "/blog",
			want:  "BLOG",
		},
		{
			name:  "multiple path segments",
			input: "/blog/posts",
			want:  "BLOG POSTS",
		},
		{
			name:  "deep path",
			input: "/blog/posts/2024",
			want:  "BLOG POSTS 2024",
		},
		{
			name:  "path with kebab-case",
			input: "/blog/my-post",
			want:  "BLOG MY-POST",
		},
		{
			name:  "root path",
			input: "/",
			want:  "",
		},
		{
			name:  "path without leading slash",
			input: "blog/posts",
			want:  "BLOG POSTS",
		},
		{
			name:  "multiple leading slashes",
			input: "//blog/posts",
			want:  "BLOG POSTS",
		},
		{
			name:  "path with consecutive slashes",
			input: "/blog//posts",
			want:  "BLOG  POSTS",
		},
		{
			name:  "path with trailing slash",
			input: "/blog/posts/",
			want:  "BLOG POSTS ",
		},
		{
			name:  "path with numbers",
			input: "/api/v1/users",
			want:  "API V1 USERS",
		},
		{
			name:  "single character segments",
			input: "/a/b/c",
			want:  "A B C",
		},
		{
			name:  "mixed case input",
			input: "/Blog/Posts",
			want:  "BLOG POSTS",
		},
		{
			name:  "path with underscores",
			input: "/my_blog_posts",
			want:  "MY_BLOG_POSTS",
		},
		{
			name:  "complex api path",
			input: "/api/v1/users/{id}/posts",
			want:  "API V1 USERS {ID} POSTS",
		},
		{
			name:  "very long path",
			input: "/this/is/a/very/long/path/with/many/segments",
			want:  "THIS IS A VERY LONG PATH WITH MANY SEGMENTS",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToTitle(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestToTitleIntegration(t *testing.T) {
	t.Run("ToTitle on BytesToString result", func(t *testing.T) {
		b := []byte("/blog/posts/2024")
		s := BytesToString(b)
		title := ToTitle(s)

		assert.Equal(t, "BLOG POSTS 2024", title)
	})

	t.Run("BytesToString on ToTitle result", func(t *testing.T) {
		path := "/blog/posts"
		title := ToTitle(path)
		b := []byte(title)
		result := BytesToString(b)

		assert.Equal(t, "BLOG POSTS", result)
	})
}

func BenchmarkToTitle(b *testing.B) {
	paths := []string{
		"/blog/posts",
		"/api/v1/users",
		"/blog/my-post",
		"/",
		"",
	}

	for _, path := range paths {
		b.Run(path, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = ToTitle(path)
			}
		})
	}
}

func BenchmarkToTitleWithBytesToString(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		path := []byte("/blog/posts/2024")
		s := BytesToString(path)
		_ = ToTitle(s)
	}
}
