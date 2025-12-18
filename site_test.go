package pages

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewSite(t *testing.T) {
	site := NewSite()

	assert.False(t, site.Created.IsZero(), "Created time should be set")

	assert.False(t, site.Updated.IsZero(), "Updated time should be set")

	assert.Equal(t, "Localhost", site.Name, "Name should be 'Localhost'")

	assert.Equal(t, "localhost", site.Host, "Host should be 'localhost'")

	assert.Equal(t, "https", site.Scheme, "Scheme should be 'https'")

	assert.Equal(t, "en", site.Locale, "Locale should be 'en'")

	assert.Equal(t, "UTC", site.Timezone, "Timezone should be 'UTC'")

	assert.Equal(t, " | ", site.Separator, "Separator should be ' | '")

	assert.Equal(t, "UTF-8", site.MetaTags.Charset, "MetaTags.Charset should be 'UTF-8'")

	assert.NotNil(t, site.Metadata, "Metadata map should be initialized")

	assert.Equal(t, Draft, site.Status, "Status should be draft by default")

	assert.False(t, site.IsDefault, "IsDefault should be false by default")
}

func TestSite_String(t *testing.T) {
	tests := []struct {
		name string
		site *Site
		want string
	}{
		{
			name: "Site with name",
			site: &Site{Name: "Test Site"},
			want: "Test Site",
		},
		{
			name: "Site without name",
			site: &Site{},
			want: "n/a",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.site.String()
			assert.Equal(t, tt.want, got, "Site.String() should return the expected value")
		})
	}
}

func TestSite_IsLocalhost(t *testing.T) {
	tests := []struct {
		name string
		site *Site
		want bool
	}{
		{
			name: "Empty host",
			site: &Site{Host: ""},
			want: true,
		},
		{
			name: "localhost",
			site: &Site{Host: "localhost"},
			want: true,
		},
		{
			name: "127.0.0.1",
			site: &Site{Host: "127.0.0.1"},
			want: true,
		},
		{
			name: "example.com",
			site: &Site{Host: "example.com"},
			want: false,
		},
		{
			name: "subdomain.example.com",
			site: &Site{Host: "subdomain.example.com"},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.site.IsLocalhost()
			assert.Equal(t, tt.want, got, "Site.IsLocalhost() should return the expected value")
		})
	}
}

func TestSite_Origin(t *testing.T) {
	tests := []struct {
		name string
		site *Site
		want string
	}{
		{
			name: "HTTPS site",
			site: &Site{Scheme: "https", Host: "example.com"},
			want: "https://example.com",
		},
		{
			name: "HTTP site",
			site: &Site{Scheme: "http", Host: "localhost"},
			want: "http://localhost",
		},
		{
			name: "Custom scheme",
			site: &Site{Scheme: "ftp", Host: "files.example.com"},
			want: "ftp://files.example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.site.Origin()
			assert.Equal(t, tt.want, got, "Site.Origin() should return the expected value")
		})
	}
}

func TestSite_URL(t *testing.T) {
	tests := []struct {
		name string
		site *Site
		want string
	}{
		{
			name: "Site with empty relative path",
			site: &Site{Scheme: "https", Host: "example.com", RelativePath: ""},
			want: "https://example.com",
		},
		{
			name: "Site with root relative path",
			site: &Site{Scheme: "https", Host: "example.com", RelativePath: "/"},
			want: "https://example.com",
		},
		{
			name: "Site with relative path starting with /",
			site: &Site{Scheme: "https", Host: "example.com", RelativePath: "/blog"},
			want: "https://example.com/blog",
		},
		{
			name: "Site with relative path without /",
			site: &Site{Scheme: "https", Host: "example.com", RelativePath: "shop"},
			want: "https://example.com/shop",
		},
		{
			name: "Site with nested relative path without /",
			site: &Site{Scheme: "https", Host: "example.com", RelativePath: "api/v1"},
			want: "https://example.com/api/v1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.site.URL()
			assert.Equal(t, tt.want, got, "Site.URL() should return the expected value")
		})
	}
}

