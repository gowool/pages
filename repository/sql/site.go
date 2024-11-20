package sql

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/gowool/cr"

	"github.com/gowool/pages/model"
	"github.com/gowool/pages/repository"
)

var _ repository.Site = (*SiteRepository)(nil)

type SiteRepository struct {
	Repository[model.Site, int64]
}

func NewSiteRepository(db *sql.DB, driver Driver) *SiteRepository {
	return &SiteRepository{
		Repository[model.Site, int64]{
			DB:     db,
			Driver: driver,
			Metadata: Metadata{
				TableName: "sites",
				Columns: []string{
					"id", "name", "title", "separator", "host", "locale", "relative_path", "is_default",
					"javascript", "stylesheet", "metas", "metadata", "created", "updated", "published", "expired",
				},
			},
			RowScan: func(row interface{ Scan(...any) error }, m *model.Site) error {
				var (
					title        sql.NullString
					locale       sql.NullString
					relativePath sql.NullString
					javascript   sql.NullString
					stylesheet   sql.NullString
					metas        Metas
					metadata     StrMap
				)

				if err := row.Scan(&m.ID, &m.Name, &title, &m.Separator, &m.Host, &locale, &relativePath,
					&m.IsDefault, &javascript, &stylesheet, &metas, &metadata, &m.Created, &m.Updated,
					&m.Published, &m.Expired); err != nil {
					return err
				}

				m.Title = title.String
				m.Locale = locale.String
				m.RelativePath = relativePath.String
				m.Javascript = javascript.String
				m.Stylesheet = stylesheet.String
				m.Metas = metas
				m.Metadata = metadata
				return nil
			},
			InsertValues: func(m *model.Site) map[string]any {
				now := time.Now()
				return map[string]any{
					"name":          m.Name,
					"title":         sql.NullString{String: m.Title, Valid: m.Title != ""},
					"separator":     m.Separator,
					"host":          m.Host,
					"locale":        sql.NullString{String: m.Locale, Valid: m.Locale != ""},
					"relative_path": sql.NullString{String: m.RelativePath, Valid: m.RelativePath != ""},
					"is_default":    m.IsDefault,
					"javascript":    sql.NullString{String: m.Javascript, Valid: m.Javascript != ""},
					"stylesheet":    sql.NullString{String: m.Stylesheet, Valid: m.Stylesheet != ""},
					"metas":         NewMetas(m.Metas),
					"metadata":      NewStrMap(m.Metadata),
					"created":       now,
					"updated":       now,
					"published":     m.Published,
					"expired":       m.Expired,
				}
			},
			UpdateValues: func(m *model.Site) map[string]any {
				return map[string]any{
					"name":          m.Name,
					"title":         sql.NullString{String: m.Title, Valid: m.Title != ""},
					"separator":     m.Separator,
					"host":          m.Host,
					"locale":        sql.NullString{String: m.Locale, Valid: m.Locale != ""},
					"relative_path": sql.NullString{String: m.RelativePath, Valid: m.RelativePath != ""},
					"is_default":    m.IsDefault,
					"javascript":    sql.NullString{String: m.Javascript, Valid: m.Javascript != ""},
					"stylesheet":    sql.NullString{String: m.Stylesheet, Valid: m.Stylesheet != ""},
					"metas":         NewMetas(m.Metas),
					"metadata":      NewStrMap(m.Metadata),
					"updated":       time.Now(),
					"published":     m.Published,
					"expired":       m.Expired,
				}
			},
			OnError: func(err error) error {
				if errors.Is(err, sql.ErrNoRows) {
					return errors.Join(repository.ErrSiteNotFound, err)
				}
				return err
			},
		},
	}
}

func (r *SiteRepository) FindByHosts(ctx context.Context, hosts []string, now time.Time) ([]model.Site, error) {
	conditions := []any{cr.Condition{Column: "host", Operator: cr.OpIN, Value: hosts}}

	if !now.IsZero() {
		conditions = append(conditions, repository.LifeSpanConditions("", now)...)
	}

	return r.Find(ctx, &cr.Criteria{
		Filter: cr.Filter{
			Conditions: conditions,
		},
		SortBy: cr.SortBy{cr.Sort{Column: "is_default", Order: "DESC"}},
	})
}
