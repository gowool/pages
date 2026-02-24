package pages

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestFuncMap_PageURLFunction(t *testing.T) {
	tests := []struct {
		name          string
		setupMock     func(*MockURLGenerator)
		ctx           context.Context
		site          *Site
		arg           any
		args          []any
		expectedURL   string
		expectedError bool
	}{
		{
			name: "successful URL generation",
			setupMock: func(m *MockURLGenerator) {
				m.On("Generate", context.Background(), &Site{}, "test-page", []any(nil)).Return("https://example.com/test", nil)
			},
			ctx:         context.Background(),
			site:        &Site{},
			arg:         "test-page",
			expectedURL: "https://example.com/test",
		},
		{
			name: "URL generation with additional args",
			setupMock: func(m *MockURLGenerator) {
				m.On("Generate", context.Background(), &Site{}, "blog-post", []any{"category", "tech", "slug", "golang"}).Return("https://example.com/blog/tech/golang", nil)
			},
			ctx:         context.Background(),
			site:        &Site{},
			arg:         "blog-post",
			args:        []any{"category", "tech", "slug", "golang"},
			expectedURL: "https://example.com/blog/tech/golang",
		},
		{
			name: "URL generation returns error",
			setupMock: func(m *MockURLGenerator) {
				m.On("Generate", context.Background(), &Site{}, "invalid-page", []any(nil)).Return("", assert.AnError)
			},
			ctx:           context.Background(),
			site:          &Site{},
			arg:           "invalid-page",
			expectedURL:   "",
			expectedError: true,
		},
		{
			name: "nil context",
			setupMock: func(m *MockURLGenerator) {
				m.On("Generate", mock.Anything, &Site{}, "test-page", []any(nil)).Return("https://example.com/test", nil)
			},
			ctx:         nil,
			site:        &Site{},
			arg:         "test-page",
			expectedURL: "https://example.com/test",
		},
		{
			name: "nil site",
			setupMock: func(m *MockURLGenerator) {
				m.On("Generate", context.Background(), (*Site)(nil), "test-page", []any(nil)).Return("https://example.com/test", nil)
			},
			ctx:         context.Background(),
			site:        nil,
			arg:         "test-page",
			expectedURL: "https://example.com/test",
		},
		{
			name: "different argument types",
			setupMock: func(m *MockURLGenerator) {
				m.On("Generate", context.Background(), &Site{}, 123, []any(nil)).Return("https://example.com/numeric", nil)
			},
			ctx:         context.Background(),
			site:        &Site{},
			arg:         123,
			expectedURL: "https://example.com/numeric",
		},
		{
			name: "string argument with special characters",
			setupMock: func(m *MockURLGenerator) {
				m.On("Generate", context.Background(), &Site{}, "/path/with-special-chars?query=value", []any(nil)).Return("https://example.com/path/with-special-chars?query=value", nil)
			},
			ctx:         context.Background(),
			site:        &Site{},
			arg:         "/path/with-special-chars?query=value",
			expectedURL: "https://example.com/path/with-special-chars?query=value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockURLGenerator := &MockURLGenerator{}
			tt.setupMock(mockURLGenerator)

			funcMap := FuncMap(mockURLGenerator)
			pageURLFunc := funcMap["page_url"].(func(context.Context, *Site, any, ...any) string)

			result := pageURLFunc(tt.ctx, tt.site, tt.arg, tt.args...)

			if tt.expectedError {
				assert.Equal(t, tt.expectedURL, result, "Expected empty result due to error")
			} else {
				assert.Equal(t, tt.expectedURL, result, "page_url function should return correct URL")
			}

			mockURLGenerator.AssertExpectations(t)
		})
	}
}

