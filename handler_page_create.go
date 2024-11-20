package pages

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/gowool/pages/model"
	"github.com/gowool/pages/repository"
)

type PageCreateRequest struct {
	SiteID   int64  `json:"site_id,omitempty" form:"site_id,omitempty" validate:"required"`
	URL      string `json:"url,omitempty" form:"url,omitempty" validate:"required"`
	Template string `json:"template,omitempty" form:"template,omitempty" validate:"required"`
	Title    string `json:"title,omitempty" form:"title,omitempty" validate:"max=254"`
}

type PageCreateHandler struct {
	validator Validator
	pageRepo  repository.Page
}

func NewPageCreateHandler(validator Validator, pageRepo repository.Page) *PageCreateHandler {
	if validator == nil {
		panic("validator is not specified")
	}
	if pageRepo == nil {
		panic("page repository is not specified")
	}
	return &PageCreateHandler{
		validator: validator,
		pageRepo:  pageRepo,
	}
}

func (h *PageCreateHandler) Handle(c echo.Context) error {
	if c.Request().Method != http.MethodPost {
		c.Response().Header().Set(echo.HeaderAllow, http.MethodPost)
		return echo.ErrMethodNotAllowed
	}

	if !CtxEditor(c.Request().Context()) {
		return echo.ErrForbidden
	}

	var dto PageCreateRequest
	if err := c.Bind(&dto); err != nil {
		return err
	}

	if err := h.validator.ValidateCtx(c.Request().Context(), dto); err != nil {
		return err
	}

	page := model.Page{
		SiteID:    dto.SiteID,
		CustomURL: dto.URL,
		Template:  dto.Template,
		Title:     dto.Title,
		Name:      strings.ToTitle(strings.ReplaceAll(dto.URL, "/", " ")),
		Pattern:   model.PageCMS,
		Decorate:  true,
	}

	if err := h.pageRepo.Create(c.Request().Context(), &page); err != nil {
		return err
	}

	return c.Redirect(http.StatusSeeOther, page.URL)
}
