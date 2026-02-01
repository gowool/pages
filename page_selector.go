package pages

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
)

var _ PageSelector = (*DefaultPageSelector)(nil)

type PageSelector interface {
	Select(r *http.Request) error
}

type PatternArgsFunc func(r *http.Request) []any

func PatternArgs() PatternArgsFunc {
	return func(r *http.Request) (args []any) {
		pattern := getPattern(r)

		n := strings.Count(pattern, "{")
		if n == 0 {
			return
		}

		args = make([]any, 0, n*2)

		var key strings.Builder
		for _, c := range pattern {
			switch c {
			case '{':
				key.Reset()
				key.WriteRune(c)
			case '.':
				continue
			case '}':
				if key.Len() == 0 {
					panic("invalid dynamic page pattern")
				}
				key.WriteRune(c)
				param := key.String()
				args = append(args, param, r.PathValue(param[1:len(param)-1]))
			default:
				if key.Len() > 0 {
					key.WriteRune(c)
				}
			}
		}
		return
	}
}

type DefaultPageSelector struct {
	retriever   PageRetriever
	patternArgs PatternArgsFunc
}

func NewDefaultPageSelector(
	retriever PageRetriever,
	patternArgs PatternArgsFunc,
) *DefaultPageSelector {
	if retriever == nil {
		panic("page selector: retriever is required")
	}
	if patternArgs == nil {
		patternArgs = PatternArgs()
	}

	return &DefaultPageSelector{
		retriever:   retriever,
		patternArgs: patternArgs,
	}
}

func (s *DefaultPageSelector) Select(r *http.Request) error {
	ctx := r.Context()
	c := FromContext(ctx)
	if c == nil {
		panic("page selector: context is required")
	}

	if !c.HasSite() {
		return fmt.Errorf("page selector: %w", ErrSiteNotFound)
	}

	site := c.Site()
	page, err := s.retriever.Retrieve(r, site)
	if err != nil {
		if errors.Is(err, ErrPageNotFound) {
			return fmt.Errorf("page selector: %w", err)
		}
		return fmt.Errorf("page selector: %w", errors.Join(err, ErrPageNotFound))
	}

	if page == nil {
		return fmt.Errorf("page selector: %w", ErrPageNotFound)
	}

	if page.Site == nil {
		page.Site = site
		page.SiteID = site.ID
	}

	var args []any
	if page.IsDynamic() {
		args = s.patternArgs(r)
	}

	c.SetPage(page, args...)

	return nil
}
