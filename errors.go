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
	authorizer PageAuthorizer[T],
	patternFinder func(ctx context.Context, status int) (string, error),
	logger *slog.Logger,
	skippers ...middleware.Skipper[T],
) func(T, *wo.HTTPError) {
	if handler == nil {
		panic("error renderer: page handler is required")
	}

	if manager == nil {
		panic("error renderer: page manager is required")
	}

	if authorizer == nil {
		authorizer = DenyPageAuthorizer[T]{}
	}

	if patternFinder == nil {
		patternFinder = func(_ context.Context, status int) (string, error) {
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

		if httpErr.Status == http.StatusNotFound {
			if d, _ := authorizer.Authorize(e, CreatePage); d == Allow {
				pattern = PageInternalCreate
			}
		}

		if pattern == "" {
			if pattern, _ = patternFinder(e.Request().Context(), httpErr.Status); pattern == "" {
				pattern = PageError5xx
			}
		}

		page, err := manager.GetByPattern(e.Request().Context(), e.Site(), pattern)
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
