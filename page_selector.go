package pages

import (
	"net/http"
)

var _ PageSelector = (*DefaultPageSelector)(nil)

type PageSelector interface {
	Retrieve(r *http.Request, site *Site) (*Page, error)
}

type DefaultPageSelector struct {
	manager PageManager
}

func NewPageSelector(manager PageManager) *DefaultPageSelector {
	if manager == nil {
		panic("page selector: manager is required")
	}

	return &DefaultPageSelector{
		manager: manager,
	}
}

func (s *DefaultPageSelector) Retrieve(r *http.Request, site *Site) (*Page, error) {
	if r == nil {
		panic("page selector: request is required")
	}

	if site == nil {
		panic("page selector: site is required")
	}

	pattern := getPattern(r)

	pathInfo := r.URL.Path
	if pathInfo == "" {
		pathInfo = "/"
	}

	var (
		page *Page
		err  error
	)

	if pattern == PageCMSPattern {
		page, err = s.manager.GetByURL(r.Context(), site, pathInfo)
	} else {
		page, err = s.manager.GetByPattern(r.Context(), site, pattern)
	}

	if err != nil {
		return nil, err
	}

	if page == nil {
		return nil, ErrPageNotFound
	}

	return page, nil
}
