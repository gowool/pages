package pages

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrors(t *testing.T) {
	t.Run("ErrSiteNotFound is defined", func(t *testing.T) {
		assert.NotNil(t, ErrSiteNotFound)
		assert.Equal(t, "site not found", ErrSiteNotFound.Error())
	})

	t.Run("ErrPageNotFound is defined", func(t *testing.T) {
		assert.NotNil(t, ErrPageNotFound)
		assert.Equal(t, "page not found", ErrPageNotFound.Error())
	})

	t.Run("ErrPageForbidden is defined", func(t *testing.T) {
		assert.NotNil(t, ErrPageForbidden)
		assert.Equal(t, "page forbidden", ErrPageForbidden.Error())
	})

	t.Run("ErrPageUnauthorized is defined", func(t *testing.T) {
		assert.NotNil(t, ErrPageUnauthorized)
		assert.Equal(t, "page unauthorized", ErrPageUnauthorized.Error())
	})

	t.Run("ErrUniqueViolation is defined", func(t *testing.T) {
		assert.NotNil(t, ErrUniqueViolation)
		assert.Equal(t, "unique violation", ErrUniqueViolation.Error())
	})

	t.Run("ErrTemplateEmpty is defined", func(t *testing.T) {
		assert.NotNil(t, ErrTemplateEmpty)
		assert.Equal(t, "template is empty", ErrTemplateEmpty.Error())
	})
}

func TestNewRedirectError(t *testing.T) {
	tests := []struct {
		name   string
		url    string
		status int
	}{
		{
			name:   "301 Moved Permanently",
			url:    "/new-location",
			status: http.StatusMovedPermanently,
		},
		{
			name:   "302 Found",
			url:    "/temp-location",
			status: http.StatusFound,
		},
		{
			name:   "307 Temporary Redirect",
			url:    "/temp-redirect",
			status: http.StatusTemporaryRedirect,
		},
		{
			name:   "308 Permanent Redirect",
			url:    "/perm-redirect",
			status: http.StatusPermanentRedirect,
		},
		{
			name:   "external URL",
			url:    "https://example.com",
			status: http.StatusSeeOther,
		},
		{
			name:   "relative URL",
			url:    "/path/to/page",
			status: http.StatusMovedPermanently,
		},
		{
			name:   "URL with query string",
			url:    "/page?param=value",
			status: http.StatusFound,
		},
		{
			name:   "URL with fragment",
			url:    "/page#section",
			status: http.StatusSeeOther,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewRedirectError(tt.url, tt.status)

			assert.NotNil(t, err)
			assert.Equal(t, tt.url, err.url)
			assert.Equal(t, tt.status, err.status)
		})
	}
}

func TestRedirectError_Error(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		status   int
		expected string
	}{
		{
			name:     "301 redirect",
			url:      "/new",
			status:   http.StatusMovedPermanently,
			expected: "[301] /new",
		},
		{
			name:     "302 redirect",
			url:      "/temp",
			status:   http.StatusFound,
			expected: "[302] /temp",
		},
		{
			name:     "external URL",
			url:      "https://example.com",
			status:   http.StatusSeeOther,
			expected: "[303] https://example.com",
		},
		{
			name:     "URL with query",
			url:      "/page?foo=bar",
			status:   http.StatusTemporaryRedirect,
			expected: "[307] /page?foo=bar",
		},
		{
			name:     "URL with fragment",
			url:      "/page#section",
			status:   http.StatusPermanentRedirect,
			expected: "[308] /page#section",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewRedirectError(tt.url, tt.status)
			got := err.Error()

			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestRedirectError_Redirect(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		status     int
		wantURL    string
		wantStatus int
	}{
		{
			name:       "301 redirect",
			url:        "/permanent",
			status:     http.StatusMovedPermanently,
			wantURL:    "/permanent",
			wantStatus: http.StatusMovedPermanently,
		},
		{
			name:       "302 redirect",
			url:        "/temporary",
			status:     http.StatusFound,
			wantURL:    "/temporary",
			wantStatus: http.StatusFound,
		},
		{
			name:       "303 redirect",
			url:        "https://example.com",
			status:     http.StatusSeeOther,
			wantURL:    "https://example.com",
			wantStatus: http.StatusSeeOther,
		},
		{
			name:       "307 redirect",
			url:        "/preserve-method",
			status:     http.StatusTemporaryRedirect,
			wantURL:    "/preserve-method",
			wantStatus: http.StatusTemporaryRedirect,
		},
		{
			name:       "308 redirect",
			url:        "/permanent-method",
			status:     http.StatusPermanentRedirect,
			wantURL:    "/permanent-method",
			wantStatus: http.StatusPermanentRedirect,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewRedirectError(tt.url, tt.status)
			gotURL, gotStatus := err.Redirect()

			assert.Equal(t, tt.wantURL, gotURL)
			assert.Equal(t, tt.wantStatus, gotStatus)
		})
	}
}

