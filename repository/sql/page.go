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

var _ repository.Page = (*PageRepository)(nil)

type PageRepository struct {
	Repository[model.Page, int64]
}

func NewPageRepository(db *sql.DB, driver Driver) *PageRepository {
	return &PageRepository{
		Repository[model.Page, int64]{
			DB:     db,
			Driver: driver,
			Metadata: Metadata{
				TableName: "pages",
				Columns: []string{
					"id", "site_id", "parent_id", "name", "title", "pattern", "alias", "slug", "url", "custom_url",
					"javascript", "stylesheet", "template", "decorate", "position", "headers", "metas", "metadata",
					"created", "updated", "published", "expired",
				},
			},
			RowScan: func(row interface{ Scan(...any) error }, m *model.Page) error {
				var (
					title      sql.NullString
					alias      sql.NullString
					slug       sql.NullString
					url        sql.NullString
					customURL  sql.NullString
					javascript sql.NullString
					stylesheet sql.NullString
					metas      Metas
					metadata   StrMap
					headers    StrMap
				)

				if err := row.Scan(&m.ID, &m.SiteID, &m.ParentID, &m.Name, &title, &m.Pattern, &alias, &slug,
					&url, &customURL, &javascript, &stylesheet, &m.Template, &m.Decorate, &m.Position, &headers,
					&metas, &metadata, &m.Created, &m.Updated, &m.Published, &m.Expired); err != nil {
					return err
				}

				m.Title = title.String
				m.Alias = alias.String
				m.Slug = slug.String
				m.URL = url.String
				m.CustomURL = customURL.String
				m.Javascript = javascript.String
				m.Stylesheet = stylesheet.String
				m.Metas = metas
				m.Metadata = metadata
				m.Headers = headers
				return nil
			},
			InsertValues: func(m *model.Page) map[string]any {
				now := time.Now()
				return map[string]any{
					"site_id":    m.SiteID,
					"parent_id":  m.ParentID,
					"name":       m.Name,
					"title":      sql.NullString{String: m.Title, Valid: m.Title != ""},
					"pattern":    m.Pattern,
					"alias":      sql.NullString{String: m.Alias, Valid: m.Alias != ""},
					"slug":       sql.NullString{String: m.Slug, Valid: m.Slug != ""},
					"url":        sql.NullString{String: m.URL, Valid: m.URL != ""},
					"custom_url": sql.NullString{String: m.CustomURL, Valid: m.CustomURL != ""},
					"javascript": sql.NullString{String: m.Javascript, Valid: m.Javascript != ""},
					"stylesheet": sql.NullString{String: m.Stylesheet, Valid: m.Stylesheet != ""},
					"template":   m.Template,
					"decorate":   m.Decorate,
					"position":   m.Position,
					"headers":    NewStrMap(m.Headers),
					"metas":      NewMetas(m.Metas),
					"metadata":   NewStrMap(m.Metadata),
					"created":    now,
					"updated":    now,
					"published":  m.Published,
					"expired":    m.Expired,
				}
			},
			UpdateValues: func(m *model.Page) map[string]any {
				metas := Metas(m.Metas)
				if metas == nil {
					metas = Metas{}
				}
				return map[string]any{
					"site_id":    m.SiteID,
					"parent_id":  m.ParentID,
					"name":       m.Name,
					"title":      sql.NullString{String: m.Title, Valid: m.Title != ""},
					"pattern":    m.Pattern,
					"alias":      sql.NullString{String: m.Alias, Valid: m.Alias != ""},
					"slug":       sql.NullString{String: m.Slug, Valid: m.Slug != ""},
					"url":        sql.NullString{String: m.URL, Valid: m.URL != ""},
					"custom_url": sql.NullString{String: m.CustomURL, Valid: m.CustomURL != ""},
					"javascript": sql.NullString{String: m.Javascript, Valid: m.Javascript != ""},
					"stylesheet": sql.NullString{String: m.Stylesheet, Valid: m.Stylesheet != ""},
					"template":   m.Template,
					"decorate":   m.Decorate,
					"position":   m.Position,
					"headers":    NewStrMap(m.Headers),
					"metas":      metas,
					"metadata":   NewStrMap(m.Metadata),
					"updated":    time.Now(),
					"published":  m.Published,
					"expired":    m.Expired,
				}
			},
			OnError: func(err error) error {
				if errors.Is(err, sql.ErrNoRows) {
					return errors.Join(repository.ErrPageNotFound, err)
				}
				return err
			},
		},
	}
}

func (r *PageRepository) FindByParentID(ctx context.Context, parentID int64, now time.Time) ([]model.Page, error) {
	conditions := []any{cr.Condition{Column: "parent_id", Value: parentID}}

	if !now.IsZero() {
		conditions = append(conditions, repository.LifeSpanConditions("", now)...)
	}

	return r.Find(ctx, &cr.Criteria{
		Filter: cr.Filter{
			Conditions: conditions,
		},
		SortBy: cr.SortBy{cr.Sort{Column: "position", Order: "ASC"}},
	})
}

func (r *PageRepository) FindByPattern(ctx context.Context, siteID int64, pattern string, now time.Time) (model.Page, error) {
	return r.findBy(ctx, siteID, "pattern", pattern, now)
}

func (r *PageRepository) FindByAlias(ctx context.Context, siteID int64, alias string, now time.Time) (model.Page, error) {
	return r.findBy(ctx, siteID, "alias", alias, now)
}

func (r *PageRepository) FindByURL(ctx context.Context, siteID int64, url string, now time.Time) (model.Page, error) {
	return r.findBy(ctx, siteID, "url", url, now)
}

func (r *PageRepository) findBy(ctx context.Context, siteID int64, column, value string, now time.Time) (model.Page, error) {
	conditions := []any{cr.Condition{Column: column, Value: value}}

	if siteID > 0 {
		conditions = append(conditions, cr.Condition{Column: "site_id", Value: siteID})
	}

	if !now.IsZero() {
		conditions = append(conditions, repository.LifeSpanConditions("", now)...)
	}

	data, err := r.Find(ctx, cr.New().SetFilter(cr.Filter{Conditions: conditions}).SetSize(1))
	if err != nil {
		return model.Page{}, err
	}
	if len(data) == 0 {
		return model.Page{}, r.Error(sql.ErrNoRows)
	}
	return data[0], nil
}

func (r *PageRepository) Create(ctx context.Context, m *model.Page) error {
	if err := r.fixURL(ctx, m); err != nil {
		return err
	}
	return r.Repository.Create(ctx, m)
}

func (r *PageRepository) Update(ctx context.Context, m *model.Page) error {
	if err := r.fixURL(ctx, m); err != nil {
		return err
	}
	return r.Repository.Update(ctx, m)
}

func (r *PageRepository) fixURL(ctx context.Context, m *model.Page) error {
	if m == nil || m.ParentID == nil || *m.ParentID == 0 {
		return nil
	}

	p, err := r.FindByID(ctx, *m.ParentID)
	if err != nil {
		return fmt.Errorf("failed to find parent page: %w", err)
	}

	m.Parent = &p
	*m = m.WithFixedURL()
	return nil
}
