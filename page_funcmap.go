package pages

import (
	"context"
	"html/template"
)

func PageFuncMap(urlGenerator URLGenerator) template.FuncMap {
	return template.FuncMap{
		"page_url": func(ctx context.Context, site *Site, arg any, args ...any) string {
			url, _ := urlGenerator.Generate(ctx, site, arg, args...)
			return url
		},
	}
}
