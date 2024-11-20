package seo

import (
	"strings"
	"time"

	"github.com/gowool/pages/model"
)

type SEO interface {
	Site(site *model.Site) SEO
	Page(page *model.Page) SEO
	Title() string
	ReverseTitle() string
	FirstTitle() string
	SetTitle(title string) SEO
	AddTitle(title string) SEO
	Separator() string
	SetSeparator(separator string) SEO
	Metas() map[string]map[string]string
	SetMetas(metas map[string]map[string]string) SEO
	AddMeta(typ, name, content string) SEO
	RemoveMeta(typ, name string) SEO
	HasMeta(typ, name string) bool
	HTMLAttributes() map[string]string
	SetHTMLAttributes(attrs map[string]string) SEO
	AddHTMLAttribute(name, content string) SEO
	RemoveHTMLAttribute(name string) SEO
	HasHTMLAttribute(name string) bool
	HeadAttributes() map[string]string
	SetHeadAttributes(attrs map[string]string) SEO
	AddHeadAttribute(name, content string) SEO
	RemoveHeadAttribute(name string) SEO
	HasHeadAttribute(name string) bool
	BodyAttributes() map[string]string
	SetBodyAttributes(attrs map[string]string) SEO
	AddBodyAttribute(name, content string) SEO
	RemoveBodyAttribute(name string) SEO
	HasBodyAttribute(name string) bool
	LinkCanonical() string
	SetLinkCanonical(link string) SEO
	RemoveLinkCanonical() SEO
	LangAlternates() map[string]string
	SetLangAlternates(langAlternates map[string]string) SEO
	AddLangAlternate(href, hreflang string) SEO
	RemoveLangAlternate(href string) SEO
	HasLangAlternate(href string) bool
	OEmbedLinks() map[string]string
	AddOEmbedLink(title, link string) SEO
}

type pageSEO struct {
	titles         []string
	separator      string
	linkCanonical  string
	metas          map[string]map[string]string
	htmlAttrs      map[string]string
	headAttrs      map[string]string
	bodyAttrs      map[string]string
	langAlternates map[string]string
	oembedLinks    map[string]string
}

func NewSEO() SEO {
	return &pageSEO{
		separator:      " - ",
		metas:          map[string]map[string]string{},
		htmlAttrs:      map[string]string{"dir": "ltr", "lang": "en"},
		headAttrs:      map[string]string{},
		bodyAttrs:      map[string]string{},
		langAlternates: map[string]string{},
		oembedLinks:    map[string]string{},
	}
}

func (s *pageSEO) Site(site *model.Site) SEO {
	if site == nil {
		return s
	}

	s.AddHTMLAttribute("prefix", "og: https://ogp.me/ns#")

	if site.Title != "" {
		s.SetTitle(site.Title)
		s.AddMeta(model.MetaProperty.String(), "of:site_name", site.Title)
	}

	if site.Separator != "" {
		s.SetSeparator(site.Separator)
	}

	if site.Locale != "" {
		locale := strings.ReplaceAll(site.Locale, "_", "-")
		s.AddHTMLAttribute("lang", locale)
		s.AddMeta(model.MetaProperty.String(), "og:locale", locale)
	}

	s.AddMeta(model.MetaProperty.String(), "og:url", site.URL())
	s.AddMeta(model.MetaProperty.String(), "og:type", "website")

	return s.setMetas(site.Metas)
}

func (s *pageSEO) Page(page *model.Page) SEO {
	if page == nil {
		return s
	}

	if page.Title != "" {
		s.AddTitle(page.Title)
	}

	if !page.IsInternal() {
		s.AddMeta(model.MetaProperty.String(), "og:type", "article")

		if page.Published != nil && !page.Published.IsZero() {
			s.AddMeta(model.MetaProperty.String(), "article:published_time", page.Published.Format(time.RFC3339))
		}
		if !page.Updated.IsZero() {
			s.AddMeta(model.MetaProperty.String(), "article:modified_time", page.Updated.Format(time.RFC3339))
		}
		if page.Expired != nil && !page.Expired.IsZero() {
			s.AddMeta(model.MetaProperty.String(), "article:expiration_time", page.Expired.Format(time.RFC3339))
		}
	}

	return s.setMetas(page.Metas)
}

func (s *pageSEO) setMetas(metas []model.Meta) SEO {
	var (
		ogDescAdded bool
		desc        string
	)

	for _, meta := range metas {
		if meta.Content == "" && (meta.Key == "keywords" || meta.Key == "description") {
			continue
		}

		if meta.Key == "description" {
			desc = meta.Content
		}

		if meta.Key == "og:description" {
			ogDescAdded = true
		}

		s.AddMeta(meta.Type.String(), meta.Key, meta.Content)
	}

	if !ogDescAdded && desc != "" {
		s.AddMeta(model.MetaProperty.String(), "og:description", desc)
	}
	return s
}

