package sql

import (
	"context"
	"database/sql"
	"time"

	"github.com/gowool/pages/model"
	"github.com/gowool/pages/repository"
)

var _ repository.Menu = (*MenuRepository)(nil)

type MenuRepository struct {
	Repository[model.Menu, int64]
}

func NewMenuRepository(db *sql.DB, driver Driver) *MenuRepository {
	return &MenuRepository{
		Repository[model.Menu, int64]{
			DB:     db,
			Driver: driver,
			Metadata: Metadata{
				TableName: "menus",
				Columns:   []string{"id", "node_id", "name", "handle", "enabled", "created", "updated"},
			},
			RowScan: func(row interface{ Scan(...any) error }, m *model.Menu) error {
				return row.Scan(&m.ID, &m.NodeID, &m.Name, &m.Handle, &m.Enabled, &m.Created, &m.Updated)
			},
			InsertValues: func(m *model.Menu) map[string]any {
				now := time.Now()
				return map[string]any{
					"node_id": m.NodeID,
					"name":    m.Name,
					"handle":  m.Handle,
					"enabled": m.Enabled,
					"created": now,
					"updated": now,
				}
			},
			UpdateValues: func(m *model.Menu) map[string]any {
				return map[string]any{
					"node_id": m.NodeID,
					"name":    m.Name,
					"handle":  m.Handle,
					"enabled": m.Enabled,
					"updated": time.Now(),
				}
			},
		},
	}
}

func (r *MenuRepository) FindByHandle(ctx context.Context, handle string) (model.Menu, error) {
	menu, err := r.FindBy(ctx, "handle", handle)
	if err != nil {
		return model.Menu{}, err
	}
	if !menu.Enabled {
		return model.Menu{}, r.Error(sql.ErrNoRows)
	}
	return menu, nil
}

func (r *MenuRepository) Create(ctx context.Context, m *model.Menu) error {
	r.fixHandle(m)
	return r.Repository.Create(ctx, m)
}

func (r *MenuRepository) Update(ctx context.Context, m *model.Menu) error {
	r.fixHandle(m)
	return r.Repository.Update(ctx, m)
}

func (r *MenuRepository) fixHandle(m *model.Menu) {
	if m != nil {
		*m = m.WithFixedHandle()
	}
}
