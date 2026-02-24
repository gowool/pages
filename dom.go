package pages

import (
	"fmt"
	"html/template"
	"maps"
	"strings"

	"golang.org/x/net/html"
)

type Attr map[string]string

func NewAttr(attr ...string) (a Attr) {
	for i := 0; i < len(attr); i += 2 {
		a.Add(attr[i], attr[i+1])
	}
	return
}

func (attr *Attr) Add(key, value string) {
	if *attr == nil {
		*attr = make(Attr)
	}
	(*attr)[key] = value
}

func (attr *Attr) HTML() template.HTMLAttr {
	return template.HTMLAttr(attr.String())
}

func (attr *Attr) String() string {
	if *attr == nil {
		return ""
	}
	var b strings.Builder
	for key, value := range *attr {
		b.WriteByte(' ')
		b.WriteString(key)
		b.WriteString(`="`)
		b.WriteString(html.EscapeString(value))
		b.WriteByte('"')
	}
	return b.String()
}

func (attr *Attr) Copy() Attr {
	if *attr == nil {
		return nil
	}
	return maps.Clone(*attr)
}

func (attr *Attr) With(other Attr) Attr {
	if *attr == nil {
		return other.Copy()
	}

	newAttr := attr.Copy()
	maps.Copy(newAttr, other)
	return newAttr
}

type Node struct {
	Tag  string `env:"TAG" json:"tag,omitempty" yaml:"tag,omitempty"`
	Text string `env:"TEXT" json:"text,omitempty" yaml:"text,omitempty"`
	Attr Attr   `env:"ATTR" json:"attr,omitempty" yaml:"attr,omitempty"`
}

func (n Node) HTML() template.HTML {
	return template.HTML(n.String())
}

func (n Node) String() string {
	if n.Tag == "" {
		panic("tag is required")
	}
	if n.Text == "" {
		return fmt.Sprintf(`<%s%s />`, n.Tag, n.Attr.String())
	}
	return fmt.Sprintf(`<%s%s>%s</%s>`, n.Tag, n.Attr.String(), n.Text, n.Tag)
}

func (n Node) Copy() Node {
	return Node{
		Tag:  n.Tag,
		Text: n.Text,
		Attr: maps.Clone(n.Attr),
	}
}

type Nodes []Node

func (nodes Nodes) HTML() template.HTML {
	return template.HTML(nodes.String())
}

func (nodes Nodes) String() string {
	var b strings.Builder
	for _, node := range nodes {
		b.WriteString(node.String())
	}
	return b.String()
}

func (nodes Nodes) Copy() Nodes {
	newNodes := make(Nodes, len(nodes))
	for i, node := range nodes {
		newNodes[i] = node.Copy()
	}
	return newNodes
}

type Head struct {
	Attr  Attr  `envPrefix:"ATTR_" json:"attr,omitempty" yaml:"attr,omitempty"`
	Nodes Nodes `envPrefix:"NODE_" json:"nodes,omitempty" yaml:"nodes,omitempty"`
}

func (head *Head) Add(nodes ...Node) {
	head.Nodes = append(head.Nodes, nodes...)
}

func (head *Head) Copy() Head {
	return Head{
		Attr:  head.Attr.Copy(),
		Nodes: head.Nodes.Copy(),
	}
}

func (head *Head) With(other Head) (newHead Head) {
	newHead.Attr = head.Attr.With(other.Attr)
	newHead.Nodes = make(Nodes, 0, len(head.Nodes)+len(other.Nodes))
	for _, node := range head.Nodes {
		newHead.Nodes = append(newHead.Nodes, node.Copy())
	}
	for _, node := range other.Nodes {
		newHead.Nodes = append(newHead.Nodes, node.Copy())
	}
	return
}

