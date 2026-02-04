package pages

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gowool/pages/internal"
	"github.com/invopop/validation"
)

const (
	headerContentType       = "Content-Type"
	mimeApplicationJSON     = "application/json"
	mimeTextHTML            = "text/html"
	mimeTextHTMLCharsetUTF8 = mimeTextHTML + "; charset=UTF-8"
	errorTemplate           = `<!DOCTYPE html>
<html dir="ltr" lang="en">
<head>
	<meta charset="utf-8" />
	<style type="text/css">
		h1 {
		  font-size: 15vmin;
		  margin-bottom: 0;
		}
		h2 {
		  font-size: 5vmin;
		  margin-top: 0;
		  margin-bottom: 40px;
		}
		
		body {
		  height: 100vh;
		  display: flex;
		  flex-direction: column;
		  background-color: white;
		  align-items: center;
		  justify-content: center;
		  overflow: hidden;
		}
	</style>
	<title>{{.Status}} - {{.Title}}</title>
</head>
<body>
	<h1>{{.Title}}!</h1>
	<h2>Code {{.Status}}</h2>
</body>
</html>`
)

// Theme defines an interface for rendering templates with a specific context and writing the output to an io.Writer.
type Theme interface {
	Write(ctx context.Context, w io.Writer, template string, data any) error
}

// PageCtxFunc return a page context for template.
type PageCtxFunc func(*http.Request, *Context) any

func PageHandler(theme Theme, pageCtx PageCtxFunc, logger *slog.Logger) Handler {
	if theme == nil {
		panic("page handler: theme is required")
	}

	if pageCtx == nil {
		pageCtx = PageCtx()
	}

	if logger == nil {
		logger = slog.New(slog.DiscardHandler)
	}

	logger = logger.WithGroup("page_handler")

	return HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		c := FromContext(r.Context())
		if c == nil {
			panic("page handler: cms: context is required")
		}

		if !c.HasSite() {
			return fmt.Errorf("page handler: cms: %w", ErrSiteNotFound)
		}

		if !c.HasPage() {
			return fmt.Errorf("page handler: cms: %w", ErrPageNotFound)
		}

		page := c.Page()

		for key, values := range page.Header {
			for i, value := range values {
				if i == 0 {
					w.Header().Set(key, value)
				} else {
					w.Header().Add(key, value)
				}
			}
		}

		if page.Template == "" {
			w.WriteHeader(c.Status())
			return nil
		}

		var buf bytes.Buffer
		if err := theme.Write(r.Context(), &buf, page.Template, pageCtx(r, c)); err != nil {
			return fmt.Errorf("page handler: cms: theme write error: %w", err)
		}

		ct := w.Header().Get(headerContentType)
		if ct == "" {
			ct = mimeTextHTMLCharsetUTF8
		}

		w.Header().Set(headerContentType, ct)
		w.WriteHeader(c.Status())

		if _, err := w.Write(buf.Bytes()); err != nil {
			logger.Error("write response error", "error", err)
		}

		return nil
	})
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

// BeforeSaveFunc called before page save.
type BeforeSaveFunc func(context.Context, *Page) error

func PageCreateHandler(store PageStore, authorizer PageAuthorizer, beforeSave BeforeSaveFunc) Handler {
	if store == nil {
		panic("page create handler: store is required")
	}
	if authorizer == nil {
		panic("page create handler: authorizer is required")
	}

	if beforeSave == nil {
		beforeSave = func(ctx context.Context, page *Page) error { return nil }
	}

	return HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		if r.Method != http.MethodPost {
			return fmt.Errorf("page create handler: create: %w", ErrPageNotFound)
		}

		ctx := r.Context()
		c := FromContext(ctx)
		if c == nil {
			panic("page create handler: create: context is required")
		}

		if !c.HasSite() {
			return fmt.Errorf("page create handler: create: %w", ErrSiteNotFound)
		}

		if c.Guest() {
			return fmt.Errorf("page create handler: create: %w", ErrPageUnauthorized)
		}

		if authorizer.Authorize(ctx, CreatePage) == Deny {
			return fmt.Errorf("page create handler: create: %w", ErrPageForbidden)
		}

		var dto PageCreateRequest
		if strings.Contains(r.Header.Get(headerContentType), mimeApplicationJSON) {
			if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
				return fmt.Errorf("page create handler: create: json decode: %w", err)
			}
		} else {
			if err := r.ParseForm(); err != nil {
				return fmt.Errorf("page create handler: create: parse form: %w", err)
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
			return fmt.Errorf("page create handler: create: validate: %w", err)
		}

		site := c.Site()

		page := NewPage()
		page.Pattern = PageCMS
		page.Name = internal.ToTitle(dto.URL)
		page.Site = site
		page.SiteID = site.ID
		page.Title = dto.Title
		page.CustomURL = dto.URL
		page.Template = dto.Template
		page.Decorate = true

		if dto.URL != "/" {
			if index := strings.LastIndex(dto.URL, "/"); index > 0 {
				if p, err := store.FindByURL(ctx, site.ID, dto.URL[:index]); err == nil {
					page.ParentID = &p.ID
					page.Parent = p
				}
			}
		}

		page.FixURL()

		if err := beforeSave(ctx, page); err != nil {
			return fmt.Errorf("page create handler: create: before save: %w", err)
		}

		if err := store.Save(ctx, page); err != nil {
			return fmt.Errorf("page create handler: create: save: %w", err)
		}

		http.Redirect(w, r, page.AbsURL(), http.StatusFound)

		return nil
	})
}

