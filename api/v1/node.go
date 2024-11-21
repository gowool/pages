package v1

import (
	"context"

	"github.com/gowool/echox/api"

	"github.com/gowool/pages/model"
	"github.com/gowool/pages/repository"
)

type NodeBody struct {
	ParentID           int64             `json:"parentID,omitempty" yaml:"parentID,omitempty" required:"false"`
	Name               string            `json:"name,omitempty" yaml:"name,omitempty" required:"true"`
	Label              string            `json:"label,omitempty" yaml:"label,omitempty" required:"false"`
	URI                string            `json:"uri,omitempty" yaml:"uri,omitempty" required:"false"`
	Position           int               `json:"position,omitempty" yaml:"position,omitempty" required:"false"`
	DisplayChildren    bool              `json:"displayChildren,omitempty" yaml:"displayChildren,omitempty" required:"false"`
	Display            bool              `json:"display,omitempty" yaml:"display,omitempty" required:"false"`
	Attributes         map[string]string `json:"attributes,omitempty" yaml:"attributes,omitempty" required:"false"`
	LinkAttributes     map[string]string `json:"linkAttributes,omitempty" yaml:"linkAttributes,omitempty" required:"false"`
	ChildrenAttributes map[string]string `json:"childrenAttributes,omitempty" yaml:"childrenAttributes,omitempty" required:"false"`
	LabelAttributes    map[string]string `json:"labelAttributes,omitempty" yaml:"labelAttributes,omitempty" required:"false"`
	Metadata           map[string]string `json:"metadata,omitempty" yaml:"metadata,omitempty" required:"false"`
}

func (dto NodeBody) Decode(_ context.Context, m *model.Node) error {
	m.ParentID = dto.ParentID
	m.Name = dto.Name
	m.Label = dto.Label
	m.URI = dto.URI
	m.Position = dto.Position
	m.DisplayChildren = dto.DisplayChildren
	m.Display = dto.Display
	m.Attributes = dto.Attributes
	m.LinkAttributes = dto.LinkAttributes
	m.ChildrenAttributes = dto.ChildrenAttributes
	m.LabelAttributes = dto.LabelAttributes
	m.Metadata = dto.Metadata
	return nil
}

type Node struct {
	api.CRUD[NodeBody, NodeBody, model.Node, int64]
}

func NewNode(repo repository.Node, errorTransformer api.ErrorTransformerFunc, options ...api.Option) Node {
	opts := make([]api.Option, 0, len(options)+2)
	opts = append(opts, options...)
	opts = append(opts, api.WithPath("/nodes"), api.WithAddTags("node"))

	op := api.Operation(opts...)

	return Node{
		CRUD: api.CRUD[NodeBody, NodeBody, model.Node, int64]{
			Info:       Info,
			List:       api.NewList(repo.FindAndCount, errorTransformer, op(api.WithSummary("Get nodes"))),
			Read:       api.NewRead(repo.FindByID, errorTransformer, op(api.WithSummary("Get node"), api.WithAddPath("/{id}"))),
			Create:     api.NewCreate[NodeBody](repo.Create, errorTransformer, op(api.WithPost, api.WithSummary("Create node"))),
			Update:     api.NewUpdate[NodeBody](repo.FindByID, repo.Update, errorTransformer, op(api.WithPut, api.WithSummary("Update node"), api.WithAddPath("/{id}"))),
			Delete:     api.NewDelete(repo.Delete, errorTransformer, op(api.WithDelete, api.WithSummary("Delete node"), api.WithAddPath("/{id}"))),
			DeleteMany: api.NewDeleteMany(repo.Delete, errorTransformer, op(api.WithDelete, api.WithSummary("Delete nodes"))),
		},
	}
}
