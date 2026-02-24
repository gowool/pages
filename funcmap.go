package pages

import (
	"context"
	"fmt"
	"html"
	"html/template"
	"regexp"
	"slices"
	"strings"
)

var re = regexp.MustCompile(`<(.|\n)*?>`)

const defaultSeparator = " - "

func FuncMap(urlGenerator URLGenerator) template.FuncMap {
	return template.FuncMap{
		"page_url":      pageURL(urlGenerator),
		"html_attr":     htmlAttr,
		"body_attr":     bodyAttr,
		"head_attr":     headAttr,
		"head_tags":     headTags,
		"title_tag":     titleTag,
		"rev_title_tag": reverseTitleTag,
	}
}

func pageURL(urlGenerator URLGenerator) func(context.Context, *Site, any, ...any) string {
	return func(ctx context.Context, site *Site, arg any, args ...any) string {
		url, _ := urlGenerator.Generate(ctx, site, arg, args...)
		return url
	}
}

func htmlAttr(ctx context.Context) template.HTMLAttr {
	return mergeAttr(ctx, func(dom DOM) Attr {
		return dom.HTML.Attr
	})
}

func bodyAttr(ctx context.Context) template.HTMLAttr {
	return mergeAttr(ctx, func(dom DOM) Attr {
		return dom.Body.Attr
	})
}

func headAttr(ctx context.Context) template.HTMLAttr {
	return mergeAttr(ctx, func(dom DOM) Attr {
		return dom.Head.Attr
	})
}

func mergeAttr(ctx context.Context, fn func(DOM) Attr) template.HTMLAttr {
	c := MustContext(ctx)

	var attr Attr
	if c.HasSite() {
		attr = attr.With(fn(c.Site().DOM))
	}
	if c.HasPage() {
		attr = attr.With(fn(c.Page().DOM))
	}
	attr = attr.With(fn(*c.DOM()))

	return attr.HTML()
}

func headTags(ctx context.Context) template.HTML {
	c := MustContext(ctx)

	n := len(c.DOM().Head.Nodes)
	if c.HasSite() {
		n += len(c.Site().DOM.Head.Nodes)
	}
	if c.HasPage() {
		n += len(c.Page().DOM.Head.Nodes)
	}
	if n == 0 {
		return ""
	}

	nodes := make(Nodes, 0, n)
	if c.HasSite() {
		nodes = append(nodes, c.Site().DOM.Head.Nodes...)
	}
	if c.HasPage() {
		nodes = append(nodes, c.Page().DOM.Head.Nodes...)
	}
	nodes = append(nodes, c.DOM().Head.Nodes...)

	return nodes.HTML()
}

func titleTag(ctx context.Context, args ...string) template.HTML {
	data, sep := titleData(ctx, args...)

	return template.HTML(fmt.Sprintf("<title>%s</title>", stripTags(strings.Join(data, sep))))
}

func reverseTitleTag(ctx context.Context, args ...string) template.HTML {
	data, sep := titleData(ctx, args...)
	slices.Reverse(data)

	return template.HTML(fmt.Sprintf("<title>%s</title>", stripTags(strings.Join(data, sep))))
}

func titleData(ctx context.Context, args ...string) (data []string, sep string) {
	c := MustContext(ctx)

	sep = defaultSeparator
	if c.HasSite() && c.Site().Separator != "" {
		sep = c.Site().Separator
	}

	n := len(args)
	if c.HasSite() && c.Site().Title != "" {
		n++
	}
	if c.HasPage() && c.Page().Title != "" {
		n++
	}

	data = make([]string, 0, n)
	if c.HasSite() && c.Site().Title != "" {
		data = append(data, c.Site().Title)
	}
	if c.HasPage() && c.Page().Title != "" {
		data = append(data, c.Page().Title)
	}
	data = append(data, args...)

	return
}

func stripTags(content string) string {
	return re.ReplaceAllString(html.UnescapeString(content), "")
}
