package model

import (
	"fmt"
	"time"
)

type Site struct {
	ID           int64             `json:"id,omitempty" yaml:"id,omitempty" required:"true"`
	Name         string            `json:"name,omitempty" yaml:"name,omitempty" required:"true"`
	Title        string            `json:"title,omitempty" yaml:"title,omitempty" required:"false"`
	Separator    string            `json:"separator,omitempty" yaml:"separator,omitempty" required:"true"`
	Host         string            `json:"host,omitempty" yaml:"host,omitempty" required:"true"`
	Locale       string            `json:"locale,omitempty" yaml:"locale,omitempty" required:"false"`
	RelativePath string            `json:"relativePath,omitempty" yaml:"relativePath,omitempty" required:"false"`
	IsDefault    bool              `json:"isDefault,omitempty" yaml:"isDefault,omitempty" required:"false"`
	Javascript   string            `json:"javascript,omitempty" yaml:"javascript,omitempty" required:"false"`
	Stylesheet   string            `json:"stylesheet,omitempty" yaml:"stylesheet,omitempty" required:"false"`
	Metas        []Meta            `json:"metas,omitempty" yaml:"metas,omitempty" required:"false"`
	Metadata     map[string]string `json:"metadata,omitempty" yaml:"metadata,omitempty" required:"false"`
	Created      time.Time         `json:"created,omitempty" yaml:"created,omitempty" required:"true"`
	Updated      time.Time         `json:"updated,omitempty" yaml:"updated,omitempty" required:"true"`
	Published    *time.Time        `json:"published,omitempty" yaml:"published,omitempty" required:"false"`
	Expired      *time.Time        `json:"expired,omitempty" yaml:"expired,omitempty" required:"false"`

	scheme string
	host   string
}

func (s Site) GetID() int64 {
	return s.ID
}

func (s Site) String() string {
	if s.Name == "" {
		return "n/a"
	}
	return s.Name
}

func (s Site) IsEnabled(now time.Time) bool {
	now = now.Truncate(60 * time.Second)
	return s.Published != nil &&
		!s.Published.IsZero() &&
		(s.Published.Before(now) || s.Published.Equal(now)) &&
		(s.Expired == nil || s.Expired.IsZero() || s.Expired.After(now))
}

func (s Site) IsLocalhost() bool {
	if s.host == "" {
		return s.Host == "localhost"
	}
	return s.host == "localhost"
}

func (s Site) URL() string {
	return fmt.Sprintf("%s//%s%s", s.scheme, s.Host, s.RelativePath)
}

func (s Site) WithHost(scheme, host string) Site {
	s.host = s.Host
	s.scheme = scheme + ":"
	s.Host = host
	return s
}
