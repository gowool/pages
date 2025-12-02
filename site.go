package pages

import (
	"time"

	"golang.org/x/text/language"
)

type Site struct {
	ID ID `json:"id,omitempty" yaml:"id,omitempty"`

	Created time.Time `json:"created,omitzero" yaml:"created,omitempty"`
	Updated time.Time `json:"updated,omitzero" yaml:"updated,omitempty"`

	MetaTags *MetaTags      `json:"metaTags,omitempty" yaml:"metaTags,omitempty"`
	Metadata map[string]any `json:"metadata,omitempty" yaml:"metadata,omitempty"`

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
	Enabled      bool   `json:"enabled,omitempty" yaml:"enabled,omitempty"`

	location *time.Location
	tag      *language.Tag
	isRoot   bool
}

func NewSite() *Site {
	return &Site{
		Created:   time.Now().UTC(),
		Updated:   time.Now().UTC(),
		Name:      "Localhost",
		Host:      "localhost",
		Scheme:    "https",
		Locale:    "en",
		Timezone:  "UTC",
		Separator: " | ",
		MetaTags:  NewMetaTags(DefaultCharset),
		Metadata:  make(map[string]any),
	}
}

func (s *Site) String() string {
	if s.Name == "" {
		return "n/a"
	}
	return s.Name
}

func (s *Site) IsLocalhost() bool {
	return s.Host == "" || s.Host == "localhost" || s.Host == "127.0.0.1"
}

func (s *Site) Home() string {
	if s.isRoot {
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