func TestSite_Home(t *testing.T) {
	tests := []struct {
		name   string
		site   *Site
		isRoot bool
		want   string
	}{
		{
			name: "Root site",
			site: &Site{
				Scheme: "https",
				Host:   "example.com",
				isRoot: true,
			},
			isRoot: true,
			want:   "https://example.com",
		},
		{
			name: "Non-root site with empty relative path",
			site: &Site{
				Scheme:       "https",
				Host:         "example.com",
				RelativePath: "",
				isRoot:       false,
			},
			isRoot: false,
			want:   "https://example.com",
		},
		{
			name: "Non-root site with relative path",
			site: &Site{
				Scheme:       "https",
				Host:         "example.com",
				RelativePath: "/blog",
				isRoot:       false,
			},
			isRoot: false,
			want:   "https://example.com/blog",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.site.isRoot = tt.isRoot
			got := tt.site.Home()
			assert.Equal(t, tt.want, got, "Site.Home() should return the expected value")
		})
	}
}

func TestSite_Location(t *testing.T) {
	tests := []struct {
		name     string
		site     *Site
		wantErr  bool
		wantName string
	}{
		{
			name:     "Valid timezone UTC",
			site:     &Site{Timezone: "UTC"},
			wantErr:  false,
			wantName: "UTC",
		},
		{
			name:     "Valid timezone America/New_York",
			site:     &Site{Timezone: "America/New_York"},
			wantErr:  false,
			wantName: "America/New_York",
		},
		{
			name:     "Invalid timezone",
			site:     &Site{Timezone: "Invalid/Timezone"},
			wantErr:  true,
			wantName: "UTC", // Should fallback to UTC
		},
		{
			name:     "Empty timezone",
			site:     &Site{Timezone: ""},
			wantErr:  true,
			wantName: "UTC", // Should fallback to UTC
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.site.Location()
			assert.Equal(t, tt.wantName, got.String(), "Site.Location() should return the expected timezone name")
		})
	}
}

func TestSite_Tag(t *testing.T) {
	tests := []struct {
		name     string
		site     *Site
		wantLang string
	}{
		{
			name:     "Valid locale en",
			site:     &Site{Locale: "en"},
			wantLang: "en",
		},
		{
			name:     "Valid locale en-US",
			site:     &Site{Locale: "en-US"},
			wantLang: "en-US",
		},
		{
			name:     "Locale with underscore",
			site:     &Site{Locale: "en_US"},
			wantLang: "en-US", // Should convert underscore to hyphen
		},
		{
			name:     "Invalid locale",
			site:     &Site{Locale: "invalid-locale"},
			wantLang: "en", // Should fallback to English
		},
		{
			name:     "Empty locale",
			site:     &Site{Locale: ""},
			wantLang: "en", // Should fallback to English
		},
		{
			name:     "Complex locale zh-Hans-CN",
			site:     &Site{Locale: "zh-Hans-CN"},
			wantLang: "zh-Hans-CN",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.site.Tag()
			assert.Equal(t, tt.wantLang, got.String(), "Site.Tag() should return the expected language tag")
		})
	}
}

func TestSite_LocationCaching(t *testing.T) {
	site := &Site{Timezone: "UTC"}

	// First call should load the location
	loc1 := site.Location()

	// Second call should return cached location
	loc2 := site.Location()

	assert.Equal(t, loc1, loc2, "Location() should return cached location on subsequent calls")

	// Verify location is cached
	assert.NotNil(t, site.location, "Location() should cache the location in the site struct")
}

func TestSite_TagCaching(t *testing.T) {
	site := &Site{Locale: "en"}

	// First call should parse the tag
	tag1 := site.Tag()

	// Second call should return cached tag
	tag2 := site.Tag()

	assert.Equal(t, tag1, tag2, "Tag() should return cached tag on subsequent calls")

	// Verify tag is cached
	assert.NotNil(t, site.tag, "Tag() should cache the tag in the site struct")
}