func TestFuncMap_PageURLFunctionWithRealSite(t *testing.T) {
	mockURLGenerator := &MockURLGenerator{}

	// Create a realistic test site
	site := &Site{
		ID:     ID(uuid.New().String()),
		Scheme: "https",
		Host:   "test.example.com",
		Name:   "Test Site",
	}

	mockURLGenerator.On("Generate", context.Background(), site, "home-page", []any(nil)).Return("https://test.example.com/home", nil)

	funcMap := FuncMap(mockURLGenerator)
	pageURLFunc := funcMap["page_url"].(func(context.Context, *Site, any, ...any) string)

	result := pageURLFunc(context.Background(), site, "home-page")

	assert.Equal(t, "https://test.example.com/home", result, "page_url should work with realistic site data")
	mockURLGenerator.AssertExpectations(t)
}

func TestFuncMap_MultipleCalls(t *testing.T) {
	mockURLGenerator := &MockURLGenerator{}
	site := NewTestSite()
	ctx := context.Background()

	// Set up expectations for multiple calls
	mockURLGenerator.On("Generate", ctx, site, "page1", []any(nil)).Return("https://example.com/page1", nil)
	mockURLGenerator.On("Generate", ctx, site, "page2", []any(nil)).Return("https://example.com/page2", nil)
	mockURLGenerator.On("Generate", ctx, site, "page3", []any{"param1", "value1"}).Return("https://example.com/page3/value1", nil)

	funcMap := FuncMap(mockURLGenerator)
	pageURLFunc := funcMap["page_url"].(func(context.Context, *Site, any, ...any) string)

	// Make multiple calls
	result1 := pageURLFunc(ctx, site, "page1")
	result2 := pageURLFunc(ctx, site, "page2")
	result3 := pageURLFunc(ctx, site, "page3", "param1", "value1")

	assert.Equal(t, "https://example.com/page1", result1, "First call should return correct URL")
	assert.Equal(t, "https://example.com/page2", result2, "Second call should return correct URL")
	assert.Equal(t, "https://example.com/page3/value1", result3, "Third call with parameters should return correct URL")

	mockURLGenerator.AssertExpectations(t)
}

func TestFuncMap_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*MockURLGenerator)
		ctx         context.Context
		site        *Site
		arg         any
		args        []any
		expectPanic bool
	}{
		{
			name: "empty string argument",
			setupMock: func(m *MockURLGenerator) {
				m.On("Generate", mock.Anything, mock.Anything, "", []any(nil)).Return("", assert.AnError)
			},
			ctx:  context.Background(),
			site: &Site{},
			arg:  "",
		},
		{
			name: "nil argument",
			setupMock: func(m *MockURLGenerator) {
				m.On("Generate", mock.Anything, mock.Anything, (*string)(nil), []any(nil)).Return("", assert.AnError)
			},
			ctx:  context.Background(),
			site: &Site{},
			arg:  (*string)(nil),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockURLGenerator := &MockURLGenerator{}
			tt.setupMock(mockURLGenerator)

			funcMap := FuncMap(mockURLGenerator)
			pageURLFunc := funcMap["page_url"].(func(context.Context, *Site, any, ...any) string)

			if tt.expectPanic {
				assert.Panics(t, func() {
					pageURLFunc(tt.ctx, tt.site, tt.arg, tt.args...)
				}, "Expected panic for this test case")
			} else {
				// Should not panic, even if result is empty due to error
				result := pageURLFunc(tt.ctx, tt.site, tt.arg, tt.args...)
				// Result can be empty when URLGenerator returns error
				_ = result
			}

			mockURLGenerator.AssertExpectations(t)
		})
	}
}

