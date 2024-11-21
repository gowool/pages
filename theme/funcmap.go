package theme

import (
	"context"
	"fmt"
	"html"
	"html/template"
	"maps"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/gowool/cr"
	"github.com/gowool/theme"
	"github.com/spf13/cast"

	"github.com/gowool/pages"
	"github.com/gowool/pages/model"
	"github.com/gowool/pages/repository"
	"github.com/gowool/pages/seo"
)

var (
	re      = regexp.MustCompile(`<(.|\n)*?>`)
	escaper = strings.NewReplacer(`"`, "&quot;")
)

type FuncMap struct {
	pageRepo    repository.Page
	menuService pages.Menu
	matcher     pages.Matcher
}

func NewFuncMap(pageRepo repository.Page, menu pages.Menu, matcher pages.Matcher) *FuncMap {
	return &FuncMap{
		pageRepo:    pageRepo,
		menuService: menu,
		matcher:     matcher,
	}
}

func (fm *FuncMap) FuncMap(t theme.Theme) template.FuncMap {
	return template.FuncMap{
		"menu":              fm.menu(t),
		"node_is_current":   fm.matcher.IsCurrent,
		"node_is_ancestor":  fm.matcher.IsAncestor,
		"strip_tags":        stripTags,
		"escape_double_q":   escapeDoubleQuotes,
		"reverse_title_tag": reverseTitleTag,
		"title_tag":         titleTag,
		"meta_tags":         metaTags,
		"html_attrs":        htmlAttrs,
		"head_attrs":        headAttrs,
		"body_attrs":        bodyAttrs,
		"link_canonical":    linkCanonical,
		"lang_alternates":   langAlternates,
		"oembed_links":      oEmbedLinks,
		"page_url":          fm.pageURL,
		"page_by_id":        fm.findPage,
		"page_children":     fm.pageChildren,
		"pages_by_criteria": fm.pagesByCriteria,
		"js": func(str string) template.JS {
			return template.JS(str)
		},
		"css": func(str string) template.CSS {
			return template.CSS(str)
		},
	}
}

func (fm *FuncMap) menu(t theme.Theme) func(context.Context, string, string, map[string]any) template.HTML {
	return func(ctx context.Context, handle, templateName string, data map[string]any) template.HTML {
		m, err := fm.menuService.Get(ctx, handle)
		if err != nil {
			return ""
		}

		data = maps.Clone(data)
		data["node"] = m.Node

		str, err := t.HTML(ctx, templateName, data)
		if err != nil {
			return ""
		}
		return template.HTML(str)
	}
}

func (fm *FuncMap) pageURL(ctx context.Context, name any, args ...any) string {
	switch name := name.(type) {
	case string:
		if strings.HasPrefix(name, model.PageAliasPrefix) {
			return fm.pageURLByAlias(ctx, name, args...)
		}
		if name == model.PageCMS {
			return fm.pageURLByPath(ctx, args...)
		}
	case model.Page:
		return fm.pageURLByPage(ctx, name, args...)
	case *model.Page:
		return fm.pageURLByPage(ctx, *name, args...)
	}
	return ""
}

func (fm *FuncMap) pageURLByAlias(ctx context.Context, alias string, args ...any) string {
	site := pages.CtxSite(ctx)
	if site == nil {
		return ""
	}
	page, err := fm.pageRepo.FindByAlias(ctx, site.ID, alias, time.Now())
	if err != nil {
		return ""
	}
	page.Site = site
	return fm.pageURLByPage(ctx, page, args...)
}

func (fm *FuncMap) pageURLByPath(ctx context.Context, args ...any) string {
	site := pages.CtxSite(ctx)
	if site == nil {
		return ""
	}

	q := make(url.Values)
	for i := 0; i < len(args); i += 2 {
		q.Add(fmt.Sprintf("%v", args[i]), fmt.Sprintf("%v", args[i+1]))
	}
	path := q.Get("path")
	q.Del("path")

	var link strings.Builder
	link.WriteString(strings.TrimSuffix(site.URL(), "/"))
	link.WriteRune('/')
	if path != "" && path != "/" {
		link.WriteString(strings.TrimPrefix(path, "/"))
	}

	if len(q) > 0 {
		link.WriteRune('?')
		link.WriteString(q.Encode())
	}
	return link.String()
}

