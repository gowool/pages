package pages

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gowool/wo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Test helper functions
func createTestSite(name, host, locale string, isDefault bool, countries ...string) *Site {
	site := NewSite()
	site.Name = name
	site.Host = host
	site.Locale = locale
	site.IsDefault = isDefault
	site.Countries = countries
	site.Enabled = true
	return site
}

func createTestRequest(method, url string, headers map[string]string) *http.Request {
	req := httptest.NewRequest(method, url, nil)
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	return req
}

func TestNewSiteSelector(t *testing.T) {
	t.Run("Valid parameters", func(t *testing.T) {
		storage := &MockSiteStorage{}
		countryFunc := func(r *http.Request) (string, error) { return "US", nil }
		errorFunc := func(r *http.Request, err error) (*Site, error) { return nil, err }

		selector := NewSiteSelector(storage, countryFunc, errorFunc)

		assert.NotNil(t, selector)
		// Test that it implements the interface
		assert.Implements(t, (*SiteSelector)(nil), selector)
	})

	t.Run("Nil storage should panic", func(t *testing.T) {
		assert.Panics(t, func() {
			NewSiteSelector(nil, nil, nil)
		})
	})

	t.Run("Nil countryFunc should use default", func(t *testing.T) {
		storage := &MockSiteStorage{}
		selector := NewSiteSelector(storage, nil, nil)

		req := createTestRequest("GET", "http://example.com", map[string]string{
			wo.HeaderCFIPCountry: "US",
		})

		// Test that the selector works with default country function
		storage.On("FindEnabled", mock.Anything).Return([]*Site{createTestSite("Test", "example.com", "en", true)}, nil)
		site, _, err := selector.Retrieve(req)

		assert.NoError(t, err)
		assert.NotNil(t, site)
		storage.AssertExpectations(t)
	})

	t.Run("Nil errorFunc should use default", func(t *testing.T) {
		storage := &MockSiteStorage{}
		selector := NewSiteSelector(storage, nil, nil)

		req := createTestRequest("GET", "http://example.com", nil)
		testErr := errors.New("test error")

		// Test that the selector returns error directly when no custom error func is provided
		storage.On("FindEnabled", mock.Anything).Return([]*Site{}, testErr)
		site, _, err := selector.Retrieve(req)

		assert.Error(t, err)
		assert.Equal(t, testErr, err)
		assert.Nil(t, site)
		storage.AssertExpectations(t)
	})
}

