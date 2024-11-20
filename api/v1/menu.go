package v1

import (
	"context"

	"github.com/gowool/echox/api"

	"github.com/gowool/pages/model"
	"github.com/gowool/pages/repository"
)

type MenuBody struct {
	NodeID  *int64 `json:"nodeID,omitempty" yaml:"nodeID,omitempty" required:"false"`
	Name    string `json:"name,omitempty" yaml:"name,omitempty" required:"true"`
	Handle  string `json:"handle,omitempty" yaml:"handle,omitempty" required:"false"`
	Enabled bool   `json:"enabled,omitempty" yaml:"enabled,omitempty" required:"false"`
}

func (dto MenuBody) Decode(_ context.Context, m *model.Menu) error {
	m.NodeID = dto.NodeID
	m.Name = dto.Name
	m.Handle = dto.Handle
	m.Enabled = dto.Enabled
	return nil
}

type Menu struct {
	api.CRUD[MenuBody, MenuBody, model.Menu, int64]
}

func NewMenu(repo repository.Menu, errorTransformer api.ErrorTransformerFunc, options ...api.Option) Menu {
	opts := make([]api.Option, 0, len(options)+2)
	opts = append(opts, options...)
	opts = append(opts, api.WithPath("/menus"), api.WithTags("menu"))

	op := api.Operation(opts...)

	return Menu{
		CRUD: api.CRUD[MenuBody, MenuBody, model.Menu, int64]{
			Info:       Info,
			List:       api.NewList(repo.FindAndCount, errorTransformer, op(api.WithSummary("Get menus"))),
			Read:       api.NewRead(repo.FindByID, errorTransformer, op(api.WithSummary("Get menu"), api.WithAddPath("/{id}"))),
			Create:     api.NewCreate[MenuBody](repo.Create, errorTransformer, op(api.WithPost, api.WithSummary("Create menu"))),
			Update:     api.NewUpdate[MenuBody](repo.FindByID, repo.Update, errorTransformer, op(api.WithPut, api.WithSummary("Update menu"), api.WithAddPath("/{id}"))),
			Delete:     api.NewDelete(repo.Delete, errorTransformer, op(api.WithDelete, api.WithSummary("Delete menu"), api.WithAddPath("/{id}"))),
			DeleteMany: api.NewDeleteMany(repo.Delete, errorTransformer, op(api.WithDelete, api.WithSummary("Delete menus"))),
		},
	}
}
