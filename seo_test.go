package pages

import (
	"reflect"
	"testing"
	"time"
)

func TestNewSEO(t *testing.T) {
	seo := NewSEO()
	if seo == nil {
		t.Fatal("NewSEO() should not return nil")
	}

	// Check default separator
	if seo.separator != " - " {
		t.Errorf("Separator = %v, want %v", seo.separator, " - ")
	}

	// Check default HTML attributes
	if seo.htmlAttrs == nil {
		t.Error("htmlAttrs should be initialized")
	}

	// Check default meta tags
	if seo.metaTags == nil {
		t.Error("MetaTags should be initialized")
	}

	// Check that the default meta tags have expected og:type
	if len(seo.metaTags.Property["og:type"]) != 1 || seo.metaTags.Property["og:type"][0] != "website" {
		t.Errorf("og:type = %v, want %v", seo.metaTags.Property["og:type"], []string{"website"})
	}

	// Check other initializations
	if seo.titles != nil {
		t.Error("titles should be nil initially")
	}
	if seo.siteURL != "" {
		t.Error("siteURL should be empty initially")
	}
	if seo.separator != " - " {
		t.Errorf("separator = %v, want %v", seo.separator, " - ")
	}
	if seo.htmlAttrs == nil {
		t.Error("headAttrs should be initialized")
	}
	if seo.bodyAttrs == nil {
		t.Error("bodyAttrs should be initialized")
	}
	if seo.langAlternates == nil {
		t.Error("langAlternates should be initialized")
	}
	if len(seo.headLinks) != 0 {
		t.Error("headLinks should be nil initially")
	}
}

func TestSEO_Reset(t *testing.T) {
	seo := NewSEO()

	// Modify some values
	seo.titles = []string{"Test Title"}
	seo.siteURL = "https://example.com"
	seo.separator = " | "
	seo.htmlAttrs["test"] = "value"
	seo.headAttrs["test"] = "value"
	seo.bodyAttrs["test"] = "value"
	seo.langAlternates["test"] = "value"
	seo.headLinks = []HeadLink{{Rel: "test"}}

	// Reset
	seo.Reset()

	// Check values are reset to defaults
	if len(seo.titles) != 0 {
		t.Error("titles should be nil after reset")
	}
	if seo.siteURL != "" {
		t.Errorf("siteURL = %v, want empty", seo.siteURL)
	}
	if seo.separator != " - " {
		t.Errorf("separator = %v, want %v", seo.separator, " - ")
	}
	if seo.htmlAttrs["test"] == "value" || seo.headAttrs["test"] == "value" {
		t.Error("HTML attributes should be reset to defaults")
	}
	if seo.bodyAttrs["test"] == "value" {
		t.Error("Body attributes should be reset to defaults")
	}
	if seo.langAlternates["test"] == "value" {
		t.Error("Lang alternates should be reset to defaults")
	}
}

func TestSEO_Site(t *testing.T) {
	// Test with complete site
	site1 := &Site{
		Title:        "Test Site",
		Separator:    " | ",
		Locale:       "en_US",
		Scheme:       "https",
		Host:         "example.com",
		RelativePath: "",
		MetaTags:     NewMetaTags("UTF-8"),
	}
	want1 := "Test Site"

	// Test with minimal site
	site2 := &Site{
		Scheme:       "https",
		Host:         "example.com",
		RelativePath: "",
		MetaTags:     NewMetaTags("UTF-8"),
	}
	want2 := ""

	tests := []struct {
		name string
		site *Site
		want string
	}{
		{
			name: "Complete site",
			site: site1,
			want: want1,
		},
		{
			name: "Minimal site",
			site: site2,
			want: want2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			seo := NewSEO()
			if tt.site.MetaTags != nil {
				tt.site.MetaTags.SetName("description", "Site description")
			}
			seo.Site(tt.site)

			if tt.want != "" && (len(seo.titles) != 1 || seo.titles[0] != tt.want) {
				t.Errorf("Site() titles = %v, want %v", seo.titles, []string{tt.want})
			}

			if tt.site.Separator != "" && seo.separator != tt.site.Separator {
				t.Errorf("Site() separator = %v, want %v", seo.separator, tt.site.Separator)
			}

			if tt.site.Locale != "" && seo.htmlAttrs["lang"] != "en-US" {
				t.Errorf("Site() lang = %v, want %v", seo.htmlAttrs["lang"], "en-US")
			}

			expectedHome := tt.site.Origin()
			if len(seo.metaTags.Property["og:url"]) == 0 || seo.metaTags.Property["og:url"][0] != expectedHome {
				t.Errorf("Site() og:url = %v, want %v", seo.metaTags.Property["og:url"], []string{expectedHome})
			}

			if tt.site.MetaTags != nil {
				if len(seo.metaTags.Name["description"]) == 0 || seo.metaTags.Name["description"][0] != tt.site.MetaTags.Name["description"][0] {
					t.Errorf("Site() description = %v, want %v", seo.metaTags.Name["description"], tt.site.MetaTags.Name["description"][0])
				}
			}
		})
	}
}

