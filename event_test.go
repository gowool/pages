package pages

import (
	"html/template"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gowool/wo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockParamFunc is a mock parameter function for testing
func MockParamFunc(param string) string {
	switch param {
	case "slug":
		return "test-slug"
	case "id":
		return "123"
	case "category":
		return "tech"
	case "param":
		return "test-slug"
	case "another":
		return "123"
	default:
		return "mock-value"
	}
}

func TestEvent_Reset(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*Event)
		validate func(*testing.T, *Event)
	}{
		{
			name: "reset with initialized event",
			setup: func(e *Event) {
				e.seo = NewSEO()
				e.seo.SetTitle("Existing Title")
				e.theme = &MockTheme{}
			},
			validate: func(t *testing.T, e *Event) {
				assert.NotNil(t, e.seo)
				assert.Empty(t, e.seo.Title(), "SEO title should be reset")
				assert.NotNil(t, e.theme, "Theme should be set")
			},
		},
		{
			name: "reset with nil event",
			setup: func(e *Event) {
				e.seo = nil
				e.theme = nil
			},
			validate: func(t *testing.T, e *Event) {
				assert.NotNil(t, e.seo, "SEO should be initialized")
				assert.NotNil(t, e.theme, "Theme should be set")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/", nil)
			theme := &MockTheme{}

			e := &Event{}
			tt.setup(e)

			e.Reset(&wo.Response{ResponseWriter: w}, r, theme)
			tt.validate(t, e)
		})
	}
}

func TestEvent_IsRoot(t *testing.T) {
	tests := []struct {
		name   string
		url    string
		isRoot bool
	}{
		{
			name:   "root path",
			url:    "/",
			isRoot: true,
		},
		{
			name:   "non-root path",
			url:    "/about",
			isRoot: false,
		},
		{
			name:   "nested path",
			url:    "/blog/post/123",
			isRoot: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", tt.url, nil)
			// Set RawPath explicitly since httptest doesn't set it
			r.URL.RawPath = tt.url
			theme := &MockTheme{}

			e := &Event{}
			e.Reset(w, r, theme)

			assert.Equal(t, tt.isRoot, e.IsRoot())
		})
	}
}

func TestEvent_SEO(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*Event)
		validate func(*testing.T, *SEO)
	}{
		{
			name: "initialize new SEO",
			setup: func(e *Event) {
				e.seo = nil
			},
			validate: func(t *testing.T, seo *SEO) {
				assert.NotNil(t, seo)
				assert.Empty(t, seo.Title())
			},
		},
		{
			name: "return existing SEO",
			setup: func(e *Event) {
				e.seo = NewSEO()
				e.seo.SetTitle("Existing Title")
			},
			validate: func(t *testing.T, seo *SEO) {
				assert.NotNil(t, seo)
				assert.Equal(t, "Existing Title", seo.Title())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/", nil)
			theme := &MockTheme{}

			e := &Event{}
			e.Reset(w, r, theme)
			tt.setup(e)

			seo := e.SEO()
			tt.validate(t, seo)
		})
	}
}

func TestEvent_Pattern(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		expected string
	}{
		{
			name:     "simple pattern",
			pattern:  "GET /about",
			expected: "/about",
		},
		{
			name:     "pattern with space",
			pattern:  "POST /api/users",
			expected: "/api/users",
		},
		{
			name:     "pattern with multiple spaces",
			pattern:  "DELETE /api/users/{id}",
			expected: "/api/users/{id}",
		},
		{
			name:     "pattern without space",
			pattern:  "/contact",
			expected: "/contact",
		},
		{
			name:     "empty pattern",
			pattern:  "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/", nil)
			r.Pattern = tt.pattern
			theme := &MockTheme{}

			e := &Event{}
			e.Reset(w, r, theme)

			assert.Equal(t, tt.expected, e.Pattern())
		})
	}
}

func TestEvent_SetGuest_IsGuest(t *testing.T) {
	tests := []struct {
		name     string
		auth     bool // The value passed to SetGuest (represents authentication status)
		expected bool // Expected result from IsGuest()
	}{
		{
			name:     "set as guest (not authenticated)",
			auth:     false, // SetGuest(false) means not authenticated
			expected: true,  // IsGuest() should return true
		},
		{
			name:     "set as authenticated",
			auth:     true,  // SetGuest(true) means authenticated
			expected: false, // IsGuest() should return false
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/", nil)
			theme := &MockTheme{}

			e := &Event{}
			e.Reset(w, r, theme)

			e.SetGuest(tt.auth)
			assert.Equal(t, tt.expected, e.IsGuest())
		})
	}
}

