package pages

import (
	"context"
	"embed"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
)

var _ ErrorHandler = (*HTTPErrorHandler)(nil)

//go:embed error.gohtml
var ErrorTemplateFS embed.FS

// ErrorStatusFunc return an error status code.
type ErrorStatusFunc func(context.Context, error) int

type HTTPErrorHandlerConfig struct {
	// FallbackTemplate is a template name for fallback error page.
	// The fallback template is used when the error page is not found
	// by pattern, also the Site and Page variables could not be provided.
	FallbackTemplate string

	// StatusFunc returns an error status code.
	StatusFunc ErrorStatusFunc

	// JSONHandler is a handler for JSON error responses.
	JSONHandler http.Handler

	// Logger is used for logging.
	Logger *slog.Logger
}

func (cfg *HTTPErrorHandlerConfig) SetDefaults() {
	if cfg.FallbackTemplate == "" {
		cfg.FallbackTemplate = "error.gohtml"
	}

	if cfg.StatusFunc == nil {
		cfg.StatusFunc = ErrorStatus
	}

	if cfg.JSONHandler == nil {
		cfg.JSONHandler = http.HandlerFunc(cfg.jsonHandler)
	}

	if cfg.Logger == nil {
		cfg.Logger = slog.New(slog.DiscardHandler)
	}

	cfg.Logger = cfg.Logger.WithGroup("http_error_handler")
}

func (cfg *HTTPErrorHandlerConfig) jsonHandler(w http.ResponseWriter, r *http.Request) {
	c := MustContext(r.Context())

	w.Header().Set(HeaderContentType, MIMEApplicationJSON)
	w.WriteHeader(c.Status())

	data := map[string]any{
		"status":  c.Status(),
		"message": http.StatusText(c.Status()),
	}

	if c.HasSite() {
		data["site"] = map[string]any{
			"id":   c.Site().ID,
			"name": c.Site().Name,
			"url":  c.Site().Home(),
		}
	}

	if c.HasError() {
		if c.Status() == http.StatusUnprocessableEntity {
			data["data"] = c.Error()
		} else if c.Debug() {
			data["error"] = c.Error().Error()
		}
	}

	if err := json.NewEncoder(w).Encode(data); err != nil {
		cfg.Logger.Error("write response error", "error", err, "data", data)
	}
}

type HTTPErrorHandler struct {
	fallbackTemplate string
	jsonHandler      http.Handler
	pageHandler      Handler
	manager          PageManager
	errPattern       ErrorPattern
	errStatusFunc    ErrorStatusFunc
	logger           *slog.Logger
}

func NewHTTPErrorHandler(pageHandler Handler, manager PageManager, errPattern ErrorPattern) *HTTPErrorHandler {
	return NewHTTPErrorHandlerWithConfig(pageHandler, manager, errPattern, HTTPErrorHandlerConfig{})
}

func NewHTTPErrorHandlerWithConfig(pageHandler Handler, manager PageManager, errPattern ErrorPattern, config HTTPErrorHandlerConfig) *HTTPErrorHandler {
	if pageHandler == nil {
		panic("http error handler: page handler is required")
	}
	if manager == nil {
		panic("http error handler: page manager is required")
	}
	if errPattern == nil {
		panic("http error handler: error pattern is required")
	}
	config.SetDefaults()

	return &HTTPErrorHandler{
		fallbackTemplate: config.FallbackTemplate,
		jsonHandler:      config.JSONHandler,
		pageHandler:      pageHandler,
		manager:          manager,
		errPattern:       errPattern,
		errStatusFunc:    config.StatusFunc,
		logger:           config.Logger,
	}
}

func (h *HTTPErrorHandler) ServeHTTP(w http.ResponseWriter, r *http.Request, e error) {
	ctx := r.Context()
	status := h.errStatusFunc(ctx, e)

	if h.redirect(w, r, status, e) {
		return
	}

	c := MustContext(ctx)
	c.SetError(e)
	c.SetPage(nil)
	c.SetStatus(status)

	defer h.logger.Error("request failed",
		slog.Int("status", c.Status()),
		slog.String("method", r.Method),
		slog.String("protocol", r.Proto),
		slog.String("host", r.Host),
		slog.String("pattern", r.Pattern),
		slog.String("uri", r.RequestURI),
		slog.String("path", r.URL.Path),
		slog.String("remote_addr", r.RemoteAddr),
		slog.String("referer", r.Referer()),
		slog.String("user_agent", r.UserAgent()),
		slog.Bool("debug", c.Debug()),
		slog.Bool("guest", c.Guest()),
		slog.Any("error", e),
	)

	if r.Method == http.MethodHead {
		w.WriteHeader(c.Status())
		return
	}

	if strings.Contains(r.Header.Get(HeaderAccept), MIMEApplicationJSON) {
		h.jsonHandler.ServeHTTP(w, r)
		return
	}

	if !c.HasSite() {
		h.logger.Error("no site found in context", "error", e)

		h.serveHTTP(w, r)
		return
	}

	pattern := h.errPattern.Pattern(r, c.Status(), e)
	if pattern == "" {
		pattern = PageError5xx
	}

	page, err := h.manager.GetByPattern(ctx, c.Site(), pattern)
	if err != nil {
		h.logger.Error("find page by pattern return error", "error", err, "pattern", pattern)

		h.serveHTTP(w, r)
		return
	}

	c.SetPage(page)

	h.serveHTTP(w, r)
}

func (h *HTTPErrorHandler) redirect(w http.ResponseWriter, r *http.Request, status int, err error) bool {
	switch status {
	case http.StatusMovedPermanently, http.StatusFound, http.StatusTemporaryRedirect, http.StatusPermanentRedirect:
		var uri string
	Loop:
		for {
			switch t := err.(type) {
			case interface{ Redirect() (string, int) }:
				uri, _ = t.Redirect()
				break Loop
			case interface{ URI() string }:
				uri = t.URI()
				break Loop
			case interface{ URL() string }:
				uri = t.URL()
				break Loop
			case interface{ Unwrap() error }:
				err = t.Unwrap()
				continue
			default:
				break Loop
			}
		}
		if uri == "" {
			uri = "/"
		}

		http.Redirect(w, r, uri, status)

		return true
	default:
		return false
	}
}

func (h *HTTPErrorHandler) serveHTTP(w http.ResponseWriter, r *http.Request) {
	c := MustContext(r.Context())
	if !c.HasTemplate() {
		c.SetTemplate(h.fallbackTemplate)
	}

	if err := h.pageHandler.ServeHTTP(w, r); err != nil {
		h.logger.Error("page handler error", "error", err)
	}
}
