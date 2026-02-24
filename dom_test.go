package pages

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAttr(t *testing.T) {
	t.Run("NewAttr", func(t *testing.T) {
		attr := NewAttr("class", "hero", "data-id", "42")

		assert.Equal(t, "hero", attr["class"])
		assert.Equal(t, "42", attr["data-id"])
	})

	t.Run("Add initializes map", func(t *testing.T) {
		var attr Attr
		attr.Add("lang", "en")

		assert.Equal(t, Attr{"lang": "en"}, attr)
	})

	t.Run("String on nil map", func(t *testing.T) {
		var attr Attr
		assert.Equal(t, "", attr.String())
		assert.Equal(t, "", string(attr.HTML()))
	})

	t.Run("String escapes values", func(t *testing.T) {
		attr := NewAttr("title", `a "quoted" & <tag>`)
		got := attr.String()

		assert.Contains(t, got, ` title="a &#34;quoted&#34; &amp; &lt;tag&gt;"`)
		assert.Equal(t, got, string(attr.HTML()))
	})

	t.Run("Copy", func(t *testing.T) {
		t.Run("nil map", func(t *testing.T) {
			var attr Attr
			assert.Nil(t, attr.Copy())
		})

		t.Run("clones map", func(t *testing.T) {
			attr := NewAttr("a", "1")
			cp := attr.Copy()
			cp["a"] = "2"

			assert.Equal(t, "1", attr["a"])
			assert.Equal(t, "2", cp["a"])
		})
	})

	t.Run("With", func(t *testing.T) {
		t.Run("nil receiver returns copy of other", func(t *testing.T) {
			var base Attr
			other := NewAttr("x", "y")

			got := base.With(other)
			got["x"] = "z"

			assert.Equal(t, "y", other["x"])
			assert.Equal(t, "z", got["x"])
		})

		t.Run("merge with override", func(t *testing.T) {
			base := NewAttr("a", "1", "b", "2")
			other := NewAttr("b", "3", "c", "4")

			got := base.With(other)

			assert.Equal(t, "1", got["a"])
			assert.Equal(t, "3", got["b"])
			assert.Equal(t, "4", got["c"])
			assert.Equal(t, "2", base["b"])
		})
	})
}

func TestNode(t *testing.T) {
	t.Run("String requires tag", func(t *testing.T) {
		assert.PanicsWithValue(t, "tag is required", func() {
			_ = (Node{}).String()
		})
	})

	t.Run("self closing", func(t *testing.T) {
		node := Node{
			Tag:  "meta",
			Attr: NewAttr("charset", "UTF-8"),
		}

		got := node.String()
		assert.Contains(t, got, "<meta")
		assert.Contains(t, got, ` charset="UTF-8"`)
		assert.Contains(t, got, " />")
		assert.Equal(t, got, string(node.HTML()))
	})

	t.Run("open close tag", func(t *testing.T) {
		node := Node{Tag: "title", Text: "Hello"}

		assert.Equal(t, "<title>Hello</title>", node.String())
		assert.Equal(t, "<title>Hello</title>", string(node.HTML()))
	})

	t.Run("Copy", func(t *testing.T) {
		node := Node{
			Tag:  "meta",
			Text: "x",
			Attr: NewAttr("name", "description"),
		}

		cp := node.Copy()
		cp.Attr["name"] = "keywords"

		assert.Equal(t, "description", node.Attr["name"])
		assert.Equal(t, "keywords", cp.Attr["name"])
		assert.Equal(t, node.Tag, cp.Tag)
		assert.Equal(t, node.Text, cp.Text)
	})
}

func TestNodes(t *testing.T) {
	nodes := Nodes{
		{Tag: "title", Text: "A"},
		MetaNode("charset", "UTF-8"),
	}

	got := nodes.String()
	assert.Contains(t, got, "<title>A</title>")
	assert.Contains(t, got, "<meta")
	assert.Equal(t, got, string(nodes.HTML()))

	cp := nodes.Copy()
	cp[0].Text = "B"
	cp[1].Attr["charset"] = "ISO-8859-1"

	assert.Equal(t, "A", nodes[0].Text)
	assert.Equal(t, "UTF-8", nodes[1].Attr["charset"])
}

