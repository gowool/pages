package v1

import (
	"context"
	"time"

	"github.com/gowool/echox/api"

	"github.com/gowool/pages/model"
	"github.com/gowool/pages/repository"
)

type SiteBody struct {
	Name         string            `json:"name,omitempty" yaml:"name,omitempty" required:"true"`
	Title        string            `json:"title,omitempty" yaml:"title,omitempty" required:"false"`
	Separator    string            `json:"separator,omitempty" yaml:"separator,omitempty" required:"true"`
	Host         string            `json:"host,omitempty" yaml:"host,omitempty" required:"true"`
	Locale       string            `json:"locale,omitempty" yaml:"locale,omitempty" required:"false"`
	RelativePath string            `json:"relativePath,omitempty" yaml:"relativePath,omitempty" required:"false"`
	IsDefault    bool              `json:"isDefault,omitempty" yaml:"isDefault,omitempty" required:"false"`
	Javascript   string            `json:"javascript,omitempty" yaml:"javascript,omitempty" required:"false"`
	Stylesheet   string            `json:"stylesheet,omitempty" yaml:"stylesheet,omitempty" required:"false"`
	Metas        []model.Meta      `json:"metas,omitempty" yaml:"metas,omitempty" required:"false"`
	Metadata     map[string]string `json:"metadata,omitempty" yaml:"metadata,omitempty" required:"false"`
	Published    *time.Time        `json:"published,omitempty" yaml:"published,omitempty" required:"false"`
	Expired      *time.Time        `json:"expired,omitempty" yaml:"expired,omitempty" required:"false"`
}

func (dto SiteBody) Decode(_ context.Context, m *model.Site) error {
	m.Name = dto.Name
	m.Title = dto.Title
	m.Separator = dto.Separator
	m.Host = dto.Host
	m.Locale = dto.Locale
	m.RelativePath = dto.RelativePath
	m.IsDefault = dto.IsDefault
	m.Javascript = dto.Javascript
	m.Stylesheet = dto.Stylesheet
	m.Metas = dto.Metas
	m.Metadata = dto.Metadata
	m.Published = dto.Published
	m.Expired = dto.Expired
	return nil
}

type Site struct {
	api.CRUD[SiteBody, SiteBody, model.Site, int64]
}

func NewSite(repo repository.Site, errorTransformer api.ErrorTransformerFunc, options ...api.Option) Site {
	opts := make([]api.Option, 0, len(options)+2)
	opts = append(opts, options...)
	opts = append(opts, api.WithPath("/sites"), api.WithTags("site"))

	op := api.Operation(opts...)

	return Site{
		CRUD: api.CRUD[SiteBody, SiteBody, model.Site, int64]{
			Info:       Info,
			List:       api.NewList(repo.FindAndCount, errorTransformer, op(api.WithSummary("Get sites"))),
			Read:       api.NewRead(repo.FindByID, errorTransformer, op(api.WithSummary("Get site"), api.WithAddPath("/{id}"))),
			Create:     api.NewCreate[SiteBody](repo.Create, errorTransformer, op(api.WithPost, api.WithSummary("Create site"))),
			Update:     api.NewUpdate[SiteBody](repo.FindByID, repo.Update, errorTransformer, op(api.WithPut, api.WithSummary("Update site"), api.WithAddPath("/{id}"))),
			Delete:     api.NewDelete(repo.Delete, errorTransformer, op(api.WithDelete, api.WithSummary("Delete site"), api.WithAddPath("/{id}"))),
			DeleteMany: api.NewDeleteMany(repo.Delete, errorTransformer, op(api.WithDelete, api.WithSummary("Delete sites"))),
		},
	}
}