func TestDefaultSiteSelector_Retrieve(t *testing.T) {
	t.Run("Nil request should panic", func(t *testing.T) {
		storage := &MockSiteStorage{}
		selector := NewSiteSelector(storage, nil, nil)

		assert.Panics(t, func() {
			selector.Retrieve(nil)
		})
	})

	t.Run("Country function error", func(t *testing.T) {
		storage := &MockSiteStorage{}
		testErr := errors.New("country error")
		countryFunc := func(r *http.Request) (string, error) { return "", testErr }
		errorFunc := func(r *http.Request, err error) (*Site, error) {
			return createTestSite("Error Site", "example.com", "en", false), nil
		}

		selector := NewSiteSelector(storage, countryFunc, errorFunc)
		req := createTestRequest("GET", "http://example.com", nil)

		site, _, err := selector.Retrieve(req)

		assert.NoError(t, err)
		assert.NotNil(t, site)
		assert.Equal(t, "Error Site", site.Name)
	})

	t.Run("Storage error", func(t *testing.T) {
		storage := &MockSiteStorage{}
		testErr := errors.New("storage error")
		countryFunc := func(r *http.Request) (string, error) { return "US", nil }
		errorFunc := func(r *http.Request, err error) (*Site, error) {
			return createTestSite("Error Site", "example.com", "en", false), nil
		}

		storage.On("FindEnabled", mock.Anything).Return([]*Site{}, testErr)

		selector := NewSiteSelector(storage, countryFunc, errorFunc)
		req := createTestRequest("GET", "http://example.com", nil)

		site, _, err := selector.Retrieve(req)

		assert.NoError(t, err)
		assert.NotNil(t, site)
		assert.Equal(t, "Error Site", site.Name)

		storage.AssertExpectations(t)
	})

	t.Run("Successful site selection", func(t *testing.T) {
		storage := &MockSiteStorage{}
		sites := []*Site{
			createTestSite("Default Site", "example.com", "en", true),
			createTestSite("French Site", "example.com", "fr", false),
			createTestSite("US Site", "example.com", "en-US", false, "US"),
		}

		storage.On("FindEnabled", mock.Anything).Return(sites, nil)

		selector := NewSiteSelector(storage, nil, nil)
		req := createTestRequest("GET", "http://example.com", map[string]string{
			wo.HeaderAcceptLanguage: "fr-FR,fr;q=0.9,en;q=0.8",
		})

		site, _, err := selector.Retrieve(req)

		assert.NoError(t, err)
		assert.NotNil(t, site)
		assert.Equal(t, "French Site", site.Name)

		storage.AssertExpectations(t)
	})

	t.Run("Site selection with country", func(t *testing.T) {
		storage := &MockSiteStorage{}
		sites := []*Site{
			createTestSite("Default Site", "example.com", "en", true),
			createTestSite("US Site", "example.com", "en-US", false, "US"),
			createTestSite("French Site", "example.com", "fr", false),
		}

		storage.On("FindEnabled", mock.Anything).Return(sites, nil)

		countryFunc := func(r *http.Request) (string, error) { return "US", nil }
		selector := NewSiteSelector(storage, countryFunc, nil)
		req := createTestRequest("GET", "http://example.com/test", map[string]string{wo.HeaderAcceptLanguage: "en-US;q=0.9,en;q=0.8"})

		site, _, err := selector.Retrieve(req)

		assert.NoError(t, err)
		assert.NotNil(t, site)
		assert.Equal(t, "US Site", site.Name)

		storage.AssertExpectations(t)
	})

	t.Run("No matching sites returns error", func(t *testing.T) {
		storage := &MockSiteStorage{}
		sites := []*Site{
			createTestSite("Different Host", "other.com", "en", false),
		}

		storage.On("FindEnabled", mock.Anything).Return(sites, nil)

		errorFunc := func(r *http.Request, err error) (*Site, error) {
			return nil, err
		}

		selector := NewSiteSelector(storage, nil, errorFunc)
		req := createTestRequest("GET", "http://example.com", nil)

		site, _, err := selector.Retrieve(req)

		assert.Error(t, err)
		assert.Equal(t, ErrSiteNotFound, err)
		assert.Nil(t, site)

		storage.AssertExpectations(t)
	})

	t.Run("Host matching", func(t *testing.T) {
		storage := &MockSiteStorage{}
		sites := []*Site{
			createTestSite("Correct Host", "example.com", "en", true),
			createTestSite("Wrong Host", "other.com", "en", false),
		}

		storage.On("FindEnabled", mock.Anything).Return(sites, nil)

		selector := NewSiteSelector(storage, nil, nil)
		req := createTestRequest("GET", "http://example.com", nil)

		site, _, err := selector.Retrieve(req)

		assert.NoError(t, err)
		assert.NotNil(t, site)
		assert.Equal(t, "Correct Host", site.Name)

		storage.AssertExpectations(t)
	})
}

