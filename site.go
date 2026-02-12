package pages

import (
	"maps"
	"slices"
	"strings"
	"time"

	"golang.org/x/text/language"
)

type Site struct {
	ID ID `json:"id,omitempty" yaml:"id,omitempty"`

	Created time.Time `json:"created,omitzero" yaml:"created,omitempty"`
	Updated time.Time `json:"updated,omitzero" yaml:"updated,omitempty"`

	Status Status `json:"code,omitempty" yaml:"code,omitempty"`

	MetaTags *MetaTags `json:"metaTags,omitempty" yaml:"metaTags,omitempty"`
	Metadata Metadata  `json:"metadata,omitempty" yaml:"metadata,omitempty"`

	Name      string   `json:"name,omitempty" yaml:"name,omitempty"`
	Title     string   `json:"title,omitempty" yaml:"title,omitempty"`
	Separator string   `json:"separator,omitempty" yaml:"separator,omitempty"`
	Locale    string   `json:"locale,omitempty" yaml:"locale,omitempty"`
	Timezone  string   `json:"timezone,omitempty" yaml:"timezone,omitempty"`
	Countries []string `json:"countries,omitempty" yaml:"countries,omitempty"`

	Scheme       string `json:"scheme,omitempty" yaml:"scheme,omitempty"`
	Host         string `json:"host,omitempty" yaml:"host,omitempty"`
	RelativePath string `json:"relativePath,omitempty" yaml:"relativePath,omitempty"`
	IsDefault    bool   `json:"isDefault,omitempty" yaml:"isDefault,omitempty"`

	IsRoot   bool `json:"-" yaml:"-"`
	location *time.Location
	tag      *language.Tag
}

func NewSite() *Site {
	t := time.Now().UTC()

	return &Site{
		Created:   t,
		Updated:   t,
		Name:      "Localhost",
		Host:      "localhost",
		Scheme:    "https",
		Locale:    "en",
		Timezone:  "UTC",
		Separator: " | ",
		MetaTags:  NewMetaTags(DefaultCharset),
		Metadata:  NewMetadata(nil),
		Status:    Draft,
	}
}

func (s *Site) String() string {
	if s.Name == "" {
		return "n/a"
	}
	return s.Name
}

func (s *Site) IsLocalhost() bool {
	host, _, _ := strings.Cut(s.Host, ":")
	return host == "" || host == "localhost" || host == "127.0.0.1"
}

func (s *Site) Home() string {
	if s.IsRoot {
		return s.Origin()
	}
	return s.URL()
}

func (s *Site) Origin() string {
	return s.Scheme + "://" + s.Host
}

func (s *Site) URL() string {
	if s.RelativePath == "" || s.RelativePath == "/" {
		return s.Origin()
	}

	if s.RelativePath[0] != '/' {
		return s.Origin() + "/" + s.RelativePath
	}

	return s.Origin() + s.RelativePath
}

func (s *Site) Location() *time.Location {
	if s.location == nil {
		var err error
		s.location, err = time.LoadLocation(s.Timezone)
		if err != nil {
			s.location = time.UTC // Fallback to UTC if the location cannot be loaded
		}
	}
	return s.location
}

func (s *Site) Tag() language.Tag {
	if s.tag == nil {
		tag, err := language.Parse(s.Locale)
		if err != nil {
			tag = language.English // Fallback to English if the locale cannot be parsed
		}
		s.tag = &tag
	}
	return *s.tag
}

func (s *Site) Copy() *Site {
	site := *s
	site.Metadata = maps.Clone(s.Metadata)
	site.Countries = slices.Clone(s.Countries)
	site.location = nil
	site.tag = nil

	if s.MetaTags != nil {
		site.MetaTags = NewMetaTags(s.MetaTags.Charset, s.MetaTags)
	}

	return &site
}
