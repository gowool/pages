package pages

import (
	"errors"
	"html/template"
	"net/http"
	"unsafe"

	"github.com/gowool/wo"
	"github.com/gowool/wo/middleware"
)

func PageMiddleware[T Resolver](
	handler func(e T) error,
	selector PageSelector,
	strategy PageDecoratorStrategy,
	authorizer PageAuthorizer[T],
	skippers ...middleware.Skipper[T],
) func(T) error {
	if handler == nil {
		panic("page middleware: page handler is required")
	}

	if selector == nil {
		panic("page middleware: selector is required")
	}

	if strategy == nil {
		strategy = DefaultPageDecoratorStrategy
	}

	if authorizer == nil {
		authorizer = DenyPageAuthorizer[T]{}
	}

	skippers = append(skippers, func(e T) bool {
		ok, _ := strategy.IsDecorable(e.Request().Context(), e.Request().Pattern, e.Request().URL.Path)
		return !ok
	})

	skip := middleware.ChainSkipper[T](skippers...)

	return func(e T) error {
		if skip(e) {
			return e.Next()
		}

		if !e.HasSite() {
			return ErrSiteNotFound
		}

		page, err := selector.Retrieve(e.Request(), e.Site())
		if err != nil {
			if errors.Is(err, ErrPageNotFound) {
				return wo.ErrNotFound.WithInternal(err)
			}
			return wo.ErrNotFound.WithInternal(errors.Join(err, ErrPageNotFound))
		}

		if page == nil {
			return wo.ErrNotFound.WithInternal(ErrPageNotFound)
		}

		if page.Site == nil {
			page.Site = e.Site()
		}

		e.SetPage(page)

		if page.Status == Draft {
			decision, err := authorizer.Authorize(e, ViewDraftPage)
			if err != nil {
				return wo.ErrNotFound.WithInternal(err)
			}

			if decision == Deny {
				return wo.ErrNotFound.WithInternal(ErrPageNotFound)
			}
		}

		if page.Visibility == Private {
			if e.IsGuest() {
				return wo.ErrUnauthorized.WithInternal(ErrPrivatePage)
			}

			decision, err := authorizer.Authorize(e, ViewPrivatePage)
			if err != nil {
				return wo.ErrForbidden.WithInternal(err)
			}

			if decision == Deny {
				return wo.ErrForbidden.WithInternal(ErrPrivatePage)
			}
		}

		if !page.IsHybrid() {
			return e.Next()
		}

		if page.Decorate {
			e.Response().Header().Set(HeaderXPageDecorable, "1")
		} else {
			e.Response().Header().Set(HeaderXPageNotDecorable, "1")
		}

		e.Response().Buffering = true

		err = e.Next()

		e.Response().Written = false
		e.Response().Buffering = false

		if err != nil {
			return err
		}

		decorable := e.IsDecorable()
		buffer := e.Response().Buffer()

		httpStatus := e.Response().Status
		if httpStatus > 0 {
			e.SetStatus(httpStatus)
		} else {
			httpStatus = e.Status()
		}

		e.Response().Status = 0

		if !decorable {
			if len(buffer) > 0 {
				e.Response().WriteHeader(httpStatus)
				_, err = e.Response().Write(buffer)
				return err
			}

			return e.NoContent(http.StatusNoContent)
		}

		if l := len(buffer); l > 0 {
			e.SetContent(template.HTML(unsafe.String(unsafe.SliceData(buffer), l)))
		}

		return handler(e)
	}
}