// Integration tests that indirectly test selectedSite functionality
func TestDefaultSiteSelector_LanguageMatching(t *testing.T) {
	t.Run("Language preference matching", func(t *testing.T) {
		storage := &MockSiteStorage{}

		englishSite := createTestSite("English", "example.com", "en", false)
		frenchSite := createTestSite("French", "example.com", "fr", false)
		spanishSite := createTestSite("Spanish", "example.com", "es", false)

		sites := []*Site{englishSite, frenchSite, spanishSite}
		storage.On("FindEnabled", mock.Anything).Return(sites, nil)

		selector := NewSiteSelector(storage, nil, nil)

		// Test French preference
		req := createTestRequest("GET", "http://example.com", map[string]string{
			wo.HeaderAcceptLanguage: "fr-FR,fr;q=0.9,en;q=0.8",
		})

		site, _, err := selector.Retrieve(req)

		require.NoError(t, err)
		assert.Equal(t, "French", site.Name)

		storage.AssertExpectations(t)
	})

	t.Run("Fallback to parent language", func(t *testing.T) {
		storage := &MockSiteStorage{}

		englishSite := createTestSite("English", "example.com", "en", false)

		sites := []*Site{englishSite}
		storage.On("FindEnabled", mock.Anything).Return(sites, nil)

		selector := NewSiteSelector(storage, nil, nil)

		// Test with en-US which should fallback to en
		req := createTestRequest("GET", "http://example.com", map[string]string{
			wo.HeaderAcceptLanguage: "en-US,en;q=0.9",
		})

		site, _, err := selector.Retrieve(req)

		require.NoError(t, err)
		assert.Equal(t, "English", site.Name)

		storage.AssertExpectations(t)
	})

	t.Run("Invalid accept language header", func(t *testing.T) {
		storage := &MockSiteStorage{}

		defaultSite := createTestSite("Default", "example.com", "en", true)

		sites := []*Site{defaultSite}
		storage.On("FindEnabled", mock.Anything).Return(sites, nil)

		selector := NewSiteSelector(storage, nil, nil)

		// Test with invalid Accept-Language header
		req := createTestRequest("GET", "http://example.com", map[string]string{
			wo.HeaderAcceptLanguage: "invalid-language-header",
		})

		site, _, err := selector.Retrieve(req)

		require.NoError(t, err)
		assert.Equal(t, "Default", site.Name)

		storage.AssertExpectations(t)
	})
}

func TestRegexpPath(t *testing.T) {
	t.Run("Valid path pattern", func(t *testing.T) {
		path := "/blog/([a-z0-9-]+)"
		re, err := regexpPath(path)

		assert.NoError(t, err)
		assert.NotNil(t, re)

		// Test the compiled regex
		match, err := re.FindStringMatch("/blog/test-post")
		assert.NoError(t, err)
		assert.NotNil(t, match)
	})

	t.Run("Empty path", func(t *testing.T) {
		re, err := regexpPath("")

		assert.NoError(t, err)
		assert.NotNil(t, re)
	})

	t.Run("Root path", func(t *testing.T) {
		re, err := regexpPath("/")

		assert.NoError(t, err)
		assert.NotNil(t, re)
	})

	t.Run("Invalid regex pattern", func(t *testing.T) {
		path := "/blog/([a-z0-9-+" // Invalid regex
		re, err := regexpPath(path)

		assert.Error(t, err)
		assert.Nil(t, re)
	})

	t.Run("Cached result", func(t *testing.T) {
		path := "/test"

		// First call
		re1, err1 := regexpPath(path)
		assert.NoError(t, err1)
		assert.NotNil(t, re1)

		// Second call should return cached result
		re2, err2 := regexpPath(path)
		assert.NoError(t, err2)
		assert.Equal(t, re1, re2)
	})

	t.Run("Cached error result", func(t *testing.T) {
		path := "/invalid/([a-z0-9-+"

		// First call should cache error
		re1, err1 := regexpPath(path)
		assert.Error(t, err1)
		assert.Nil(t, re1)

		// Second call should return cached error
		re2, err2 := regexpPath(path)
		assert.Error(t, err2)
		assert.Nil(t, re2)
	})
}

