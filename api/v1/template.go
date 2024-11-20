package v1

import (
	"context"

	"github.com/gowool/echox/api"

	"github.com/gowool/pages/model"
	"github.com/gowool/pages/repository"
)

type TemplateBody struct {
	Name    string `json:"name,omitempty" yaml:"name,omitempty" required:"true"`
	Content string `json:"content,omitempty" yaml:"content,omitempty" required:"false"`
	Enabled bool   `json:"enabled,omitempty" yaml:"enabled,omitempty" required:"false"`
}

func (dto TemplateBody) Decode(_ context.Context, m *model.Template) error {
	m.Name = dto.Name
	m.Content = dto.Content
	m.Enabled = dto.Enabled
	return nil
}

type Template struct {
	api.CRUD[TemplateBody, TemplateBody, model.Template, int64]
}

func NewTemplate(repo repository.Template, errorTransformer api.ErrorTransformerFunc, options ...api.Option) Template {
	opts := make([]api.Option, 0, len(options)+2)
	opts = append(opts, options...)
	opts = append(opts, api.WithPath("/templates"), api.WithTags("template"))

	op := api.Operation(opts...)

	return Template{
		CRUD: api.CRUD[TemplateBody, TemplateBody, model.Template, int64]{
			Info:       Info,
			List:       api.NewList(repo.FindAndCount, errorTransformer, op(api.WithSummary("Get templates"))),
			Read:       api.NewRead(repo.FindByID, errorTransformer, op(api.WithSummary("Get template"), api.WithAddPath("/{id}"))),
			Create:     api.NewCreate[TemplateBody](repo.Create, errorTransformer, op(api.WithPost, api.WithSummary("Create template"))),
			Update:     api.NewUpdate[TemplateBody](repo.FindByID, repo.Update, errorTransformer, op(api.WithPut, api.WithSummary("Update template"), api.WithAddPath("/{id}"))),
			Delete:     api.NewDelete(repo.Delete, errorTransformer, op(api.WithDelete, api.WithSummary("Delete template"), api.WithAddPath("/{id}"))),
			DeleteMany: api.NewDeleteMany(repo.Delete, errorTransformer, op(api.WithDelete, api.WithSummary("Delete templates"))),
		},
	}
}