type DOM struct {
	HTML struct {
		Attr Attr `envPrefix:"ATTR_" json:"attr,omitempty" yaml:"attr,omitempty"`
	} `envPrefix:"HTML_" json:"html,omitempty" yaml:"html,omitempty"`

	Head Head `envPrefix:"HEAD_" json:"head,omitempty" yaml:"head,omitempty"`

	Body struct {
		Attr Attr `envPrefix:"ATTR_" json:"attr,omitempty" yaml:"attr,omitempty"`
	} `envPrefix:"BODY_" json:"body,omitempty" yaml:"body,omitempty"`
}

func (dom DOM) Copy() (newDOM DOM) {
	newDOM.HTML.Attr = dom.HTML.Attr.Copy()
	newDOM.Body.Attr = dom.Body.Attr.Copy()
	newDOM.Head = dom.Head.Copy()
	return
}

func (dom DOM) With(other DOM) (newDOM DOM) {
	newDOM.HTML.Attr = dom.HTML.Attr.With(other.HTML.Attr)
	newDOM.Body.Attr = dom.Body.Attr.With(other.Body.Attr)
	newDOM.Head = dom.Head.With(other.Head)
	return
}

func TitleNode(text string) Node {
	return Node{
		Tag:  "title",
		Text: text,
	}
}

func MetaNode(attr ...string) Node {
	return Node{
		Tag:  "meta",
		Attr: NewAttr(attr...),
	}
}

const DefaultCharset = "UTF-8"

func CharsetMetaNode(charset string) Node {
	return MetaNode("charset", charset)
}

func NameMetaNode(name, content string) Node {
	return MetaNode("name", name, "content", content)
}

func PropertyMetaNode(property, content string) Node {
	return MetaNode("property", property, "content", content)
}

func HTTPEquivMetaNode(name, content string) Node {
	return MetaNode("http-equiv", name, "content", content)
}

const (
	LinkRelAlternate  = "alternate"
	LinkRelAuthor     = "author"
	LinkRelCanonical  = "canonical"
	LinkRelLicense    = "license"
	LinkRelNext       = "next"
	LinkRelPrev       = "prev"
	LinkRelStylesheet = "stylesheet"
	LinkRelIcon       = "icon"
)

const (
	ReferrerPolicyNoReferrer              = "no-referrer"
	ReferrerPolicyNoReferrerWhenDowngrade = "no-referrer-when-downgrade"
	ReferrerPolicyOrigin                  = "origin"
	ReferrerPolicyOriginWhenCrossOrigin   = "origin-when-cross-origin"
	ReferrerPolicySameOrigin              = "same-origin"
	ReferrerPolicyStrictOrigin            = "strict-origin"
	ReferrerPolicyUnsafeUrl               = "unsafe-url"
)

type HeadLink struct {
	// CrossOrigin Specifies how the element handles cross-origin requests
	CrossOrigin string
	// Href Specifies the location of the linked document
	Href string
	// HrefLang Specifies the language of the text in the linked document
	HrefLang string
	// Media Specifies on what device the linked document will be displayed
	Media string
	// Rel REQUIRED Specifies the relationship between the current document and the linked document
	Rel string
	// Sizes Specifies the size of the linked resource. Only for rel="icon"
	Sizes string
	// Title Defines a preferred or an alternate stylesheet
	Title string
	// Type Specifies the media type of the linked document
	Type string
}

func HeadLinkNode(link HeadLink) (node Node) {
	node.Tag = "link"
	node.Attr.Add("rel", link.Rel)

	if link.CrossOrigin != "" {
		node.Attr.Add("crossorigin", link.CrossOrigin)
	}

	if link.Href != "" {
		node.Attr.Add("href", link.Href)
	}

	if link.HrefLang != "" {
		node.Attr.Add("hreflang", link.HrefLang)
	}

	if link.Media != "" {
		node.Attr.Add("media", link.Media)
	}

	if link.Sizes != "" {
		node.Attr.Add("sizes", link.Sizes)
	}

	if link.Title != "" {
		node.Attr.Add("title", link.Title)
	}

	if link.Type != "" {
		node.Attr.Add("type", link.Type)
	}

	return
}