func TestSEO_Prefix(t *testing.T) {
	seo := NewSEO()

	// Test adding prefix to existing prefix
	originalPrefix := seo.htmlAttrs["prefix"]
	newPrefix := "fb: https://ogp.me/ns/fb#"

	seo.AddHTMLPrefixAttribute(newPrefix)
	expectedPrefix := originalPrefix + " " + newPrefix
	if seo.htmlAttrs["prefix"] != expectedPrefix {
		t.Errorf("AddHTMLPrefixAttribute() = %v, want %v", seo.htmlAttrs["prefix"], expectedPrefix)
	}

	// Test adding prefix to empty prefix
	seo.Reset()
	seo.AddHTMLPrefixAttribute("test: http://example.com")

	if seo.htmlAttrs["prefix"] != "og: https://ogp.me/ns# test: http://example.com" {
		t.Errorf("AddHTMLPrefixAttribute() to empty = %v, want %v", seo.htmlAttrs["prefix"], "og: https://ogp.me/ns# test: http://example.com")
	}

	// Test adding prefix that starts with http
	seo.Reset()
	seo.AddHTMLPrefixAttribute("http://example.com")

	if seo.htmlAttrs["prefix"] != "og: https://ogp.me/ns# http://example.com" {
		t.Errorf("AddHTMLPrefixAttribute() to http = %v, want %v", seo.htmlAttrs["prefix"], "og: https://ogp.me/ns# http://example.com")
	}
}

func TestSEO_AttributeMethods(t *testing.T) {
	seo := NewSEO()

	// Test HTML attributes
	htmlAttrs := seo.HTMLAttributes()
	if htmlAttrs == nil {
		t.Error("HTMLAttributes() should not return nil")
	}

	// Test setting attributes
	seo.SetHTMLAttributes(map[string]string{
		"test": "value",
	})
	if seo.htmlAttrs["test"] != "value" {
		t.Errorf("SetHTMLAttributes() = %v, want %v", seo.htmlAttrs, map[string]string{"test": "value"})
	}

	// Test setting single attribute
	seo.SetHTMLAttribute("new", "attr")
	if seo.htmlAttrs["new"] != "attr" {
		t.Errorf("SetHTMLAttribute() = %v, want %v", seo.htmlAttrs, map[string]string{"test": "value", "new": "attr"})
	}

	// Test removing attribute
	seo.RemoveHTMLAttribute("test")
	if _, exists := seo.htmlAttrs["test"]; exists {
		t.Error("RemoveHTMLAttribute() should remove attribute")
	}

	// Test has attribute
	if !seo.HasHTMLAttribute("new") {
		t.Error("HasHTMLAttribute() should return true for existing attribute")
	}

	if seo.HasHTMLAttribute("nonexistent") {
		t.Error("HasHTMLAttribute() should return false for non-existent attribute")
	}

	// Test head attributes
	headAttrs := seo.HeadAttributes()
	if headAttrs == nil {
		t.Error("HeadAttributes() should not return nil")
	}

	seo.SetHeadAttribute("head", "attr")
	if headAttrs["head"] != "attr" {
		t.Errorf("SetHeadAttribute() = %v, want %v", headAttrs, map[string]string{"head": "attr"})
	}

	seo.RemoveHeadAttribute("head")
	if _, exists := seo.headAttrs["head"]; exists {
		t.Error("RemoveHeadAttribute() should remove attribute")
	}

	// Test body attributes
	bodyAttrs := seo.BodyAttributes()
	if bodyAttrs == nil {
		t.Error("BodyAttributes() should not return nil")
	}

	seo.SetBodyAttribute("body", "attr")
	if bodyAttrs["body"] != "attr" {
		t.Errorf("SetBodyAttribute() = %v, want %v", bodyAttrs, map[string]string{"body": "attr"})
	}

	seo.RemoveBodyAttribute("body")
	if _, exists := seo.bodyAttrs["body"]; exists {
		t.Error("RemoveBodyAttribute() should remove attribute")
	}

	// Test has attributes
	if !seo.HasHTMLAttribute("new") {
		t.Error("HasHTMLAttribute() should return true for existing attribute")
	}

	if seo.HasHTMLAttribute("nonexistent") {
		t.Error("HasHTMLAttribute() should return false for non-existent attribute")
	}
}

