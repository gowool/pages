package pages

import (
	"strings"
	"time"
)

const (
	LinkRelAlternate  = "alternate"
	LinkRelAuthor     = "author"
	LinkRelCanonical  = "canonical"
	LinkRelLicense    = "license"
	LinkRelNext       = "next"
	LinkRelPrev       = "prev"
	LinkRelStylesheet = "stylesheet"
	LinkRelIcon       = "icon"
)

const (
	ReferrerPolicyNoReferrer              = "no-referrer"
	ReferrerPolicyNoReferrerWhenDowngrade = "no-referrer-when-downgrade"
	ReferrerPolicyOrigin                  = "origin"
	ReferrerPolicyOriginWhenCrossOrigin   = "origin-when-cross-origin"
	ReferrerPolicySameOrigin              = "same-origin"
	ReferrerPolicyStrictOrigin            = "strict-origin"
	ReferrerPolicyUnsafeUrl               = "unsafe-url"
)

type HeadLink struct {
	// CrossOrigin Specifies how the element handles cross-origin requests
	CrossOrigin string
	// Href Specifies the location of the linked document
	Href string
	// HrefLang Specifies the language of the text in the linked document
	HrefLang string
	// Media Specifies on what device the linked document will be displayed
	Media string
	// Rel REQUIRED Specifies the relationship between the current document and the linked document
	Rel string
	// Sizes Specifies the size of the linked resource. Only for rel="icon"
	Sizes string
	// Title Defines a preferred or an alternate stylesheet
	Title string
	// Type Specifies the media type of the linked document
	Type string
}

type SEO struct {
	headLinks      []HeadLink
	titles         []string
	separator      string
	siteURL        string
	metaTags       *MetaTags
	htmlAttrs      map[string]string
	headAttrs      map[string]string
	bodyAttrs      map[string]string
	langAlternates map[string]string
}

func NewSEO() *SEO {
	var seo SEO
	seo.Reset()
	return &seo
}

func (s *SEO) Reset() {
	s.headLinks = nil
	s.titles = nil
	s.siteURL = ""
	s.separator = " - "
	s.htmlAttrs = map[string]string{
		"dir":    "ltr",
		"lang":   "en",
		"prefix": "og: https://ogp.me/ns#",
	}
	s.metaTags = &MetaTags{
		Property: map[string][]string{
			"og:type": {"website"},
		},
	}
	s.headAttrs = map[string]string{}
	s.bodyAttrs = map[string]string{}
	s.langAlternates = map[string]string{}
}

func (s *SEO) Site(site *Site) {
	if site == nil {
		return
	}

	s.siteURL = site.URL()

	if site.Title != "" {
		s.SetTitle(site.Title)
		s.metaTags.SetProperty("og:site_name", site.Title)
	}

	if site.Separator != "" {
		s.SetSeparator(site.Separator)
	}

	if site.Locale != "" {
		locale := strings.ReplaceAll(site.Locale, "_", "-")
		s.SetHTMLAttribute("lang", locale)
		s.metaTags.SetProperty("og:locale", locale)
	}

	s.SetOgURL(site.Home())

	s.MergeMetaTags(site.MetaTags)
}

func (s *SEO) Page(page *Page, args ...any) {
	if page == nil {
		return
	}

	if page.Title != "" {
		s.AddTitle(page.Title)
	}

	s.SetOgURL(page.AbsURL(args...))

	s.MergeMetaTags(page.MetaTags)
}

func (s *SEO) AddHTMLPrefixAttribute(prefix string) {
	if oldPrefix, ok := s.htmlAttrs["prefix"]; ok && oldPrefix != "" {
		prefix = oldPrefix + " " + prefix
	}
	s.SetHTMLAttribute("prefix", prefix)
}

func (s *SEO) SetWebsiteType() {
	s.SetOGType("website")
}

func (s *SEO) SetArticleType() {
	s.SetOGType("article")
}

func (s *SEO) SetOGType(t string) {
	s.metaTags.SetProperty("og:type", t)
}

func (s *SEO) AddArticleSection(section string) {
	if section != "" {
		s.metaTags.AppendProperty("article:section", section)
	}
}

func (s *SEO) AddArticleTag(tag string) {
	if tag != "" {
		s.metaTags.AppendProperty("article:tag", tag)
	}
}

func (s *SEO) SetArticlePublishedTime(published time.Time) {
	if !published.IsZero() {
		s.metaTags.SetProperty("article:published_time", published.Format(time.RFC3339))
	}
}

func (s *SEO) SetArticleExpirationTime(expired time.Time) {
	if !expired.IsZero() {
		s.metaTags.SetProperty("article:expiration_time", expired.Format(time.RFC3339))
	}
}

