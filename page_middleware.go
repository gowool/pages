package pages

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"unsafe"
)

const (
	HeaderXPageDecorable    = "X-Page-Decorable"
	HeaderXPageNotDecorable = "X-Page-Not-Decorable"
	headerXRequestedWith    = "X-Requested-With"
	xmlHTTPRequest          = "XMLHttpRequest"
)

type PageMiddleware struct {
	pageHandler PageHandler
	authorizer  PageAuthorizer
	strategy    PageDecoratorStrategy
	pool        *sync.Pool
	logger      *slog.Logger
}

func NewPageMiddleware(pageHandler PageHandler, authorizer PageAuthorizer, strategy PageDecoratorStrategy, logger *slog.Logger) *PageMiddleware {
	if pageHandler == nil {
		panic("page middleware: page handler is required")
	}
	if authorizer == nil {
		panic("page middleware: authorizer is required")
	}
	if strategy == nil {
		panic("page middleware: strategy is required")
	}

	if logger == nil {
		logger = slog.New(slog.DiscardHandler)
	}

	return &PageMiddleware{
		authorizer:  authorizer,
		pageHandler: pageHandler,
		strategy:    strategy,
		logger:      logger.WithGroup("page_middleware"),
		pool:        &sync.Pool{New: func() any { return new(delayedWriter) }},
	}
}

func (m *PageMiddleware) Middleware(w http.ResponseWriter, r *http.Request, next func(http.ResponseWriter, *http.Request) error) error {
	ctx := r.Context()

	if ok, err := m.strategy.IsDecorable(ctx, r.Pattern, r.URL.Path); !ok || err != nil {
		return next(w, r)
	}

	c := FromContext(ctx)
	if c == nil {
		panic("page hybrid middleware: context is required")
	}

	if !c.HasPage() {
		return fmt.Errorf("page hybrid middleware: %w", ErrPageNotFound)
	}

	page := c.Page()

	if page.Status == Draft {
		if err := m.allow(ctx, ViewDraftPage, c.Guest()); err != nil {
			return fmt.Errorf("page selector: %w", errors.Join(err, ErrPageNotFound))
		}
	}

	if page.Visibility == Private {
		if err := m.allow(ctx, ViewPrivatePage, c.Guest()); err != nil {
			return fmt.Errorf("page selector: %w", errors.Join(err, ErrPrivatePage))
		}
	}

	if !page.IsHybrid() {
		return next(w, r)
	}

	if page.Decorate {
		w.Header().Set(HeaderXPageDecorable, "1")
	} else {
		w.Header().Set(HeaderXPageNotDecorable, "1")
	}

	response := m.pool.Get().(*delayedWriter)
	response.reset(w)
	defer func() {
		response.reset(nil)
		m.pool.Put(response)
	}()

	if err := next(response, r); err != nil {
		return fmt.Errorf("page hybrid middleware: %w", err)
	}

	httpStatus := response.status
	if httpStatus > 0 {
		c.SetStatus(httpStatus)
	} else {
		httpStatus = c.Status()
	}

	buffer := response.buffer.Bytes()

	if !IsDecorable(response, r) {
		if len(buffer) > 0 {
			w.WriteHeader(httpStatus)
			if _, err := w.Write(buffer); err != nil {
				m.logger.Error("write response error", "error", err)
			}
			return nil
		}

		w.WriteHeader(http.StatusNoContent)
		return nil
	}

	if l := len(buffer); l > 0 {
		c.SetContent(template.HTML(unsafe.String(unsafe.SliceData(buffer), l)))
	}

	return m.pageHandler.Handle(w, r)
}

func (m *PageMiddleware) allow(ctx context.Context, action PageAction, isGuest bool) error {
	if isGuest {
		return errors.New("guest is not allowed")
	}

	decision, err := m.authorizer.Authorize(ctx, action)
	if err != nil {
		return err
	}

	if decision == Deny {
		return errors.New("access denied")
	}

	return nil
}

func IsDecorable(w *delayedWriter, r *http.Request) bool {
	contentType := w.Header().Get(headerContentType)

	if contentType != "" && !strings.HasPrefix(contentType, mimeTextHTML) {
		return false
	}

	if w.Header().Get(HeaderXPageNotDecorable) == "1" {
		return false
	}

	if w.Header().Get(HeaderXPageDecorable) == "1" {
		return true
	}

	if w.status != http.StatusOK {
		return false
	}

	return r.Header.Get(headerXRequestedWith) != xmlHTTPRequest
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

func (w *delayedWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}