func TestRedirectError_ImplementsErrorInterface(t *testing.T) {
	var _ error = (*RedirectError)(nil)
}

func TestRedirectError_ConstantValues(t *testing.T) {
	t.Run("HTTP status codes are correct", func(t *testing.T) {
		assert.Equal(t, 301, http.StatusMovedPermanently)
		assert.Equal(t, 302, http.StatusFound)
		assert.Equal(t, 303, http.StatusSeeOther)
		assert.Equal(t, 307, http.StatusTemporaryRedirect)
		assert.Equal(t, 308, http.StatusPermanentRedirect)
	})
}

func TestRedirectError_ZeroValues(t *testing.T) {
	t.Run("zero struct returns empty error message", func(t *testing.T) {
		err := &RedirectError{}
		got := err.Error()

		assert.Equal(t, "[0] ", got)
	})

	t.Run("zero struct returns empty redirect", func(t *testing.T) {
		err := &RedirectError{}
		url, status := err.Redirect()

		assert.Equal(t, "", url)
		assert.Equal(t, 0, status)
	})
}

func TestRedirectError_ValueAndPointerReceiver(t *testing.T) {
	t.Run("Error() works with both value and pointer", func(t *testing.T) {
		valueErr := RedirectError{url: "/test", status: 302}
		pointerErr := NewRedirectError("/test", 302)

		assert.Equal(t, valueErr.Error(), pointerErr.Error())
	})

	t.Run("Redirect() works with both value and pointer", func(t *testing.T) {
		valueErr := RedirectError{url: "/test", status: 302}
		pointerErr := NewRedirectError("/test", 302)

		valueURL, valueStatus := valueErr.Redirect()
		pointerURL, pointerStatus := pointerErr.Redirect()

		assert.Equal(t, valueURL, pointerURL)
		assert.Equal(t, valueStatus, pointerStatus)
	})
}

func TestRedirectError_ConstructorConsistency(t *testing.T) {
	t.Run("NewRedirectError matches struct initialization", func(t *testing.T) {
		constructorErr := NewRedirectError("/test", 302)
		structErr := &RedirectError{url: "/test", status: 302}

		assert.Equal(t, structErr.url, constructorErr.url)
		assert.Equal(t, structErr.status, constructorErr.status)
		assert.Equal(t, structErr.Error(), constructorErr.Error())
		structURL, structStatus := structErr.Redirect()
		constructorURL, constructorStatus := constructorErr.Redirect()
		assert.Equal(t, structURL, constructorURL)
		assert.Equal(t, structStatus, constructorStatus)
	})
}

func TestRedirectError_Format(t *testing.T) {
	t.Run("Error() uses Sprintf format", func(t *testing.T) {
		err := NewRedirectError("/path", 307)
		expected := fmt.Sprintf("[%d] %s", 307, "/path")
		got := err.Error()

		assert.Equal(t, expected, got)
	})
}

func TestErrorVariableTypes(t *testing.T) {
	t.Run("all errors are of type error", func(t *testing.T) {
		var err error

		err = ErrSiteNotFound
		assert.NotNil(t, err)

		err = ErrPageNotFound
		assert.NotNil(t, err)

		err = ErrPageForbidden
		assert.NotNil(t, err)

		err = ErrPageUnauthorized
		assert.NotNil(t, err)

		err = ErrUniqueViolation
		assert.NotNil(t, err)

		err = ErrTemplateEmpty
		assert.NotNil(t, err)
	})

	t.Run("RedirectError implements error interface", func(t *testing.T) {
		var err error
		err = NewRedirectError("/test", 302)
		assert.NotNil(t, err)
		assert.Equal(t, "[302] /test", err.Error())
	})
}

func TestErrorVariableUniqueness(t *testing.T) {
	t.Run("all error variables are unique", func(t *testing.T) {
		errors := []error{
			ErrSiteNotFound,
			ErrPageNotFound,
			ErrPageForbidden,
			ErrPageUnauthorized,
			ErrUniqueViolation,
			ErrTemplateEmpty,
		}

		for i, e1 := range errors {
			for j, e2 := range errors {
				if i == j {
					continue
				}
				assert.NotEqual(t, e1, e2, "errors at indices %d and %d should be different", i, j)
			}
		}
	})
}
