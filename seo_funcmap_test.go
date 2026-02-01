package pages

import (
	"fmt"
	"html/template"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSEOFuncMap(t *testing.T) {
	funcMap := SEOFuncMap

	// Test that all expected functions are present
	expectedFunctions := []string{
		"strip_tags",
		"escape_double_q",
		"reverse_title_tag",
		"title_tag",
		"meta_tags",
		"html_attrs",
		"head_attrs",
		"body_attrs",
		"lang_alternates",
		"head_links",
	}

	for _, funcName := range expectedFunctions {
		t.Run(funcName+"_exists", func(t *testing.T) {
			assert.Contains(t, funcMap, funcName, "SEOFuncMap should contain %s function", funcName)
			assert.NotNil(t, funcMap[funcName], "%s function should not be nil", funcName)
		})
	}
}

func TestStripTags(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Simple HTML tags",
			input:    "<p>Hello <b>world</b></p>",
			expected: "Hello world",
		},
		{
			name:     "No HTML tags",
			input:    "Plain text",
			expected: "Plain text",
		},
		{
			name:     "Nested HTML tags",
			input:    "<div><span>Hello</span> <em>world</em></div>",
			expected: "Hello world",
		},
		{
			name:     "Self-closing tags",
			input:    "Hello<br/>world<hr/>test",
			expected: "Helloworldtest",
		},
		{
			name:     "HTML with attributes",
			input:    `<a href="https://example.com">Link</a>`,
			expected: "Link",
		},
		{
			name:     "HTML entities",
			input:    "Hello &amp; world &lt;test&gt;",
			expected: "Hello & world ",
		},
		{
			name:     "Multiline HTML",
			input:    "<div>\n  <p>Hello</p>\n  <span>world</span>\n</div>",
			expected: "\n  Hello\n  world\n",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "Only HTML tags",
			input:    "<div><p></p></div>",
			expected: "",
		},
		{
			name:     "Complex HTML with script/style",
			input:    `<script>alert("test")</script><style>body { color: red; }</style><p>Content</p>`,
			expected: `alert("test")body { color: red; }Content`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stripTags(tt.input)
			assert.Equal(t, tt.expected, result, "stripTags() should produce expected output")
		})
	}
}

func TestEscapeDoubleQuotes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Single double quote",
			input:    `Hello "world"`,
			expected: "Hello &quot;world&quot;",
		},
		{
			name:     "Multiple double quotes",
			input:    `"Hello" "world" "test"`,
			expected: "&quot;Hello&quot; &quot;world&quot; &quot;test&quot;",
		},
		{
			name:     "No double quotes",
			input:    "Hello world",
			expected: "Hello world",
		},
		{
			name:     "Mixed quotes",
			input:    `Hello 'world' "test"`,
			expected: "Hello 'world' &quot;test&quot;",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "Only double quotes",
			input:    `"""`,
			expected: "&quot;&quot;&quot;",
		},
		{
			name:     "Double quotes with HTML",
			input:    `<a href="https://example.com">Link "text"</a>`,
			expected: `<a href=&quot;https://example.com&quot;>Link &quot;text&quot;</a>`,
		},
		{
			name:     "Adjacent double quotes",
			input:    `Hello""world`,
			expected: "Hello&quot;&quot;world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := escapeDoubleQuotes(tt.input)
			assert.Equal(t, tt.expected, result, "escapeDoubleQuotes() should produce expected output")
		})
	}
}

