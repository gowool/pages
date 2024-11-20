package model

import (
	"time"

	"github.com/gosimple/slug"
)

type Menu struct {
	ID      int64     `json:"id,omitempty" yaml:"id,omitempty" required:"true"`
	NodeID  *int64    `json:"nodeID,omitempty" yaml:"nodeID,omitempty" required:"false"`
	Name    string    `json:"name,omitempty" yaml:"name,omitempty" required:"true"`
	Handle  string    `json:"handle,omitempty" yaml:"handle,omitempty" required:"true"`
	Enabled bool      `json:"enabled,omitempty" yaml:"enabled,omitempty" required:"true"`
	Created time.Time `json:"created,omitempty" yaml:"created,omitempty" required:"true"`
	Updated time.Time `json:"updated,omitempty" yaml:"updated,omitempty" required:"true"`
	Node    *Node     `json:"-" yaml:"-"`
}

func (m Menu) GetID() int64 {
	return m.ID
}

func (m Menu) String() string {
	if m.Name == "" {
		return "n/a"
	}
	return m.Name
}

func (m Menu) WithFixedHandle() Menu {
	if m.Handle == "" {
		m.Handle = slug.Make(m.Name)
	}
	return m
}
