package pages

import (
	"fmt"
	"html"
	"html/template"
	"iter"
	"regexp"
	"strings"

	"github.com/spf13/cast"
)

var (
	re      = regexp.MustCompile(`<(.|\n)*?>`)
	escaper = strings.NewReplacer(`"`, "&quot;")
)

var SEOFuncMap = template.FuncMap{
	"strip_tags":        stripTags,
	"escape_double_q":   escapeDoubleQuotes,
	"reverse_title_tag": reverseTitleTag,
	"title_tag":         titleTag,
	"meta_tags":         metaTags,
	"html_attrs":        htmlAttrs,
	"head_attrs":        headAttrs,
	"body_attrs":        bodyAttrs,
	"lang_alternates":   langAlternates,
	"head_links":        headLinks,
}

func titleTag(seo *SEO, args ...string) template.HTML {
	return template.HTML(
		fmt.Sprintf(
			"<title>%s</title>",
			stripTags(strings.Join(append([]string{seo.Title()}, args...), seo.Separator())),
		),
	)
}

func reverseTitleTag(seo *SEO, args ...string) template.HTML {
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

func metaTags(seo *SEO) template.HTML {
	var b strings.Builder

	mt := seo.MetaTags()

	if mt.Charset != "" {
		b.WriteString(`<meta charset="`)
		b.WriteString(normalize(mt.Charset))
		b.WriteString("\" />\n")
	}

	for typ, metas := range iter.Seq2[string, map[string][]string](func(yield func(string, map[string][]string) bool) {
		yield("name", mt.Name)
		yield("property", mt.Property)
		yield("http-equiv", mt.HTTPEquiv)
	}) {
		for name, contents := range metas {
			for _, content := range contents {
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
	}
	return template.HTML(b.String())
}

func htmlAttrs(seo *SEO, rest ...any) template.HTMLAttr {
	return attrs(seo.HTMLAttributes(), rest...)
}

func headAttrs(seo *SEO, rest ...any) template.HTMLAttr {
	return attrs(seo.HeadAttributes(), rest...)
}

func bodyAttrs(seo *SEO, rest ...any) template.HTMLAttr {
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

func langAlternates(seo *SEO) template.HTML {
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

func headLinks(seo *SEO) template.HTML {
	var b strings.Builder
	for _, link := range seo.Links() {
		if link.Rel == "" {
			continue
		}

		b.WriteString(`<link rel="`)
		b.WriteString(html.EscapeString(link.Rel))
		b.WriteString(`" `)

		if link.CrossOrigin != "" {
			b.WriteString(`crossorigin="`)
			b.WriteString(html.EscapeString(link.CrossOrigin))
			b.WriteString(`" `)
		}

		if link.Href != "" {
			b.WriteString(`href="`)
			b.WriteString(html.EscapeString(link.Href))
			b.WriteString(`" `)
		}

		if link.HrefLang != "" {
			b.WriteString(`hreflang="`)
			b.WriteString(html.EscapeString(link.HrefLang))
			b.WriteString(`" `)
		}

		if link.Media != "" {
			b.WriteString(`media="`)
			b.WriteString(html.EscapeString(link.Media))
			b.WriteString(`" `)
		}

		if link.Sizes != "" {
			b.WriteString(`sizes="`)
			b.WriteString(html.EscapeString(link.Sizes))
			b.WriteString(`" `)
		}

		if link.Title != "" {
			b.WriteString(`title="`)
			b.WriteString(html.EscapeString(link.Title))
			b.WriteString(`" `)
		}

		if link.Type != "" {
			b.WriteString(`type="`)
			b.WriteString(html.EscapeString(link.Type))
			b.WriteString(`" `)
		}

		b.WriteString("/>\n")
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