func TestTitleTag(t *testing.T) {
	tests := []struct {
		name     string
		seo      *SEO
		args     []string
		expected string
	}{
		{
			name:     "Basic title",
			seo:      createTestSEO("Test Title"),
			args:     nil,
			expected: "<title>Test Title</title>",
		},
		{
			name:     "Title with additional args",
			seo:      createTestSEO("Page Title"),
			args:     []string{"Additional", "Info"},
			expected: "<title>Page Title - Additional - Info</title>",
		},
		{
			name:     "Empty title with args",
			seo:      createTestSEO(""),
			args:     []string{"Only", "Args"},
			expected: "<title> - Only - Args</title>",
		},
		{
			name:     "Title with HTML tags",
			seo:      createTestSEO("<h1>Page Title</h1>"),
			args:     nil,
			expected: "<title>Page Title</title>",
		},
		{
			name:     "Title with args containing HTML",
			seo:      createTestSEO("Page"),
			args:     []string{"<b>Bold</b>", "Info"},
			expected: "<title>Page - Bold - Info</title>",
		},
		{
			name:     "Nil SEO should panic",
			seo:      nil,
			args:     []string{"Test"},
			expected: "should panic",
		},
		{
			name:     "Custom separator",
			seo:      createTestSEOWithSeparator("Page", " | "),
			args:     []string{"Additional"},
			expected: "<title>Page | Additional</title>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.seo == nil {
				assert.Panics(t, func() {
					titleTag(tt.seo, tt.args...)
				}, "titleTag() should panic with nil SEO")
				return
			}
			result := titleTag(tt.seo, tt.args...)
			assert.Equal(t, template.HTML(tt.expected), result, "titleTag() should produce expected output")
		})
	}
}

func TestReverseTitleTag(t *testing.T) {
	tests := []struct {
		name     string
		seo      *SEO
		args     []string
		expected string
	}{
		{
			name:     "Basic reverse title",
			seo:      createTestSEOWithTitle("Site Name", "Page Title"),
			args:     nil,
			expected: "<title>Page Title - Site Name</title>",
		},
		{
			name:     "Reverse title with args",
			seo:      createTestSEOWithTitle("Site", "Page"),
			args:     []string{"First", "Second"},
			expected: "<title>Second - First - Page - Site</title>",
		},
		{
			name:     "Empty title with args",
			seo:      createTestSEO(""),
			args:     []string{"First", "Last"},
			expected: "<title>Last - First - </title>",
		},
		{
			name:     "Single title with args",
			seo:      createTestSEO("Only Title"),
			args:     []string{"First", "Second"},
			expected: "<title>Second - First - Only Title</title>",
		},
		{
			name:     "Nil SEO with args",
			seo:      nil,
			args:     []string{"First", "Last"},
			expected: "<title>Last - First</title>",
		},
		{
			name:     "Args with HTML",
			seo:      createTestSEO("Site"),
			args:     []string{"<b>Bold</b>", "Plain"},
			expected: "<title>Plain - Bold - Site</title>",
		},
		{
			name:     "No args no title",
			seo:      createTestSEO(""),
			args:     nil,
			expected: "<title></title>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.seo == nil {
				assert.Panics(t, func() {
					reverseTitleTag(tt.seo, tt.args...)
				}, "reverseTitleTag() should panic with nil SEO")
				return
			}
			result := reverseTitleTag(tt.seo, tt.args...)
			assert.Equal(t, template.HTML(tt.expected), result, "reverseTitleTag() should produce expected output")
		})
	}
}

