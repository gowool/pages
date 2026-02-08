package pages

import (
	"embed"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gowool/keratin"
	"github.com/gowool/keratin/middleware"
)

var _ ErrorHandler = (*HTTPErrorHandler)(nil)

//go:embed error.gohtml
var ErrorTemplateFS embed.FS

type HTTPErrorHandlerConfig struct {
	// FallbackTemplate is a template name for fallback error page.
	// The fallback template is used when the error page is not found
	// by pattern, also the Site and Page variables could not be provided.
	FallbackTemplate string

	// StatusFunc returns an error code code.
	StatusFunc middleware.ErrorStatusFunc

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

	w.Header().Set(keratin.HeaderContentType, keratin.MIMEApplicationJSON)
	w.WriteHeader(c.Status())

	data := map[string]any{
		"code":    c.Status(),
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
		var httpErr *keratin.HTTPError
		if errors.As(c.Error(), &httpErr) && httpErr.Message != "" {
			data["message"] = httpErr.Message
		}

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
	errStatusFunc    middleware.ErrorStatusFunc
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
	if committer := keratin.ResponseCommitter(w); committer != nil && committer.Committed() {
		h.logger.Warn("response is committed, skip error handler", "error", e)
		return
	}

	ctx := r.Context()
	status := h.errStatusFunc(ctx, e)

	c := MustContext(ctx)
	c.SetError(e)
	c.SetPage(nil)
	c.SetStatus(status)

	defer h.logger.Error("request failed",
		slog.Int("code", c.Status()),
		slog.String("method", r.Method),
		slog.Int("status_code", status),
		slog.String("path", r.URL.Path),
		slog.Bool("debug", c.Debug()),
		slog.Bool("guest", c.Guest()),
		slog.Any("error", e),
	)

	if r.Method == http.MethodHead {
		w.WriteHeader(c.Status())
		return
	}

	if strings.Contains(r.Header.Get(keratin.HeaderAccept), keratin.MIMEApplicationJSON) {
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

func (h *HTTPErrorHandler) serveHTTP(w http.ResponseWriter, r *http.Request) {
	c := MustContext(r.Context())
	if !c.HasTemplate() {
		c.SetTemplate(h.fallbackTemplate)
	}

	if err := h.pageHandler.ServeHTTP(w, r); err != nil {
		h.logger.Error("page handler error", "error", err)
	}
}
