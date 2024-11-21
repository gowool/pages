package pages

import (
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/dlclark/regexp2"

	"github.com/gowool/pages/internal"
	"github.com/gowool/pages/model"
	"github.com/gowool/pages/repository"
)

var _ SiteSelector = (*DefaultSiteSelector)(nil)

type RedirectError struct {
	Status int
	URL    string
}

func (r RedirectError) Error() string {
	return fmt.Sprintf("[%d] %s", r.Status, r.URL)
}

type SiteSelector interface {
	Retrieve(*http.Request) (m *model.Site, urlPath string, err error)
}

type DefaultSiteSelector struct {
	cfgRepository  repository.Configuration
	siteRepository repository.Site
}

func NewDefaultSiteSelector(cfgRepository repository.Configuration, siteRepository repository.Site) *DefaultSiteSelector {
	if cfgRepository == nil {
		panic("configuration repository is not specified")
	}
	if siteRepository == nil {
		panic("site repository is not specified")
	}
	return &DefaultSiteSelector{cfgRepository, siteRepository}
}

func (s *DefaultSiteSelector) Retrieve(r *http.Request) (*model.Site, string, error) {
	cfg, err := s.cfgRepository.Load(r.Context())
	if err != nil {
		return nil, "", err
	}

	switch cfg.Multisite {
	case model.Host:
		return s.HostRetrieve(r, cfg.FallbackLocale)
	case model.HostByLocale:
		return s.HostByLocaleRetrieve(r, cfg.FallbackLocale)
	case model.HostWithPath:
		return s.HostPathRetrieve(r, cfg.FallbackLocale)
	case model.HostWithPathByLocale:
		return s.HostPathByLocaleRetrieve(r, cfg.FallbackLocale)
	default:
		panic(fmt.Errorf("unknown multisite strategy: %s", cfg.Multisite))
	}
}

func (s *DefaultSiteSelector) HostRetrieve(r *http.Request, fallbackLocale string) (*model.Site, string, error) {
	host := Host(r)
	sites, err := s.siteRepository.FindByHosts(r.Context(), hosts(host), time.Now())
	if err != nil {
		return nil, "", err
	}

	var site *model.Site
	for _, s := range sites {
		site = &s
		if !s.IsLocalhost() {
			break
		}
	}

	if site != nil {
		site = internal.Ptr(site.WithHost(Scheme(r), host))
		if site.Locale == "" {
			site.Locale = getLocale(r, fallbackLocale)
		}
	}
	return site, r.URL.Path, nil
}

func (s *DefaultSiteSelector) HostByLocaleRetrieve(r *http.Request, fallbackLocale string) (*model.Site, string, error) {
	host := Host(r)
	sites, err := s.siteRepository.FindByHosts(r.Context(), hosts(host), time.Now())
	if err != nil {
		return nil, "", err
	}

	if index := slices.IndexFunc(sites, func(s model.Site) bool {
		return s.IsLocalhost()
	}); index >= 0 {
		sites = sites[:index+1]
	}

	if site := preferredSite(r, sites, fallbackLocale); site != nil {
		return site, r.URL.Path, nil
	}
	return nil, "", ErrSiteNotFound
}

func (s *DefaultSiteSelector) HostPathRetrieve(r *http.Request, fallbackLocale string) (*model.Site, string, error) {
	var (
		defaultSite *model.Site
		site        *model.Site
		pathInfo    string
	)

	host := Host(r)
	sites, err := s.siteRepository.FindByHosts(r.Context(), hosts(host), time.Now())
	if err != nil {
		return nil, "", err
	}

	for _, item := range sites {
		if defaultSite == nil && item.IsDefault {
			defaultSite = &item
		}

		match, err := matchRequest(r, item)
		if err != nil {
			continue
		}

		site = &item
		pathInfo = match

		if !item.IsLocalhost() {
			break
		}
	}

	if site != nil {
		site = internal.Ptr(site.WithHost(Scheme(r), host))
		if site.Locale == "" {
			site.Locale = getLocale(r, fallbackLocale)
		}

		if pathInfo == "" {
			pathInfo = "/"
		}
		return site, pathInfo, nil
	}

	if defaultSite != nil {
		defaultSite = internal.Ptr(defaultSite.WithHost(Scheme(r), host))
		return nil, "", RedirectError{Status: http.StatusMovedPermanently, URL: defaultSite.URL()}
	}
	return nil, "", ErrSiteNotFound
}

