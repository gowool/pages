package pages

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gowool/wo"
	"github.com/gowool/wo/middleware"
)

var (
	ErrSiteNotFound    = errors.New("site not found")
	ErrPageNotFound    = errors.New("page not found")
	ErrPrivatePage     = errors.New("page is private")
	ErrUniqueViolation = errors.New("unique violation")
)

type ErrorPatternFinderFunc func(ctx context.Context, status int) (string, error)

func ErrorPatternFinder() ErrorPatternFinderFunc {
	return func(_ context.Context, status int) (string, error) {
		switch status {
		case http.StatusUnauthorized:
			return PageErrorUnauthorized, nil
		case http.StatusForbidden:
			return PageErrorForbidden, nil
		case http.StatusNotFound:
			return PageErrorNotFound, nil
		default:
			if status >= 400 && status < 500 {
				return PageError4xx, nil
			}
			return PageError5xx, nil
		}
	}
}

func ErrorMapper(err error) *wo.HTTPError {
	if he := wo.AsHTTPError(err); he != nil {
		return he
	}

	if errors.Is(err, ErrSiteNotFound) {
		return wo.ErrInternalServerError.WithInternal(err)
	}

	if errors.Is(err, ErrPageNotFound) {
		return wo.ErrNotFound.WithInternal(err)
	}

	if errors.Is(err, ErrPrivatePage) {
		return wo.ErrForbidden.WithInternal(err)
	}

	if errors.Is(err, ErrUniqueViolation) {
		return wo.ErrConflict.WithInternal(err)
	}

	return nil
}

func ErrorRenderer[T Resolver](
	handler func(e T) error,
	manager PageManager,
	strategy PageDecoratorStrategy,
	authorizer PageAuthorizer[T],
	patternFinder ErrorPatternFinderFunc,
	logger *slog.Logger,
	skippers ...middleware.Skipper[T],
) func(T, *wo.HTTPError) {
	if handler == nil {
		panic("error renderer: page handler is required")
	}

	if manager == nil {
		panic("error renderer: page manager is required")
	}

	if strategy == nil {
		strategy = DefaultPageDecoratorStrategy
	}

	if authorizer == nil {
		authorizer = DenyPageAuthorizer[T]{}
	}

	if patternFinder == nil {
		patternFinder = ErrorPatternFinder()
	}

	if logger == nil {
		logger = slog.New(slog.DiscardHandler)
	}

	skip := middleware.ChainSkipper[T](skippers...)

	return func(e T, httpErr *wo.HTTPError) {
		if !strings.Contains(e.Request().Header.Get(wo.HeaderAccept), wo.MIMETextHTML) {
			return
		}

		if !e.HasSite() {
			return
		}

		if skip(e) {
			return
		}

		var (
			pattern string
			err     error
		)

		req := e.Request()
		ctx := req.Context()

		if httpErr.Status == http.StatusNotFound {
			if ok, err := strategy.IsDecorable(ctx, req.Pattern, req.URL.Path); ok && err == nil {
				if d, err := authorizer.Authorize(e, CreatePage); d == Allow && err == nil {
					pattern = PageInternalCreate
				}
			}
		}

		if pattern == "" {
			if pattern, _ = patternFinder(ctx, httpErr.Status); pattern == "" {
				pattern = PageError5xx
			}
		}

		page, err := manager.GetByPattern(ctx, e.Site(), pattern)
		if err != nil {
			logger.Error("error renderer: find error page", "error", err, "pattern", pattern)
			return
		}

		e.SetStatus(httpErr.Status)
		e.SetError(httpErr)
		e.SetPage(page)

		if err = handler(e); err != nil {
			logger.Error("error renderer: html render return error", "error", err, "pattern", pattern, "pageID", page.ID)
		}
	}
}