func TestHead(t *testing.T) {
	head := Head{
		Attr:  NewAttr("data-a", "1"),
		Nodes: Nodes{{Tag: "title", Text: "A"}},
	}

	head.Add(Node{Tag: "meta", Attr: NewAttr("charset", "UTF-8")})
	assert.Len(t, head.Nodes, 2)

	cp := head.Copy()
	cp.Attr["data-a"] = "2"
	cp.Nodes[0].Text = "B"

	assert.Equal(t, "1", head.Attr["data-a"])
	assert.Equal(t, "A", head.Nodes[0].Text)

	other := Head{
		Attr:  NewAttr("data-a", "3", "data-b", "4"),
		Nodes: Nodes{{Tag: "script", Text: "console.log(1)"}},
	}

	merged := head.With(other)
	assert.Equal(t, "3", merged.Attr["data-a"])
	assert.Equal(t, "4", merged.Attr["data-b"])
	assert.Len(t, merged.Nodes, 3)

	merged.Nodes[0].Text = "changed"
	assert.Equal(t, "A", head.Nodes[0].Text)
}

func TestDOM(t *testing.T) {
	base := DOM{}
	base.HTML.Attr = NewAttr("lang", "en")
	base.Body.Attr = NewAttr("class", "base")
	base.Head = Head{
		Attr:  NewAttr("data-head", "a"),
		Nodes: Nodes{{Tag: "title", Text: "Base"}},
	}

	other := DOM{}
	other.HTML.Attr = NewAttr("lang", "fr", "dir", "ltr")
	other.Body.Attr = NewAttr("class", "other", "id", "app")
	other.Head = Head{
		Attr:  NewAttr("data-head", "b", "x", "y"),
		Nodes: Nodes{{Tag: "meta", Attr: NewAttr("charset", "UTF-8")}},
	}

	cp := base.Copy()
	cp.HTML.Attr["lang"] = "de"
	cp.Body.Attr["class"] = "changed"
	cp.Head.Attr["data-head"] = "changed"
	cp.Head.Nodes[0].Text = "Changed"

	assert.Equal(t, "en", base.HTML.Attr["lang"])
	assert.Equal(t, "base", base.Body.Attr["class"])
	assert.Equal(t, "a", base.Head.Attr["data-head"])
	assert.Equal(t, "Base", base.Head.Nodes[0].Text)

	merged := base.With(other)
	assert.Equal(t, "fr", merged.HTML.Attr["lang"])
	assert.Equal(t, "ltr", merged.HTML.Attr["dir"])
	assert.Equal(t, "other", merged.Body.Attr["class"])
	assert.Equal(t, "app", merged.Body.Attr["id"])
	assert.Equal(t, "b", merged.Head.Attr["data-head"])
	assert.Equal(t, "y", merged.Head.Attr["x"])
	assert.Len(t, merged.Head.Nodes, 2)
}

func TestNodeFactories(t *testing.T) {
	title := TitleNode("Hello")
	assert.Equal(t, Node{Tag: "title", Text: "Hello"}, title)

	meta := MetaNode("name", "description", "content", "about")
	assert.Equal(t, "meta", meta.Tag)
	assert.Equal(t, "description", meta.Attr["name"])
	assert.Equal(t, "about", meta.Attr["content"])

	charset := CharsetMetaNode(DefaultCharset)
	assert.Equal(t, "UTF-8", charset.Attr["charset"])

	name := NameMetaNode("description", "content")
	assert.Equal(t, "description", name.Attr["name"])
	assert.Equal(t, "content", name.Attr["content"])

	property := PropertyMetaNode("og:title", "Title")
	assert.Equal(t, "og:title", property.Attr["property"])
	assert.Equal(t, "Title", property.Attr["content"])

	httpEquiv := HTTPEquivMetaNode("refresh", "30")
	assert.Equal(t, "refresh", httpEquiv.Attr["http-equiv"])
	assert.Equal(t, "30", httpEquiv.Attr["content"])
}

func TestHeadLinkNode(t *testing.T) {
	link := HeadLink{
		Rel:         LinkRelCanonical,
		CrossOrigin: "anonymous",
		Href:        "https://example.com",
		HrefLang:    "en",
		Media:       "screen",
		Sizes:       "16x16",
		Title:       "Canonical",
		Type:        "text/html",
	}

	node := HeadLinkNode(link)

	assert.Equal(t, "link", node.Tag)
	assert.Equal(t, LinkRelCanonical, node.Attr["rel"])
	assert.Equal(t, "anonymous", node.Attr["crossorigin"])
	assert.Equal(t, "https://example.com", node.Attr["href"])
	assert.Equal(t, "en", node.Attr["hreflang"])
	assert.Equal(t, "screen", node.Attr["media"])
	assert.Equal(t, "16x16", node.Attr["sizes"])
	assert.Equal(t, "Canonical", node.Attr["title"])
	assert.Equal(t, "text/html", node.Attr["type"])

	minimal := HeadLinkNode(HeadLink{Rel: LinkRelNext})
	assert.Equal(t, Attr{"rel": LinkRelNext}, minimal.Attr)
}
