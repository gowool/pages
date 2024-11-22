package pages

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

var _ PageHandler = (*DefaultPageHandler)(nil)

type PageHandler interface {
	Handle(echo.Context) error
}

type DefaultPageHandler struct{}

func NewDefaultPageHandler() *DefaultPageHandler {
	return &DefaultPageHandler{}
}

func (h *DefaultPageHandler) Handle(c echo.Context) error {
	ctx := c.Request().Context()

	site := CtxSite(ctx)
	if site == nil {
		return fmt.Errorf("page handler: %w", ErrSiteNotFound)
	}

	page := CtxPage(ctx)
	if page == nil {
		return fmt.Errorf("page handler: %w", ErrPageNotFound)
	}

	status := http.StatusOK
	if page.Status > 0 {
		status = page.Status
	} else if s, ok := CtxData(ctx)["status"].(int); ok {
		status = s
	}

	if page.ContentType == "" {
		return c.Render(status, page.Template, nil)
	}

	buf := new(bytes.Buffer)
	if err := c.Echo().Renderer.Render(buf, page.Template, nil, c); err != nil {
		return err
	}
	return c.Blob(status, page.ContentType, buf.Bytes())
}
