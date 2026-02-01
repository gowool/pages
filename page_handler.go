package pages

import (
	"bytes"
	"fmt"
	"log/slog"
	"net/http"
)

var _ PageHandler = (*DefaultPageHandler)(nil)

const (
	headerContentType       = "Content-Type"
	mimeTextHTML            = "text/html"
	mimeTextHTMLCharsetUTF8 = mimeTextHTML + "; charset=UTF-8"
)

type PageHandler interface {
	Handle(w http.ResponseWriter, r *http.Request) error
}

type PageContext struct {
	*Context
	Request *http.Request
}

func (c PageContext) Value(key any) any {
	return c.Request.Context().Value(key)
}

type ViewCtxFunc func(*Context, *http.Request) any

func ViewCtx() ViewCtxFunc {
	return func(ctx *Context, r *http.Request) any {
		return PageContext{Context: ctx, Request: r}
	}
}

type DefaultPageHandler struct {
	theme       Theme
	viewCtxFunc ViewCtxFunc
	logger      *slog.Logger
}

func NewDefaultPageHandler(theme Theme, viewCtxFunc ViewCtxFunc, logger *slog.Logger) *DefaultPageHandler {
	if theme == nil {
		panic("page handler: theme is required")
	}
	if viewCtxFunc == nil {
		viewCtxFunc = ViewCtx()
	}
	if logger == nil {
		logger = slog.New(slog.DiscardHandler)
	}

	return &DefaultPageHandler{
		theme:       theme,
		viewCtxFunc: viewCtxFunc,
		logger:      logger.WithGroup("page_handler"),
	}
}

func (h *DefaultPageHandler) Handle(w http.ResponseWriter, r *http.Request) error {
	c := FromContext(r.Context())
	if c == nil {
		panic("page handler: context is required")
	}

	if !c.HasSite() {
		return fmt.Errorf("page handler: %w", ErrSiteNotFound)
	}

	if !c.HasPage() {
		return fmt.Errorf("page handler: %w", ErrPageNotFound)
	}

	page := c.Page()

	for key, values := range page.Header {
		for i, value := range values {
			if i == 0 {
				w.Header().Set(key, value)
			} else {
				w.Header().Add(key, value)
			}
		}
	}

	if page.Template == "" {
		w.WriteHeader(c.Status())
		return nil
	}

	var buf bytes.Buffer
	if err := h.theme.Write(r.Context(), &buf, page.Template, h.viewCtxFunc(c, r)); err != nil {
		return fmt.Errorf("page handler: theme write error: %w", err)
	}

	ct := w.Header().Get(headerContentType)
	if ct == "" {
		ct = mimeTextHTMLCharsetUTF8
	}

	w.Header().Set(headerContentType, ct)
	w.WriteHeader(c.Status())

	if _, err := w.Write(buf.Bytes()); err != nil {
		h.logger.Error("write response error", "error", err)
	}

	return nil
}