func TestEvent_SetSite_HasSite(t *testing.T) {
	tests := []struct {
		name     string
		site     *Site
		expected bool
		validate func(*testing.T, *Event)
	}{
		{
			name:     "set nil site",
			site:     nil,
			expected: false,
		},
		{
			name: "set valid site",
			site: &Site{
				Host: "example.com",
			},
			expected: true,
			validate: func(t *testing.T, e *Event) {
				assert.Equal(t, "example.com", e.Site().Host)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/", nil)
			theme := &MockTheme{}

			e := &Event{}
			e.Reset(w, r, theme)

			e.SetSite(tt.site)
			assert.Equal(t, tt.expected, e.HasSite())

			if tt.validate != nil {
				tt.validate(t, e)
			}
		})
	}
}

func TestEvent_SetPage_HasPage(t *testing.T) {
	tests := []struct {
		name     string
		page     *Page
		expected bool
		validate func(*testing.T, *Event)
	}{
		{
			name:     "set nil page",
			page:     nil,
			expected: false,
		},
		{
			name: "set static page",
			page: &Page{
				ID:    "1",
				Title: "About Us",
				URL:   "/about",
			},
			expected: true,
			validate: func(t *testing.T, e *Event) {
				assert.Equal(t, ID("1"), e.Page().ID)
				assert.Equal(t, "About Us", e.Page().Title)
			},
		},
		{
			name: "set dynamic page",
			page: &Page{
				ID:      "2",
				Title:   "Blog Post",
				URL:     "/blog/test-slug",
				Pattern: "/blog/{slug}",
			},
			expected: true,
			validate: func(t *testing.T, e *Event) {
				assert.Equal(t, ID("2"), e.Page().ID)
				assert.True(t, e.Page().IsDynamic())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/blog/test-slug", nil)
			r.Pattern = "/blog/{slug}"

			theme := &MockTheme{}

			e := &Event{}
			e.Reset(w, r, theme)

			// Set the site first for the page
			site := &Site{Host: "example.com"}
			e.SetSite(site)

			e.SetPage(tt.page)
			assert.Equal(t, tt.expected, e.HasPage())

			if tt.validate != nil {
				tt.validate(t, e)
			}
		})
	}
}

func TestEvent_SetStatus_Status(t *testing.T) {
	tests := []struct {
		name     string
		status   int
		expected int
	}{
		{
			name:     "set OK status",
			status:   http.StatusOK,
			expected: http.StatusOK,
		},
		{
			name:     "set Not Found status",
			status:   http.StatusNotFound,
			expected: http.StatusNotFound,
		},
		{
			name:     "set Internal Server Error status",
			status:   http.StatusInternalServerError,
			expected: http.StatusInternalServerError,
		},
		{
			name:     "default status when not set",
			status:   0,
			expected: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/", nil)
			theme := &MockTheme{}

			e := &Event{}
			e.Reset(w, r, theme)

			if tt.status > 0 {
				e.SetStatus(tt.status)
			}

			assert.Equal(t, tt.expected, e.Status())
		})
	}
}

func TestEvent_SetError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected error
	}{
		{
			name:     "set nil error",
			err:      nil,
			expected: nil,
		},
		{
			name:     "set custom error",
			err:      assert.AnError,
			expected: assert.AnError,
		},
		{
			name:     "set error message",
			err:      assert.AnError,
			expected: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/", nil)
			theme := &MockTheme{}

			e := &Event{}
			e.Reset(w, r, theme)

			e.SetError(tt.err)
			assert.Equal(t, tt.expected, e.Error())
		})
	}
}

func TestEvent_SetContent_Content(t *testing.T) {
	tests := []struct {
		name     string
		content  template.HTML
		expected template.HTML
	}{
		{
			name:     "set HTML content",
			content:  "<h1>Hello World</h1>",
			expected: "<h1>Hello World</h1>",
		},
		{
			name:     "set empty content",
			content:  "",
			expected: "",
		},
		{
			name:     "set complex HTML",
			content:  template.HTML("<div><p>Paragraph</p></div>"),
			expected: template.HTML("<div><p>Paragraph</p></div>"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/", nil)
			theme := &MockTheme{}

			e := &Event{}
			e.Reset(w, r, theme)

			e.SetContent(tt.content)
			assert.Equal(t, tt.expected, e.Content())
		})
	}
}

