package pages

import (
	"context"
	"net/url"

	"github.com/gowool/pages/model"
	"github.com/gowool/pages/seo"
)

type (
	ctxDebugKey       struct{}
	skipSelectSiteKey struct{}
	skipSelectPageKey struct{}
	siteKey           struct{}
	pageKey           struct{}
	seoKey            struct{}
	editorKey         struct{}
	dataKey           struct{}
	urlKey            struct{}
)

func WithDebug(ctx context.Context, debug bool) context.Context {
	return context.WithValue(ctx, ctxDebugKey{}, debug)
}

func CtxDebug(ctx context.Context) bool {
	debug, _ := ctx.Value(ctxDebugKey{}).(bool)
	return debug
}

func SetSkipSelectSite(ctx context.Context) context.Context {
	return context.WithValue(ctx, skipSelectSiteKey{}, true)
}

func SkipSelectSite(ctx context.Context) bool {
	return ctx.Value(skipSelectSiteKey{}) != nil
}

func SetSkipSelectPage(ctx context.Context) context.Context {
	return context.WithValue(ctx, skipSelectPageKey{}, true)
}

func SkipSelectPage(ctx context.Context) bool {
	return ctx.Value(skipSelectPageKey{}) != nil
}

func WithSite(ctx context.Context, m *model.Site) context.Context {
	return context.WithValue(ctx, siteKey{}, m)
}

func CtxSite(ctx context.Context) *model.Site {
	m, _ := ctx.Value(siteKey{}).(*model.Site)
	return m
}

func WithPage(ctx context.Context, m *model.Page) context.Context {
	return context.WithValue(ctx, pageKey{}, m)
}

func CtxPage(ctx context.Context) *model.Page {
	m, _ := ctx.Value(pageKey{}).(*model.Page)
	return m
}

func WithSEO(ctx context.Context, s seo.SEO) context.Context {
	return context.WithValue(ctx, seoKey{}, s)
}

func CtxSEO(ctx context.Context) seo.SEO {
	if s, ok := ctx.Value(seoKey{}).(seo.SEO); ok {
		return s
	}
	return seo.NewSEO()
}

func WithEditor(ctx context.Context, isEditor bool) context.Context {
	return context.WithValue(ctx, editorKey{}, isEditor)
}

func CtxEditor(ctx context.Context) bool {
	isEditor, _ := ctx.Value(editorKey{}).(bool)
	return isEditor
}

func WithData(ctx context.Context, data map[string]any) context.Context {
	return context.WithValue(ctx, dataKey{}, data)
}

func CtxData(ctx context.Context) map[string]any {
	if data, ok := ctx.Value(dataKey{}).(map[string]any); ok {
		return data
	}
	return make(map[string]any)
}

func WithURL(ctx context.Context, u url.URL) context.Context {
	return context.WithValue(ctx, urlKey{}, u)
}

func CtxURL(ctx context.Context) url.URL {
	u, _ := ctx.Value(urlKey{}).(url.URL)
	return u
}