func TestSite_FieldDefaults(t *testing.T) {
	site := &Site{}

	// Test that uninitialized fields don't panic
	_ = site.String()
	_ = site.IsLocalhost()
	_ = site.Origin()
	_ = site.URL()
	_ = site.Home()
	_ = site.Location()
	_ = site.Tag()

	// Should not panic even with empty fields
	assert.True(t, site.IsLocalhost(), "Empty host should be considered localhost")

	assert.Equal(t, "://", site.Origin(), "Empty scheme and host should produce '://'")
}

func TestSite_CompleteExample(t *testing.T) {
	site := &Site{
		ID:           "test-site",
		Created:      time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		Updated:      time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
		Name:         "Example Website",
		Title:        "Example Site Title",
		Separator:    " - ",
		Locale:       "en-US",
		Timezone:     "America/New_York",
		Countries:    []string{"US", "CA"},
		Scheme:       "https",
		Host:         "example.com",
		RelativePath: "/blog",
		IsDefault:    true,
		Status:       Published,
		Metadata:     map[string]any{"theme": "dark"},
	}

	// Test all methods
	assert.Equal(t, "Example Website", site.String(), "String() should return 'Example Website'")

	assert.False(t, site.IsLocalhost(), "Should not be localhost")

	assert.Equal(t, "https://example.com", site.Origin(), "Origin() should return 'https://example.com'")

	assert.Equal(t, "https://example.com/blog", site.URL(), "URL() should return 'https://example.com/blog'")

	assert.Equal(t, "https://example.com/blog", site.Home(), "Home() should return 'https://example.com/blog'")

	location := site.Location()
	assert.Equal(t, "America/New_York", location.String(), "Location() should return 'America/New_York'")

	tag := site.Tag()
	assert.Equal(t, "en-US", tag.String(), "Tag() should return 'en-US'")
}

func TestSite_EdgeCases(t *testing.T) {
	t.Run("Site with special characters in host", func(t *testing.T) {
		site := &Site{Scheme: "https", Host: "sub-domain.example.co.uk"}
		assert.Equal(t, "https://sub-domain.example.co.uk", site.Origin(), "Origin() with special characters should return the expected value")
	})

	t.Run("Site with port", func(t *testing.T) {
		site := &Site{Scheme: "http", Host: "localhost:8080"}
		assert.Equal(t, "http://localhost:8080", site.Origin(), "Origin() with port should return the expected value")
	})

	t.Run("Site with IPv6", func(t *testing.T) {
		site := &Site{Scheme: "https", Host: "[::1]"}
		assert.Equal(t, "https://[::1]", site.Origin(), "Origin() with IPv6 should return the expected value")
	})

	t.Run("Complex relative path", func(t *testing.T) {
		site := &Site{Scheme: "https", Host: "example.com", RelativePath: "/api/v1/users"}
		assert.Equal(t, "https://example.com/api/v1/users", site.URL(), "URL() with complex path should return the expected value")
	})

	t.Run("URL with trailing slash in relative path", func(t *testing.T) {
		site := &Site{Scheme: "https", Host: "example.com", RelativePath: "/blog/"}
		assert.Equal(t, "https://example.com/blog/", site.URL(), "URL() should preserve trailing slash in relative path")
	})

	t.Run("URL with query-like relative path", func(t *testing.T) {
		site := &Site{Scheme: "https", Host: "example.com", RelativePath: "search?q=test"}
		assert.Equal(t, "https://example.com/search?q=test", site.URL(), "URL() should handle query-like relative paths")
	})

	t.Run("URL with hash-like relative path", func(t *testing.T) {
		site := &Site{Scheme: "https", Host: "example.com", RelativePath: "section#anchor"}
		assert.Equal(t, "https://example.com/section#anchor", site.URL(), "URL() should handle hash-like relative paths")
	})
}