func TestEvent_IsDecorable(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*Event)
		expected bool
	}{
		{
			name: "HTML content with OK status is decorable",
			setup: func(e *Event) {
				e.Response().Header().Set(wo.HeaderContentType, wo.MIMETextHTML)
				e.Response().WriteHeader(http.StatusOK)
				e.Request().Header.Del(wo.HeaderXRequestedWith)
			},
			expected: true,
		},
		{
			name: "JSON content is not decorable",
			setup: func(e *Event) {
				e.Response().Header().Set(wo.HeaderContentType, wo.MIMEApplicationJSON)
				e.Response().WriteHeader(http.StatusOK)
			},
			expected: false,
		},
		{
			name: "Explicitly not decorable",
			setup: func(e *Event) {
				e.Response().Header().Set(HeaderXPageNotDecorable, "1")
				e.Response().WriteHeader(http.StatusOK)
			},
			expected: false,
		},
		{
			name: "Explicitly decorable",
			setup: func(e *Event) {
				e.Response().Header().Set(HeaderXPageDecorable, "1")
				e.Response().WriteHeader(http.StatusOK)
			},
			expected: true,
		},
		{
			name: "Non-OK status is not decorable",
			setup: func(e *Event) {
				e.Response().Header().Set(wo.HeaderContentType, wo.MIMETextHTML)
				e.Response().WriteHeader(http.StatusNotFound)
			},
			expected: false,
		},
		{
			name: "Ajax request is not decorable",
			setup: func(e *Event) {
				e.Response().Header().Set(wo.HeaderContentType, wo.MIMETextHTML)
				e.Response().WriteHeader(http.StatusOK)
				e.Request().Header.Set(wo.HeaderXRequestedWith, wo.XMLHTTPRequest)
			},
			expected: false,
		},
		{
			name: "Empty content type with OK status is decorable",
			setup: func(e *Event) {
				e.Response().WriteHeader(http.StatusOK)
				e.Request().Header.Del(wo.HeaderXRequestedWith)
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/", nil)
			theme := &MockTheme{}

			e := &Event{}
			e.Reset(w, r, theme)

			tt.setup(e)

			assert.Equal(t, tt.expected, e.IsDecorable())
		})
	}
}

func TestGetPattern(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		expected string
	}{
		{
			name:     "pattern with HTTP method",
			pattern:  "GET /home",
			expected: "/home",
		},
		{
			name:     "pattern with multiple words",
			pattern:  "POST /api/v1/users",
			expected: "/api/v1/users",
		},
		{
			name:     "pattern without space",
			pattern:  "/contact",
			expected: "/contact",
		},
		{
			name:     "empty pattern",
			pattern:  "",
			expected: "",
		},
		{
			name:     "pattern with only space",
			pattern:  " ",
			expected: "",
		},
		{
			name:     "complex pattern",
			pattern:  "PUT /api/users/{id}/profile",
			expected: "/api/users/{id}/profile",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", "/", nil)
			r.Pattern = tt.pattern

			assert.Equal(t, tt.expected, getPattern(r))
		})
	}
}

func TestPatternArgs(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		expected []any
	}{
		{
			name:     "static pattern",
			pattern:  "/about",
			expected: nil,
		},
		{
			name:     "single parameter",
			pattern:  "/blog/{slug}",
			expected: []any{"{slug}", "test-slug"},
		},
		{
			name:     "multiple parameters",
			pattern:  "/category/{category}/post/{id}",
			expected: []any{"{category}", "tech", "{id}", "123"},
		},
		{
			name:     "parameter with dots",
			pattern:  "/api/{.id}",
			expected: []any{"{id}", "123"},
		},
		{
			name:     "mixed pattern",
			pattern:  "/static/{param}/more/{another}/fixed",
			expected: []any{"{param}", "test-slug", "{another}", "123"},
		},
		{
			name:    "empty braces should panic",
			pattern: "/api/}{",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "empty braces should panic" {
				assert.Panics(t, func() {
					patternArgs(MockParamFunc, tt.pattern)
				})
				return
			}

			args := patternArgs(MockParamFunc, tt.pattern)
			assert.Equal(t, tt.expected, args)
		})
	}
}

