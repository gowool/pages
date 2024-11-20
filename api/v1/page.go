package v1

import (
	"context"
	"maps"
	"slices"
	"strings"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/gowool/echox/api"
	"github.com/labstack/echo/v4"

	"github.com/gowool/pages/model"
	"github.com/gowool/pages/repository"
)

type PageBody struct {
	SiteID     int64             `json:"siteID,omitempty" yaml:"siteID,omitempty" required:"true"`
	ParentID   *int64            `json:"parentID,omitempty" yaml:"parentID,omitempty" required:"false"`
	Name       string            `json:"name,omitempty" yaml:"name,omitempty" required:"true"`
	Title      string            `json:"title,omitempty" yaml:"title,omitempty" required:"false"`
	Pattern    string            `json:"pattern,omitempty" yaml:"pattern,omitempty" required:"true"`
	Alias      string            `json:"alias,omitempty" yaml:"alias,omitempty" required:"false"`
	Slug       string            `json:"slug,omitempty" yaml:"slug,omitempty" required:"false"`
	CustomURL  string            `json:"customURL,omitempty" yaml:"customURL,omitempty" required:"false"`
	Javascript string            `json:"javascript,omitempty" yaml:"javascript,omitempty" required:"false"`
	Stylesheet string            `json:"stylesheet,omitempty" yaml:"stylesheet,omitempty" required:"false"`
	Template   string            `json:"template,omitempty" yaml:"template,omitempty" required:"true"`
	Decorate   bool              `json:"decorate,omitempty" yaml:"decorate,omitempty" required:"false"`
	Position   int               `json:"position,omitempty" yaml:"position,omitempty" required:"false"`
	Headers    map[string]string `json:"headers,omitempty" yaml:"headers,omitempty" required:"false"`
	Metas      []model.Meta      `json:"metas,omitempty" yaml:"metas,omitempty" required:"false"`
	Metadata   map[string]string `json:"metadata,omitempty" yaml:"metadata,omitempty" required:"false"`
	Published  *time.Time        `json:"published,omitempty" yaml:"published,omitempty" required:"false"`
	Expired    *time.Time        `json:"expired,omitempty" yaml:"expired,omitempty" required:"false"`
}

func (dto PageBody) Decode(_ context.Context, m *model.Page) error {
	m.SiteID = dto.SiteID
	m.ParentID = dto.ParentID
	m.Name = dto.Name
	m.Title = dto.Title
	m.Pattern = dto.Pattern
	m.Alias = dto.Alias
	m.Slug = dto.Slug
	m.CustomURL = dto.CustomURL
	m.Javascript = dto.Javascript
	m.Stylesheet = dto.Stylesheet
	m.Template = dto.Template
	m.Decorate = dto.Decorate
	m.Position = dto.Position
	m.Headers = dto.Headers
	m.Metas = dto.Metas
	m.Metadata = dto.Metadata
	m.Published = dto.Published
	m.Expired = dto.Expired
	return nil
}

type Page struct {
	api.CRUD[PageBody, PageBody, model.Page, int64]
	cfgRepo         repository.Configuration
	hybridOperation huma.Operation
}

func NewPage(pageRepo repository.Page, cfgRepo repository.Configuration, errorTransformer api.ErrorTransformerFunc, options ...api.Option) Page {
	opts := make([]api.Option, 0, len(options)+2)
	opts = append(opts, options...)
	opts = append(opts, api.WithPath("/pages"), api.WithTags("page"))

	op := api.Operation(opts...)

	return Page{
		CRUD: api.CRUD[PageBody, PageBody, model.Page, int64]{
			Info:       Info,
			List:       api.NewList(pageRepo.FindAndCount, errorTransformer, op(api.WithSummary("Get pages"))),
			Read:       api.NewRead(pageRepo.FindByID, errorTransformer, op(api.WithSummary("Get page"), api.WithAddPath("/{id}"))),
			Create:     api.NewCreate[PageBody](pageRepo.Create, errorTransformer, op(api.WithPost, api.WithSummary("Create page"))),
			Update:     api.NewUpdate[PageBody](pageRepo.FindByID, pageRepo.Update, errorTransformer, op(api.WithPut, api.WithSummary("Update page"), api.WithAddPath("/{id}"))),
			Delete:     api.NewDelete(pageRepo.Delete, errorTransformer, op(api.WithDelete, api.WithSummary("Delete page"), api.WithAddPath("/{id}"))),
			DeleteMany: api.NewDeleteMany(pageRepo.Delete, errorTransformer, op(api.WithDelete, api.WithSummary("Delete pages"))),
		},
		cfgRepo:         cfgRepo,
		hybridOperation: op(api.WithSummary("Get hybrid patterns"), api.WithAddPath("/hybrid-patterns")),
	}
}

func (h Page) Register(e *echo.Echo, humaAPI huma.API) {
	h.CRUD.Register(e, humaAPI)
	api.Register(humaAPI, api.Transform(h.List.ErrorTransformer, h.HybridPatterns(e)), h.hybridOperation)
}

type Route struct {
	Pattern string   `json:"pattern" yaml:"pattern" required:"true"`
	Methods []string `json:"methods,omitempty" yaml:"methods,omitempty" required:"false"`
}

func (h Page) HybridPatterns(e *echo.Echo) func(context.Context, *struct{}) (*api.Response[[]*Route], error) {
	return func(ctx context.Context, _ *struct{}) (*api.Response[[]*Route], error) {
		cfg, err := h.cfgRepo.Load(ctx)
		if err != nil {
			return nil, err
		}

		routes := map[string]*Route{}
		for _, r := range e.Routes() {
			if r.Method == echo.RouteNotFound {
				continue
			}
			if cfg.IgnorePattern(r.Path) {
				continue
			}

			p, ok := routes[r.Path]
			if !ok {
				p = &Route{Pattern: r.Path}
				routes[r.Path] = p
			}
			if r.Method != "" {
				p.Methods = append(p.Methods, r.Method)
			}
		}
		return &api.Response[[]*Route]{
			Body: slices.SortedFunc(maps.Values(routes), func(i, j *Route) int {
				return strings.Compare(i.Pattern, j.Pattern)
			}),
		}, nil
	}
}
