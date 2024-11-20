package sql

import (
	"context"
	"database/sql"
	"time"

	"github.com/gowool/pages/model"
	"github.com/gowool/pages/repository"
)

var _ repository.Template = (*TemplateRepository)(nil)

type TemplateRepository struct {
	Repository[model.Template, int64]
}

func NewTemplateRepository(db *sql.DB, driver Driver) *TemplateRepository {
	return &TemplateRepository{
		Repository[model.Template, int64]{
			DB:     db,
			Driver: driver,
			Metadata: Metadata{
				TableName: "templates",
				Columns:   []string{"id", "name", "content", "enabled", "created", "updated"},
			},
			RowScan: func(row interface{ Scan(...any) error }, m *model.Template) error {
				m.Type = model.TemplateDB
				return row.Scan(&m.ID, &m.Name, &m.Content, &m.Enabled, &m.Created, &m.Updated)
			},
			InsertValues: func(m *model.Template) map[string]any {
				now := time.Now()
				return map[string]any{
					"name":    m.Name,
					"content": m.Content,
					"enabled": m.Enabled,
					"created": now,
					"updated": now,
				}
			},
			UpdateValues: func(m *model.Template) map[string]any {
				return map[string]any{
					"name":    m.Name,
					"content": m.Content,
					"enabled": m.Enabled,
					"updated": time.Now(),
				}
			},
		},
	}
}

func (r *TemplateRepository) FindByName(ctx context.Context, name string) (model.Template, error) {
	return r.FindBy(ctx, "name", name)
}
