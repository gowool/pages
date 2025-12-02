package pages

import "slices"

const DefaultCharset = "UTF-8"

type MetaTags struct {
	Charset   string              `json:"charset,omitempty" yaml:"charset,omitempty"`
	Name      map[string][]string `json:"name,omitempty" yaml:"name,omitempty"`
	Property  map[string][]string `json:"property,omitempty" yaml:"property,omitempty"`
	HTTPEquiv map[string][]string `json:"httpEquiv,omitempty" yaml:"httpEquiv,omitempty"`
}

func NewMetaTags(charset string, other ...*MetaTags) *MetaTags {
	if len(other) == 0 {
		return &MetaTags{
			Charset:   charset,
			Name:      make(map[string][]string),
			Property:  make(map[string][]string),
			HTTPEquiv: make(map[string][]string),
		}
	}

	mt := &MetaTags{Charset: charset}
	mt.Set(other[0])
	for _, m := range other[1:] {
		mt.Append(m)
	}
	return mt
}

func (m *MetaTags) With(other *MetaTags) *MetaTags {
	return NewMetaTags(m.Charset, m, other)
}

func (m *MetaTags) Set(other *MetaTags) {
	if other == nil {
		return
	}

	m.Charset = other.Charset

	m.Name = make(map[string][]string)
	for k, v := range other.Name {
		m.Name[k] = slices.Clone(v)
	}

	m.Property = make(map[string][]string)
	for k, v := range other.Property {
		m.Property[k] = slices.Clone(v)
	}

	m.HTTPEquiv = make(map[string][]string)
	for k, v := range other.HTTPEquiv {
		m.HTTPEquiv[k] = slices.Clone(v)
	}
}

func (m *MetaTags) Append(other *MetaTags) {
	if other == nil {
		return
	}

	for k, v := range other.Name {
		m.Name[k] = append(m.Name[k], v...)
	}

	for k, v := range other.Property {
		m.Property[k] = append(m.Property[k], v...)
	}

	for k, v := range other.HTTPEquiv {
		m.HTTPEquiv[k] = append(m.HTTPEquiv[k], v...)
	}
}

func (m *MetaTags) SetName(name string, content ...string) {
	m.Name[name] = content
}

func (m *MetaTags) AppendName(name string, content ...string) {
	m.Name[name] = append(m.Name[name], content...)
}

func (m *MetaTags) SetProperty(name string, content ...string) {
	m.Property[name] = content
}

func (m *MetaTags) AppendProperty(name string, content ...string) {
	m.Property[name] = append(m.Property[name], content...)
}

func (m *MetaTags) SetHTTPEquiv(name string, content ...string) {
	m.HTTPEquiv[name] = content
}

func (m *MetaTags) AppendHTTPEquiv(name string, content ...string) {
	m.HTTPEquiv[name] = append(m.HTTPEquiv[name], content...)
}
