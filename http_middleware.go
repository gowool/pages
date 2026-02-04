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

	"github.com/gowool/pages/internal"
)

func SelectSiteMiddleware(retriever SiteRetriever, skippers ...Skipper) MiddlewareFunc {
	if retriever == nil {
		panic("middleware: select site: retriever is required")
	}

	skip := ChainSkipper(skippers...)

	return func(next Handler) Handler {
		return HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
			if skip(r) {
				return next.ServeHTTP(w, r)
			}

			c := FromContext(r.Context())
			if c == nil {
				panic("middleware: select site: context is required")
			}

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

			site.Scheme = internal.Scheme(r)
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

type PatternArgsFunc func(*http.Request) []any

func SelectPageMiddleware(manager PageManager, authorizer PageAuthorizer, patternArgs PatternArgsFunc, skippers ...Skipper) MiddlewareFunc {
	if manager == nil {
		panic("middleware: select page: manager is required")
	}
	if authorizer == nil {
		panic("middleware: select page: authorizer is required")
	}

	if patternArgs == nil {
		patternArgs = PatternArgs()
	}

	skip := ChainSkipper(skippers...)

	findPage := func(r *http.Request, site *Site) (page *Page, err error) {
		if pattern := internal.Pattern(r); pattern != PageCMSPattern {
			page, err = manager.GetByPattern(r.Context(), site, pattern)
		} else {
			page, err = manager.GetByURL(r.Context(), site, r.URL.Path)
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

	return func(next Handler) Handler {
		return HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
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

			if page.Site == nil {
				page.Site = c.Site()
				page.SiteID = c.Site().ID
			}

			var args []any
			if page.IsDynamic() {
				args = patternArgs(r)
			}

			c.SetPage(page, args...)

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

func HybridPageMiddleware(pageHandler Handler, logger *slog.Logger, skippers ...Skipper) MiddlewareFunc {
	if pageHandler == nil {
		panic("middleware: hybrid page: page handler is required")
	}

	if logger == nil {
		logger = slog.New(slog.DiscardHandler)
	}

	logger = logger.WithGroup("hybrid_page")

	skippers = append(skippers, func(r *http.Request) bool {
		c := FromContext(r.Context())
		if c == nil {
			panic("middleware: hybrid page: context is required")
		}
		return c.HasPage() && !c.Page().IsHybrid()
	})

	skip := ChainSkipper(skippers...)

	pool := &sync.Pool{
		New: func() any {
			return new(delayedWriter)
		},
	}

	return func(next Handler) Handler {
		return HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
			if skip(r) {
				return next.ServeHTTP(w, r)
			}

			c := FromContext(r.Context())
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

			if response.status > 0 {
				c.SetStatus(response.status)
			}

			buffer := response.buffer.Bytes()
			c.SetContent(template.HTML(internal.BytesToString(buffer)))

			if !IsDecorable(response, r) {
				w.WriteHeader(c.Status())

				if len(buffer) > 0 {
					if _, err := w.Write(buffer); err != nil {
						logger.Error("write response error", "error", err)
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
	buffer   bytes.Buffer
	commited bool
	status   int
}

func (w *delayedWriter) reset(rw http.ResponseWriter) {
	w.ResponseWriter = rw
	w.buffer.Reset()
	w.commited = false
	w.status = http.StatusOK
}

func (w *delayedWriter) WriteHeader(statusCode int) {
	w.status = statusCode
	w.commited = true
}

func (w *delayedWriter) Write(data []byte) (int, error) {
	if !w.commited {
		w.WriteHeader(http.StatusOK)
	}
	return w.buffer.Write(data)
}

func (w *delayedWriter) Status() int {
	return w.status
}

func (w *delayedWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}