func TestSEO_Links(t *testing.T) {
	seo := NewSEO()

	// Test Links returns nil initially
	links := seo.Links()
	if links != nil {
		t.Error("Links() should return nil initially")
	}

	// Test setting links
	seo.SetLinks([]HeadLink{{Rel: "stylesheet", Href: "/style1.css"}})
	if len(seo.headLinks) != 1 {
		t.Errorf("SetLinks() length = %v, want 1", len(seo.headLinks))
	}

	// Test replacing links with a different one
	seo.SetLinks([]HeadLink{{Rel: "stylesheet", Href: "/style2.css"}})
	if len(seo.headLinks) != 1 {
		t.Errorf("SetLinks() length = %v, want 1", len(seo.headLinks))
	}

	seo.AddCanonicalLink("https://example.com/canonical")
	seo.AddPrevLink("https://example.com/prev")
	seo.AddNextLink("https://example.com/next")

	// Check specific links
	foundCanonical := false
	foundPrev := false
	foundNext := false
	for _, link := range seo.headLinks {
		switch link.Rel {
		case "canonical":
			foundCanonical = true
		case "prev":
			foundPrev = true
		case "next":
			foundNext = true
		}
	}

	if !foundCanonical {
		t.Error("AddCanonicalLink() should add canonical link")
	}

	if !foundPrev {
		t.Error("AddPrevLink() should add prev link")
	}

	if !foundNext {
		t.Error("AddNextLink() should add next link")
	}

	// Test that AddLink appends to existing links
	originalCount := len(seo.headLinks)
	seo.AddLink(HeadLink{Rel: "test", Href: "/test"})
	if len(seo.headLinks) != originalCount+1 {
		t.Errorf("AddLink() should append to existing links, got %v, want %v", len(seo.headLinks), originalCount+1)
	}

	// Test replacing links
	seo.SetLinks([]HeadLink{{Rel: "stylesheet", Href: "/new.css"}})
	if len(seo.headLinks) != 1 {
		t.Errorf("SetLinks() should replace existing links, got %v, want 1", len(seo.headLinks))
	}

	if seo.headLinks[0].Rel != "stylesheet" || seo.headLinks[0].Href != "/new.css" {
		t.Errorf("SetLinks() should replace links with new ones")
	}
}

func TestSEO_RemoveLinks(t *testing.T) {
	seo := NewSEO()

	// Test replacing links with single stylesheet
	seo.SetLinks([]HeadLink{{Rel: "stylesheet", Href: "/style2.css"}})
	if len(seo.headLinks) != 1 {
		t.Errorf("SetLinks() should set single link, got %v, want 1", len(seo.headLinks))
	}

	if len(seo.headLinks) == 1 {
		if seo.headLinks[0].Rel != "stylesheet" || seo.headLinks[0].Href != "/style2.css" {
			t.Error("SetLinks() should set correct stylesheet link")
		}
	}
}

