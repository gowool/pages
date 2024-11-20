package model

var MetaTypes = []MetaType{MetaName, MetaEquiv, MetaProperty}

const (
	MetaName     = MetaType("name")
	MetaEquiv    = MetaType("http-equiv")
	MetaProperty = MetaType("property")
)

type MetaType string

func (t MetaType) IsZero() bool {
	return t == ""
}

func (t MetaType) String() string {
	return string(t)
}

type Meta struct {
	Type    MetaType `json:"type,omitempty" yaml:"type,omitempty" required:"true" enum:"name,http-equiv,property"`
	Key     string   `json:"key,omitempty" yaml:"key,omitempty" required:"true"`
	Content string   `json:"content,omitempty" yaml:"content,omitempty" required:"false"`
}

func (m Meta) Equal(another Meta) bool {
	return m.Type == another.Type && m.Key == another.Key
}
