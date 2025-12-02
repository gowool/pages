package pages

import (
	"testing"
	"time"
)

func TestNewSite(t *testing.T) {
	site := NewSite()

	if site.Created.IsZero() {
		t.Error("Created time should be set")
	}

	if site.Updated.IsZero() {
		t.Error("Updated time should be set")
	}

	if site.Name != "Localhost" {
		t.Errorf("Name = %v, want %v", site.Name, "Localhost")
	}

	if site.Host != "localhost" {
		t.Errorf("Host = %v, want %v", site.Host, "localhost")
	}

	if site.Scheme != "https" {
		t.Errorf("Scheme = %v, want %v", site.Scheme, "https")
	}

	if site.Locale != "en" {
		t.Errorf("Locale = %v, want %v", site.Locale, "en")
	}

	if site.Timezone != "UTC" {
		t.Errorf("Timezone = %v, want %v", site.Timezone, "UTC")
	}

	if site.Separator != " | " {
		t.Errorf("Separator = %v, want %v", site.Separator, " | ")
	}

	if site.MetaTags.Charset != "UTF-8" {
		t.Errorf("MetaTags.Charset = %v, want %v", site.MetaTags.Charset, "UTF-8")
	}

	if site.Metadata == nil {
		t.Error("Metadata map should be initialized")
	}

	if site.Enabled {
		t.Error("Enabled should be false by default")
	}

	if site.IsDefault {
		t.Error("IsDefault should be false by default")
	}
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
			if got := tt.site.String(); got != tt.want {
				t.Errorf("Site.String() = %v, want %v", got, tt.want)
			}
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
			if got := tt.site.IsLocalhost(); got != tt.want {
				t.Errorf("Site.IsLocalhost() = %v, want %v", got, tt.want)
			}
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
			if got := tt.site.Origin(); got != tt.want {
				t.Errorf("Site.Origin() = %v, want %v", got, tt.want)
			}
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
			if got := tt.site.URL(); got != tt.want {
				t.Errorf("Site.URL() = %v, want %v", got, tt.want)
			}
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
			if got := tt.site.Home(); got != tt.want {
				t.Errorf("Site.Home() = %v, want %v", got, tt.want)
			}
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
			if got.String() != tt.wantName {
				t.Errorf("Site.Location() = %v, want %v", got.String(), tt.wantName)
			}
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
			if got.String() != tt.wantLang {
				t.Errorf("Site.Tag() = %v, want %v", got.String(), tt.wantLang)
			}
		})
	}
}

func TestSite_LocationCaching(t *testing.T) {
	site := &Site{Timezone: "UTC"}

	// First call should load the location
	loc1 := site.Location()

	// Second call should return cached location
	loc2 := site.Location()

	if loc1 != loc2 {
		t.Error("Location() should return cached location on subsequent calls")
	}

	// Verify location is cached
	if site.location == nil {
		t.Error("Location() should cache the location in the site struct")
	}
}

func TestSite_TagCaching(t *testing.T) {
	site := &Site{Locale: "en"}

	// First call should parse the tag
	tag1 := site.Tag()

	// Second call should return cached tag
	tag2 := site.Tag()

	if tag1 != tag2 {
		t.Error("Tag() should return cached tag on subsequent calls")
	}

	// Verify tag is cached
	if site.tag == nil {
		t.Error("Tag() should cache the tag in the site struct")
	}
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
	if !site.IsLocalhost() {
		t.Error("Empty host should be considered localhost")
	}

	if site.Origin() != "://" {
		t.Errorf("Empty scheme and host should produce '://', got %v", site.Origin())
	}
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
	if site.String() != "Example Website" {
		t.Errorf("String() = %v, want %v", site.String(), "Example Website")
	}

	if site.IsLocalhost() {
		t.Error("Should not be localhost")
	}

	if site.Origin() != "https://example.com" {
		t.Errorf("Origin() = %v, want %v", site.Origin(), "https://example.com")
	}

	if site.URL() != "https://example.com/blog" {
		t.Errorf("URL() = %v, want %v", site.URL(), "https://example.com/blog")
	}

	if site.Home() != "https://example.com/blog" {
		t.Errorf("Home() = %v, want %v", site.Home(), "https://example.com/blog")
	}

	location := site.Location()
	if location.String() != "America/New_York" {
		t.Errorf("Location() = %v, want %v", location.String(), "America/New_York")
	}

	tag := site.Tag()
	if tag.String() != "en-US" {
		t.Errorf("Tag() = %v, want %v", tag.String(), "en-US")
	}
}

func TestSite_EdgeCases(t *testing.T) {
	t.Run("Site with special characters in host", func(t *testing.T) {
		site := &Site{Scheme: "https", Host: "sub-domain.example.co.uk"}
		if site.Origin() != "https://sub-domain.example.co.uk" {
			t.Errorf("Origin() with special characters = %v", site.Origin())
		}
	})

	t.Run("Site with port", func(t *testing.T) {
		site := &Site{Scheme: "http", Host: "localhost:8080"}
		if site.Origin() != "http://localhost:8080" {
			t.Errorf("Origin() with port = %v", site.Origin())
		}
	})

	t.Run("Site with IPv6", func(t *testing.T) {
		site := &Site{Scheme: "https", Host: "[::1]"}
		if site.Origin() != "https://[::1]" {
			t.Errorf("Origin() with IPv6 = %v", site.Origin())
		}
	})

	t.Run("Complex relative path", func(t *testing.T) {
		site := &Site{Scheme: "https", Host: "example.com", RelativePath: "/api/v1/users"}
		if site.URL() != "https://example.com/api/v1/users" {
			t.Errorf("URL() with complex path = %v", site.URL())
		}
	})
}