func (s *DefaultSiteSelector) HostPathByLocaleRetrieve(r *http.Request, fallbackLocale string) (*model.Site, string, error) {
	var (
		enabledSites []model.Site
		site         *model.Site
		pathInfo     string
	)

	host := Host(r)
	sites, err := s.siteRepository.FindByHosts(r.Context(), hosts(host), time.Now())
	if err != nil {
		return nil, "", err
	}

	enabledSites = make([]model.Site, 0, len(sites))
	for _, item := range sites {
		enabledSites = append(enabledSites, item)

		match, err := matchRequest(r, item)
		if err != nil {
			continue
		}

		site = &item
		pathInfo = match

		if !item.IsLocalhost() {
			break
		}
	}

	if site != nil {
		site = internal.Ptr(site.WithHost(Scheme(r), host))

		if pathInfo == "" {
			pathInfo = "/"
		}
		return site, pathInfo, nil
	}

	if len(enabledSites) > 0 {
		if defaultSite := preferredSite(r, enabledSites, fallbackLocale); defaultSite != nil {
			return nil, "", RedirectError{Status: http.StatusFound, URL: defaultSite.URL()}
		}
	}
	return nil, "", ErrSiteNotFound
}

func matchRequest(r *http.Request, site model.Site) (string, error) {
	var (
		re    *regexp2.Regexp
		match *regexp2.Match
		err   error
	)

	if site.RelativePath == "" || site.RelativePath == "/" {
		re = internal.ReNoPath
	} else if re, err = internal.RegexpPath(site.RelativePath); err != nil {
		return "", err
	}

	if match, err = re.FindStringMatch(r.URL.Path); err != nil {
		return "", err
	}

	if match == nil {
		return "", fmt.Errorf("invalid path %s", r.URL.Path)
	}

	groups := match.Groups()

	if len(groups) < 3 {
		return "", fmt.Errorf("invalid match path %s", r.URL.Path)
	}

	return groups[2].String(), nil
}

func hosts(host string) []string {
	return []string{host, "localhost", "127.0.0.1"}
}

func preferredSite(r *http.Request, sites []model.Site, fallbackLocale string) (site *model.Site) {
	locales := internal.FilterMap(sites, func(item model.Site) (string, bool) {
		return item.Locale, item.Locale != ""
	})

	locale := preferredLanguage(Languages(r), locales, fallbackLocale)
	host := Host(r)
	sHosts := hosts(host)

	if index := slices.IndexFunc(sites, func(item model.Site) bool {
		return item.Locale == locale && slices.Contains(sHosts, item.Host)
	}); index > -1 {
		item := sites[index]
		site = internal.Ptr(item.WithHost(Scheme(r), host))
	}
	return
}

func preferredLanguage(headerLanguages, siteLocales []string, fallback string) string {
	if len(siteLocales) == 0 {
		if len(headerLanguages) > 0 {
			return headerLanguages[0]
		}
		return fallback
	}

	if len(headerLanguages) == 0 {
		return siteLocales[0]
	}

	mLocales := make(map[string]struct{})
	for _, locale := range siteLocales {
		mLocales[locale] = struct{}{}
	}

	for _, language := range headerLanguages {
		if _, ok := mLocales[language]; ok {
			return language
		}

		if codes := strings.Split(language, "_"); len(codes) > 1 {
			if _, ok := mLocales[codes[0]]; ok {
				return codes[0]
			}
		}
	}
	return fallback
}
