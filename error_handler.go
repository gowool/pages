package pages

import (
	"context"
	"html/template"
	"log/slog"
	"net/http"
)

var _ ErrorHandler = (*DefaultErrorHandler)(nil)

const errorTemplate = `<!DOCTYPE html>
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

var ErrorTemplate = template.Must(template.New("error_template").Parse(errorTemplate))

type (
	ErrorPatternFinderFunc func(ctx context.Context, status int) string
	ErrorStatusFinderFunc  func(ctx context.Context, err error) int
)

func ErrorPatternFinder() ErrorPatternFinderFunc {
	return func(_ context.Context, status int) string {
		switch status {
		case http.StatusUnauthorized:
			return PageErrorUnauthorized
		case http.StatusForbidden:
			return PageErrorForbidden
		case http.StatusNotFound:
			return PageErrorNotFound
		default:
			if status >= 400 && status < 500 {
				return PageError4xx
			}
			return PageError5xx
		}
	}
}

type ErrorHandler interface {
	Handle(w http.ResponseWriter, r *http.Request, err error)
}

type DefaultErrorHandler struct {
	pageHandler   PageHandler
	manager       PageManager
	authorizer    PageAuthorizer
	strategy      PageDecoratorStrategy
	statusFinder  ErrorStatusFinderFunc
	patternFinder ErrorPatternFinderFunc
	logger        *slog.Logger
}

func NewDefaultErrorHandler(
	pageHandler PageHandler,
	manager PageManager,
	authorizer PageAuthorizer,
	strategy PageDecoratorStrategy,
	statusFinder ErrorStatusFinderFunc,
	patternFinder ErrorPatternFinderFunc,
	logger *slog.Logger,
) *DefaultErrorHandler {
	if pageHandler == nil {
		panic("error handler: page handler is required")
	}
	if manager == nil {
		panic("error handler: page manager is required")
	}
	if authorizer == nil {
		panic("error handler: page authorizer is required")
	}
	if strategy == nil {
		panic("error handler: page decorator strategy is required")
	}
	if statusFinder == nil {
		panic("error handler: status finder is required")
	}
	if patternFinder == nil {
		panic("error handler: pattern finder is required")
	}
	if logger == nil {
		logger = slog.New(slog.DiscardHandler)
	}

	return &DefaultErrorHandler{
		pageHandler:   pageHandler,
		manager:       manager,
		authorizer:    authorizer,
		strategy:      strategy,
		statusFinder:  statusFinder,
		patternFinder: patternFinder,
		logger:        logger.WithGroup("error_handler"),
	}
}

func (h *DefaultErrorHandler) Handle(w http.ResponseWriter, r *http.Request, e error) {
	ctx := r.Context()
	c := FromContext(ctx)
	if c == nil {
		panic("error handler: context is required")
	}

	if !c.HasSite() {
		h.internal(ctx, w, http.StatusInternalServerError, e)
		return
	}

	status := h.statusFinder(ctx, e)
	if status < http.StatusBadRequest {
		status = http.StatusInternalServerError
	}

	pattern := h.getPattern(r, status)

	page, err := h.manager.GetByPattern(ctx, c.Site(), pattern)
	if err != nil {
		h.logger.Error("find page by pattern return error", "error", err, "pattern", pattern)
		h.internal(ctx, w, status, e)
		return
	}

	c.SetError(e)
	c.SetStatus(status)
	c.SetPage(page)

	if err = h.pageHandler.Handle(w, r); err != nil {
		h.logger.Error("page handler return error", "error", err, "pattern", pattern)
		h.internal(ctx, w, status, e)
	}
}

func (h *DefaultErrorHandler) getPattern(r *http.Request, status int) string {
	ctx := r.Context()

	if status == http.StatusNotFound {
		if ok, err := h.strategy.IsDecorable(ctx, r.Pattern, r.URL.Path); ok && err == nil {
			if decision, err := h.authorizer.Authorize(ctx, CreatePage); decision == Allow && err == nil {
				return PageInternalCreate
			}
		}
	}

	if pattern := h.patternFinder(ctx, status); pattern != "" {
		return pattern
	}

	return PageError5xx
}

func (h *DefaultErrorHandler) internal(ctx context.Context, w http.ResponseWriter, status int, e error) {
	data := map[string]any{
		"Title":   http.StatusText(status),
		"Context": ctx,
		"Status":  status,
		"Error":   e,
	}

	w.Header().Set(headerContentType, mimeTextHTMLCharsetUTF8)
	w.WriteHeader(status)

	if err := ErrorTemplate.Execute(w, data); err != nil {
		h.logger.Error("write response error", "error", err, "data", data)
	}
}
