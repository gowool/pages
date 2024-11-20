package model

import (
	"fmt"
	"time"
)

type Node struct {
	ID                 int64             `json:"id,omitempty" yaml:"id,omitempty" required:"true"`
	ParentID           int64             `json:"parentID,omitempty" yaml:"parentID,omitempty" required:"true"`
	Name               string            `json:"name,omitempty" yaml:"name,omitempty" required:"true"`
	Label              string            `json:"label,omitempty" yaml:"label,omitempty" required:"false"`
	URI                string            `json:"uri,omitempty" yaml:"uri,omitempty" required:"false"`
	Path               string            `json:"path,omitempty" yaml:"path,omitempty" required:"true"`
	Level              int               `json:"level,omitempty" yaml:"level,omitempty" required:"false"`
	Position           int               `json:"position,omitempty" yaml:"position,omitempty" required:"false"`
	DisplayChildren    bool              `json:"displayChildren,omitempty" yaml:"displayChildren,omitempty" required:"false"`
	Display            bool              `json:"display,omitempty" yaml:"display,omitempty" required:"false"`
	Attributes         map[string]string `json:"attributes,omitempty" yaml:"attributes,omitempty" required:"false"`
	LinkAttributes     map[string]string `json:"linkAttributes,omitempty" yaml:"linkAttributes,omitempty" required:"false"`
	ChildrenAttributes map[string]string `json:"childrenAttributes,omitempty" yaml:"childrenAttributes,omitempty" required:"false"`
	LabelAttributes    map[string]string `json:"labelAttributes,omitempty" yaml:"labelAttributes,omitempty" required:"false"`
	Metadata           map[string]string `json:"metadata,omitempty" yaml:"metadata,omitempty" required:"false"`
	Created            time.Time         `json:"created,omitempty" yaml:"created,omitempty" required:"true"`
	Updated            time.Time         `json:"updated,omitempty" yaml:"updated,omitempty" required:"true"`
	Current            bool              `json:"-" yaml:"-"`
	Ancestor           bool              `json:"-" yaml:"-"`
	Parent             *Node             `json:"-" yaml:"-"`
	Menu               *Menu             `json:"-" yaml:"-"`
	Children           []*Node           `json:"-" yaml:"-"`
}

func (n Node) GetID() int64 {
	return n.ID
}

func (n Node) String() string {
	if n.Name == "" {
		return "n/a"
	}
	return n.Name
}

func (n Node) IsRoot() bool {
	return n.ParentID == 0
}

func (n Node) HasChildren() bool {
	return len(n.Children) > 0
}

func (n Node) WithFixedPathAndLevel() Node {
	if n.ParentID == 0 || n.Parent == nil {
		n.ParentID = 0
		n.Level = 0
		n.Path = fmt.Sprintf("%d", n.ID)
		return n
	}

	n.Level = n.Parent.Level + 1
	n.Path = fmt.Sprintf("%s/%d", n.Parent.Path, n.ID)
	return n
}