func TestSEO_LangAlternates(t *testing.T) {
	seo := NewSEO()

	// Test lang alternates returns empty map initially
	langAlts := seo.LangAlternates()
	if len(langAlts) != 0 {
		t.Errorf("LangAlternates() should return empty map initially, got %v", langAlts)
	}

	// Test setting lang alternates
	newLangAlts := map[string]string{
		"https://example.com/fr": "fr",
		"https://example.com/es": "es",
	}
	seo.SetLangAlternates(newLangAlts)

	// Test values are properly set by getting fresh reference
	currentLangAlts := seo.LangAlternates()
	if !reflect.DeepEqual(currentLangAlts, newLangAlts) {
		t.Errorf("SetLangAlternates() = %v, want %v", currentLangAlts, newLangAlts)
	}

	// Test adding lang alternate
	seo.AddLangAlternate("https://example.com/de", "de")

	// Test has lang alternate
	if !seo.HasLangAlternate("https://example.com/fr") {
		t.Error("AddLangAlternate() should add lang alternate")
	}

	// Test getting lang alternates
	retrievedLangAlts := seo.LangAlternates()
	if len(retrievedLangAlts) != 3 {
		t.Errorf("LangAlternates() = %v, want 3", len(retrievedLangAlts))
	}

	// Test contains lang alternate
	if !seo.HasLangAlternate("https://example.com/de") {
		t.Error("LangAlternates should contain de alternate")
	}

	// Test removing lang alternate
	seo.RemoveLangAlternate("https://example.com/fr")
	if _, exists := seo.langAlternates["https://example.com/fr"]; exists {
		t.Error("RemoveLangAlternate() should remove lang alternate")
	}

	// Test has non-existent lang alternate
	if seo.HasLangAlternate("https://example.com/it") {
		t.Error("HasLangAlternate() should return false for non-existent alternate")
	}
}

func TestSEO_MergeMetaTags(t *testing.T) {
	seo := NewSEO()

	// Test that MergeMetaTags copies description to og:description if not present
	newMetaTags := &MetaTags{
		Name: map[string][]string{
			"description": {"New description"},
		},
		Property: map[string][]string{
			"og:description": {"New OG description"},
		},
	}

	seo.MergeMetaTags(newMetaTags)
	if len(seo.metaTags.Name["description"]) == 0 || seo.metaTags.Name["description"][0] != "New description" {
		t.Errorf("MergeMetaTags() description = %v, want %v", seo.metaTags.Name["description"][0], "New description")
	}

	if len(seo.metaTags.Property["og:description"]) == 0 || seo.metaTags.Property["og:description"][0] != "New OG description" {
		t.Errorf("MergeMetaTags() og:description = %v, want %v", seo.metaTags.Property["og:description"][0], "New OG description")
	}
}

func TestSEO_MergeMetaTags_DescriptionCopy(t *testing.T) {
	// Test that description is copied to og:description if not present
	seo := NewSEO()
	newMetaTags := &MetaTags{
		Name: map[string][]string{
			"description": {"Test description"},
		},
		Property: map[string][]string{},
	}

	seo.MergeMetaTags(newMetaTags)
	if len(seo.metaTags.Name["description"]) == 0 || seo.metaTags.Name["description"][0] != "Test description" {
		t.Errorf("MergeMetaTags() should copy description to og:description, got %v", seo.metaTags.Name["description"][0])
	}

	if len(seo.metaTags.Property["og:description"]) != 1 || seo.metaTags.Property["og:description"][0] != "Test description" {
		t.Errorf("MergeMetaTags() should copy description to og:description, got %v", seo.metaTags.Property["og:description"])
	}
}

func TestSEO_MergeMetaTags_DescriptionOverride(t *testing.T) {
	// Test that description is NOT copied to og:description when present
	seo := NewSEO()
	newMetaTags := &MetaTags{
		Name: map[string][]string{
			"description": {"Test description"},
		},
		Property: map[string][]string{
			"og:description": {"Existing OG description"},
		},
	}

	seo.MergeMetaTags(newMetaTags)
	if len(seo.metaTags.Property["og:description"]) == 0 || seo.metaTags.Property["og:description"][0] != "Existing OG description" {
		t.Errorf("MergeMetaTags() should NOT copy description to og:description when present, got %v", seo.metaTags.Property["og:description"][0])
	}
}