var ErrorTemplate = template.Must(template.New("error_template").Parse(errorTemplate))

// ErrorStatusFunc return an error status code.
type ErrorStatusFunc func(context.Context, error) int

// ErrorPatternFunc return an error page pattern.
type ErrorPatternFunc func(context.Context, int) string

func PageErrorHandler(
	pageHandler Handler,
	manager PageManager,
	authorizer PageAuthorizer,
	strategy PageDecoratorStrategy,
	errorStatusFunc ErrorStatusFunc,
	errorPatternFunc ErrorPatternFunc,
	logger *slog.Logger,
) ErrorHandler {
	if pageHandler == nil {
		panic("page error handler: page handler is required")
	}
	if manager == nil {
		panic("page error handler: manager is required")
	}
	if authorizer == nil {
		panic("page error handler: authorizer is required")
	}
	if strategy == nil {
		panic("page error handler: decorator strategy is required")
	}
	if errorStatusFunc == nil {
		panic("page error handler: error status finder is required")
	}

	if errorPatternFunc == nil {
		errorPatternFunc = ErrorPattern()
	}

	if logger == nil {
		logger = slog.New(slog.DiscardHandler)
	}

	logger = logger.WithGroup("page_error_handler")

	errorPattern := func(r *http.Request, status int) string {
		ctx := r.Context()

		if status == http.StatusNotFound && strategy.IsURIDecorable(ctx, r.URL.Path) && authorizer.Authorize(ctx, CreatePage) == Allow {
			return PageInternalCreate
		}

		if pattern := errorPatternFunc(ctx, status); pattern != "" {
			return pattern
		}

		return PageError5xx
	}

	fallback := func(ctx context.Context, w http.ResponseWriter, status int, e error) {
		data := map[string]any{
			"Title":   http.StatusText(status),
			"Context": ctx,
			"Status":  status,
			"Error":   e,
		}

		w.Header().Set(headerContentType, mimeTextHTMLCharsetUTF8)
		w.WriteHeader(status)

		if err := ErrorTemplate.Execute(w, data); err != nil {
			logger.Error("write response error", "error", err, "data", data)
		}
	}

	return ErrorHandlerFunc(func(w http.ResponseWriter, r *http.Request, e error) {
		ctx := r.Context()
		c := FromContext(ctx)
		if c == nil {
			panic("page error handler: context is required")
		}

		if !c.HasSite() {
			fallback(ctx, w, http.StatusInternalServerError, e)
			return
		}

		status := errorStatusFunc(ctx, e)
		if status < http.StatusBadRequest {
			status = http.StatusInternalServerError
		}

		pattern := errorPattern(r, status)

		page, err := manager.GetByPattern(ctx, c.Site(), pattern)
		if err != nil {
			logger.Error("find page by pattern return error", "error", err, "pattern", pattern)
			fallback(ctx, w, status, e)
			return
		}

		c.SetError(e)
		c.SetStatus(status)
		c.SetPage(page)

		if err = pageHandler.ServeHTTP(w, r); err != nil {
			logger.Error("page handler return error", "error", err, "pattern", pattern)
			fallback(ctx, w, status, e)
		}
	})
}
