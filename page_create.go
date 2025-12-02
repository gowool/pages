package pages

import (
	"context"
	"strings"

	"github.com/gowool/wo"
	"github.com/invopop/validation"
)

type PageCreateRequest struct {
	URL      string `json:"url,omitempty" form:"url,omitempty"`
	Template string `json:"template,omitempty" form:"template,omitempty"`
	Title    string `json:"title,omitempty" form:"title,omitempty"`
}

func (r *PageCreateRequest) Validate() error {
	return validation.ValidateStruct(r,
		validation.Field(&r.URL, validation.Required, validation.Length(1, 254)),
		validation.Field(&r.Template, validation.Required, validation.Length(1, 254)),
		validation.Field(&r.Title, validation.Length(0, 254)),
	)
}

type PageCreate[T Resolver] struct {
	storage    PageStorage
	beforeSave func(ctx context.Context, page *Page) error
}

func NewPageCreate[T Resolver](storage PageStorage, beforeSave func(ctx context.Context, page *Page) error) *PageCreate[T] {
	return &PageCreate[T]{
		storage:    storage,
		beforeSave: beforeSave,
	}
}

func (h *PageCreate[T]) Handle(e T) error {
	var dto PageCreateRequest
	if err := e.BindBody(&dto); err != nil {
		return err
	}

	if err := dto.Validate(); err != nil {
		return err
	}

	dto.URL = strings.TrimRight(dto.URL, "/")
	if len(dto.URL) == 0 || dto.URL[0] != '/' {
		dto.URL = "/" + dto.URL
	}

	ctx := e.Request().Context()
	site := e.Site()

	page := NewPage()
	page.Pattern = PageCMS
	page.Name = strings.ToTitle(strings.ReplaceAll(strings.TrimLeft(dto.URL, "/"), "/", " "))
	page.Site = site
	page.SiteID = site.ID
	page.Title = dto.Title
	page.CustomURL = dto.URL
	page.Template = dto.Template
	page.Decorate = true

	if dto.URL != "/" {
		if index := strings.LastIndex(dto.URL, "/"); index > 0 {
			if p, err := h.storage.FindByURL(ctx, site.ID, dto.URL[:index]); err == nil {
				page.ParentID = &p.ID
				page.Parent = p
			}
		}
	}

	page.FixURL()

	if h.beforeSave != nil {
		if err := h.beforeSave(ctx, page); err != nil {
			return err
		}
	}

	if err := h.storage.Save(ctx, page); err != nil {
		return err
	}

	return wo.NewFoundRedirectError(page.AbsURL())
}