func TestMetaTags(t *testing.T) {
	tests := []struct {
		name               string
		seo                *SEO
		expectedExact      string            // For simple cases where exact match is expected
		expectedMetaTags   []expectedMetaTag // For complex cases where order doesn't matter
		expectedSubstrings []string          // Substrings that should be present
	}{
		{
			name: "Basic meta tags",
			seo:  createTestSEOWithMeta(),
			expectedExact: `<meta charset="UTF-8" />
<meta name="description" content="Test description" />
`,
		},
		{
			name: "Meta tags with charset only",
			seo:  createTestSEOWithCharset("UTF-8"),
			expectedExact: `<meta charset="UTF-8" />
`,
		},
		{
			name: "Empty meta tags",
			seo:  &SEO{metaTags: NewMetaTags("")},
			expectedExact: `<meta charset="UTF-8" />
`,
		},
		{
			name: "Complete meta tags",
			seo:  createTestSEOCompleteMeta(),
			expectedMetaTags: []expectedMetaTag{
				{Type: "charset", Value: "UTF-8"},
				{Type: "name", Name: "description", Content: "Test description"},
				{Type: "name", Name: "keywords", Content: "test"},
				{Type: "name", Name: "keywords", Content: "keywords"},
				{Type: "property", Name: "og:title", Content: "Test Title"},
				{Type: "http-equiv", Name: "refresh", Content: "30"},
			},
		},
		{
			name: "Meta tags with special characters",
			seo:  createTestSEOSpecialChars(),
			expectedMetaTags: []expectedMetaTag{
				{Type: "charset", Value: "UTF-8"},
				{Type: "name", Name: "description", Content: `Test & description &quot;with quotes&quot;`},
			},
		},
		{
			name:          "Nil SEO",
			seo:           nil,
			expectedExact: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.seo == nil {
				assert.Panics(t, func() {
					metaTags(tt.seo)
				}, "metaTags() should panic with nil SEO")
				return
			}
			result := metaTags(tt.seo)
			resultStr := string(result)

			if tt.expectedExact != "" {
				assert.Equal(t, template.HTML(tt.expectedExact), result, "metaTags() should produce expected exact output")
			} else if len(tt.expectedMetaTags) > 0 {
				// Check that charset is always first (if present)
				for _, expected := range tt.expectedMetaTags {
					if expected.Type == "charset" {
						assert.True(t, strings.HasPrefix(resultStr, fmt.Sprintf(`<meta charset="%s" />`, expected.Value)),
							"metaTags() should start with charset meta tag")
						break
					}
				}

				// Check that all expected meta tags are present (order doesn't matter for non-charset)
				for _, expected := range tt.expectedMetaTags {
					switch expected.Type {
					case "charset":
						assert.Contains(t, resultStr, fmt.Sprintf(`<meta charset="%s" />`, expected.Value),
							"metaTags() should contain charset meta tag")
					case "name":
						assert.Contains(t, resultStr, fmt.Sprintf(`<meta name="%s" content="%s" />`, expected.Name, expected.Content),
							"metaTags() should contain name meta tag")
					case "property":
						assert.Contains(t, resultStr, fmt.Sprintf(`<meta property="%s" content="%s" />`, expected.Name, expected.Content),
							"metaTags() should contain property meta tag")
					case "http-equiv":
						assert.Contains(t, resultStr, fmt.Sprintf(`<meta http-equiv="%s" content="%s" />`, expected.Name, expected.Content),
							"metaTags() should contain http-equiv meta tag")
					}
				}

				// Validate that meta tags are properly formatted
				metaTagPattern := `<meta\s+(?:charset="[^"]*"|name="[^"]*"|property="[^"]*"|http-equiv="[^"]*")(?:\s+content="[^"]*")?\s*/>`
				assert.Regexp(t, metaTagPattern, resultStr, "metaTags() should produce properly formatted meta tags")
			}

			// Check for expected substrings
			for _, substr := range tt.expectedSubstrings {
				assert.Contains(t, resultStr, substr, "metaTags() should contain substring: %s", substr)
			}
		})
	}
}

// Helper struct for expected meta tags
type expectedMetaTag struct {
	Type    string // "charset", "name", "property", "http-equiv"
	Name    string // for name, property, http-equiv types
	Content string // for name, property, http-equiv types
	Value   string // for charset type
}

