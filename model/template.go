package model

import (
	"time"

	"github.com/gowool/pages/internal"
)

var TemplateTypes = []TemplateType{TemplateDB, TemplateFS}

const (
	TemplateDB = TemplateType("db")
	TemplateFS = TemplateType("fs")
)

type TemplateType string

func (t TemplateType) IsZero() bool {
	return t == ""
}

func (t TemplateType) String() string {
	return string(t)
}

type Template struct {
	ID      int64        `json:"id,omitempty" yaml:"id,omitempty" required:"true"`
	Name    string       `json:"name,omitempty" yaml:"name,omitempty" required:"true"`
	Content string       `json:"content,omitempty" yaml:"content,omitempty" required:"false"`
	Type    TemplateType `json:"type,omitempty" yaml:"type,omitempty" required:"true" enum:"db,fs"`
	Enabled bool         `json:"enabled,omitempty" yaml:"enabled,omitempty" required:"false"`
	Created time.Time    `json:"created,omitempty" yaml:"created,omitempty" required:"true"`
	Updated time.Time    `json:"updated,omitempty" yaml:"updated,omitempty" required:"true"`
}

func (t Template) GetID() int64 {
	return t.ID
}

func (t Template) String() string {
	if t.Name == "" {
		return "n/a"
	}
	return t.Name
}

func (t Template) Code() []byte {
	return internal.Bytes(t.Content)
}

func (t Template) Changed() time.Time {
	return t.Updated
}
