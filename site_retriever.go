package pages

import (
	"maps"
	"net/http"
	"slices"
	"strings"

	"github.com/gowool/pages/internal"
	"golang.org/x/text/language"
)

var _ SiteRetriever = (*DefaultSiteRetriever)(nil)

const (
	headerAcceptLanguage = "Accept-Language"
	headerCFIPCountry    = "CF-IPCountry"
)

type SiteRetriever interface {
	Retrieve(r *http.Request) (*Site, string, error)
}

type DefaultSiteRetriever struct {
	store       SiteStore
	countryFunc func(*http.Request) (string, error)
	errorFunc   func(*http.Request, error) (*Site, error)
}

func NewDefaultSiteRetriever(
	store SiteStore,
	countryFunc func(*http.Request) (string, error),
	errorFunc func(*http.Request, error) (*Site, error),
) *DefaultSiteRetriever {
	if store == nil {
		panic("site retriever: store is required")
	}

	if countryFunc == nil {
		countryFunc = func(r *http.Request) (string, error) {
			return strings.ToUpper(r.Header.Get(headerCFIPCountry)), nil
		}
	}

	if errorFunc == nil {
		errorFunc = func(_ *http.Request, err error) (*Site, error) {
			return nil, err
		}
	}

	return &DefaultSiteRetriever{
		store:       store,
		countryFunc: countryFunc,
		errorFunc:   errorFunc,
	}
}

type candidate struct {
	site *Site
	path string
}

func (s *DefaultSiteRetriever) Retrieve(r *http.Request) (*Site, string, error) {
	if r == nil {
		panic("site retriever: request is required")
	}

	country, err := s.countryFunc(r)
	if err != nil {
		if site, pathInfo, err := s.resolveError(r, err); site != nil || err != nil {
			return site, pathInfo, err
		}
	}

	host := internal.Host(r)

	localhosts := make([]candidate, 0, 1)
	defaults := make([]candidate, 0, 1)
	candidates := make([]candidate, 0, 5)
	sites := make([]candidate, 0, 10)

	for site, err := range s.store.FindPublished(r.Context()) {
		if err != nil {
			if site, pathInfo, err := s.resolveError(r, err); site != nil || err != nil {
				return site, pathInfo, err
			}
			continue
		}

		if site.IsLocalhost() {
			localhosts = append(localhosts, candidate{site: site})
		}

		if site.Host != host {
			continue
		}

		sites = append(sites, candidate{site: site})

		if site.IsDefault {
			defaults = append(defaults, candidate{site: site})
		}

		pathInfo, err := internal.MatchRequest(r, site.RelativePath)
		if err != nil {
			continue
		}

		candidates = append(candidates, candidate{site: site, path: pathInfo})
	}

	switch len(candidates) {
	case 0:
		if site, pathInfo := s.candidate(r, sites, country); site != nil {
			return site, pathInfo, nil
		}
	case 1:
		return candidates[0].site, candidates[0].path, nil
	}

	if site, pathInfo := s.candidate(r, candidates, country); site != nil {
		return site, pathInfo, nil
	}

	switch len(defaults) {
	case 0:
		if len(localhosts) > 0 {
			return localhosts[0].site, "", nil
		}
		return nil, "", ErrSiteNotFound
	case 1:
		return defaults[0].site, "", nil
	}

	if site, pathInfo := s.candidate(r, defaults, country); site != nil {
		return site, pathInfo, nil
	}
	return defaults[0].site, "", nil
}

func (s *DefaultSiteRetriever) candidate(r *http.Request, candidates []candidate, country string) (*Site, string) {
	var defaultCountry, defaultNoCountry candidate
	countryTags, noCountryTags := make(map[language.Tag]candidate), make(map[language.Tag]candidate)

	for _, c := range candidates {
		if country != "" && len(c.site.Countries) > 0 && !slices.Contains(c.site.Countries, country) {
			continue
		}
		if len(c.site.Countries) == 0 {
			noCountryTags[c.site.Tag()] = c
			defaultNoCountry = c
		} else if country != "" {
			countryTags[c.site.Tag()] = c
			defaultCountry = c
		}
	}

	switch len(countryTags) {
	case 1:
		return defaultCountry.site, defaultCountry.path
	case 0:
		switch len(noCountryTags) {
		case 1:
			return defaultNoCountry.site, defaultNoCountry.path
		case 0:
			return nil, ""
		}
		return s.language(r, noCountryTags, defaultNoCountry)
	}

	return s.language(r, countryTags, defaultCountry)
}

func (s *DefaultSiteRetriever) language(r *http.Request, candidateTags map[language.Tag]candidate, defaultCandidate candidate) (*Site, string) {
	t, _, err := language.ParseAcceptLanguage(r.Header.Get(headerAcceptLanguage))
	if err != nil || len(t) == 0 {
		return defaultCandidate.site, defaultCandidate.path
	}

	matcher := language.NewMatcher(slices.Collect(maps.Keys(candidateTags)))
	tag, _, _ := matcher.Match(t...)

	for !tag.IsRoot() {
		if c, ok := candidateTags[tag]; ok {
			return c.site, c.path
		}
		tag = tag.Parent()
	}

	return defaultCandidate.site, defaultCandidate.path
}

func (s *DefaultSiteRetriever) resolveError(r *http.Request, err error) (*Site, string, error) {
	site, err := s.errorFunc(r, err)
	if err != nil {
		return nil, "", err
	}

	if site == nil {
		return nil, "", nil
	}

	pathInfo, _ := internal.MatchRequest(r, site.RelativePath)

	return site, pathInfo, nil
}