func TestSite_Copy(t *testing.T) {
	originalTime := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)

	t.Run("Copy basic site", func(t *testing.T) {
		original := &Site{
			ID:           "test-id",
			Created:      originalTime,
			Updated:      originalTime,
			Name:         "Test Site",
			Title:        "Test Title",
			Separator:    " - ",
			Locale:       "en-US",
			Timezone:     "America/New_York",
			Countries:    []string{"US", "CA"},
			Scheme:       "https",
			Host:         "example.com",
			RelativePath: "/blog",
			IsDefault:    true,
			Status:       Published,
			Metadata:     map[string]any{"theme": "dark", "version": 1},
			MetaTags:     NewMetaTags("UTF-8"),
			isRoot:       false,
		}

		// Add some meta tags
		original.MetaTags.SetName("description", "Test description")
		original.MetaTags.SetProperty("og:title", "Test OG Title")

		copy := original.Copy()

		// Verify all fields are copied but not the same reference
		assert.NotSame(t, original, copy, "Copy should return a different instance")
		assert.Equal(t, original.ID, copy.ID, "ID should be copied")
		assert.Equal(t, original.Created, copy.Created, "Created should be copied")
		assert.Equal(t, original.Updated, copy.Updated, "Updated should be copied")
		assert.Equal(t, original.Name, copy.Name, "Name should be copied")
		assert.Equal(t, original.Title, copy.Title, "Title should be copied")
		assert.Equal(t, original.Separator, copy.Separator, "Separator should be copied")
		assert.Equal(t, original.Locale, copy.Locale, "Locale should be copied")
		assert.Equal(t, original.Timezone, copy.Timezone, "Timezone should be copied")
		assert.Equal(t, original.Scheme, copy.Scheme, "Scheme should be copied")
		assert.Equal(t, original.Host, copy.Host, "Host should be copied")
		assert.Equal(t, original.RelativePath, copy.RelativePath, "RelativePath should be copied")
		assert.Equal(t, original.IsDefault, copy.IsDefault, "IsDefault should be copied")
		assert.Equal(t, original.Status, copy.Status, "Status should be copied")

		// Verify slices are cloned
		assert.Equal(t, original.Countries, copy.Countries, "Countries should be copied")
		// Test independence by modifying copy
		copy.Countries = append(copy.Countries, "FR")
		assert.Len(t, original.Countries, 2, "Original countries should not be modified")
		// Restore for later tests
		copy.Countries = copy.Countries[:len(copy.Countries)-1]

		// Verify maps are cloned
		assert.Equal(t, original.Metadata, copy.Metadata, "Metadata should be copied")
		// Test independence by modifying copy
		copy.Metadata["test"] = "value"
		assert.NotContains(t, original.Metadata, "test", "Original metadata should be independent")
		// Remove test key
		delete(copy.Metadata, "test")

		// Verify MetaTags is copied correctly
		assert.NotNil(t, copy.MetaTags, "MetaTags should be copied")
		assert.Equal(t, original.MetaTags.Charset, copy.MetaTags.Charset, "MetaTags charset should be copied")
		assert.Equal(t, original.MetaTags.Name, copy.MetaTags.Name, "MetaTags name should be copied")
		assert.NotSame(t, original.MetaTags, copy.MetaTags, "MetaTags should be a new instance")

		// Verify cached fields are reset
		assert.Nil(t, copy.location, "Location cache should be reset")
		assert.Nil(t, copy.tag, "Tag cache should be reset")
		assert.False(t, copy.isRoot, "isRoot should be reset to false")
	})

	t.Run("Copy site with nil MetaTags", func(t *testing.T) {
		original := &Site{
			Name: "Test Site",
			Host: "example.com",
		}
		original.MetaTags = nil

		copy := original.Copy()

		assert.Nil(t, copy.MetaTags, "MetaTags should remain nil when original has nil MetaTags")
	})

	t.Run("Copy site with empty slices and maps", func(t *testing.T) {
		original := &Site{
			Name:      "Test Site",
			Countries: []string{},
			Metadata:  make(map[string]any),
			MetaTags:  NewMetaTags("UTF-8"),
		}

		copy := original.Copy()

		assert.Empty(t, copy.Countries, "Countries slice should be empty")
		// For empty slices, they're both empty but should be different instances
		copy.Countries = append(copy.Countries, "test")
		assert.Empty(t, original.Countries, "Original countries slice should remain empty")

		assert.Empty(t, copy.Metadata, "Metadata map should be empty")
		// For empty maps, they're both empty but should be different instances
		copy.Metadata["test"] = "value"
		assert.Empty(t, original.Metadata, "Original metadata map should remain empty")
	})

	t.Run("Modify copy doesn't affect original", func(t *testing.T) {
		original := &Site{
			Name:      "Original Site",
			Countries: []string{"US"},
			Metadata:  map[string]any{"key": "original"},
			MetaTags:  NewMetaTags("UTF-8"),
		}

		copy := original.Copy()

		// Modify the copy
		copy.Name = "Modified Site"
		copy.Countries = append(copy.Countries, "CA")
		copy.Metadata["key"] = "modified"
		copy.Metadata["newKey"] = "new value"
		copy.MetaTags.Charset = "ISO-8859-1"

		// Verify original is unchanged
		assert.Equal(t, "Original Site", original.Name, "Original name should be unchanged")
		assert.Equal(t, []string{"US"}, original.Countries, "Original countries should be unchanged")
		assert.Equal(t, "original", original.Metadata["key"], "Original metadata should be unchanged")
		assert.NotContains(t, original.Metadata, "newKey", "Original should not have new metadata key")
		assert.Equal(t, "UTF-8", original.MetaTags.Charset, "Original MetaTags charset should be unchanged")
	})
}

