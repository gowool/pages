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
	if charset == "" {
		charset = DefaultCharset
	}

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
	charset := DefaultCharset
	mt := make([]*MetaTags, 0, 2)

	if m != nil {
		charset = m.Charset
		mt = append(mt, m)
	}
	if other != nil {
		if other.Charset != "" {
			charset = other.Charset
		}
		mt = append(mt, other)
	}

	return NewMetaTags(charset, mt...)
}

func (m *MetaTags) Set(other *MetaTags) {
	if m == nil || other == nil {
		return
	}

	if other.Charset != "" {
		m.Charset = other.Charset
	}

	if len(other.Name) > 0 {
		m.Name = make(map[string][]string)
		for k, v := range other.Name {
			m.SetName(k, slices.Clone(v)...)
		}
	}

	if len(other.Property) > 0 {
		m.Property = make(map[string][]string)
		for k, v := range other.Property {
			m.SetProperty(k, slices.Clone(v)...)
		}
	}

	if len(other.HTTPEquiv) > 0 {
		m.HTTPEquiv = make(map[string][]string)
		for k, v := range other.HTTPEquiv {
			m.SetHTTPEquiv(k, slices.Clone(v)...)
		}
	}
}

func (m *MetaTags) Append(other *MetaTags) {
	if m == nil || other == nil {
		return
	}

	for k, v := range other.Name {
		m.AppendName(k, v...)
	}

	for k, v := range other.Property {
		m.AppendProperty(k, v...)
	}

	for k, v := range other.HTTPEquiv {
		m.AppendHTTPEquiv(k, v...)
	}
}

func (m *MetaTags) SetName(name string, content ...string) {
	if name != "" {
		m.Name[name] = content
	}
}

func (m *MetaTags) AppendName(name string, content ...string) {
	if name != "" {
		m.Name[name] = append(m.Name[name], content...)
	}
}

func (m *MetaTags) SetProperty(name string, content ...string) {
	if name != "" {
		m.Property[name] = content
	}
}

func (m *MetaTags) AppendProperty(name string, content ...string) {
	if name != "" {
		m.Property[name] = append(m.Property[name], content...)
	}
}

func (m *MetaTags) SetHTTPEquiv(name string, content ...string) {
	if name != "" {
		m.HTTPEquiv[name] = content
	}
}

func (m *MetaTags) AppendHTTPEquiv(name string, content ...string) {
	if name != "" {
		m.HTTPEquiv[name] = append(m.HTTPEquiv[name], content...)
	}
}
