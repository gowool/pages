package sql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/gowool/cr"

	"github.com/gowool/pages/model"
	"github.com/gowool/pages/repository"
)

var _ repository.Node = (*NodeRepository)(nil)

type NodeRepository struct {
	Repository[model.Node, int64]
	sequenceRepo repository.SequenceNode
}

func NewNodeRepository(db *sql.DB, sequenceRepo repository.SequenceNode, driver Driver) *NodeRepository {
	return &NodeRepository{
		sequenceRepo: sequenceRepo,
		Repository: Repository[model.Node, int64]{
			DB:     db,
			Driver: driver,
			Metadata: Metadata{
				TableName: "nodes",
				MinSize:   50,
				Columns: []string{
					"id", "parent_id", "name", "label", "uri", "path", "level", "position", "display_children",
					"display", "attributes", "link_attributes", "children_attributes", "label_attributes", "metadata",
					"created", "updated",
				},
			},
			RowScan: func(row interface{ Scan(...any) error }, m *model.Node) error {
				var (
					label              sql.NullString
					uri                sql.NullString
					attributes         StrMap
					linkAttributes     StrMap
					childrenAttributes StrMap
					labelAttributes    StrMap
					metadata           StrMap
				)

				if err := row.Scan(&m.ID, &m.ParentID, &m.Name, &label, &uri, &m.Path, &m.Level, &m.Position,
					&m.DisplayChildren, &m.Display, &attributes, &linkAttributes, &childrenAttributes, &labelAttributes,
					&metadata, &m.Created, &m.Updated); err != nil {
					return err
				}
				m.Label = label.String
				m.URI = uri.String
				m.Attributes = attributes
				m.LinkAttributes = linkAttributes
				m.ChildrenAttributes = childrenAttributes
				m.LabelAttributes = labelAttributes
				m.Metadata = metadata
				return nil
			},
			InsertValues: func(m *model.Node) map[string]any {
				now := time.Now()
				return map[string]any{
					"id":                  m.ID,
					"parent_id":           m.ParentID,
					"name":                m.Name,
					"label":               sql.NullString{String: m.Label, Valid: m.Label != ""},
					"uri":                 sql.NullString{String: m.URI, Valid: m.URI != ""},
					"path":                m.Path,
					"level":               m.Level,
					"position":            m.Position,
					"display_children":    m.DisplayChildren,
					"display":             m.Display,
					"attributes":          NewStrMap(m.Attributes),
					"link_attributes":     NewStrMap(m.LinkAttributes),
					"children_attributes": NewStrMap(m.ChildrenAttributes),
					"label_attributes":    NewStrMap(m.LabelAttributes),
					"metadata":            NewStrMap(m.Metadata),
					"created":             now,
					"updated":             now,
				}
			},
			UpdateValues: func(m *model.Node) map[string]any {
				return map[string]any{
					"parent_id":           m.ParentID,
					"name":                m.Name,
					"label":               sql.NullString{String: m.Label, Valid: m.Label != ""},
					"uri":                 sql.NullString{String: m.URI, Valid: m.URI != ""},
					"path":                m.Path,
					"level":               m.Level,
					"position":            m.Position,
					"display_children":    m.DisplayChildren,
					"display":             m.Display,
					"attributes":          NewStrMap(m.Attributes),
					"link_attributes":     NewStrMap(m.LinkAttributes),
					"children_attributes": NewStrMap(m.ChildrenAttributes),
					"label_attributes":    NewStrMap(m.LabelAttributes),
					"metadata":            NewStrMap(m.Metadata),
					"updated":             time.Now(),
				}
			},
		},
	}
}

func (r *NodeRepository) FindWithChildren(ctx context.Context, id int64) ([]model.Node, error) {
	criteria := cr.New().
		SetSortBy(cr.ParseSort("path")...).
		SetFilter(cr.Filter{
			Operator: cr.OpOR,
			Conditions: []any{
				cr.Condition{Column: "path", Operator: cr.OpLIKE, Value: fmt.Sprintf("%%/%d", id)},
				cr.Condition{Column: "path", Operator: cr.OpLIKE, Value: fmt.Sprintf("%%/%d/%%", id)},
			},
		})

	return r.Find(ctx, criteria)
}

func (r *NodeRepository) Create(ctx context.Context, m *model.Node) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return r.Error(err)
	}

	defer func() {
		if err == nil {
			err = r.Error(tx.Commit())
		} else {
			err = errors.Join(err, r.Error(tx.Rollback()))
		}
	}()

	ctx = WithTx(ctx, tx)

	id, err := r.sequenceRepo.Create(ctx)
	if err != nil {
		return r.Error(err)
	}
	m.ID = id

	if err = r.fixPath(ctx, m); err != nil {
		return err
	}
	return r.Repository.Create(ctx, m)
}

func (r *NodeRepository) Update(ctx context.Context, m *model.Node) error {
	if err := r.fixPath(ctx, m); err != nil {
		return err
	}
	return r.Repository.Update(ctx, m)
}

func (r *NodeRepository) Delete(ctx context.Context, ids ...int64) error {
	return r.Error(r.sequenceRepo.Delete(ctx, ids...))
}

func (r *NodeRepository) fixPath(ctx context.Context, m *model.Node) (err error) {
	if m == nil {
		return nil
	}

	if m.ParentID != 0 && m.Parent == nil {
		var parent model.Node
		if parent, err = r.FindByID(ctx, m.ParentID); err != nil {
			return err
		}
		m.Parent = &parent
	}

	*m = m.WithFixedPathAndLevel()
	return
}