func TestMatchRequest(t *testing.T) {
	t.Run("Empty relative path", func(t *testing.T) {
		req := createTestRequest("GET", "http://example.com/test", nil)

		pathInfo, err := matchRequest(req, "")

		assert.NoError(t, err)
		assert.Equal(t, "/test", pathInfo)
	})

	t.Run("Root relative path", func(t *testing.T) {
		req := createTestRequest("GET", "http://example.com/test", nil)

		pathInfo, err := matchRequest(req, "/")

		assert.NoError(t, err)
		assert.Equal(t, "/test", pathInfo)
	})

	t.Run("Matching pattern", func(t *testing.T) {
		req := createTestRequest("GET", "http://example.com/blog/test-post", nil)

		pathInfo, err := matchRequest(req, "/blog")

		assert.NoError(t, err)
		assert.Equal(t, "/test-post", pathInfo)
	})

	t.Run("Complex matching pattern", func(t *testing.T) {
		req := createTestRequest("GET", "http://example.com/blog/2023/test-post", nil)

		pathInfo, err := matchRequest(req, "/blog/([0-9]{4})")

		assert.NoError(t, err)
		assert.Equal(t, "2023", pathInfo)
	})

	t.Run("No match", func(t *testing.T) {
		req := createTestRequest("GET", "http://example.com/other/path", nil)

		pathInfo, err := matchRequest(req, "/blog")

		assert.Error(t, err)
		assert.Equal(t, "", pathInfo)
	})

	t.Run("Invalid regex pattern", func(t *testing.T) {
		req := createTestRequest("GET", "http://example.com/test", nil)

		pathInfo, err := matchRequest(req, "/invalid/([a-z0-9-+")

		assert.Error(t, err)
		assert.Equal(t, "", pathInfo)
	})

	t.Run("Root path request", func(t *testing.T) {
		req := createTestRequest("GET", "http://example.com/", nil)

		pathInfo, err := matchRequest(req, "/")

		assert.NoError(t, err)
		assert.Equal(t, "/", pathInfo)
	})
}

func TestGetHost(t *testing.T) {
	t.Run("Host without port", func(t *testing.T) {
		req := createTestRequest("GET", "http://example.com", nil)
		req.Host = "example.com"

		host := getHost(req)
		assert.Equal(t, "example.com", host)
	})

	t.Run("Host with port", func(t *testing.T) {
		req := createTestRequest("GET", "http://example.com", nil)
		req.Host = "example.com:8080"

		host := getHost(req)
		assert.Equal(t, "example.com", host)
	})

	t.Run("IPv6 host with port", func(t *testing.T) {
		req := createTestRequest("GET", "http://[::1]", nil)
		req.Host = "[::1]:8080"

		host := getHost(req)
		assert.Equal(t, "::1", host)
	})

	t.Run("Invalid host format", func(t *testing.T) {
		req := createTestRequest("GET", "http://example.com", nil)
		req.Host = "invalid-host-format"

		host := getHost(req)
		assert.Equal(t, "invalid-host-format", host)
	})
}

func TestDefaultSiteSelector_Integration(t *testing.T) {
	t.Run("Complete site selection workflow", func(t *testing.T) {
		// Create test sites
		defaultSite := createTestSite("Default", "example.com", "en", true)
		frenchSite := createTestSite("French", "example.com", "fr", false)
		germanSite := createTestSite("German", "example.com", "de", false)
		usSite := createTestSite("US", "example.com", "en-US", false, "US")

		sites := []*Site{defaultSite, frenchSite, germanSite, usSite}

		storage := &MockSiteStorage{}
		storage.On("FindEnabled", mock.Anything).Return(sites, nil)

		// Create selector
		countryFunc := func(r *http.Request) (string, error) { return "US", nil }
		selector := NewSiteSelector(storage, countryFunc, nil)

		t.Run("French language preference", func(t *testing.T) {
			req := createTestRequest("GET", "http://example.com", map[string]string{
				wo.HeaderAcceptLanguage: "fr-FR,fr;q=0.9,en;q=0.8",
			})

			site, _, err := selector.Retrieve(req)

			require.NoError(t, err)
			require.NotNil(t, site)
			assert.Equal(t, "French", site.Name)
		})

		t.Run("US country with path", func(t *testing.T) {
			req := createTestRequest("GET", "http://example.com/blog/test", map[string]string{
				wo.HeaderAcceptLanguage: "en-US,en;q=0.9",
			})

			site, pathInfo, err := selector.Retrieve(req)

			require.NoError(t, err)
			require.NotNil(t, site)
			assert.Equal(t, "US", site.Name)
			assert.Equal(t, "/blog/test", pathInfo)
		})

		t.Run("Fallback to default", func(t *testing.T) {
			req := createTestRequest("GET", "http://example.com", map[string]string{
				wo.HeaderAcceptLanguage: "es-ES,es;q=0.9",
			})

			site, _, err := selector.Retrieve(req)

			require.NoError(t, err)
			require.NotNil(t, site)
			assert.Equal(t, "Default", site.Name)
		})

		storage.AssertExpectations(t)
	})

	t.Run("Multi-host environment", func(t *testing.T) {
		site1 := createTestSite("Site1", "site1.com", "en", true)
		site2 := createTestSite("Site2", "site2.com", "fr", true)

		sites := []*Site{site1, site2}

		storage := &MockSiteStorage{}
		storage.On("FindEnabled", mock.Anything).Return(sites, nil)

		selector := NewSiteSelector(storage, nil, nil)

		// Test site1.com
		req1 := createTestRequest("GET", "http://site1.com", nil)
		site, _, err := selector.Retrieve(req1)
		require.NoError(t, err)
		assert.Equal(t, "Site1", site.Name)

		// Test site2.com
		req2 := createTestRequest("GET", "http://site2.com", nil)
		site, _, err = selector.Retrieve(req2)
		require.NoError(t, err)
		assert.Equal(t, "Site2", site.Name)

		storage.AssertExpectations(t)
	})
}