func TestHTMLAttrs(t *testing.T) {
	tests := []struct {
		name               string
		seo                *SEO
		rest               []any
		expectedAttributes map[string]string // key-value pairs that should be present
		shouldContain      []string          // substrings that should be present
		shouldNotContain   []string          // substrings that should not be present
	}{
		{
			name: "Basic HTML attributes",
			seo:  createTestSEOWithHTMLAttrs(),
			expectedAttributes: map[string]string{
				"dir":    "ltr",
				"lang":   "en",
				"prefix": "og: https://ogp.me/ns#",
			},
		},
		{
			name: "HTML attributes with additional",
			seo:  createTestSEOWithHTMLAttrs(),
			rest: []any{"data-test", "value", "class", "container"},
			expectedAttributes: map[string]string{
				"dir":       "ltr",
				"lang":      "en",
				"prefix":    "og: https://ogp.me/ns#",
				"data-test": "value",
				"class":     "container",
			},
		},
		{
			name: "Empty HTML attributes with rest",
			seo:  &SEO{htmlAttrs: map[string]string{}},
			rest: []any{"id", "test"},
			expectedAttributes: map[string]string{
				"id": "test",
			},
		},
		{
			name: "HTML attributes with special characters",
			seo:  &SEO{htmlAttrs: map[string]string{"title": `Test "with quotes"`}},
			expectedAttributes: map[string]string{
				"title": `Test &#34;with quotes&#34;`,
			},
		},
		{
			name: "Nil SEO with rest",
			seo:  nil,
			rest: []any{"id", "test"},
			expectedAttributes: map[string]string{
				"id": "test",
			},
		},
		{
			name: "Odd number of rest parameters",
			seo:  createTestSEOWithHTMLAttrs(),
			rest: []any{"incomplete"},
			expectedAttributes: map[string]string{
				"dir":    "ltr",
				"lang":   "en",
				"prefix": "og: https://ogp.me/ns#",
			},
			shouldNotContain: []string{"incomplete"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.seo == nil {
				assert.Panics(t, func() {
					htmlAttrs(tt.seo, tt.rest...)
				}, "htmlAttrs() should panic with nil SEO")
				return
			}
			result := htmlAttrs(tt.seo, tt.rest...)
			resultStr := string(result)

			// Check that all expected attributes are present with correct values
			for attrName, expectedValue := range tt.expectedAttributes {
				expectedPattern := fmt.Sprintf(`%s="%s"`, attrName, expectedValue)
				assert.Contains(t, resultStr, expectedPattern, "htmlAttrs() should contain attribute %s with value %s", attrName, expectedValue)
			}

			// Check that specified substrings are present
			for _, substr := range tt.shouldContain {
				assert.Contains(t, resultStr, substr, "htmlAttrs() should contain substring: %s", substr)
			}

			// Check that specified substrings are not present
			for _, substr := range tt.shouldNotContain {
				assert.NotContains(t, resultStr, substr, "htmlAttrs() should not contain substring: %s", substr)
			}

			// Validate that attributes are properly formatted by checking the whole string
			if resultStr != "" {
				// Each attribute should be in the format name="value"
				attributes := strings.Fields(resultStr)
				for _, attr := range attributes {
					// Skip validation if this doesn't look like a complete attribute
					if !strings.Contains(attr, "=") {
						continue
					}
					// Check if it starts with attribute name format
					parts := strings.SplitN(attr, "=", 2)
					if len(parts) == 2 && strings.HasPrefix(parts[1], `"`) && strings.HasSuffix(parts[1], `"`) {
						// Attribute name should be valid
						assert.Regexp(t, `^[\w:-]+$`, parts[0], "htmlAttrs() should produce valid attribute names")
					}
				}
			}
		})
	}
}

func TestHeadAttrs(t *testing.T) {
	tests := []struct {
		name     string
		seo      *SEO
		rest     []any
		expected string
	}{
		{
			name:     "Basic head attributes",
			seo:      createTestSEOWithHeadAttrs(),
			rest:     nil,
			expected: `profile="http://example.com"`,
		},
		{
			name:     "Head attributes with additional",
			seo:      createTestSEOWithHeadAttrs(),
			rest:     []any{"data-theme", "dark"},
			expected: `profile="http://example.com" data-theme="dark"`,
		},
		{
			name:     "Empty head attributes",
			seo:      &SEO{headAttrs: map[string]string{}},
			rest:     nil,
			expected: "",
		},
		{
			name:     "Nil SEO with rest",
			seo:      nil,
			rest:     []any{"style", "color: red"},
			expected: `style="color: red"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.seo == nil {
				assert.Panics(t, func() {
					headAttrs(tt.seo, tt.rest...)
				}, "headAttrs() should panic with nil SEO")
				return
			}
			result := headAttrs(tt.seo, tt.rest...)
			assert.Equal(t, template.HTMLAttr(tt.expected), result, "headAttrs() should produce expected output")
		})
	}
}

func TestBodyAttrs(t *testing.T) {
	tests := []struct {
		name     string
		seo      *SEO
		rest     []any
		expected string
	}{
		{
			name:     "Basic body attributes",
			seo:      createTestSEOWithBodyAttrs(),
			rest:     nil,
			expected: `class="main-body"`,
		},
		{
			name:     "Body attributes with additional",
			seo:      createTestSEOWithBodyAttrs(),
			rest:     []any{"data-loaded", "true"},
			expected: `class="main-body" data-loaded="true"`,
		},
		{
			name:     "Empty body attributes",
			seo:      &SEO{bodyAttrs: map[string]string{}},
			rest:     nil,
			expected: "",
		},
		{
			name:     "Nil SEO with rest",
			seo:      nil,
			rest:     []any{"onload", "init()"},
			expected: `onload="init()"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.seo == nil {
				assert.Panics(t, func() {
					bodyAttrs(tt.seo, tt.rest...)
				}, "bodyAttrs() should panic with nil SEO")
				return
			}
			result := bodyAttrs(tt.seo, tt.rest...)
			assert.Equal(t, template.HTMLAttr(tt.expected), result, "bodyAttrs() should produce expected output")
		})
	}
}

