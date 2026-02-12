package pages

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gowool/keratin"
	"github.com/invopop/validation"
)

var _ keratin.Handler = (*PageCreateHandler)(nil)

// BeforeSaveFunc called before page save.
type BeforeSaveFunc func(context.Context, *Page) error

type PageCreateHandlerConfig struct {
	GeneratorFunc  IDGeneratorFunc
	BeforeSaveFunc BeforeSaveFunc
}

func (c *PageCreateHandlerConfig) SetDefaults() {
	if c.GeneratorFunc == nil {
		c.GeneratorFunc = IDGenerator()
	}

	if c.BeforeSaveFunc == nil {
		c.BeforeSaveFunc = c.beforeSave
	}
}

func (c *PageCreateHandlerConfig) beforeSave(context.Context, *Page) error {
	return nil
}

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
	store          PageStore
	authorizer     PageAuthorizer
	generatorFunc  IDGeneratorFunc
	beforeSaveFunc BeforeSaveFunc
}

func NewPageCreateHandler(store PageStore, authorizer PageAuthorizer) *PageCreateHandler {
	return NewPageCreateHandlerWithConfig(store, authorizer, PageCreateHandlerConfig{})
}

func NewPageCreateHandlerWithConfig(store PageStore, authorizer PageAuthorizer, cfg PageCreateHandlerConfig) *PageCreateHandler {
	if store == nil {
		panic("page create handler: store is required")
	}
	if authorizer == nil {
		panic("page create handler: authorizer is required")
	}
	cfg.SetDefaults()

	return &PageCreateHandler{
		store:          store,
		authorizer:     authorizer,
		generatorFunc:  cfg.GeneratorFunc,
		beforeSaveFunc: cfg.BeforeSaveFunc,
	}
}

func (h *PageCreateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodPost {
		return fmt.Errorf("page create handler: %w", ErrPageNotFound)
	}

	ctx := r.Context()
	c := MustContext(ctx)

	if !c.HasSite() {
		return fmt.Errorf("page create handler: %w", ErrSiteNotFound)
	}

	if c.Guest() {
		return fmt.Errorf("page create handler: %w", ErrPageUnauthorized)
	}

	if h.authorizer.Authorize(ctx, CreatePage) == Deny {
		return fmt.Errorf("page create handler: %w", ErrPageForbidden)
	}

	var dto PageCreateRequest
	if strings.Contains(r.Header.Get(keratin.HeaderContentType), keratin.MIMEApplicationJSON) {
		if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
			return fmt.Errorf("page create handler: json decode: %w", err)
		}
	} else {
		if err := r.ParseForm(); err != nil {
			return fmt.Errorf("page create handler: parse form: %w", err)
		}

		dto.Template = r.PostFormValue("template")
		dto.Title = r.PostFormValue("title")
		dto.URL = r.PostFormValue("url")
	}

	dto.URL = strings.TrimRight(dto.URL, "/")
	if len(dto.URL) == 0 || dto.URL[0] != '/' {
		dto.URL = "/" + dto.URL
	}

	if err := dto.Validate(); err != nil {
		return fmt.Errorf("page create handler: validate: %w", err)
	}

	pageID, err := h.generatorFunc(ctx)
	if err != nil {
		return fmt.Errorf("page create handler: generate id: %w", err)
	}

	site := c.Site()

	page := NewPage()
	page.ID = pageID
	page.Pattern = PageCMS
	page.Name = site.Name + ": " + dto.URL
	page.Site = site
	page.SiteID = site.ID
	page.Title = dto.Title
	page.CustomURL = dto.URL
	page.Template = dto.Template
	page.Decorate = true

	if dto.URL != "/" {
		if index := strings.LastIndex(dto.URL, "/"); index > 0 {
			if p, err := h.store.FindByURL(ctx, site.ID, dto.URL[:index]); err == nil {
				page.ParentID = &p.ID
				page.Parent = p
				page.CustomURL = dto.URL[index+1:]
			}
		}
	}

	page.FixURL()

	if err := h.beforeSaveFunc(ctx, page); err != nil {
		return fmt.Errorf("page create handler: before save: %w", err)
	}

	if err := h.store.Save(ctx, page); err != nil {
		return fmt.Errorf("page create handler: save: %w", err)
	}

	http.Redirect(w, r, page.AbsURL(), http.StatusFound)

	return nil
}
