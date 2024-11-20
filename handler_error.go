package pages

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"net/http"
	"reflect"
	"slices"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"github.com/gowool/pages/internal"
	"github.com/gowool/pages/model"
	"github.com/gowool/pages/repository"
)

var ErrInternal = errors.New("internal server error")

type ErrorResolverFunc func(*http.Request, error) (statusCode int, data any)

type ErrorHandler struct {
	Logger         *zap.Logger
	Resolver       ErrorResolverFunc
	SiteSelector   SiteSelector
	PageHandler    PageHandler
	PageRepository repository.Page
	CfgRepository  repository.Configuration
}

func NewErrorHandler(
	pageHandler PageHandler,
	resolver ErrorResolverFunc,
	siteSelector SiteSelector,
	pageRepository repository.Page,
	cfgRepository repository.Configuration,
	logger *zap.Logger,
) *ErrorHandler {
	if pageHandler == nil {
		panic("page handler is not specified")
	}
	if resolver == nil {
		panic("error resolver is not specified")
	}
	if siteSelector == nil {
		panic("site selector is not specified")
	}
	if pageRepository == nil {
		panic("page repository is not specified")
	}
	if cfgRepository == nil {
		panic("configuration repository is not specified")
	}
	if logger == nil {
		logger = zap.NewNop()
	}
	return &ErrorHandler{
		Logger:         logger,
		Resolver:       resolver,
		SiteSelector:   siteSelector,
		PageHandler:    pageHandler,
		PageRepository: pageRepository,
		CfgRepository:  cfgRepository,
	}
}

func (h *ErrorHandler) Handle(err error, c echo.Context) {
	if c.Response().Committed {
		return
	}

	r := c.Request()

	status, data := h.Resolver(r, err)

	// Send response
	if c.Request().Method == http.MethodHead { // Issue #608
		if err = c.NoContent(status); err != nil {
			h.Logger.Error("cms: error handler", zap.Error(err))
		}
		return
	} else if strings.Contains(r.Header.Get(echo.HeaderAccept), "json") {
		if err = c.JSON(status, data); err != nil {
			h.Logger.Error("cms: error handler", zap.Error(err))
		}
		return
	}

	cfg, err := h.CfgRepository.Load(r.Context())
	if err != nil {
		h.Logger.Error("find configuration", zap.Error(err))
		if err = c.String(status, http.StatusText(status)); err != nil {
			h.Logger.Error("cms: error handler", zap.Error(err))
		}
		return
	}

	var htmlData map[string]any
	if tmp, ok := data.(map[string]any); ok && tmp != nil {
		htmlData = tmp
	} else {
		hdType := reflect.TypeOf(htmlData)
		if reflect.TypeOf(data).ConvertibleTo(hdType) {
			htmlData = reflect.ValueOf(data).Convert(hdType).Interface().(map[string]any)
		} else {
			htmlData = map[string]any{"data": data}
		}
	}

	maps.Copy(htmlData, CtxData(r.Context()))

	htmlData["err"] = err
	htmlData["status"] = status

	if SkipSelectSite(r.Context()) || cfg.IgnoreURI(r.URL.Path) {
		h.internal(c, htmlData, cfg)
		return
	}

	if errors.Is(err, ErrInternal) {
		h.internal(c, htmlData, cfg)
		return
	}

	if CtxSite(r.Context()) == nil {
		if site, urlPath, errSite := h.SiteSelector.Retrieve(r); errSite == nil {
			path := r.URL.Path
			rawPath := r.URL.RawPath
			defer func() {
				r.URL.Path = path
				r.URL.RawPath = rawPath
			}()

			r.URL.Path = urlPath
			r.URL.RawPath = ""

			ctx := WithSite(r.Context(), site)
			r = r.WithContext(ctx)
			c.SetRequest(r)
		}
	}

	if SkipSelectPage(r.Context()) || cfg.IgnorePattern(r.Pattern) {
		h.internal(c, htmlData, cfg)
		return
	}

	if (errors.Is(err, repository.ErrPageNotFound) || status == http.StatusNotFound) && CtxEditor(r.Context()) {
		h.create(c, htmlData, cfg)
		return
	}

	h.native(c, status, htmlData, cfg)
}

