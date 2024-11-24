package theme

import (
	"fmt"
	"html"
	"html/template"
	"regexp"
	"strings"

	"github.com/gowool/theme"
	"github.com/spf13/cast"

	"github.com/gowool/pages/seo"
)

var (
	re      = regexp.MustCompile(`<(.|\n)*?>`)
	escaper = strings.NewReplacer(`"`, "&quot;")
)

type FuncMap struct{}

func NewFuncMap() *FuncMap {
	return &FuncMap{}
}

func (fm *FuncMap) FuncMap(_ theme.Theme) template.FuncMap {
	return template.FuncMap{
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
		"js": func(str string) template.JS {
			return template.JS(str)
		},
		"css": func(str string) template.CSS {
			return template.CSS(str)
		},
	}
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
