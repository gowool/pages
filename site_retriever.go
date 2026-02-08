package pages

import (
	"maps"
	"net/http"
	"slices"
	"strings"

	"github.com/gowool/gor"
	"github.com/gowool/pages/internal"
	"golang.org/x/text/language"
)

var _ SiteRetriever = (*HTTPSiteRetriever)(nil)

type SiteRetriever interface {
	Retrieve(r *http.Request) (*Site, string, error)
}

type HTTPSiteRetrieverConfig struct {
	// CountryFunc determines the country based on the provided
	// HTTP request and returns it as a string along with any error.
	CountryFunc func(*http.Request) (string, error)

	// ErrorFunc handles errors by taking an HTTP request and an error,
	// returning a Site instance or an error.
	ErrorFunc func(*http.Request, error) (*Site, error)
}

func (c *HTTPSiteRetrieverConfig) SetDefaults() {
	if c.CountryFunc == nil {
		c.CountryFunc = func(r *http.Request) (string, error) {
			return strings.ToUpper(r.Header.Get(gor.HeaderCFIPCountry)), nil
		}
	}

	if c.ErrorFunc == nil {
		c.ErrorFunc = func(_ *http.Request, err error) (*Site, error) {
			return nil, err
		}
	}
}

type HTTPSiteRetriever struct {
	store       SiteStore
	countryFunc func(*http.Request) (string, error)
	errorFunc   func(*http.Request, error) (*Site, error)
}

func NewHTTPSiteRetriever(store SiteStore) *HTTPSiteRetriever {
	return NewHTTPSiteRetrieverWithConfig(store, HTTPSiteRetrieverConfig{})
}

func NewHTTPSiteRetrieverWithConfig(store SiteStore, config HTTPSiteRetrieverConfig) *HTTPSiteRetriever {
	if store == nil {
		panic("site retriever: store is required")
	}
	config.SetDefaults()

	return &HTTPSiteRetriever{
		store:       store,
		countryFunc: config.CountryFunc,
		errorFunc:   config.ErrorFunc,
	}
}

type candidate struct {
	site *Site
	path string
}

func (s *HTTPSiteRetriever) Retrieve(r *http.Request) (*Site, string, error) {
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

func (s *HTTPSiteRetriever) candidate(r *http.Request, candidates []candidate, country string) (*Site, string) {
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

func (s *HTTPSiteRetriever) language(r *http.Request, candidateTags map[language.Tag]candidate, defaultCandidate candidate) (*Site, string) {
	t, _, err := language.ParseAcceptLanguage(r.Header.Get(gor.HeaderAcceptLanguage))
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

func (s *HTTPSiteRetriever) resolveError(r *http.Request, err error) (*Site, string, error) {
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