func TestSEO_MergeMetaTags_OriginalNotModified(t *testing.T) {
	seo := NewSEO()
	// Test that original is not modified
	originalMetaTags := &MetaTags{
		Name: map[string][]string{
			"description": {"Original description"},
		},
		Property: map[string][]string{},
	}

	seo.MergeMetaTags(originalMetaTags)
	if len(seo.metaTags.Name["description"]) != 1 || seo.metaTags.Name["description"][0] != "Original description" {
		t.Errorf("MergeMetaTags() should not modify original")
	}
}

func TestSEO_ResetDefaults(t *testing.T) {
	// Test Reset restores defaults
	seo := NewSEO()
	seo.Reset()
	if len(seo.metaTags.Name) != 0 || len(seo.metaTags.Property) != 1 {
		t.Error("Reset() should restore default meta tags")
	}
	if len(seo.metaTags.Property["og:type"]) != 1 || seo.metaTags.Property["og:type"][0] != "website" {
		t.Error("Reset() should restore default og:type")
	}
}

func TestSEO_IntegrationExample(t *testing.T) {
	// Create a complete SEO setup
	siteMetaTags := NewMetaTags("UTF-8")
	siteMetaTags.SetName("description", "A blog about technology")

	site := &Site{
		Title:        "My Blog",
		Separator:    " | ",
		Locale:       "en_US",
		Scheme:       "https",
		Host:         "myblog.com",
		RelativePath: "",
		MetaTags:     siteMetaTags,
	}

	pageMetaTags := NewMetaTags("UTF-8")
	pageMetaTags.SetName("description", "A comprehensive guide to Go testing")

	page := &Page{
		Title:    "Understanding Go Testing",
		Site:     site,
		Pattern:  "/go-testing",
		MetaTags: pageMetaTags,
	}

	published := time.Date(2023, 6, 15, 10, 0, 0, 0, time.UTC)
	modified := time.Date(2023, 6, 16, 20, 0, 0, 0, time.UTC)
	expired := time.Date(2023, 12, 31, 23, 59, 0, 0, time.UTC)

	seo := NewSEO()
	seo.Site(site)
	seo.Page(page)
	seo.SetArticleType()
	seo.AddArticleSection("Programming")
	seo.AddArticleTag("Go")
	seo.SetArticlePublishedTime(published)
	seo.SetArticleModifiedTime(modified)
	seo.SetArticleExpirationTime(expired)
	seo.AddCanonicalLink("https://myblog.com/go-testing")

	expectedTitle := "My Blog | Understanding Go Testing"
	expectedHome := "https://myblog.com/go-testing"
	if seo.Title() != expectedTitle {
		t.Errorf("Integration test title = %v, want %v", seo.Title(), expectedTitle)
	}

	if len(seo.titles) != 2 {
		t.Errorf("Integration test titles = %v, want 2", len(seo.titles))
	}

	if len(seo.Links()) != 1 || seo.Links()[0].Rel != LinkRelCanonical {
		t.Error("Integration test canonical link not found")
	}

	if len(seo.titles) != 2 || seo.titles[0] != "My Blog" {
		t.Errorf("Integration test site name not in titles, got %v, want %v", seo.titles, []string{"My Blog", "Understanding Go Testing"})
	}

	if seo.metaTags.Property["og:site_name"][0] != "My Blog" {
		t.Errorf("Integration test site name = %v, want %v", seo.metaTags.Property["og:site_name"][0], "My Blog")
	}

	if len(seo.metaTags.Property["og:url"]) == 0 || seo.metaTags.Property["og:url"][0] != expectedHome {
		t.Errorf("Integration test og:url = %v, want %v", seo.metaTags.Property["og:url"], []string{expectedHome})
	}

	if len(seo.metaTags.Property["og:type"]) != 1 || seo.metaTags.Property["og:type"][0] != "article" {
		t.Errorf("Integration test type = %v, want %v", seo.metaTags.Property["og:type"][0], "article")
	}

	if len(seo.metaTags.Property["article:section"]) != 1 || seo.metaTags.Property["article:section"][0] != "Programming" {
		t.Errorf("Integration test section = %v, want %v", seo.metaTags.Property["article:section"][0], "Programming")
	}

	if len(seo.metaTags.Property["article:tag"]) != 1 || seo.metaTags.Property["article:tag"][0] != "Go" {
		t.Errorf("Integration test tags length = %v, want 1", len(seo.metaTags.Property["article:tag"]))
	}

	if seo.metaTags.Property["article:published_time"][0] != published.Format(time.RFC3339) {
		t.Errorf("Integration test published time = %v, want %v", seo.metaTags.Property["article:published_time"][0], published.Format(time.RFC3339))
	}

	if seo.metaTags.Property["article:modified_time"][0] != modified.Format(time.RFC3339) {
		t.Errorf("Integration test modified time = %v, want %v", seo.metaTags.Property["article:modified_time"][0], modified.Format(time.RFC3339))
	}

	if seo.metaTags.Property["article:expiration_time"][0] != expired.Format(time.RFC3339) {
		t.Errorf("Integration test expiration time = %v, want %v", seo.metaTags.Property["article:expiration_time"][0], expired.Format(time.RFC3339))
	}
}

