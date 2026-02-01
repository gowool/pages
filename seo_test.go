package pages

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSEO(t *testing.T) {
	seo := NewSEO()
	require.NotNil(t, seo, "NewSEO() should not return nil")

	// Check default separator
	assert.Equal(t, " - ", seo.separator, "Separator should be default")

	// Check default HTML attributes
	assert.NotNil(t, seo.htmlAttrs, "htmlAttrs should be initialized")

	// Check default meta tags
	assert.NotNil(t, seo.metaTags, "MetaTags should be initialized")

	// Check other initializations
	assert.Nil(t, seo.titles, "titles should be nil initially")
	assert.Empty(t, seo.siteURL, "siteURL should be empty initially")
	assert.Equal(t, " - ", seo.separator, "separator should be default")
	assert.NotNil(t, seo.htmlAttrs, "headAttrs should be initialized")
	assert.NotNil(t, seo.bodyAttrs, "bodyAttrs should be initialized")
	assert.NotNil(t, seo.langAlternates, "langAlternates should be initialized")
	assert.Empty(t, seo.headLinks, "headLinks should be empty initially")
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
	assert.Empty(t, seo.titles, "titles should be empty after reset")
	assert.Empty(t, seo.siteURL, "siteURL should be empty after reset")
	assert.Equal(t, " - ", seo.separator, "separator should be reset to default")
	assert.NotEqual(t, "value", seo.htmlAttrs["test"], "HTML attributes should be reset to defaults")
	assert.NotEqual(t, "value", seo.headAttrs["test"], "Head attributes should be reset to defaults")
	assert.NotEqual(t, "value", seo.bodyAttrs["test"], "Body attributes should be reset to defaults")
	assert.NotEqual(t, "value", seo.langAlternates["test"], "Lang alternates should be reset to defaults")
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

			if tt.want != "" {
				assert.Equal(t, []string{tt.want}, seo.titles, "Site() titles should match expected")
			}

			if tt.site.Separator != "" {
				assert.Equal(t, tt.site.Separator, seo.separator, "Site() separator should match site separator")
			}

			if tt.site.Locale != "" {
				assert.Equal(t, "en-US", seo.htmlAttrs["lang"], "Site() lang should be en-US for US locale")
			}

			expectedHome := tt.site.Origin()
			assert.Equal(t, []string{expectedHome}, seo.metaTags.Property["og:url"], "Site() og:url should match site origin")

			if tt.site.MetaTags != nil {
				assert.Equal(t, tt.site.MetaTags.Name["description"], seo.metaTags.Name["description"], "Site() description should match site meta description")
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
	assert.Equal(t, expectedPrefix, seo.htmlAttrs["prefix"], "AddHTMLPrefixAttribute() should append to existing prefix")

	// Test adding prefix to empty prefix
	seo.Reset()
	seo.AddHTMLPrefixAttribute("test: http://example.com")

	assert.Equal(t, "og: https://ogp.me/ns# test: http://example.com", seo.htmlAttrs["prefix"], "AddHTMLPrefixAttribute() should work with empty prefix")

	// Test adding prefix that starts with http
	seo.Reset()
	seo.AddHTMLPrefixAttribute("http://example.com")

	assert.Equal(t, "og: https://ogp.me/ns# http://example.com", seo.htmlAttrs["prefix"], "AddHTMLPrefixAttribute() should work with http prefix")
}

func TestSEO_AttributeMethods(t *testing.T) {
	seo := NewSEO()

	// Test HTML attributes
	htmlAttrs := seo.HTMLAttributes()
	assert.NotNil(t, htmlAttrs, "HTMLAttributes() should not return nil")

	// Test setting attributes
	seo.SetHTMLAttributes(map[string]string{
		"test": "value",
	})
	assert.Equal(t, "value", seo.htmlAttrs["test"], "SetHTMLAttributes() should set attributes")

	// Test setting single attribute
	seo.SetHTMLAttribute("new", "attr")
	assert.Equal(t, "attr", seo.htmlAttrs["new"], "SetHTMLAttribute() should set single attribute")

	// Test removing attribute
	seo.RemoveHTMLAttribute("test")
	assert.False(t, seo.HasHTMLAttribute("test"), "RemoveHTMLAttribute() should remove attribute")

	// Test has attribute
	assert.True(t, seo.HasHTMLAttribute("new"), "HasHTMLAttribute() should return true for existing attribute")
	assert.False(t, seo.HasHTMLAttribute("nonexistent"), "HasHTMLAttribute() should return false for non-existent attribute")

	// Test head attributes
	headAttrs := seo.HeadAttributes()
	assert.NotNil(t, headAttrs, "HeadAttributes() should not return nil")

	seo.SetHeadAttribute("head", "attr")
	assert.Equal(t, "attr", seo.headAttrs["head"], "SetHeadAttribute() should set head attribute")

	seo.RemoveHeadAttribute("head")
	assert.False(t, seo.HasHeadAttribute("head"), "RemoveHeadAttribute() should remove head attribute")

	// Test body attributes
	bodyAttrs := seo.BodyAttributes()
	assert.NotNil(t, bodyAttrs, "BodyAttributes() should not return nil")

	seo.SetBodyAttribute("body", "attr")
	assert.Equal(t, "attr", seo.bodyAttrs["body"], "SetBodyAttribute() should set body attribute")

	seo.RemoveBodyAttribute("body")
	assert.False(t, seo.HasBodyAttribute("body"), "RemoveBodyAttribute() should remove body attribute")

	// Test has attributes (repeated check)
	assert.True(t, seo.HasHTMLAttribute("new"), "HasHTMLAttribute() should return true for existing attribute")
	assert.False(t, seo.HasHTMLAttribute("nonexistent"), "HasHTMLAttribute() should return false for non-existent attribute")
}

func TestSEO_Links(t *testing.T) {
	seo := NewSEO()

	// Test Links returns nil initially
	links := seo.Links()
	assert.Empty(t, links, "Links() should return nil initially")

	// Test setting links
	seo.SetLinks([]HeadLink{{Rel: "stylesheet", Href: "/style1.css"}})
	assert.Len(t, seo.headLinks, 1, "SetLinks() should set one link")

	// Test replacing links with a different one
	seo.SetLinks([]HeadLink{{Rel: "stylesheet", Href: "/style2.css"}})
	assert.Len(t, seo.headLinks, 1, "SetLinks() should replace with one link")

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

	assert.True(t, foundCanonical, "AddCanonicalLink() should add canonical link")
	assert.True(t, foundPrev, "AddPrevLink() should add prev link")
	assert.True(t, foundNext, "AddNextLink() should add next link")

	// Test that AddLink appends to existing links
	originalCount := len(seo.headLinks)
	seo.AddLink(HeadLink{Rel: "test", Href: "/test"})
	assert.Len(t, seo.headLinks, originalCount+1, "AddLink() should append to existing links")

	// Test replacing links
	seo.SetLinks([]HeadLink{{Rel: "stylesheet", Href: "/new.css"}})
	assert.Len(t, seo.headLinks, 1, "SetLinks() should replace existing links with one link")
	assert.Equal(t, "stylesheet", seo.headLinks[0].Rel, "SetLinks() should set correct rel attribute")
	assert.Equal(t, "/new.css", seo.headLinks[0].Href, "SetLinks() should set correct href attribute")
}

func TestSEO_RemoveLinks(t *testing.T) {
	seo := NewSEO()

	// Test replacing links with single stylesheet
	seo.SetLinks([]HeadLink{{Rel: "stylesheet", Href: "/style2.css"}})
	assert.Len(t, seo.headLinks, 1, "SetLinks() should set single link")

	if len(seo.headLinks) == 1 {
		assert.Equal(t, "stylesheet", seo.headLinks[0].Rel, "SetLinks() should set correct rel attribute")
		assert.Equal(t, "/style2.css", seo.headLinks[0].Href, "SetLinks() should set correct href attribute")
	}
}

func TestSEO_LangAlternates(t *testing.T) {
	seo := NewSEO()

	// Test lang alternates returns empty map initially
	langAlts := seo.LangAlternates()
	assert.Empty(t, langAlts, "LangAlternates() should return empty map initially")

	// Test setting lang alternates
	newLangAlts := map[string]string{
		"https://example.com/fr": "fr",
		"https://example.com/es": "es",
	}
	seo.SetLangAlternates(newLangAlts)

	// Test values are properly set by getting fresh reference
	currentLangAlts := seo.LangAlternates()
	assert.Equal(t, newLangAlts, currentLangAlts, "SetLangAlternates() should set lang alternates correctly")

	// Test adding lang alternate
	seo.AddLangAlternate("https://example.com/de", "de")

	// Test has lang alternate
	assert.True(t, seo.HasLangAlternate("https://example.com/fr"), "AddLangAlternate() should add lang alternate")

	// Test getting lang alternates
	retrievedLangAlts := seo.LangAlternates()
	assert.Len(t, retrievedLangAlts, 3, "LangAlternates() should contain 3 entries")

	// Test contains lang alternate
	assert.True(t, seo.HasLangAlternate("https://example.com/de"), "LangAlternates should contain de alternate")

	// Test removing lang alternate
	seo.RemoveLangAlternate("https://example.com/fr")
	assert.False(t, seo.HasLangAlternate("https://example.com/fr"), "RemoveLangAlternate() should remove lang alternate")

	// Test has non-existent lang alternate
	assert.False(t, seo.HasLangAlternate("https://example.com/it"), "HasLangAlternate() should return false for non-existent alternate")
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
	assert.Equal(t, []string{"New description"}, seo.metaTags.Name["description"], "MergeMetaTags() should set description")
	assert.Equal(t, []string{"New OG description"}, seo.metaTags.Property["og:description"], "MergeMetaTags() should set og:description")
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
	assert.Equal(t, []string{"Test description"}, seo.metaTags.Name["description"], "MergeMetaTags() should set description")
	assert.Equal(t, []string{"Test description"}, seo.metaTags.Property["og:description"], "MergeMetaTags() should copy description to og:description")
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
	assert.Equal(t, []string{"Existing OG description"}, seo.metaTags.Property["og:description"], "MergeMetaTags() should NOT copy description to og:description when present")
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
	assert.Equal(t, []string{"Original description"}, seo.metaTags.Name["description"], "MergeMetaTags() should not modify original")
}

func TestSEO_ResetDefaults(t *testing.T) {
	// Test Reset restores defaults
	seo := NewSEO()
	seo.SetOGType("website")
	assert.Len(t, seo.metaTags.Property, 1, "Reset() should restore default meta tags properties")
	assert.Equal(t, []string{"website"}, seo.metaTags.Property["og:type"], "Reset() should restore default og:type")
	seo.Reset()
	assert.Empty(t, seo.metaTags.Name, "Reset() should restore default meta tags names")
	assert.Len(t, seo.metaTags.Property, 0, "Reset() should restore default meta tags properties")
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
	assert.Equal(t, expectedTitle, seo.Title(), "Integration test title should match expected")
	assert.Len(t, seo.titles, 2, "Integration test should have 2 titles")

	links := seo.Links()
	assert.Len(t, links, 1, "Integration test should have 1 link")
	assert.Equal(t, LinkRelCanonical, links[0].Rel, "Integration test canonical link not found")

	assert.Equal(t, []string{"My Blog", "Understanding Go Testing"}, seo.titles, "Integration test site name should be in titles")
	assert.Equal(t, []string{"My Blog"}, seo.metaTags.Property["og:site_name"], "Integration test site name should match")
	assert.Equal(t, []string{expectedHome}, seo.metaTags.Property["og:url"], "Integration test og:url should match")
	assert.Equal(t, []string{"article"}, seo.metaTags.Property["og:type"], "Integration test type should be article")
	assert.Equal(t, []string{"Programming"}, seo.metaTags.Property["article:section"], "Integration test section should match")
	assert.Equal(t, []string{"Go"}, seo.metaTags.Property["article:tag"], "Integration test tags should match")
	assert.Equal(t, []string{published.Format(time.RFC3339)}, seo.metaTags.Property["article:published_time"], "Integration test published time should match")
	assert.Equal(t, []string{modified.Format(time.RFC3339)}, seo.metaTags.Property["article:modified_time"], "Integration test modified time should match")
	assert.Equal(t, []string{expired.Format(time.RFC3339)}, seo.metaTags.Property["article:expiration_time"], "Integration test expiration time should match")
}

func TestSEO_EdgeCases(t *testing.T) {
	t.Run("Empty titles", func(t *testing.T) {
		seo := NewSEO()
		seo.titles = []string{}
		assert.Empty(t, seo.Title(), "Title() with empty slice should return empty")
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
		assert.Empty(t, seo.metaTags.Name, "Empty values should not be added")
	})

	t.Run("Multiple AddLink calls", func(t *testing.T) {
		seo := NewSEO()

		// Add three links
		seo.AddLink(HeadLink{Rel: "stylesheet", Href: "/style1.css"})
		seo.AddLink(HeadLink{Rel: "icon", Href: "/favicon.ico"})
		seo.AddLink(HeadLink{Rel: "stylesheet", Href: "/style2.css"})

		assert.Len(t, seo.Links(), 3, "AddLink() should append to existing links")

		// Replace all links with three new ones
		seo.SetLinks([]HeadLink{{Rel: "stylesheet", Href: "/new.css"}})

		assert.Len(t, seo.Links(), 1, "SetLinks() should replace existing links")
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
			assert.NotEmpty(t, rel, "Link rel constant should not be empty")
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
			assert.NotEmpty(t, policy, "Referrer policy for %s should not be empty", name)
		}
	})
}