func TestEvent_View_Render(t *testing.T) {
	tests := []struct {
		name           string
		templateName   string
		templateData   any
		themeError     error
		expectedError  bool
		expectedOutput string
		viewError      error
	}{
		{
			name:           "successful view rendering",
			templateName:   "index.html",
			templateData:   "test data",
			expectedOutput: "rendered content",
		},
		{
			name:          "theme returns error",
			templateName:  "error.html",
			themeError:    assert.AnError,
			expectedError: true,
		},
		{
			name:          "view returns error",
			templateName:  "index.html",
			expectedError: true,
			viewError:     ErrThemeRequired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/", nil)

			e := &Event{}

			var mockTheme *MockTheme
			if tt.viewError == nil {
				mockTheme = &MockTheme{}
				if tt.themeError != nil {
					mockTheme.On("Write", mock.Anything, mock.Anything, tt.templateName, mock.Anything).Return(tt.themeError)
				} else {
					mockTheme.On("Write", mock.Anything, mock.Anything, tt.templateName, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
						writer := args.Get(1).(io.Writer)
						_, _ = writer.Write([]byte(tt.expectedOutput))
					})
				}
				e.Reset(w, r, mockTheme)
			} else {
				e.Reset(w, r, nil)
			}

			data, err := e.View(tt.templateName, tt.templateData)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, data)
				if tt.viewError != nil {
					assert.ErrorIs(t, err, tt.viewError)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedOutput, string(data))
			}

			if tt.viewError == nil {
				mockTheme.AssertExpectations(t)
			}
		})
	}
}

func TestEvent_Render_RenderHTML(t *testing.T) {
	tests := []struct {
		name          string
		status        int
		contentType   string
		template      string
		expectedError bool
		setupMock     func(*MockTheme)
	}{
		{
			name:        "successful HTML render",
			status:      http.StatusOK,
			contentType: wo.MIMETextHTMLCharsetUTF8,
			template:    "index.html",
			setupMock: func(m *MockTheme) {
				m.On("Write", mock.Anything, mock.Anything, "index.html", mock.Anything).Return(nil).Run(func(args mock.Arguments) {
					writer := args.Get(1).(io.Writer)
					_, _ = writer.Write([]byte("<html><body>Test</body></html>"))
				})
			},
		},
		{
			name:          "theme write error",
			status:        http.StatusOK,
			contentType:   wo.MIMETextHTMLCharsetUTF8,
			template:      "error.html",
			expectedError: true,
			setupMock: func(m *MockTheme) {
				m.On("Write", mock.Anything, mock.Anything, "error.html", mock.Anything).Return(assert.AnError)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/", nil)

			mockTheme := &MockTheme{}
			tt.setupMock(mockTheme)

			e := &Event{}
			e.Reset(w, r, mockTheme)

			var err error
			if tt.contentType == wo.MIMETextHTMLCharsetUTF8 {
				err = e.RenderHTML(tt.status, tt.template)
			} else {
				err = e.Render(tt.status, tt.contentType, tt.template)
			}

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.status, w.Code)
				assert.Equal(t, tt.contentType, w.Header().Get(wo.HeaderContentType))
			}

			mockTheme.AssertExpectations(t)
		})
	}
}

// Test Event implements Resolver interface
func TestEvent_SetTheme_Theme(t *testing.T) {
	tests := []struct {
		name     string
		theme    Theme
		validate func(*testing.T, *Event, Theme)
	}{
		{
			name:  "set mock theme",
			theme: &MockTheme{},
			validate: func(t *testing.T, e *Event, expected Theme) {
				assert.Equal(t, expected, e.Theme())
			},
		},
		{
			name:  "set nil theme",
			theme: nil,
			validate: func(t *testing.T, e *Event, expected Theme) {
				assert.Nil(t, e.Theme())
			},
		},
		{
			name:  "set mock page theme",
			theme: &MockPageTheme{},
			validate: func(t *testing.T, e *Event, expected Theme) {
				assert.Equal(t, expected, e.Theme())
			},
		},
		{
			name:  "set theme and retrieve multiple times",
			theme: &MockTheme{},
			validate: func(t *testing.T, e *Event, expected Theme) {
				assert.Equal(t, expected, e.Theme())
				assert.Equal(t, expected, e.Theme())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/", nil)
			theme := &MockTheme{}

			e := &Event{}
			e.Reset(w, r, theme)

			e.SetTheme(tt.theme)
			tt.validate(t, e, tt.theme)
		})
	}
}

func TestEvent_ResolverImplementation(t *testing.T) {
	var _ Resolver = (*Event)(nil)
}

func BenchmarkPatternArgs(b *testing.B) {
	pattern := "/category/{category}/post/{id}/comment/{commentId}"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		patternArgs(MockParamFunc, pattern)
	}
}
