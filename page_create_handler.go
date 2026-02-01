package pages

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/invopop/validation"
)

const mimeApplicationJSON = "application/json"

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

type PageCreateHandler struct {
	store      PageStore
	beforeSave func(ctx context.Context, page *Page) error
}

func NewPageCreateHandler(store PageStore, beforeSave func(ctx context.Context, page *Page) error) *PageCreateHandler {
	if store == nil {
		panic("page create handler: store is required")
	}
	if beforeSave == nil {
		beforeSave = func(ctx context.Context, page *Page) error { return nil }
	}

	return &PageCreateHandler{
		store:      store,
		beforeSave: beforeSave,
	}
}

func (h *PageCreateHandler) Handle(w http.ResponseWriter, r *http.Request) error {
	c := FromContext(r.Context())
	if c == nil {
		panic("page create handler: context is required")
	}

	if !c.HasSite() {
		return fmt.Errorf("page create handler: %w", ErrSiteNotFound)
	}

	var dto PageCreateRequest
	if strings.Contains(r.Header.Get(headerContentType), mimeApplicationJSON) {
		if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
			return fmt.Errorf("page create handler: json decode: %w", err)
		}
	} else {
		dto.Template = r.FormValue("template")
		dto.Title = r.FormValue("title")
		dto.URL = r.FormValue("url")
	}

	dto.URL = strings.TrimRight(dto.URL, "/")
	if len(dto.URL) == 0 || dto.URL[0] != '/' {
		dto.URL = "/" + dto.URL
	}

	if err := dto.Validate(); err != nil {
		return fmt.Errorf("page create handler: validate: %w", err)
	}

	site := c.Site()

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
			if p, err := h.store.FindByURL(r.Context(), site.ID, dto.URL[:index]); err == nil {
				page.ParentID = &p.ID
				page.Parent = p
			}
		}
	}

	page.FixURL()

	if err := h.beforeSave(r.Context(), page); err != nil {
		return fmt.Errorf("page create handler: before save: %w", err)
	}

	if err := h.store.Save(r.Context(), page); err != nil {
		return fmt.Errorf("page create handler: save: %w", err)
	}

	http.Redirect(w, r, page.AbsURL(), http.StatusFound)

	return nil
}