func TestDefaultSiteSelector_CountryBasedSelection(t *testing.T) {
	t.Run("Country filtering for root path", func(t *testing.T) {
		usSite := createTestSite("US Site", "example.com", "en", false, "US")
		euSite := createTestSite("EU Site", "example.com", "en", false, "FR", "DE", "IT")
		globalSite := createTestSite("Global Site", "example.com", "en", true)

		sites := []*Site{usSite, euSite, globalSite}

		storage := &MockSiteStorage{}
		storage.On("FindEnabled", mock.Anything).Return(sites, nil)

		// Test US visitor
		countryFunc := func(r *http.Request) (string, error) { return "US", nil }
		selector := NewSiteSelector(storage, countryFunc, nil)

		req := createTestRequest("GET", "http://example.com/", nil)
		site, _, err := selector.Retrieve(req)

		require.NoError(t, err)
		require.NotNil(t, site)
		// Should have US site and EU site available (both match countries)
		assert.Contains(t, []*Site{usSite, euSite}, site)

		storage.AssertExpectations(t)
	})

	t.Run("Country restriction with no match", func(t *testing.T) {
		euOnlySite := createTestSite("EU Only", "example.com", "en", false, "FR", "DE")
		defaultSite := createTestSite("Default", "other.com", "en", true) // Different host

		sites := []*Site{euOnlySite, defaultSite}

		storage := &MockSiteStorage{}
		storage.On("FindEnabled", mock.Anything).Return(sites, nil)

		countryFunc := func(r *http.Request) (string, error) { return "US", nil }
		errorFunc := func(r *http.Request, err error) (*Site, error) { return nil, err }
		selector := NewSiteSelector(storage, countryFunc, errorFunc)

		req := createTestRequest("GET", "http://example.com/", nil)
		site, _, err := selector.Retrieve(req)

		assert.Error(t, err)
		assert.Equal(t, ErrSiteNotFound, err)
		assert.Nil(t, site)

		storage.AssertExpectations(t)
	})
}

// Benchmark tests
func BenchmarkSiteSelector_Retrieve(b *testing.B) {
	sites := make([]*Site, 100)
	for i := 0; i < 100; i++ {
		sites[i] = createTestSite(
			fmt.Sprintf("Site %d", i),
			"example.com",
			fmt.Sprintf("en-%d", i),
			i == 0, // First site is default
		)
	}

	storage := &MockSiteStorage{}
	storage.On("FindEnabled", mock.Anything).Return(sites, nil)

	selector := NewSiteSelector(storage, nil, nil)
	req := createTestRequest("GET", "http://example.com", map[string]string{
		wo.HeaderAcceptLanguage: "en-US,en;q=0.9",
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = selector.Retrieve(req)
	}
}

func BenchmarkRegexpPath(b *testing.B) {
	path := "/blog/([a-z0-9-]+)/([0-9]{4})"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = regexpPath(path)
	}
}

func BenchmarkMatchRequest(b *testing.B) {
	req := createTestRequest("GET", "http://example.com/blog/test-post/2023", nil)
	path := "/blog/([a-z0-9-]+)/([0-9]{4})"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = matchRequest(req, path)
	}
}