func TestSEO_EdgeCases(t *testing.T) {
	t.Run("Empty titles", func(t *testing.T) {
		seo := NewSEO()
		seo.titles = []string{}
		if seo.Title() != "" {
			t.Errorf("Title() with empty slice = %v, want empty", seo.Title())
		}
	})

	t.Run("Empty values", func(t *testing.T) {
		seo := NewSEO()
		seo.AddTitle("")
		seo.metaTags.SetName("", "")
		seo.metaTags.SetProperty("og:title", "")
		seo.SetArticleType()
		seo.AddArticleSection("")
		seo.AddArticleTag("")
		seo.SetArticlePublishedTime(time.Time{})
		seo.SetArticleModifiedTime(time.Time{})
		seo.SetArticleExpirationTime(time.Time{})
		seo.AddCanonicalLink("")
		seo.AddPrevLink("")
		seo.AddNextLink("")
		if len(seo.metaTags.Name) != 0 {
			t.Errorf("Empty values should not be added")
		}
	})

	t.Run("Multiple AddLink calls", func(t *testing.T) {
		seo := NewSEO()

		// Add three links
		seo.AddLink(HeadLink{Rel: "stylesheet", Href: "/style1.css"})
		seo.AddLink(HeadLink{Rel: "icon", Href: "/favicon.ico"})
		seo.AddLink(HeadLink{Rel: "stylesheet", Href: "/style2.css"})

		if len(seo.Links()) != 3 {
			t.Errorf("AddLink() should append to existing links")
		}

		// Replace all links with three new ones
		seo.SetLinks([]HeadLink{{Rel: "stylesheet", Href: "/new.css"}})

		if len(seo.Links()) != 1 {
			t.Errorf("SetLinks() should replace existing links")
		}
	})
}

func TestSEO_ConstantValues(t *testing.T) {
	t.Run("Link rel constants", func(t *testing.T) {
		expectedRels := []string{
			LinkRelAlternate,
			LinkRelAuthor,
			LinkRelCanonical,
			LinkRelLicense,
			LinkRelNext,
			LinkRelPrev,
			LinkRelStylesheet,
			LinkRelIcon,
		}

		for _, rel := range expectedRels {
			if rel == "" {
				t.Errorf("Empty link rel constant")
			}
		}
	})

	t.Run("Referrer policy constants", func(t *testing.T) {
		expectedPolicies := map[string]string{
			"ReferrerPolicyNoReferrer":              ReferrerPolicyNoReferrer,
			"ReferrerPolicyNoReferrerWhenDowngrade": ReferrerPolicyNoReferrerWhenDowngrade,
			"ReferrerPolicyOrigin":                  ReferrerPolicyOrigin,
			"ReferrerPolicyOriginWhenCrossOrigin":   ReferrerPolicyOriginWhenCrossOrigin,
			"ReferrerPolicySameOrigin":              ReferrerPolicySameOrigin,
			"ReferrerPolicyStrictOrigin":            ReferrerPolicyStrictOrigin,
			"ReferrerPolicyUnsafeUrl":               ReferrerPolicyUnsafeUrl,
		}

		for name, policy := range expectedPolicies {
			if policy == "" {
				t.Errorf("Empty referrer policy for %v", name)
			}
		}
	})
}
