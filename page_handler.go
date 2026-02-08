package pages

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/gowool/gor"
)

var _ Handler = (*PageHandler)(nil)

// Theme defines an interface for rendering templates with a specific context and writing the output to an io.Writer.
type Theme interface {
	Write(ctx context.Context, w io.Writer, template string, data any) error
}

// TemplateVarsFunc return a template variables.
type TemplateVarsFunc func(*http.Request, *Context) any

type TemplateVars struct {
	*Context
	Request *http.Request
}

func (c TemplateVars) Value(key any) any {
	return c.Request.Context().Value(key)
}

type PageHandlerConfig struct {
	// VarsFunc defines a function that generates
	// template variables from an HTTP request and context.
	VarsFunc TemplateVarsFunc

	// Logger is used for logging.
	Logger *slog.Logger
}

func (cfg *PageHandlerConfig) SetDefaults() {
	if cfg.VarsFunc == nil {
		cfg.VarsFunc = cfg.vars
	}

	if cfg.Logger == nil {
		cfg.Logger = slog.New(slog.DiscardHandler)
	}

	cfg.Logger = cfg.Logger.WithGroup("page_handler")
}

func (*PageHandlerConfig) vars(r *http.Request, c *Context) any {
	return TemplateVars{
		Context: c,
		Request: r,
	}
}

// PageHandler renders pages using a Theme.
type PageHandler struct {
	theme    Theme
	varsFunc TemplateVarsFunc
	logger   *slog.Logger
}

func NewPageHandler(theme Theme) *PageHandler {
	return NewPageHandlerWithConfig(theme, PageHandlerConfig{})
}

func NewPageHandlerWithConfig(theme Theme, config PageHandlerConfig) *PageHandler {
	if theme == nil {
		panic("page handler: theme is required")
	}
	config.SetDefaults()

	return &PageHandler{
		theme:    theme,
		varsFunc: config.VarsFunc,
		logger:   config.Logger,
	}
}

func (h *PageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) error {
	c := MustContext(r.Context())

	if !c.HasError() {
		if !c.HasSite() {
			return fmt.Errorf("page handler: %w", ErrSiteNotFound)
		}

		if !c.HasPage() {
			return fmt.Errorf("page handler: %w", ErrPageNotFound)
		}
	}

	if c.HasPage() {
		for key, values := range c.Page().Header {
			for i, value := range values {
				if i == 0 {
					w.Header().Set(key, value)
				} else {
					w.Header().Add(key, value)
				}
			}
		}
	}

	if !c.HasTemplate() {
		if !c.HasError() {
			return fmt.Errorf("page handler: %w", ErrTemplateEmpty)
		}

		h.logger.Error("template empty", "context_error", c.Error())

		http.Error(w, http.StatusText(c.Status()), c.Status())
		return nil
	}

	var buf bytes.Buffer
	if err := h.theme.Write(r.Context(), &buf, c.Template(), h.varsFunc(r, c)); err != nil {
		if !c.HasError() {
			return fmt.Errorf("page handler: theme write error: %w", err)
		}

		h.logger.Error("theme write error", "error", err, "context_error", c.Error())

		http.Error(w, http.StatusText(c.Status()), c.Status())
		return nil
	}

	ct := w.Header().Get(gor.HeaderContentType)
	if ct == "" {
		ct = gor.MIMETextHTMLCharsetUTF8
	}

	w.Header().Set(gor.HeaderContentType, ct)
	w.WriteHeader(c.Status())

	if _, err := w.Write(buf.Bytes()); err != nil {
		h.logger.Error("write response error", "error", err)
	}

	return nil
}