func TestFuncMap_AttrFunctions(t *testing.T) {
	ctx, cancel := NewContext(context.Background())
	t.Cleanup(cancel)

	c := MustContext(ctx)
	c.SetSite(&Site{
		DOM: DOM{
			HTML: struct {
				Attr Attr `envPrefix:"ATTR_" json:"attr,omitempty" yaml:"attr,omitempty"`
			}{Attr: NewAttr("lang", "en", "class", "site")},
			Head: Head{Attr: NewAttr("data-head", "site")},
			Body: struct {
				Attr Attr `envPrefix:"ATTR_" json:"attr,omitempty" yaml:"attr,omitempty"`
			}{Attr: NewAttr("class", "site-body")},
		},
	})
	c.SetPage(&Page{
		DOM: DOM{
			HTML: struct {
				Attr Attr `envPrefix:"ATTR_" json:"attr,omitempty" yaml:"attr,omitempty"`
			}{Attr: NewAttr("lang", "fr", "id", "page")},
			Head: Head{Attr: NewAttr("data-head", "page")},
			Body: struct {
				Attr Attr `envPrefix:"ATTR_" json:"attr,omitempty" yaml:"attr,omitempty"`
			}{Attr: NewAttr("class", "page-body")},
		},
	})
	c.DOM().HTML.Attr = NewAttr("lang", "de", "dir", "ltr")
	c.DOM().Head.Attr = NewAttr("data-head", "ctx", "data-ctx", "1")
	c.DOM().Body.Attr = NewAttr("class", "ctx-body", "data-body", "1")

	html := string(htmlAttr(ctx))
	assert.Contains(t, html, ` lang="de"`)
	assert.Contains(t, html, ` dir="ltr"`)
	assert.Contains(t, html, ` id="page"`)

	head := string(headAttr(ctx))
	assert.Contains(t, head, ` data-head="ctx"`)
	assert.Contains(t, head, ` data-ctx="1"`)

	body := string(bodyAttr(ctx))
	assert.Contains(t, body, ` class="ctx-body"`)
	assert.Contains(t, body, ` data-body="1"`)
}

func TestFuncMap_HeadTags(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		ctx, cancel := NewContext(context.Background())
		t.Cleanup(cancel)

		assert.Equal(t, "", string(headTags(ctx)))
	})

	t.Run("site page and context nodes", func(t *testing.T) {
		ctx, cancel := NewContext(context.Background())
		t.Cleanup(cancel)

		c := MustContext(ctx)
		c.SetSite(&Site{
			DOM: DOM{
				Head: Head{
					Nodes: Nodes{
						TitleNode("Site"),
					},
				},
			},
		})
		c.SetPage(&Page{
			DOM: DOM{
				Head: Head{
					Nodes: Nodes{
						NameMetaNode("description", "Page"),
					},
				},
			},
		})
		c.DOM().Head.Add(HeadLinkNode(HeadLink{Rel: LinkRelCanonical, Href: "https://example.com"}))

		got := string(headTags(ctx))
		assert.Contains(t, got, "<title>Site</title>")
		assert.Contains(t, got, `name="description"`)
		assert.Contains(t, got, `rel="canonical"`)
		assert.True(t, strings.Index(got, "<title>Site</title>") < strings.Index(got, `name="description"`))
	})
}

func TestFuncMap_TitleFunctions(t *testing.T) {
	t.Run("title and reverse title with defaults", func(t *testing.T) {
		ctx, cancel := NewContext(context.Background())
		t.Cleanup(cancel)

		c := MustContext(ctx)
		c.SetSite(&Site{Title: "Site"})
		c.SetPage(&Page{Title: "Page"})

		assert.Equal(t, "<title>Site - Page - Extra</title>", string(titleTag(ctx, "Extra")))
		assert.Equal(t, "<title>Extra - Page - Site</title>", string(reverseTitleTag(ctx, "Extra")))
	})

	t.Run("title separator from site and strips tags", func(t *testing.T) {
		ctx, cancel := NewContext(context.Background())
		t.Cleanup(cancel)

		c := MustContext(ctx)
		c.SetSite(&Site{Title: "<b>Site</b>", Separator: " | "})
		c.SetPage(&Page{Title: "&lt;i&gt;Page&lt;/i&gt;"})

		assert.Equal(t, "<title>Site | Page</title>", string(titleTag(ctx)))
	})

	t.Run("title data without site/page title", func(t *testing.T) {
		ctx, cancel := NewContext(context.Background())
		t.Cleanup(cancel)

		data, sep := titleData(ctx, "A", "B")
		assert.Equal(t, defaultSeparator, sep)
		assert.Equal(t, []string{"A", "B"}, data)
	})
}

func TestFuncMap_StripTags(t *testing.T) {
	got := stripTags("&lt;b&gt;Hello&lt;/b&gt; <i>World</i>")
	assert.Equal(t, "Hello World", got)
}