func TestLangAlternates(t *testing.T) {
	tests := []struct {
		name     string
		seo      *SEO
		expected string
	}{
		{
			name:     "Basic lang alternates",
			seo:      createTestSEOWithLangAlternates(),
			expected: "<link rel=\"alternate\" href=\"https://example.com/fr\" hreflang=\"fr\" />\n",
		},
		{
			name:     "Empty lang alternates",
			seo:      &SEO{langAlternates: map[string]string{}},
			expected: "",
		},
		{
			name: "Lang alternates with special characters",
			seo:  &SEO{langAlternates: map[string]string{"https://example.com/path?param=value": "en-US"}},
			expected: `<link rel="alternate" href="https://example.com/path?param=value" hreflang="en-US" />
`,
		},
		{
			name:     "Nil SEO",
			seo:      nil,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.seo == nil {
				assert.Panics(t, func() {
					langAlternates(tt.seo)
				}, "langAlternates() should panic with nil SEO")
				return
			}
			result := langAlternates(tt.seo)
			assert.Equal(t, template.HTML(tt.expected), result, "langAlternates() should produce expected output")
		})
	}
}

func TestHeadLinks(t *testing.T) {
	tests := []struct {
		name               string
		seo                *SEO
		expectedExact      string         // For simple cases where exact match is expected
		expectedLinks      []expectedLink // For complex cases where we want to validate individual links
		expectedSubstrings []string       // Substrings that should be present
		expectedLinkCount  int            // Expected number of links in output
	}{
		{
			name: "Basic head links",
			seo:  createTestSEOWithHeadLinks(),
			expectedExact: `<link rel="stylesheet" href="/style.css" />
<link rel="icon" href="/favicon.ico" />
`,
			expectedLinkCount: 2,
		},
		{
			name: "Complete head links",
			seo:  createTestSEOCompleteHeadLinks(),
			expectedLinks: []expectedLink{
				{Rel: "stylesheet", Href: "/style.css", Type: "text/css"},
				{Rel: "icon", Href: "/favicon.ico", Sizes: "32x32", Type: "image/png"},
				{Rel: "canonical", Href: "https://example.com/page"},
			},
			expectedLinkCount: 3,
		},
		{
			name:              "Empty head links",
			seo:               &SEO{headLinks: []HeadLink{}},
			expectedExact:     "",
			expectedLinkCount: 0,
		},
		{
			name: "Head links with empty rel (should be skipped)",
			seo:  &SEO{headLinks: []HeadLink{{Rel: "", Href: "/skip"}, {Rel: "valid", Href: "/keep"}}},
			expectedLinks: []expectedLink{
				{Rel: "valid", Href: "/keep"},
			},
			expectedLinkCount: 1,
		},
		{
			name:              "Nil SEO",
			seo:               nil,
			expectedExact:     "",
			expectedLinkCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.seo == nil {
				assert.Panics(t, func() {
					headLinks(tt.seo)
				}, "headLinks() should panic with nil SEO")
				return
			}
			result := headLinks(tt.seo)
			resultStr := string(result)

			if tt.expectedExact != "" {
				assert.Equal(t, template.HTML(tt.expectedExact), result, "headLinks() should produce expected exact output")
			}

			// Check expected link count
			if tt.expectedLinkCount >= 0 {
				if tt.expectedLinkCount == 0 && tt.expectedExact == "" {
					assert.Empty(t, resultStr, "headLinks() should produce empty output when no links")
				} else {
					linkCount := strings.Count(resultStr, `<link `)
					assert.Equal(t, tt.expectedLinkCount, linkCount, "headLinks() should produce correct number of links")
				}
			}

			// Check individual expected links
			for _, expectedLink := range tt.expectedLinks {
				linkPattern := buildLinkPattern(expectedLink)
				assert.Regexp(t, linkPattern, resultStr, "headLinks() should contain link with rel=%s", expectedLink.Rel)
			}

			// Check for expected substrings
			for _, substr := range tt.expectedSubstrings {
				assert.Contains(t, resultStr, substr, "headLinks() should contain substring: %s", substr)
			}

			// Validate that all links are properly formatted
			if resultStr != "" {
				lines := strings.Split(strings.TrimSpace(resultStr), "\n")
				for _, line := range lines {
					line = strings.TrimSpace(line)
					if line != "" {
						// Basic link tag validation
						assert.True(t, strings.HasPrefix(line, `<link `) && strings.HasSuffix(line, `/>`),
							"headLinks() should produce properly formatted link tags: %s", line)

						// Check that rel attribute is present (should never be empty due to the skip logic)
						assert.Regexp(t, `rel="[^"]+"`, line,
							"headLinks() should always include rel attribute: %s", line)
					}
				}
			}
		})
	}
}