func (s *pageSEO) Title() string {
	return strings.Join(s.titles, s.separator)
}

func (s *pageSEO) ReverseTitle() string {
	titles := make([]string, 0, len(s.titles))
	for i := len(s.titles) - 1; i >= 0; i-- {
		titles = append(titles, s.titles[i])
	}
	return strings.Join(titles, s.separator)
}

func (s *pageSEO) FirstTitle() string {
	if len(s.titles) == 0 {
		return ""
	}
	return s.titles[0]
}

func (s *pageSEO) SetTitle(title string) SEO {
	s.titles = []string{title}
	return s
}

func (s *pageSEO) AddTitle(title string) SEO {
	s.titles = append(s.titles, title)
	return s
}

func (s *pageSEO) Separator() string {
	return s.separator
}

func (s *pageSEO) SetSeparator(separator string) SEO {
	s.separator = separator
	return s
}

func (s *pageSEO) Metas() map[string]map[string]string {
	return s.metas
}

func (s *pageSEO) SetMetas(metas map[string]map[string]string) SEO {
	s.metas = metas
	return s
}

func (s *pageSEO) AddMeta(typ, name, content string) SEO {
	if _, ok := s.metas[typ]; !ok {
		s.metas[typ] = map[string]string{}
	}

	s.metas[typ][name] = content
	return s
}

func (s *pageSEO) RemoveMeta(typ, name string) SEO {
	delete(s.metas[typ], name)
	return s
}

func (s *pageSEO) HasMeta(typ, name string) bool {
	if _, ok := s.metas[typ]; !ok {
		return false
	}

	_, ok := s.metas[typ][name]
	return ok
}

func (s *pageSEO) HTMLAttributes() map[string]string {
	return s.htmlAttrs
}

func (s *pageSEO) SetHTMLAttributes(attrs map[string]string) SEO {
	s.htmlAttrs = attrs
	return s
}

func (s *pageSEO) AddHTMLAttribute(name, content string) SEO {
	s.htmlAttrs[name] = content
	return s
}

func (s *pageSEO) RemoveHTMLAttribute(name string) SEO {
	delete(s.htmlAttrs, name)
	return s
}

func (s *pageSEO) HasHTMLAttribute(name string) bool {
	_, ok := s.htmlAttrs[name]
	return ok
}

func (s *pageSEO) HeadAttributes() map[string]string {
	return s.headAttrs
}

func (s *pageSEO) SetHeadAttributes(attrs map[string]string) SEO {
	s.headAttrs = attrs
	return s
}

func (s *pageSEO) AddHeadAttribute(name, content string) SEO {
	s.headAttrs[name] = content
	return s
}

func (s *pageSEO) RemoveHeadAttribute(name string) SEO {
	delete(s.headAttrs, name)
	return s
}

func (s *pageSEO) HasHeadAttribute(name string) bool {
	_, ok := s.headAttrs[name]
	return ok
}

func (s *pageSEO) BodyAttributes() map[string]string {
	return s.bodyAttrs
}

func (s *pageSEO) SetBodyAttributes(attrs map[string]string) SEO {
	s.bodyAttrs = attrs
	return s
}

func (s *pageSEO) AddBodyAttribute(name, content string) SEO {
	s.bodyAttrs[name] = content
	return s
}

func (s *pageSEO) RemoveBodyAttribute(name string) SEO {
	delete(s.bodyAttrs, name)
	return s
}

func (s *pageSEO) HasBodyAttribute(name string) bool {
	_, ok := s.bodyAttrs[name]
	return ok
}

func (s *pageSEO) LinkCanonical() string {
	return s.linkCanonical
}

func (s *pageSEO) SetLinkCanonical(link string) SEO {
	s.linkCanonical = link
	return s
}

func (s *pageSEO) RemoveLinkCanonical() SEO {
	s.linkCanonical = ""
	return s
}

func (s *pageSEO) LangAlternates() map[string]string {
	return s.langAlternates
}

func (s *pageSEO) SetLangAlternates(langAlternates map[string]string) SEO {
	s.langAlternates = langAlternates
	return s
}

func (s *pageSEO) AddLangAlternate(href, hreflang string) SEO {
	s.langAlternates[href] = hreflang
	return s
}

func (s *pageSEO) RemoveLangAlternate(href string) SEO {
	delete(s.langAlternates, href)
	return s
}

func (s *pageSEO) HasLangAlternate(href string) bool {
	_, ok := s.langAlternates[href]
	return ok
}

func (s *pageSEO) OEmbedLinks() map[string]string {
	return s.oembedLinks
}

func (s *pageSEO) AddOEmbedLink(title, link string) SEO {
	s.oembedLinks[title] = link
	return s
}
