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

	assert.False(t, site.Enabled, "Enabled should be false by default")

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
		Enabled:      true,
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
}
