package pages

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"sync"

	"github.com/gowool/keratin"
	"github.com/gowool/keratin/middleware"
	"github.com/gowool/pages/internal"
)

func SelectSiteMiddleware(retriever SiteRetriever, skippers ...middleware.Skipper) func(keratin.Handler) keratin.Handler {
	if retriever == nil {
		panic("middleware: select site: retriever is required")
	}

	skip := middleware.ChainSkipper(skippers...)

	return func(next keratin.Handler) keratin.Handler {
		return keratin.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
			if skip(r) {
				return next.ServeHTTP(w, r)
			}

			c := MustContext(r.Context())

			site, pathInfo, err := retriever.Retrieve(r)
			if err != nil {
				if !errors.Is(err, ErrSiteNotFound) {
					err = errors.Join(err, ErrSiteNotFound)
				}
				return fmt.Errorf("middleware: select site: %w", err)
			}

			if site == nil {
				return ErrSiteNotFound
			}

			site.Scheme = keratin.FromContext(r.Context()).Scheme()
			site.Host = r.Host
			site.IsRoot = r.URL.Path == "/"

			c.SetSite(site)

			if pathInfo != "" {
				r.URL.Path = pathInfo
			}

			return next.ServeHTTP(w, r)
		})
	}
}

func SelectPageMiddleware(manager PageManager, authorizer PageAuthorizer, skippers ...middleware.Skipper) func(keratin.Handler) keratin.Handler {
	if manager == nil {
		panic("middleware: select page: manager is required")
	}
	if authorizer == nil {
		panic("middleware: select page: authorizer is required")
	}

	skip := middleware.ChainSkipper(skippers...)

	findPage := func(r *http.Request, site *Site) (page *Page, err error) {
		if pattern := keratin.Pattern(r); pattern == PageCMSPattern {
			page, err = manager.GetByURL(r.Context(), site, r.URL.Path)
		} else {
			page, err = manager.GetByPattern(r.Context(), site, pattern)
		}

		if err != nil {
			if !errors.Is(err, ErrPageNotFound) {
				err = errors.Join(err, ErrPageNotFound)
			}
			return
		} else if page == nil {
			err = ErrPageNotFound
		}
		return
	}

	allow := func(ctx context.Context, action PageAction, isGuest bool) error {
		if isGuest {
			return errors.New("guest is not allowed")
		}

		if authorizer.Authorize(ctx, action) == Deny {
			return errors.New("access denied")
		}

		return nil
	}

	return func(next keratin.Handler) keratin.Handler {
		return keratin.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
			if skip(r) {
				return next.ServeHTTP(w, r)
			}

			ctx := r.Context()
			c := FromContext(ctx)
			if c == nil {
				panic("middleware: select page: context is required")
			}

			if !c.HasSite() {
				return fmt.Errorf("middleware: select page: %w", ErrSiteNotFound)
			}

			page, err := findPage(r, c.Site())
			if err != nil {
				return fmt.Errorf("middleware: select page: %w", err)
			}

			c.SetPage(page)

			if page.Status == Draft {
				if err := allow(ctx, ViewDraftPage, c.Guest()); err != nil {
					return fmt.Errorf("middleware: hybrid page: %w", errors.Join(err, ErrPageNotFound))
				}
			}

			if page.Visibility == Private {
				if err := allow(ctx, ViewPrivatePage, c.Guest()); err != nil {
					return fmt.Errorf("middleware: hybrid page: %w", errors.Join(err, ErrPageForbidden))
				}
			}

			return next.ServeHTTP(w, r)
		})
	}
}

func HybridPageMiddleware(pageHandler keratin.Handler, logger *slog.Logger, skippers ...middleware.Skipper) func(keratin.Handler) keratin.Handler {
	if pageHandler == nil {
		panic("middleware: hybrid page: page handler is required")
	}

	if logger == nil {
		logger = slog.Default()
	}

	logger = logger.WithGroup("hybrid_page")

	skippers = append(skippers, func(r *http.Request) bool {
		c := MustContext(r.Context())
		return c.HasPage() && !c.Page().IsHybrid()
	})

	skip := middleware.ChainSkipper(skippers...)

	pool := &sync.Pool{
		New: func() any {
			return new(delayedWriter)
		},
	}

	return func(next keratin.Handler) keratin.Handler {
		return keratin.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
			if skip(r) {
				return next.ServeHTTP(w, r)
			}

			c := MustContext(r.Context())
			if !c.HasPage() {
				return fmt.Errorf("middleware: hybrid page: %w", ErrPageNotFound)
			}

			if c.Page().Decorate {
				w.Header().Set(HeaderXPageDecorable, "1")
			} else {
				w.Header().Set(HeaderXPageNotDecorable, "1")
			}

			response := pool.Get().(*delayedWriter)
			response.reset(w)
			defer func() {
				response.reset(nil)
				pool.Put(response)
			}()

			if err := next.ServeHTTP(response, r); err != nil {
				return fmt.Errorf("middleware: hybrid page: %w", err)
			}

			if response.code > 0 {
				c.SetStatus(response.code)
			}

			buffer := response.buffer.Bytes()
			c.SetContent(template.HTML(internal.BytesToString(buffer)))

			// Skip decoration when the response should be returned as-is.
			if !IsDecorable(response, r) {
				w.WriteHeader(c.Status())

				if len(buffer) > 0 {
					if _, err := w.Write(buffer); err != nil {
						logger.ErrorContext(r.Context(), "write response error", "error", err)
					}
				}

				return nil
			}

			return pageHandler.ServeHTTP(w, r)
		})
	}
}

type delayedWriter struct {
	http.ResponseWriter
	buffer    bytes.Buffer
	committed bool
	code      int
}

func (w *delayedWriter) reset(rw http.ResponseWriter) {
	w.ResponseWriter = rw
	w.buffer.Reset()
	w.committed = false
	w.code = http.StatusOK
}

func (w *delayedWriter) StatusCode() int {
	return w.code
}

func (w *delayedWriter) WriteHeader(statusCode int) {
	w.code = statusCode
	w.committed = true
}

func (w *delayedWriter) Write(data []byte) (int, error) {
	if !w.committed {
		w.WriteHeader(http.StatusOK)
	}
	return w.buffer.Write(data)
}

func (w *delayedWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}