// Helper struct for expected head links
type expectedLink struct {
	Rel         string
	Href        string
	HrefLang    string
	Media       string
	Sizes       string
	Title       string
	Type        string
	CrossOrigin string
}

// Helper function to build regex pattern for link validation
func buildLinkPattern(link expectedLink) string {
	pattern := `<link rel="` + regexp.QuoteMeta(link.Rel) + `"`

	attrs := []string{}
	if link.Href != "" {
		attrs = append(attrs, `\s+href="`+regexp.QuoteMeta(link.Href)+`"`)
	}
	if link.HrefLang != "" {
		attrs = append(attrs, `\s+hreflang="`+regexp.QuoteMeta(link.HrefLang)+`"`)
	}
	if link.Media != "" {
		attrs = append(attrs, `\s+media="`+regexp.QuoteMeta(link.Media)+`"`)
	}
	if link.Sizes != "" {
		attrs = append(attrs, `\s+sizes="`+regexp.QuoteMeta(link.Sizes)+`"`)
	}
	if link.Title != "" {
		attrs = append(attrs, `\s+title="`+regexp.QuoteMeta(link.Title)+`"`)
	}
	if link.Type != "" {
		attrs = append(attrs, `\s+type="`+regexp.QuoteMeta(link.Type)+`"`)
	}
	if link.CrossOrigin != "" {
		attrs = append(attrs, `\s+crossorigin="`+regexp.QuoteMeta(link.CrossOrigin)+`"`)
	}

	// Add attributes in any order
	if len(attrs) > 0 {
		for i, attr := range attrs {
			if i < len(attrs)-1 {
				pattern += `(?:` + attr + `)?`
			} else {
				pattern += attr + `?`
			}
		}
	}

	pattern += `\s+/>`
	return pattern
}

func TestNormalize(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "HTML with tags and quotes",
			input:    `<p>Test "with" HTML</p>`,
			expected: "Test &quot;with&quot; HTML",
		},
		{
			name:     "Plain text with quotes",
			input:    `Just "quotes" here`,
			expected: `Just &quot;quotes&quot; here`,
		},
		{
			name:     "HTML entities",
			input:    "Test &amp; more &lt;tags&gt;",
			expected: "Test & more ",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalize(tt.input)
			assert.Equal(t, tt.expected, result, "normalize() should produce expected output")
		})
	}
}

