package pages

import (
	"errors"
	"io"
	"maps"

	"github.com/gowool/theme"
	"github.com/labstack/echo/v4"

	"github.com/gowool/pages/model"
	"github.com/gowool/pages/repository"
)

type Renderer struct {
	theme   theme.Theme
	cfgRepo repository.Configuration
}

func NewRenderer(theme theme.Theme, cfgRepo repository.Configuration) *Renderer {
	if theme == nil {
		panic("theme is not specified")
	}
	if cfgRepo == nil {
		panic("configuration repository is not specified")
	}
	return &Renderer{theme: theme, cfgRepo: cfgRepo}
}

func (renderer *Renderer) Render(w io.Writer, template string, data any, c echo.Context) error {
	htmlData, ok := data.(map[string]any)
	if !ok {
		htmlData = map[string]any{}
		if data != nil {
			htmlData["data"] = data
		}
	}

	r := c.Request()
	ctx := r.Context()

	cfg, err := renderer.cfgRepo.Load(ctx)
	if err != nil {
		return errors.New("renderer: configuration not found")
	}

	url := *r.URL
	url.User = nil
	ctx = WithURL(ctx, url)

	site := CtxSite(ctx)
	if site == nil {
		if site, ok = htmlData["site"].(*model.Site); !ok {
			site = &model.Site{
				ID:        -1,
				Name:      "Internal",
				Separator: " - ",
			}
			ctx = WithSite(ctx, site)
		}
	}

	page := CtxPage(ctx)
	if page == nil {
		if page, ok = htmlData["page"].(*model.Page); !ok {
			page = &model.Page{
				ID:     -1,
				SiteID: site.ID,
			}
			ctx = WithPage(ctx, page)
		}
	}
	page.Site = site

	for key, value := range page.Headers {
		c.Response().Header().Set(key, value)
	}

	seo := CtxSEO(ctx).Site(site).Page(page)
	ctx = WithSEO(ctx, seo)

	if _, ok = htmlData["debug"]; !ok {
		htmlData["debug"] = cfg.Debug
	}
	htmlData["cfg"] = cfg
	htmlData["url"] = url
	htmlData["site"] = site
	htmlData["page"] = page
	htmlData["seo"] = seo
	htmlData["ctx"] = ctx
	htmlData["csrf"] = ctx.Value("csrf")

	maps.Copy(htmlData, CtxData(r.Context()))

	if t, ok := htmlData["template"].(string); ok {
		template = t
	}

	return renderer.theme.Write(ctx, w, template, htmlData)
}