func (h *ErrorHandler) internal(c echo.Context, data map[string]any, cfg model.Configuration) {
	r := c.Request()
	site := CtxSite(r.Context())
	if site == nil {
		site = siteInternal(r, cfg)
	}

	ctx := WithSite(r.Context(), site)
	page, err := h.findPageByPattern(ctx, site.ID, model.PageErrorInternal)
	if err != nil {
		p := errorPages[model.PageErrorInternal]
		p.SiteID = site.ID
		page = &p
	}

	c.SetRequest(r.WithContext(ctx))

	h.serve(c, page, data)
}

func (h *ErrorHandler) create(c echo.Context, data map[string]any, cfg model.Configuration) {
	site := CtxSite(c.Request().Context())

	page, err := h.findPageByPattern(c.Request().Context(), site.ID, model.PageInternalCreate)
	if err != nil {
		h.internal(c, data, cfg)
		return
	}

	h.serve(c, page, data)
}

func (h *ErrorHandler) native(c echo.Context, status int, data map[string]any, cfg model.Configuration) {
	page, err := h.findPageByStatus(c.Request().Context(), cfg, CtxSite(c.Request().Context()).ID, status)
	if err != nil {
		h.internal(c, data, cfg)
		return
	}

	h.serve(c, page, data)
}

func (h *ErrorHandler) serve(c echo.Context, page *model.Page, data map[string]any) {
	if page.Title == "" {
		page.Title = http.StatusText(data["status"].(int))
	}

	r := c.Request()
	ctx := WithPage(r.Context(), page)
	ctx = WithData(ctx, data)
	c.SetRequest(r.WithContext(ctx))

	if err := h.PageHandler.Handle(c); err != nil {
		if !c.Response().Committed {
			var httpErr *echo.HTTPError
			if !errors.As(err, &httpErr) {
				httpErr = echo.ErrInternalServerError
			}
			_ = c.String(httpErr.Code, fmt.Sprintf("%v", httpErr.Message))
		}
		h.Logger.Error("cms: error page handler", zap.Error(err))
	}
}

func (h *ErrorHandler) findPageByStatus(ctx context.Context, cfg model.Configuration, siteID int64, status int) (*model.Page, error) {
	var pattern string
	for key, codes := range cfg.CatchErrors {
		if slices.Contains(codes, status) {
			pattern = key
			break
		}
	}
	return h.findPageByPattern(ctx, siteID, pattern)
}

func (h *ErrorHandler) findPageByPattern(ctx context.Context, siteID int64, pattern string) (*model.Page, error) {
	if pattern == "" {
		return nil, repository.ErrPageNotFound
	}

	var now time.Time
	if !CtxEditor(ctx) {
		now = time.Now()
	}

	page, err := h.PageRepository.FindByPattern(ctx, siteID, pattern, now)
	if err != nil {
		return nil, err
	}
	return &page, nil
}

func siteInternal(r *http.Request, cfg model.Configuration) *model.Site {
	now := time.Now()

	locale := "en_US"
	if cfg.FallbackLocale != "" {
		locale = cfg.FallbackLocale
	}

	site := &model.Site{
		ID:        -1,
		Name:      "Internal",
		Separator: " - ",
		Locale:    getLocale(r, locale),
		Created:   now,
		Updated:   now,
		Published: &now,
	}

	return internal.Ptr(site.WithHost(Scheme(r), Host(r)))
}

func ErrorResolver(asHTTPError func(err error, target **echo.HTTPError)) ErrorResolverFunc {
	return func(r *http.Request, err error) (int, any) {
		he := &echo.HTTPError{
			Code:    http.StatusInternalServerError,
			Message: http.StatusText(http.StatusInternalServerError),
		}
		if errors.As(err, &he) {
			if he.Internal != nil { // max 2 levels of checks even if internal could have also internal
				errors.As(he.Internal, &he)
			}
		}

		if asHTTPError != nil {
			asHTTPError(err, &he)
		}

		return he.Code, errorData(he, CtxDebug(r.Context()))
	}
}

func errorData(he *echo.HTTPError, internal bool) any {
	code := he.Code
	title := http.StatusText(code)
	detail := he.Message

	switch m := he.Message.(type) {
	case error:
		detail = m.Error()
	}

	if internal && he.Internal != nil {
		return echo.Map{"status": code, "title": title, "detail": detail, "error": he.Internal.Error()}
	}
	return echo.Map{"status": code, "title": title, "detail": detail}
}