func TestSite_InvalidTimezoneHandling(t *testing.T) {
	t.Run("Multiple calls to Location with invalid timezone", func(t *testing.T) {
		site := &Site{Timezone: "Invalid/Timezone"}

		loc1 := site.Location()
		loc2 := site.Location()

		assert.Same(t, loc1, loc2, "Location() should return cached location even for invalid timezone")
		assert.Equal(t, "UTC", loc1.String(), "Invalid timezone should fallback to UTC")
	})

	t.Run("Valid timezone after invalid one", func(t *testing.T) {
		site := &Site{Timezone: "Invalid/Timezone"}

		// First call with invalid timezone
		loc1 := site.Location()
		assert.Equal(t, "UTC", loc1.String(), "First call with invalid timezone should return UTC")

		// Change to valid timezone
		site.Timezone = "America/New_York"
		site.location = nil // Reset cache

		loc2 := site.Location()
		assert.Equal(t, "America/New_York", loc2.String(), "Second call with valid timezone should return correct location")
	})
}

func TestSite_InvalidLocaleHandling(t *testing.T) {
	t.Run("Multiple calls to Tag with invalid locale", func(t *testing.T) {
		site := &Site{Locale: "invalid-locale"}

		tag1 := site.Tag()
		tag2 := site.Tag()

		assert.Equal(t, tag1, tag2, "Tag() should return cached tag on subsequent calls")
		assert.Equal(t, "en", tag1.String(), "Invalid locale should fallback to English")
	})

	t.Run("Valid locale after invalid one", func(t *testing.T) {
		site := &Site{Locale: "invalid-locale"}

		// First call with invalid locale
		tag1 := site.Tag()
		assert.Equal(t, "en", tag1.String(), "First call with invalid locale should return English")

		// Change to valid locale
		site.Locale = "fr-FR"
		site.tag = nil // Reset cache

		tag2 := site.Tag()
		assert.Equal(t, "fr-FR", tag2.String(), "Second call with valid locale should return correct tag")
	})
}

