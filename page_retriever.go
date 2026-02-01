package pages

import (
	"net/http"
	"strings"
)

var _ PageRetriever = (*DefaultPageRetriever)(nil)

type PageRetriever interface {
	Retrieve(r *http.Request, site *Site) (*Page, error)
}

type DefaultPageRetriever struct {
	manager PageManager
}

func NewDefaultPageRetriever(manager PageManager) *DefaultPageRetriever {
	if manager == nil {
		panic("page retriever: manager is required")
	}

	return &DefaultPageRetriever{
		manager: manager,
	}
}

func (s *DefaultPageRetriever) Retrieve(r *http.Request, site *Site) (*Page, error) {
	if r == nil {
		panic("page retriever: request is required")
	}

	if site == nil {
		panic("page retriever: site is required")
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

func getPattern(r *http.Request) string {
	pattern := r.Pattern
	if index := strings.IndexRune(pattern, ' '); index > -1 {
		pattern = pattern[index+1:]
	}
	return pattern
}