func (s *SEO) SetArticleModifiedTime(updated time.Time) {
	if !updated.IsZero() {
		s.metaTags.SetProperty("article:modified_time", updated.Format(time.RFC3339))
	}
}

func (s *SEO) Title() string {
	return strings.Join(s.titles, s.separator)
}

func (s *SEO) ReverseTitle() string {
	titles := make([]string, 0, len(s.titles))
	for i := len(s.titles) - 1; i >= 0; i-- {
		titles = append(titles, s.titles[i])
	}
	return strings.Join(titles, s.separator)
}

func (s *SEO) FirstTitle() string {
	if len(s.titles) == 0 {
		return ""
	}
	return s.titles[0]
}

func (s *SEO) SetTitle(title string) {
	s.titles = []string{title}
}

func (s *SEO) AddTitle(title string) {
	s.titles = append(s.titles, title)
}

func (s *SEO) Separator() string {
	return s.separator
}

func (s *SEO) SetSeparator(separator string) {
	s.separator = separator
}

func (s *SEO) MetaTags() *MetaTags {
	return s.metaTags
}

func (s *SEO) MergeMetaTags(metaTags *MetaTags) {
	s.metaTags.Set(metaTags)

	if content, ok := metaTags.Property["description"]; ok {
		if _, ok := s.metaTags.Property["og:description"]; !ok {
			s.metaTags.SetProperty("og:description", content...)
		}
	}
}

func (s *SEO) HTMLAttributes() map[string]string {
	return s.htmlAttrs
}

func (s *SEO) SetHTMLAttributes(attrs map[string]string) {
	s.htmlAttrs = attrs
}

func (s *SEO) SetHTMLAttribute(name, content string) {
	s.htmlAttrs[name] = content
}

func (s *SEO) RemoveHTMLAttribute(name string) {
	delete(s.htmlAttrs, name)
}

func (s *SEO) HasHTMLAttribute(name string) bool {
	_, ok := s.htmlAttrs[name]
	return ok
}

func (s *SEO) HeadAttributes() map[string]string {
	return s.headAttrs
}

func (s *SEO) SetHeadAttributes(attrs map[string]string) {
	s.headAttrs = attrs
}

func (s *SEO) SetHeadAttribute(name, content string) {
	s.headAttrs[name] = content
}

func (s *SEO) RemoveHeadAttribute(name string) {
	delete(s.headAttrs, name)
}

func (s *SEO) HasHeadAttribute(name string) bool {
	_, ok := s.headAttrs[name]
	return ok
}

func (s *SEO) BodyAttributes() map[string]string {
	return s.bodyAttrs
}

func (s *SEO) SetBodyAttributes(attrs map[string]string) {
	s.bodyAttrs = attrs
}

func (s *SEO) SetBodyAttribute(name, content string) {
	s.bodyAttrs[name] = content
}

func (s *SEO) RemoveBodyAttribute(name string) {
	delete(s.bodyAttrs, name)
}

func (s *SEO) HasBodyAttribute(name string) bool {
	_, ok := s.bodyAttrs[name]
	return ok
}

func (s *SEO) LangAlternates() map[string]string {
	return s.langAlternates
}

func (s *SEO) SetLangAlternates(langAlternates map[string]string) {
	s.langAlternates = langAlternates
}

func (s *SEO) AddLangAlternate(href, hreflang string) {
	s.langAlternates[href] = hreflang
}

func (s *SEO) RemoveLangAlternate(href string) {
	delete(s.langAlternates, href)
}

func (s *SEO) HasLangAlternate(href string) bool {
	_, ok := s.langAlternates[href]
	return ok
}

func (s *SEO) Links() []HeadLink {
	return s.headLinks
}

func (s *SEO) SetLinks(links []HeadLink) {
	s.headLinks = links
}

func (s *SEO) AddLink(link ...HeadLink) {
	s.headLinks = append(s.headLinks, link...)
}

func (s *SEO) AddCanonicalLink(href string) {
	s.AddLink(HeadLink{Rel: LinkRelCanonical, Href: href})
}

func (s *SEO) AddPrevLink(href string) {
	s.AddLink(HeadLink{Rel: LinkRelPrev, Href: href})
}

func (s *SEO) AddNextLink(href string) {
	s.AddLink(HeadLink{Rel: LinkRelNext, Href: href})
}

func (s *SEO) SetOgURL(url string) {
	if url == "" {
		return
	}

	if s.siteURL != "" && !strings.HasPrefix(url, "http") {
		url = s.siteURL + url
	}
	s.metaTags.SetProperty("og:url", url)
}