func (fm *FuncMap) pageURLByPage(ctx context.Context, page model.Page, args ...any) string {
	if page.Site == nil {
		site := pages.CtxSite(ctx)
		if site == nil {
			return ""
		}
		page.Site = site
		page.SiteID = site.ID
	}

	path := page.URL

	if page.IsHybrid() {
		path = page.Pattern
		var rest []any
		for i := 0; i < len(args); i += 2 {
			if key := fmt.Sprintf("%v", args[i]); len(key) > 2 && key[0] == '{' && key[len(key)-1] == '}' {
				path = strings.ReplaceAll(path, key, fmt.Sprintf("%v", args[i+1]))
				continue
			}
			rest = append(rest, args[i], args[i+1])
		}
		args = rest
	}

	return fm.pageURLByPath(pages.WithSite(ctx, page.Site), append([]any{"path", path}, args...)...)
}

func (fm *FuncMap) findPage(ctx context.Context, id int64) model.Page {
	page, _ := fm.pageRepo.FindByID(ctx, id)
	return page
}

func (fm *FuncMap) pageChildren(ctx context.Context, parentID int64, now time.Time) []model.Page {
	data, _ := fm.pageRepo.FindByParentID(ctx, parentID, now)
	return data
}

func (fm *FuncMap) pagesByCriteria(ctx context.Context, criteria *cr.Criteria) map[string]any {
	data, total, _ := fm.pageRepo.FindAndCount(ctx, criteria)
	return map[string]any{"pages": data, "total": total}
}

func titleTag(seo seo.SEO, args ...string) template.HTML {
	return template.HTML(
		fmt.Sprintf(
			"<title>%s</title>",
			stripTags(strings.Join(append([]string{seo.Title()}, args...), seo.Separator())),
		),
	)
}

func reverseTitleTag(seo seo.SEO, args ...string) template.HTML {
	data := make([]string, 0, len(args)+1)
	for i := len(args) - 1; i >= 0; i-- {
		data = append(data, args[i])
	}
	data = append(data, seo.ReverseTitle())
	return template.HTML(
		fmt.Sprintf(
			"<title>%s</title>",
			stripTags(strings.Join(data, seo.Separator())),
		),
	)
}

func metaTags(seo seo.SEO) template.HTML {
	var b strings.Builder
	for typ, metas := range seo.Metas() {
		for name, content := range metas {
			b.WriteString("<meta ")
			b.WriteString(typ)
			b.WriteString(`="`)
			b.WriteString(normalize(name))
			if content != "" {
				b.WriteString(`" content="`)
				b.WriteString(normalize(content))
			}
			b.WriteString("\" />\n")
		}
	}
	return template.HTML(b.String())
}

func htmlAttrs(seo seo.SEO, rest ...any) template.HTMLAttr {
	return attrs(seo.HTMLAttributes(), rest...)
}

func headAttrs(seo seo.SEO, rest ...any) template.HTMLAttr {
	return attrs(seo.HeadAttributes(), rest...)
}

func bodyAttrs(seo seo.SEO, rest ...any) template.HTMLAttr {
	return attrs(seo.BodyAttributes(), rest...)
}

func attrs(attrs map[string]string, rest ...any) template.HTMLAttr {
	var b strings.Builder
	for name, value := range attrs {
		b.WriteString(name)
		b.WriteString(`="`)
		b.WriteString(html.EscapeString(value))
		b.WriteString(`" `)
	}
	for i := 1; i < len(rest); i += 2 {
		b.WriteString(cast.ToString(rest[i-1]))
		b.WriteString(`="`)
		b.WriteString(html.EscapeString(cast.ToString(rest[i])))
		b.WriteString(`" `)
	}
	return template.HTMLAttr(strings.TrimSuffix(b.String(), " "))
}

func linkCanonical(seo seo.SEO) template.HTML {
	if seo.LinkCanonical() != "" {
		return template.HTML(fmt.Sprintf(`<link rel="canonical" href="%s" />`, html.EscapeString(seo.LinkCanonical())))
	}
	return ""
}

func langAlternates(seo seo.SEO) template.HTML {
	var b strings.Builder
	for href, hreflang := range seo.LangAlternates() {
		b.WriteString(`<link rel="alternate" href="`)
		b.WriteString(html.EscapeString(href))
		b.WriteString(`" hreflang="`)
		b.WriteString(html.EscapeString(hreflang))
		b.WriteString("\" />\n")
	}
	return template.HTML(b.String())
}

func oEmbedLinks(seo seo.SEO) template.HTML {
	var b strings.Builder
	for title, link := range seo.OEmbedLinks() {
		b.WriteString(`<link rel="alternate" type="application/json+oembed" href="`)
		b.WriteString(link)
		b.WriteString(`" title="`)
		b.WriteString(html.EscapeString(title))
		b.WriteString("\" />\n")
	}
	return template.HTML(b.String())
}

func normalize(s string) string {
	return escapeDoubleQuotes(stripTags(s))
}

func stripTags(content string) string {
	return re.ReplaceAllString(html.UnescapeString(content), "")
}

func escapeDoubleQuotes(content string) string {
	return escaper.Replace(content)
}
