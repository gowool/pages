package pages

import (
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
	if s, ok := CtxData(ctx)["status"].(int); ok {
		status = s
	}

	return c.Render(status, page.Template, nil)
}
