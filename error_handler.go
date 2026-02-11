package pages

import (
	"embed"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gowool/keratin"
	"github.com/gowool/keratin/middleware"
)

//go:embed error.gohtml
var ErrorTemplateFS embed.FS

type ErrorHandlerConfig struct {
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

func (cfg *ErrorHandlerConfig) SetDefaults() {
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

func (cfg *ErrorHandlerConfig) jsonHandler(w http.ResponseWriter, r *http.Request) {
	c := MustContext(r.Context())

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
		if httpErr, ok := errors.AsType[*keratin.HTTPError](c.Error()); ok && httpErr.Message != "" {
			data["message"] = httpErr.Message
		}

		if c.Status() == http.StatusUnprocessableEntity {
			data["data"] = c.Error()
		} else if c.Debug() {
			data["error"] = c.Error().Error()
		}
	}

	if err := keratin.JSON(w, c.Status(), data); err != nil {
		cfg.Logger.Error("write response error", "error", err, "data", data)
	}
}

type ErrorPatternFunc func(r *http.Request, status int, err error) string

func ErrorHandler(cfg ErrorHandlerConfig, pageHandler keratin.Handler, manager PageManager, errPattern ErrorPatternFunc) keratin.ErrorHandlerFunc {
	if pageHandler == nil {
		panic("http error handler: page handler is required")
	}
	if manager == nil {
		panic("http error handler: page manager is required")
	}
	if errPattern == nil {
		panic("http error handler: error pattern is required")
	}

	cfg.SetDefaults()

	serveHTTP := func(w http.ResponseWriter, r *http.Request, logger *slog.Logger) {
		if err := pageHandler.ServeHTTP(w, r); err != nil {
			logger.Error("page handler error", "error", err)
		}
	}

	return func(w http.ResponseWriter, r *http.Request, e error) {
		logger := cfg.Logger

		requestID := r.Header.Get(keratin.HeaderXRequestID)
		if requestID == "" {
			requestID = w.Header().Get(keratin.HeaderXRequestID)
		}
		if requestID != "" {
			logger = logger.With("request_id", requestID)
		}

		if committed(w) {
			logger.Warn("response is committed, skip error handler", "error", e)
			return
		}

		ctx := r.Context()
		status := cfg.StatusFunc(ctx, e)

		c := MustContext(ctx)
		c.SetError(e)
		c.SetPage(nil)
		c.SetStatus(status)
		c.SetTemplate(cfg.FallbackTemplate)

		defer logger.Error("request failed",
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
			cfg.JSONHandler.ServeHTTP(w, r)

			if committed(w) {
				return
			}
		}

		if !c.HasSite() {
			logger.Error("no site found in context", "error", e)

			serveHTTP(w, r, logger)
			return
		}

		pattern := errPattern(r, c.Status(), e)
		if pattern == "" {
			pattern = PageError5xx
		}

		page, err := manager.GetByPattern(ctx, c.Site(), pattern)
		if err != nil {
			logger.Error("find page by pattern return error", "error", err, "pattern", pattern)

			serveHTTP(w, r, logger)
			return
		}

		c.SetTemplate("")
		c.SetPage(page)

		serveHTTP(w, r, logger)
	}
}

func committed(w http.ResponseWriter) bool {
	committer := keratin.ResponseCommitter(w)

	return committer != nil && committer.Committed()
}