func TestAttrs(t *testing.T) {
	tests := []struct {
		name     string
		attrs    map[string]string
		rest     []any
		expected string
	}{
		{
			name:     "Basic attributes",
			attrs:    map[string]string{"class": "container"},
			rest:     nil,
			expected: `class="container"`,
		},
		{
			name:     "Attributes with rest",
			attrs:    map[string]string{},
			rest:     []any{"style", "color: red"},
			expected: `style="color: red"`,
		},
		{
			name:     "Empty attributes with rest",
			attrs:    map[string]string{},
			rest:     []any{"href", "/test"},
			expected: `href="/test"`,
		},
		{
			name:     "Attributes with HTML special characters",
			attrs:    map[string]string{"title": `Test "with quotes"`},
			rest:     nil,
			expected: `title="Test &#34;with quotes&#34;"`,
		},
		{
			name:     "Nil attributes with rest",
			attrs:    nil,
			rest:     []any{"src", "/image.png"},
			expected: `src="/image.png"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := attrs(tt.attrs, tt.rest...)
			assert.Equal(t, template.HTMLAttr(tt.expected), result, "attrs() should produce expected output")
		})
	}
}

// Helper functions to create test data

func createTestSEO(title string) *SEO {
	seo := NewSEO()
	if title != "" {
		seo.SetTitle(title)
	}
	return seo
}

func createTestSEOWithSeparator(title, separator string) *SEO {
	seo := NewSEO()
	if title != "" {
		seo.SetTitle(title)
	}
	if separator != "" {
		seo.SetSeparator(separator)
	}
	return seo
}

func createTestSEOWithTitle(titles ...string) *SEO {
	seo := NewSEO()
	for _, title := range titles {
		seo.AddTitle(title)
	}
	return seo
}

func createTestSEOWithMeta() *SEO {
	seo := NewSEO()
	seo.metaTags.SetName("description", "Test description")
	return seo
}

func createTestSEOWithCharset(charset string) *SEO {
	seo := NewSEO()
	seo.metaTags = NewMetaTags(charset)
	return seo
}

func createTestSEOCompleteMeta() *SEO {
	seo := NewSEO()
	seo.metaTags.SetName("description", "Test description")
	seo.metaTags.SetName("keywords", "test", "keywords")
	seo.metaTags.SetProperty("og:title", "Test Title")
	seo.metaTags.SetHTTPEquiv("refresh", "30")
	return seo
}

func createTestSEOSpecialChars() *SEO {
	seo := NewSEO()
	seo.metaTags.SetName("description", `Test & description "with quotes"`)
	return seo
}

func createTestSEOWithHTMLAttrs() *SEO {
	seo := NewSEO()
	return seo
}

func createTestSEOWithHeadAttrs() *SEO {
	seo := NewSEO()
	seo.SetHeadAttribute("profile", "http://example.com")
	return seo
}

func createTestSEOWithBodyAttrs() *SEO {
	seo := NewSEO()
	seo.SetBodyAttribute("class", "main-body")
	return seo
}

func createTestSEOWithLangAlternates() *SEO {
	seo := NewSEO()
	seo.AddLangAlternate("https://example.com/fr", "fr")
	return seo
}

func createTestSEOWithHeadLinks() *SEO {
	seo := NewSEO()
	seo.AddLink(HeadLink{Rel: "stylesheet", Href: "/style.css"})
	seo.AddLink(HeadLink{Rel: "icon", Href: "/favicon.ico"})
	return seo
}

func createTestSEOCompleteHeadLinks() *SEO {
	seo := NewSEO()
	seo.AddLink(HeadLink{Rel: "stylesheet", Href: "/style.css", Type: "text/css"})
	seo.AddLink(HeadLink{Rel: "icon", Href: "/favicon.ico", Sizes: "32x32", Type: "image/png"})
	seo.AddLink(HeadLink{Rel: "canonical", Href: "https://example.com/page"})
	return seo
}

// Integration tests

func TestSEOFuncMap_Integration(t *testing.T) {
	t.Run("Complete SEO setup", func(t *testing.T) {
		seo := NewSEO()
		seo.SetTitle("Test Page")
		seo.metaTags.SetName("description", "Test description")
		seo.SetHTMLAttribute("lang", "en")
		seo.SetHeadAttribute("profile", "test")
		seo.SetBodyAttribute("class", "test-body")
		seo.AddLangAlternate("https://example.com/fr", "fr")
		seo.AddLink(HeadLink{Rel: "stylesheet", Href: "/style.css"})

		// Test all functions work together
		title := titleTag(seo, "Additional")
		require.NotEmpty(t, title)

		meta := metaTags(seo)
		require.NotEmpty(t, meta)

		htmlAttr := htmlAttrs(seo)
		require.NotEmpty(t, htmlAttr)

		headAttr := headAttrs(seo)
		require.NotEmpty(t, headAttr)

		bodyAttr := bodyAttrs(seo)
		require.NotEmpty(t, bodyAttr)

		langAlt := langAlternates(seo)
		require.NotEmpty(t, langAlt)

		links := headLinks(seo)
		require.NotEmpty(t, links)
	})

	t.Run("Template function execution", func(t *testing.T) {
		// Test that functions can be called as template functions
		funcMap := SEOFuncMap

		// Test strip_tags function
		if stripTagsFunc, ok := funcMap["strip_tags"].(func(string) string); ok {
			result := stripTagsFunc("<p>Test</p>")
			assert.Equal(t, "Test", result)
		} else {
			t.Error("strip_tags function not found or wrong signature")
		}

		// Test escape_double_q function
		if escapeFunc, ok := funcMap["escape_double_q"].(func(string) string); ok {
			result := escapeFunc(`Test "quoted"`)
			assert.Equal(t, "Test &quot;quoted&quot;", result)
		} else {
			t.Error("escape_double_q function not found or wrong signature")
		}
	})
}

func TestSEOFuncMap_EdgeCases(t *testing.T) {
	t.Run("Nil parameters handling", func(t *testing.T) {
		// Test that all functions panic with nil SEO
		assert.Panics(t, func() {
			titleTag(nil, "test")
		}, "titleTag should panic with nil SEO")

		assert.Panics(t, func() {
			reverseTitleTag(nil, "test")
		}, "reverseTitleTag should panic with nil SEO")

		assert.Panics(t, func() {
			metaTags(nil)
		}, "metaTags should panic with nil SEO")

		assert.Panics(t, func() {
			htmlAttrs(nil)
		}, "htmlAttrs should panic with nil SEO")

		assert.Panics(t, func() {
			headAttrs(nil)
		}, "headAttrs should panic with nil SEO")

		assert.Panics(t, func() {
			bodyAttrs(nil)
		}, "bodyAttrs should panic with nil SEO")

		assert.Panics(t, func() {
			langAlternates(nil)
		}, "langAlternates should panic with nil SEO")

		assert.Panics(t, func() {
			headLinks(nil)
		}, "headLinks should panic with nil SEO")
	})

	t.Run("Empty inputs", func(t *testing.T) {
		// Test functions with empty inputs
		assert.NotPanics(t, func() {
			stripTags("")
			escapeDoubleQuotes("")
			titleTag(&SEO{}, "")
			reverseTitleTag(&SEO{}, "")
			metaTags(&SEO{metaTags: NewMetaTags("")})
			htmlAttrs(&SEO{htmlAttrs: map[string]string{}})
			headAttrs(&SEO{headAttrs: map[string]string{}})
			bodyAttrs(&SEO{bodyAttrs: map[string]string{}})
			langAlternates(&SEO{langAlternates: map[string]string{}})
			headLinks(&SEO{headLinks: []HeadLink{}})
		})
	})

	t.Run("Large inputs", func(t *testing.T) {
		// Test functions with large inputs
		longString := string(make([]byte, 10000))
		for i := range longString {
			longString = longString[:i] + "a" + longString[i+1:]
		}

		assert.NotPanics(t, func() {
			stripTags(longString)
			escapeDoubleQuotes(longString)
		})
	})
}