func TestSite_MetaTagsIntegration(t *testing.T) {
	t.Run("NewSite creates MetaTags with default charset", func(t *testing.T) {
		site := NewSite()

		assert.NotNil(t, site.MetaTags, "NewSite should create MetaTags")
		assert.Equal(t, "UTF-8", site.MetaTags.Charset, "MetaTags should have default charset")
		assert.NotNil(t, site.MetaTags.Name, "MetaTags.Name map should be initialized")
		assert.NotNil(t, site.MetaTags.Property, "MetaTags.Property map should be initialized")
		assert.NotNil(t, site.MetaTags.HTTPEquiv, "MetaTags.HTTPEquiv map should be initialized")
	})

	t.Run("Copy site with complex MetaTags", func(t *testing.T) {
		original := NewSite()

		// Add complex meta tags
		original.MetaTags.SetName("description", "Test site description")
		original.MetaTags.SetName("keywords", "test", "site", "example")
		original.MetaTags.SetProperty("og:title", "Test OG Title")
		original.MetaTags.SetProperty("og:description", "Test OG Description")
		original.MetaTags.SetHTTPEquiv("refresh", "30")
		original.MetaTags.Charset = "ISO-8859-1"

		copy := original.Copy()

		// Verify MetaTags are copied correctly
		assert.NotNil(t, copy.MetaTags, "Copy should have MetaTags")
		assert.Equal(t, "ISO-8859-1", copy.MetaTags.Charset, "Charset should be copied")
		assert.Equal(t, original.MetaTags.Name, copy.MetaTags.Name, "Name meta tags should be copied")
		assert.Equal(t, original.MetaTags.Property, copy.MetaTags.Property, "Property meta tags should be copied")
		assert.Equal(t, original.MetaTags.HTTPEquiv, copy.MetaTags.HTTPEquiv, "HTTP-Equiv meta tags should be copied")
		assert.NotSame(t, original.MetaTags, copy.MetaTags, "MetaTags should be a new instance")

		// Verify modifying copy doesn't affect original
		copy.MetaTags.SetName("new-name", "new value")
		assert.NotContains(t, original.MetaTags.Name, "new-name", "Original MetaTags should not be modified")
	})
}

func BenchmarkSite_String(b *testing.B) {
	site := &Site{Name: "Test Site"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = site.String()
	}
}

func BenchmarkSite_IsLocalhost(b *testing.B) {
	sites := []*Site{
		{Host: "localhost"},
		{Host: "127.0.0.1"},
		{Host: "example.com"},
		{Host: ""},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		site := sites[i%len(sites)]
		_ = site.IsLocalhost()
	}
}

func BenchmarkSite_Origin(b *testing.B) {
	site := &Site{Scheme: "https", Host: "example.com"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = site.Origin()
	}
}

func BenchmarkSite_URL(b *testing.B) {
	site := &Site{Scheme: "https", Host: "example.com", RelativePath: "/api/v1/users"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = site.URL()
	}
}

func BenchmarkSite_Home(b *testing.B) {
	site := &Site{
		Scheme:       "https",
		Host:         "example.com",
		RelativePath: "/blog",
		isRoot:       false,
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = site.Home()
	}
}

func BenchmarkSite_Location_Cached(b *testing.B) {
	site := &Site{Timezone: "UTC"}
	site.Location() // Pre-cache the location
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = site.Location()
	}
}

func BenchmarkSite_Location_Uncached(b *testing.B) {
	sites := []*Site{
		{Timezone: "UTC"},
		{Timezone: "America/New_York"},
		{Timezone: "Europe/London"},
		{Timezone: "Asia/Tokyo"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		site := sites[i%len(sites)]
		site.location = nil // Reset cache each time
		_ = site.Location()
	}
}

func BenchmarkSite_Tag_Cached(b *testing.B) {
	site := &Site{Locale: "en-US"}
	site.Tag() // Pre-cache the tag
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = site.Tag()
	}
}

func BenchmarkSite_Tag_Uncached(b *testing.B) {
	locales := []string{"en-US", "fr-FR", "de-DE", "es-ES", "ja-JP"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		site := &Site{Locale: locales[i%len(locales)]}
		_ = site.Tag()
	}
}

func BenchmarkSite_Copy(b *testing.B) {
	original := &Site{
		ID:        "test-id",
		Created:   time.Now(),
		Updated:   time.Now(),
		Name:      "Test Site",
		Countries: []string{"US", "CA", "MX", "GB", "FR"},
		Metadata:  map[string]any{"theme": "dark", "version": 1, "features": []string{"search", "api"}},
		MetaTags:  NewMetaTags("UTF-8"),
	}
	original.MetaTags.SetName("description", "Test description")
	original.MetaTags.SetProperty("og:title", "Test OG Title")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = original.Copy()
	}
}
