package fx

import (
	"github.com/gowool/echox/api"

	"github.com/gowool/pages/api/v1"
	"github.com/gowool/pages/repository"
)

func NewConfigurationAPI(r repository.Configuration, options ...api.Option) v1.Configuration {
	return v1.NewConfiguration(r, v1.ErrorTransformer, options...)
}

func NewPageAPI(r repository.Page, cfg repository.Configuration, options ...api.Option) v1.Page {
	return v1.NewPage(r, cfg, v1.ErrorTransformer, options...)
}

func NewSiteAPI(r repository.Site, options ...api.Option) v1.Site {
	return v1.NewSite(r, v1.ErrorTransformer, options...)
}

func NewTemplateAPI(r repository.Template, options ...api.Option) v1.Template {
	return v1.NewTemplate(r, v1.ErrorTransformer, options...)
}

func NewMenuAPI(r repository.Menu, options ...api.Option) v1.Menu {
	return v1.NewMenu(r, v1.ErrorTransformer, options...)
}

func NewNodeAPI(r repository.Node, options ...api.Option) v1.Node {
	return v1.NewNode(r, v1.ErrorTransformer, options...)
}
