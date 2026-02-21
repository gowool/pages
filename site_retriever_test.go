package pages

import (
	"errors"
	"fmt"
	"iter"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gowool/keratin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Test helper functions
func CreateTestSite(name, host, locale string, isDefault bool, countries ...string) *Site {
	site := NewSite()
	site.Name = name
	site.Host = host
	site.Locale = locale
	site.IsDefault = isDefault
	site.Countries = countries
	site.Status = Published
	return site
}

// SitesToIterator converts a slice of sites to an iterator for testing
func SitesToIterator(sites []*Site, err error) iter.Seq2[*Site, error] {
	return func(yield func(*Site, error) bool) {
		if err != nil {
			yield(nil, err)
			return
		}
		for _, site := range sites {
			if !yield(site, nil) {
				break
			}
		}
	}
}

func CreateTestRequest(method, url string, headers map[string]string) *http.Request {
	req := httptest.NewRequest(method, url, nil)
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	return req
}

func TestNewHTTPSiteRetriever(t *testing.T) {
	t.Run("Valid parameters", func(t *testing.T) {
		store := &MockSiteStore{}
		countryFunc := func(r *http.Request) (string, error) { return "US", nil }
		errorFunc := func(r *http.Request, err error) (*Site, error) { return nil, err }

		retriever := NewHTTPSiteRetrieverWithConfig(store, HTTPSiteRetrieverConfig{
			CountryFunc: countryFunc,
			ErrorFunc:   errorFunc,
		})

		assert.NotNil(t, retriever)
		// Test that it implements the interface
		assert.Implements(t, (*SiteRetriever)(nil), retriever)
	})

	t.Run("Nil store should panic", func(t *testing.T) {
		assert.Panics(t, func() {
			NewHTTPSiteRetriever(nil)
		})
	})

	t.Run("Nil countryFunc should use default", func(t *testing.T) {
		store := &MockSiteStore{}
		retriever := NewHTTPSiteRetriever(store)

		req := CreateTestRequest("GET", "http://example.com", map[string]string{
			keratin.HeaderCFIPCountry: "US",
		})

		// Test that the retriever works with default country function
		testSite := CreateTestSite("Test", "example.com", "en", true)
		store.On("FindPublished", mock.Anything).Return(SitesToIterator([]*Site{testSite}, nil))
		site, _, err := retriever.Retrieve(req)

		assert.NoError(t, err)
		assert.NotNil(t, site)
		store.AssertExpectations(t)
	})

	t.Run("Nil errorFunc should use default", func(t *testing.T) {
		store := &MockSiteStore{}
		retriever := NewHTTPSiteRetriever(store)

		req := CreateTestRequest("GET", "http://example.com", nil)
		testErr := errors.New("test error")

		// Test that the retriever returns error directly when no custom error func is provided
		store.On("FindPublished", mock.Anything).Return(SitesToIterator([]*Site{}, testErr))
		site, _, err := retriever.Retrieve(req)

		assert.Error(t, err)
		assert.Equal(t, testErr, err)
		assert.Nil(t, site)
		store.AssertExpectations(t)
	})
}

func TestHTTPSiteRetriever_Retrieve(t *testing.T) {
	t.Run("Nil request should panic", func(t *testing.T) {
		store := &MockSiteStore{}
		retriever := NewHTTPSiteRetriever(store)

		assert.Panics(t, func() {
			_, _, _ = retriever.Retrieve(nil)
		})
	})

	t.Run("Country function error", func(t *testing.T) {
		store := &MockSiteStore{}
		testErr := errors.New("country error")
		countryFunc := func(r *http.Request) (string, error) { return "", testErr }
		errorFunc := func(r *http.Request, err error) (*Site, error) {
			return CreateTestSite("Error Site", "example.com", "en", false), nil
		}

		retriever := NewHTTPSiteRetrieverWithConfig(store, HTTPSiteRetrieverConfig{
			CountryFunc: countryFunc,
			ErrorFunc:   errorFunc,
		})
		req := CreateTestRequest("GET", "http://example.com", nil)

		site, _, err := retriever.Retrieve(req)

		assert.NoError(t, err)
		assert.NotNil(t, site)
		assert.Equal(t, "Error Site", site.Name)
	})

	t.Run("Store error", func(t *testing.T) {
		store := &MockSiteStore{}
		testErr := errors.New("store error")
		countryFunc := func(r *http.Request) (string, error) { return "US", nil }
		errorFunc := func(r *http.Request, err error) (*Site, error) {
			return CreateTestSite("Error Site", "example.com", "en", false), nil
		}

		store.On("FindPublished", mock.Anything).Return(SitesToIterator([]*Site{}, testErr))

		retriever := NewHTTPSiteRetrieverWithConfig(store, HTTPSiteRetrieverConfig{
			CountryFunc: countryFunc,
			ErrorFunc:   errorFunc,
		})
		req := CreateTestRequest("GET", "http://example.com", nil)

		site, _, err := retriever.Retrieve(req)

		assert.NoError(t, err)
		assert.NotNil(t, site)
		assert.Equal(t, "Error Site", site.Name)

		store.AssertExpectations(t)
	})

	t.Run("Successful site selection", func(t *testing.T) {
		store := &MockSiteStore{}
		sites := []*Site{
			CreateTestSite("Default Site", "example.com", "en", true),
			CreateTestSite("French Site", "example.com", "fr", false),
			CreateTestSite("US Site", "example.com", "en-US", false, "US"),
		}

		store.On("FindPublished", mock.Anything).Return(SitesToIterator(sites, nil))

		retriever := NewHTTPSiteRetriever(store)
		req := CreateTestRequest("GET", "http://example.com", map[string]string{
			keratin.HeaderAcceptLanguage: "fr-FR,fr;q=0.9,en;q=0.8",
		})

		site, _, err := retriever.Retrieve(req)

		assert.NoError(t, err)
		assert.NotNil(t, site)
		assert.Equal(t, "French Site", site.Name)

		store.AssertExpectations(t)
	})

	t.Run("Site selection with country", func(t *testing.T) {
		store := &MockSiteStore{}
		sites := []*Site{
			CreateTestSite("Default Site", "example.com", "en", true),
			CreateTestSite("US Site", "example.com", "en-US", false, "US"),
			CreateTestSite("French Site", "example.com", "fr", false),
		}

		store.On("FindPublished", mock.Anything).Return(SitesToIterator(sites, nil))

		retriever := NewHTTPSiteRetrieverWithConfig(store, HTTPSiteRetrieverConfig{
			CountryFunc: func(r *http.Request) (string, error) { return "US", nil },
		})
		req := CreateTestRequest("GET", "http://example.com/test", map[string]string{keratin.HeaderAcceptLanguage: "en-US;q=0.9,en;q=0.8"})

		site, _, err := retriever.Retrieve(req)

		assert.NoError(t, err)
		assert.NotNil(t, site)
		assert.Equal(t, "US Site", site.Name)

		store.AssertExpectations(t)
	})

	t.Run("No matching sites returns error", func(t *testing.T) {
		store := &MockSiteStore{}
		sites := []*Site{
			CreateTestSite("Different Host", "other.com", "en", false),
		}

		store.On("FindPublished", mock.Anything).Return(SitesToIterator(sites, nil))

		retriever := NewHTTPSiteRetriever(store)
		req := CreateTestRequest("GET", "http://example.com", nil)

		site, _, err := retriever.Retrieve(req)

		assert.Error(t, err)
		assert.Equal(t, ErrSiteNotFound, err)
		assert.Nil(t, site)

		store.AssertExpectations(t)
	})

	t.Run("Host matching", func(t *testing.T) {
		store := &MockSiteStore{}
		sites := []*Site{
			CreateTestSite("Correct Host", "example.com", "en", true),
			CreateTestSite("Wrong Host", "other.com", "en", false),
		}

		store.On("FindPublished", mock.Anything).Return(SitesToIterator(sites, nil))

		retriever := NewHTTPSiteRetriever(store)
		req := CreateTestRequest("GET", "http://example.com", nil)

		site, _, err := retriever.Retrieve(req)

		assert.NoError(t, err)
		assert.NotNil(t, site)
		assert.Equal(t, "Correct Host", site.Name)

		store.AssertExpectations(t)
	})
}

// Integration tests that indirectly test selectedSite functionality
func TestHTTPSiteRetriever_LanguageMatching(t *testing.T) {
	t.Run("Language preference matching", func(t *testing.T) {
		store := &MockSiteStore{}

		englishSite := CreateTestSite("English", "example.com", "en", false)
		frenchSite := CreateTestSite("French", "example.com", "fr", false)
		spanishSite := CreateTestSite("Spanish", "example.com", "es", false)

		sites := []*Site{englishSite, frenchSite, spanishSite}
		store.On("FindPublished", mock.Anything).Return(SitesToIterator(sites, nil))

		retriever := NewHTTPSiteRetriever(store)

		// Test French preference
		req := CreateTestRequest("GET", "http://example.com", map[string]string{
			keratin.HeaderAcceptLanguage: "fr-FR,fr;q=0.9,en;q=0.8",
		})

		site, _, err := retriever.Retrieve(req)

		require.NoError(t, err)
		assert.Equal(t, "French", site.Name)

		store.AssertExpectations(t)
	})

	t.Run("Fallback to parent language", func(t *testing.T) {
		store := &MockSiteStore{}

		englishSite := CreateTestSite("English", "example.com", "en", false)

		sites := []*Site{englishSite}
		store.On("FindPublished", mock.Anything).Return(SitesToIterator(sites, nil))

		retriever := NewHTTPSiteRetriever(store)

		// Test with en-US which should fallback to en
		req := CreateTestRequest("GET", "http://example.com", map[string]string{
			keratin.HeaderAcceptLanguage: "en-US,en;q=0.9",
		})

		site, _, err := retriever.Retrieve(req)

		require.NoError(t, err)
		assert.Equal(t, "English", site.Name)

		store.AssertExpectations(t)
	})

	t.Run("Invalid accept language header", func(t *testing.T) {
		store := &MockSiteStore{}

		defaultSite := CreateTestSite("Default", "example.com", "en", true)

		sites := []*Site{defaultSite}
		store.On("FindPublished", mock.Anything).Return(SitesToIterator(sites, nil))

		retriever := NewHTTPSiteRetriever(store)

		// Test with invalid Accept-Language header
		req := CreateTestRequest("GET", "http://example.com", map[string]string{
			keratin.HeaderAcceptLanguage: "invalid-language-header",
		})

		site, _, err := retriever.Retrieve(req)

		require.NoError(t, err)
		assert.Equal(t, "Default", site.Name)

		store.AssertExpectations(t)
	})
}

func TestHTTPSiteRetriever_Integration(t *testing.T) {
	t.Run("Complete site selection workflow", func(t *testing.T) {
		// Create test sites
		defaultSite := CreateTestSite("Default", "example.com", "en", true)
		frenchSite := CreateTestSite("French", "example.com", "fr", false)
		germanSite := CreateTestSite("German", "example.com", "de", false)
		usSite := CreateTestSite("US", "example.com", "en-US", false, "US")

		sites := []*Site{defaultSite, frenchSite, germanSite, usSite}

		store := &MockSiteStore{}
		store.On("FindPublished", mock.Anything).Return(SitesToIterator(sites, nil))

		retriever := NewHTTPSiteRetrieverWithConfig(store, HTTPSiteRetrieverConfig{
			CountryFunc: func(r *http.Request) (string, error) { return "US", nil },
		})

		t.Run("French language preference", func(t *testing.T) {
			req := CreateTestRequest("GET", "http://example.com", map[string]string{
				keratin.HeaderAcceptLanguage: "fr-FR,fr;q=0.9,en;q=0.8",
			})

			site, _, err := retriever.Retrieve(req)

			require.NoError(t, err)
			require.NotNil(t, site)
			assert.Equal(t, "US", site.Name)
		})

		t.Run("US country with path", func(t *testing.T) {
			req := CreateTestRequest("GET", "http://example.com/blog/test", map[string]string{
				keratin.HeaderAcceptLanguage: "en-US,en;q=0.9",
			})

			site, pathInfo, err := retriever.Retrieve(req)

			require.NoError(t, err)
			require.NotNil(t, site)
			assert.Equal(t, "US", site.Name)
			assert.Equal(t, "/blog/test", pathInfo)
		})

		t.Run("Fallback to default", func(t *testing.T) {
			req := CreateTestRequest("GET", "http://example.com", map[string]string{
				keratin.HeaderAcceptLanguage: "es-ES,es;q=0.9",
			})

			site, _, err := retriever.Retrieve(req)

			require.NoError(t, err)
			require.NotNil(t, site)
			assert.Equal(t, "US", site.Name)
		})

		store.AssertExpectations(t)
	})

	t.Run("Multi-host environment", func(t *testing.T) {
		site1 := CreateTestSite("Site1", "site1.com", "en", true)
		site2 := CreateTestSite("Site2", "site2.com", "fr", true)

		sites := []*Site{site1, site2}

		store := &MockSiteStore{}
		store.On("FindPublished", mock.Anything).Return(SitesToIterator(sites, nil))

		retriever := NewHTTPSiteRetriever(store)

		// Test site1.com
		req1 := CreateTestRequest("GET", "http://site1.com", nil)
		site, _, err := retriever.Retrieve(req1)
		require.NoError(t, err)
		assert.Equal(t, "Site1", site.Name)

		// Test site2.com
		req2 := CreateTestRequest("GET", "http://site2.com", nil)
		site, _, err = retriever.Retrieve(req2)
		require.NoError(t, err)
		assert.Equal(t, "Site2", site.Name)

		store.AssertExpectations(t)
	})
}

func TestHTTPSiteRetriever_CountryBasedSelection(t *testing.T) {
	t.Run("Country filtering for root path", func(t *testing.T) {
		usSite := CreateTestSite("US Site", "example.com", "en", false, "US")
		euSite := CreateTestSite("EU Site", "example.com", "en", false, "FR", "DE", "IT")
		globalSite := CreateTestSite("Global Site", "example.com", "en", true)

		sites := []*Site{usSite, euSite, globalSite}

		store := &MockSiteStore{}
		store.On("FindPublished", mock.Anything).Return(SitesToIterator(sites, nil))

		retriever := NewHTTPSiteRetrieverWithConfig(store, HTTPSiteRetrieverConfig{
			CountryFunc: func(r *http.Request) (string, error) { return "US", nil },
		})

		req := CreateTestRequest("GET", "http://example.com/", nil)
		site, _, err := retriever.Retrieve(req)

		require.NoError(t, err)
		require.NotNil(t, site)
		// Should have US site and EU site available (both match countries)
		assert.Contains(t, []*Site{usSite, euSite}, site)

		store.AssertExpectations(t)
	})

	t.Run("Country restriction with no match", func(t *testing.T) {
		euOnlySite := CreateTestSite("EU Only", "eu-example.com", "en", false, "FR", "DE")
		defaultSite := CreateTestSite("Default", "other.com", "en", true) // Different host

		sites := []*Site{euOnlySite, defaultSite}

		store := &MockSiteStore{}
		store.On("FindPublished", mock.Anything).Return(SitesToIterator(sites, nil))

		retriever := NewHTTPSiteRetrieverWithConfig(store, HTTPSiteRetrieverConfig{
			CountryFunc: func(r *http.Request) (string, error) { return "US", nil },
		})

		req := CreateTestRequest("GET", "http://example.com/", nil)
		site, _, err := retriever.Retrieve(req)

		assert.Error(t, err)
		assert.Equal(t, ErrSiteNotFound, err)
		assert.Nil(t, site)

		store.AssertExpectations(t)
	})
}

// Benchmark tests
func BenchmarkSiteRetriever_Retrieve(b *testing.B) {
	sites := make([]*Site, 100)
	for i := range 100 {
		sites[i] = CreateTestSite(
			fmt.Sprintf("Site %d", i),
			"example.com",
			fmt.Sprintf("en-%d", i),
			i == 0, // First site is default
		)
	}

	store := &MockSiteStore{}
	store.On("FindPublished", mock.Anything).Return(SitesToIterator(sites, nil))

	retriever := NewHTTPSiteRetriever(store)
	req := CreateTestRequest("GET", "http://example.com", map[string]string{
		keratin.HeaderAcceptLanguage: "en-US,en;q=0.9",
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = retriever.Retrieve(req)
	}
}

func TestHTTPSiteRetriever_CandidateEdgeCases(t *testing.T) {
	mockStore := &MockSiteStore{}
	mockStore.On("FindPublished", mock.Anything).Return(SitesToIterator([]*Site{}, nil))
	retriever := NewHTTPSiteRetriever(mockStore)

	t.Run("ParseAcceptLanguage returns empty tags", func(t *testing.T) {
		site := CreateTestSite("Test", "example.com", "en-US", false)
		candidates := []candidate{{site: site}}
		req := CreateTestRequest("GET", "/", map[string]string{
			keratin.HeaderAcceptLanguage: "",
		})

		resultSite, path := retriever.candidate(req, candidates, "")
		assert.Equal(t, site, resultSite)
		assert.Equal(t, "", path)
	})

	t.Run("ParseAcceptLanguage returns error", func(t *testing.T) {
		site := CreateTestSite("Test", "example.com", "en-US", false)
		candidates := []candidate{{site: site}}
		req := CreateTestRequest("GET", "/", map[string]string{
			keratin.HeaderAcceptLanguage: "invalid***language",
		})

		resultSite, path := retriever.candidate(req, candidates, "")
		assert.NotNil(t, resultSite)
		assert.Equal(t, "", path)
	})
}

func TestHTTPSiteRetriever_ResolveErrorWithMatchRequestError(t *testing.T) {
	mockStore := &MockSiteStore{}
	mockStore.On("FindPublished", mock.Anything).Return(SitesToIterator([]*Site{}, nil))

	expectedSite := CreateTestSite("Test", "example.com", "en-US", false)
	expectedSite.RelativePath = "/invalid([regex"

	retriever := NewHTTPSiteRetrieverWithConfig(mockStore, HTTPSiteRetrieverConfig{
		ErrorFunc: func(r *http.Request, err error) (*Site, error) {
			return expectedSite, nil
		},
	})
	req := CreateTestRequest("GET", "http://example.com/test", nil)

	site, path, err := retriever.resolveError(req, errors.New("test error"))
	assert.NoError(t, err)
	assert.Equal(t, expectedSite, site)
	assert.Equal(t, "", path)
}

func TestHTTPSiteRetriever_Retrieve_StoreErrorContinue(t *testing.T) {
	mockStore := &MockSiteStore{}

	expectedSite := CreateTestSite("Error Site", "example.com", "en-US", true)

	var iterator iter.Seq2[*Site, error] = func(yield func(*Site, error) bool) {
		if !yield(nil, errors.New("first error")) {
			return
		}
		yield(expectedSite, nil)
	}

	mockStore.On("FindPublished", mock.Anything).Return(iterator)

	retriever := NewHTTPSiteRetrieverWithConfig(mockStore, HTTPSiteRetrieverConfig{
		ErrorFunc: func(r *http.Request, err error) (*Site, error) {
			return nil, nil
		},
	})
	req := CreateTestRequest("GET", "http://example.com", nil)

	site, _, err := retriever.Retrieve(req)
	assert.NoError(t, err)
	assert.Equal(t, expectedSite, site)

	mockStore.AssertExpectations(t)
}

func TestHTTPSiteRetriever_Retrieve_MultipleDefaults(t *testing.T) {
	mockStore := &MockSiteStore{}

	defaultSite1 := CreateTestSite("Default1", "example.com", "en-US", true)
	defaultSite2 := CreateTestSite("Default2", "example.com", "fr", true)

	sites := []*Site{defaultSite1, defaultSite2}
	mockStore.On("FindPublished", mock.Anything).Return(SitesToIterator(sites, nil))

	retriever := NewHTTPSiteRetriever(mockStore)
	req := CreateTestRequest("GET", "http://example.com", nil)

	site, _, err := retriever.Retrieve(req)
	assert.NoError(t, err)
	assert.NotNil(t, site)

	mockStore.AssertExpectations(t)
}

func TestHTTPSiteRetriever_Retrieve_LanguagePreferenceFallback(t *testing.T) {
	mockStore := &MockSiteStore{}

	site1 := CreateTestSite("Site1", "example.com", "en-US", false)
	site2 := CreateTestSite("Site2", "example.com", "fr", false)

	sites := []*Site{site1, site2}
	mockStore.On("FindPublished", mock.Anything).Return(SitesToIterator(sites, nil))

	retriever := NewHTTPSiteRetriever(mockStore)
	req := CreateTestRequest("GET", "http://example.com", map[string]string{
		keratin.HeaderAcceptLanguage: "de-DE,de;q=0.9,en-US;q=0.8,en;q=0.7",
	})

	site, _, err := retriever.Retrieve(req)
	assert.NoError(t, err)
	assert.NotNil(t, site)

	mockStore.AssertExpectations(t)
}

func TestHTTPSiteRetriever_Retrieve_SitesMatchNoCandidates(t *testing.T) {
	mockStore := &MockSiteStore{}

	site := CreateTestSite("NoMatch", "example.com", "en-US", false)
	site.RelativePath = "/nomatch"

	sites := []*Site{site}
	mockStore.On("FindPublished", mock.Anything).Return(SitesToIterator(sites, nil))

	retriever := NewHTTPSiteRetriever(mockStore)
	req := CreateTestRequest("GET", "http://example.com/test", nil)

	resultSite, path, err := retriever.Retrieve(req)
	assert.NoError(t, err)
	assert.Equal(t, site, resultSite)
	assert.Equal(t, "", path)

	mockStore.AssertExpectations(t)
}

func TestHTTPSiteRetriever_Retrieve_LocalhostFallback(t *testing.T) {
	mockStore := &MockSiteStore{}

	localhostSite := CreateTestSite("Localhost", "localhost", "en-US", false)
	localhostSite.RelativePath = "/nomatch"

	sites := []*Site{localhostSite}
	mockStore.On("FindPublished", mock.Anything).Return(SitesToIterator(sites, nil))

	retriever := NewHTTPSiteRetriever(mockStore)
	req := CreateTestRequest("GET", "http://example.com/test", nil)

	resultSite, path, err := retriever.Retrieve(req)
	assert.NoError(t, err)
	assert.Equal(t, localhostSite, resultSite)
	assert.Equal(t, "", path)

	mockStore.AssertExpectations(t)
}

func TestHTTPSiteRetriever_ExactLanguageMatch(t *testing.T) {
	mockStore := &MockSiteStore{}

	site := CreateTestSite("French", "example.com", "fr", false)

	sites := []*Site{site}
	mockStore.On("FindPublished", mock.Anything).Return(SitesToIterator(sites, nil))

	retriever := NewHTTPSiteRetriever(mockStore)
	req := CreateTestRequest("GET", "http://example.com/test", map[string]string{
		keratin.HeaderAcceptLanguage: "fr-FR,fr;q=0.9,en;q=0.8",
	})

	resultSite, path, err := retriever.Retrieve(req)
	assert.NoError(t, err)
	assert.Equal(t, site, resultSite)
	assert.Equal(t, "/test", path)

	mockStore.AssertExpectations(t)
}

func TestHTTPSiteRetriever_MultipleCountriesSameLanguage(t *testing.T) {
	mockStore := &MockSiteStore{}

	usSite := CreateTestSite("US Site", "example.com", "en", false, "US")
	gbSite := CreateTestSite("GB Site", "example.com", "en", false, "GB")

	sites := []*Site{usSite, gbSite}
	mockStore.On("FindPublished", mock.Anything).Return(SitesToIterator(sites, nil))

	retriever := NewHTTPSiteRetrieverWithConfig(mockStore, HTTPSiteRetrieverConfig{
		CountryFunc: func(r *http.Request) (string, error) { return "US", nil },
	})
	req := CreateTestRequest("GET", "http://example.com/test", nil)

	resultSite, path, err := retriever.Retrieve(req)
	assert.NoError(t, err)
	assert.Equal(t, usSite, resultSite)
	assert.Equal(t, "/test", path)

	mockStore.AssertExpectations(t)
}

func TestHTTPSiteRetriever_ParentTagFallback(t *testing.T) {
	mockStore := &MockSiteStore{}

	site := CreateTestSite("English Generic", "example.com", "en", false)

	sites := []*Site{site}
	mockStore.On("FindPublished", mock.Anything).Return(SitesToIterator(sites, nil))

	retriever := NewHTTPSiteRetriever(mockStore)
	req := CreateTestRequest("GET", "http://example.com/test", map[string]string{
		keratin.HeaderAcceptLanguage: "en-GB,en;q=0.9",
	})

	resultSite, path, err := retriever.Retrieve(req)
	assert.NoError(t, err)
	assert.Equal(t, site, resultSite)
	assert.Equal(t, "/test", path)

	mockStore.AssertExpectations(t)
}
