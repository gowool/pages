package pages

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPage_AbsURL(t *testing.T) {
	type fields struct {
		ID        ID
		SiteID    ID
		Site      *Site
		ParentID  *ID
		Parent    *Page
		Children  []*Page
		Created   time.Time
		Updated   time.Time
		Status    Status
		MetaTags  *MetaTags
		Metadata  map[string]any
		Header    map[string][]string
		Name      string
		Title     string
		Pattern   string
		Alias     string
		Slug      string
		URL       string
		CustomURL string
		Template  string
		Position  int
		Decorate  bool
	}
	type args struct {
		args []any
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name:   "Error page URL",
			fields: fields{Pattern: PageError5xx},
			want:   "",
		},
		{
			name:   "Internal page URL",
			fields: fields{Pattern: PageInternalCreate},
			want:   "",
		},
		{
			name:   "CMS home without site",
			fields: fields{Pattern: PageCMS, URL: "/"},
			want:   "/",
		},
		{
			name:   "CMS home with site",
			fields: fields{Pattern: PageCMS, URL: "/", Site: NewSite()},
			want:   "https://localhost",
		},
		{
			name: "CMS /en/foo/boo",
			fields: fields{Pattern: PageCMS, URL: "/foo/boo", Site: func() *Site {
				s := NewSite()
				s.RelativePath = "/en"
				return s
			}()},
			want: "https://localhost/en/foo/boo",
		},
		{
			name:   "CMS with args",
			fields: fields{Pattern: PageCMS, URL: "/foo/boo", Site: NewSite()},
			args:   args{args: []any{"a", "1", "b", "2"}},
			want:   "https://localhost/foo/boo?a=1&b=2",
		},
		{
			name:   "Pattern /foo/boo",
			fields: fields{Pattern: "/foo/boo", Site: NewSite()},
			want:   "https://localhost/foo/boo",
		},
		{
			name:   "Pattern /foo/{name}",
			fields: fields{Pattern: "/foo/{name}", Site: NewSite()},
			args:   args{args: []any{"{name}", "boo"}},
			want:   "https://localhost/foo/boo",
		},
		{
			name:   "Pattern /foo/{name} with query args",
			fields: fields{Pattern: "/foo/{name}", Site: NewSite()},
			args:   args{args: []any{"{name}", "boo", "a", "1", "b", "2"}},
			want:   "https://localhost/foo/boo?a=1&b=2",
		},
		{
			name:   "Pattern /test/{name}/{name2}",
			fields: fields{Pattern: "/test/{name}/{name2}", Site: NewSite()},
			args:   args{args: []any{"{name}", "foo", "{name2}", "boo"}},
			want:   "https://localhost/test/foo/boo",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Page{
				ID:        tt.fields.ID,
				SiteID:    tt.fields.SiteID,
				Site:      tt.fields.Site,
				ParentID:  tt.fields.ParentID,
				Parent:    tt.fields.Parent,
				Children:  tt.fields.Children,
				Created:   tt.fields.Created,
				Updated:   tt.fields.Updated,
				Status:    tt.fields.Status,
				MetaTags:  tt.fields.MetaTags,
				Metadata:  tt.fields.Metadata,
				Header:    tt.fields.Header,
				Name:      tt.fields.Name,
				Title:     tt.fields.Title,
				Pattern:   tt.fields.Pattern,
				Alias:     tt.fields.Alias,
				Slug:      tt.fields.Slug,
				URL:       tt.fields.URL,
				CustomURL: tt.fields.CustomURL,
				Template:  tt.fields.Template,
				Position:  tt.fields.Position,
				Decorate:  tt.fields.Decorate,
			}
			got := p.AbsURL(tt.args.args...)
			assert.Equal(t, tt.want, got, "AbsURL() should return the expected URL")
		})
	}
}

func TestPage_FixURL(t *testing.T) {
	type fields struct {
		ID        ID
		SiteID    ID
		Site      *Site
		ParentID  *ID
		Parent    *Page
		Children  []*Page
		Created   time.Time
		Updated   time.Time
		Status    Status
		MetaTags  *MetaTags
		Metadata  map[string]any
		Header    map[string][]string
		Name      string
		Title     string
		Pattern   string
		Alias     string
		Slug      string
		URL       string
		CustomURL string
		Template  string
		Position  int
		Decorate  bool
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{name: "Internal create", fields: fields{Pattern: PageInternalCreate}, want: ""},
		{name: "Hybrid", fields: fields{Pattern: "/foo/boo"}, want: ""},
		{name: "CMS home", fields: fields{Pattern: PageCMS}, want: "/"},
		{name: "CMS custom url: /foo", fields: fields{Pattern: PageCMS, CustomURL: "/foo"}, want: "/foo"},
		{name: "CMS slug: /foo", fields: fields{Pattern: PageCMS, Slug: "/foo"}, want: "/foo"},
		{name: "CMS name: /foo", fields: fields{Pattern: PageCMS, Name: "Foo"}, want: "/foo"},
		{name: "CMS with parent: /parent/foo", fields: fields{Pattern: PageCMS, Name: "Foo", Parent: &Page{Pattern: PageCMS, URL: "/parent"}}, want: "/parent/foo"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Page{
				ID:        tt.fields.ID,
				SiteID:    tt.fields.SiteID,
				Site:      tt.fields.Site,
				ParentID:  tt.fields.ParentID,
				Parent:    tt.fields.Parent,
				Children:  tt.fields.Children,
				Created:   tt.fields.Created,
				Updated:   tt.fields.Updated,
				Status:    tt.fields.Status,
				MetaTags:  tt.fields.MetaTags,
				Metadata:  tt.fields.Metadata,
				Header:    tt.fields.Header,
				Name:      tt.fields.Name,
				Title:     tt.fields.Title,
				Pattern:   tt.fields.Pattern,
				Alias:     tt.fields.Alias,
				Slug:      tt.fields.Slug,
				URL:       tt.fields.URL,
				CustomURL: tt.fields.CustomURL,
				Template:  tt.fields.Template,
				Position:  tt.fields.Position,
				Decorate:  tt.fields.Decorate,
			}
			p.FixURL()
			assert.Equal(t, tt.want, p.URL, "URL should be correctly fixed")
		})
	}
}

func TestPage_IsCMS(t *testing.T) {
	type fields struct {
		ID        ID
		SiteID    ID
		Site      *Site
		ParentID  *ID
		Parent    *Page
		Children  []*Page
		Created   time.Time
		Updated   time.Time
		Status    Status
		MetaTags  *MetaTags
		Metadata  map[string]any
		Header    map[string][]string
		Name      string
		Title     string
		Pattern   string
		Alias     string
		Slug      string
		URL       string
		CustomURL string
		Template  string
		Position  int
		Decorate  bool
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{name: "CMS", fields: fields{Pattern: PageCMS}, want: true},
		{name: "Internal create", fields: fields{Pattern: PageInternalCreate}, want: false},
		{name: "Error", fields: fields{Pattern: PageError5xx}, want: false},
		{name: "Dynamic", fields: fields{Pattern: "/test/{foo}"}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Page{
				ID:        tt.fields.ID,
				SiteID:    tt.fields.SiteID,
				Site:      tt.fields.Site,
				ParentID:  tt.fields.ParentID,
				Parent:    tt.fields.Parent,
				Children:  tt.fields.Children,
				Created:   tt.fields.Created,
				Updated:   tt.fields.Updated,
				Status:    tt.fields.Status,
				MetaTags:  tt.fields.MetaTags,
				Metadata:  tt.fields.Metadata,
				Header:    tt.fields.Header,
				Name:      tt.fields.Name,
				Title:     tt.fields.Title,
				Pattern:   tt.fields.Pattern,
				Alias:     tt.fields.Alias,
				Slug:      tt.fields.Slug,
				URL:       tt.fields.URL,
				CustomURL: tt.fields.CustomURL,
				Template:  tt.fields.Template,
				Position:  tt.fields.Position,
				Decorate:  tt.fields.Decorate,
			}
			assert.Equal(t, tt.want, p.IsCMS(), "IsCMS() should return the expected boolean value")
		})
	}
}

func TestPage_IsDynamic(t *testing.T) {
	type fields struct {
		ID        ID
		SiteID    ID
		Site      *Site
		ParentID  *ID
		Parent    *Page
		Children  []*Page
		Created   time.Time
		Updated   time.Time
		Status    Status
		MetaTags  *MetaTags
		Metadata  map[string]any
		Header    map[string][]string
		Name      string
		Title     string
		Pattern   string
		Alias     string
		Slug      string
		URL       string
		CustomURL string
		Template  string
		Position  int
		Decorate  bool
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{name: "Dynamic", fields: fields{Pattern: "/test/{foo}"}, want: true},
		{name: "No Dynamic", fields: fields{Pattern: "/test/foo"}, want: false},
		{name: "CMS", fields: fields{Pattern: PageCMS}, want: false},
		{name: "Internal create", fields: fields{Pattern: PageInternalCreate}, want: false},
		{name: "Error", fields: fields{Pattern: PageError5xx}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Page{
				ID:        tt.fields.ID,
				SiteID:    tt.fields.SiteID,
				Site:      tt.fields.Site,
				ParentID:  tt.fields.ParentID,
				Parent:    tt.fields.Parent,
				Children:  tt.fields.Children,
				Created:   tt.fields.Created,
				Updated:   tt.fields.Updated,
				Status:    tt.fields.Status,
				MetaTags:  tt.fields.MetaTags,
				Metadata:  tt.fields.Metadata,
				Header:    tt.fields.Header,
				Name:      tt.fields.Name,
				Title:     tt.fields.Title,
				Pattern:   tt.fields.Pattern,
				Alias:     tt.fields.Alias,
				Slug:      tt.fields.Slug,
				URL:       tt.fields.URL,
				CustomURL: tt.fields.CustomURL,
				Template:  tt.fields.Template,
				Position:  tt.fields.Position,
				Decorate:  tt.fields.Decorate,
			}
			assert.Equal(t, tt.want, p.IsDynamic(), "IsDynamic() should return the expected boolean value")
		})
	}
}

func TestPage_IsError(t *testing.T) {
	type fields struct {
		ID        ID
		SiteID    ID
		Site      *Site
		ParentID  *ID
		Parent    *Page
		Children  []*Page
		Created   time.Time
		Updated   time.Time
		Status    Status
		MetaTags  *MetaTags
		Metadata  map[string]any
		Header    map[string][]string
		Name      string
		Title     string
		Pattern   string
		Alias     string
		Slug      string
		URL       string
		CustomURL string
		Template  string
		Position  int
		Decorate  bool
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{name: "Error", fields: fields{Pattern: PageError5xx}, want: true},
		{name: "Error (foo)", fields: fields{Pattern: PageErrorPrefix + "foo"}, want: true},
		{name: "Dynamic", fields: fields{Pattern: "/test/{foo}"}, want: false},
		{name: "No Dynamic", fields: fields{Pattern: "/test/foo"}, want: false},
		{name: "CMS", fields: fields{Pattern: PageCMS}, want: false},
		{name: "Internal create", fields: fields{Pattern: PageInternalCreate}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Page{
				ID:        tt.fields.ID,
				SiteID:    tt.fields.SiteID,
				Site:      tt.fields.Site,
				ParentID:  tt.fields.ParentID,
				Parent:    tt.fields.Parent,
				Children:  tt.fields.Children,
				Created:   tt.fields.Created,
				Updated:   tt.fields.Updated,
				Status:    tt.fields.Status,
				MetaTags:  tt.fields.MetaTags,
				Metadata:  tt.fields.Metadata,
				Header:    tt.fields.Header,
				Name:      tt.fields.Name,
				Title:     tt.fields.Title,
				Pattern:   tt.fields.Pattern,
				Alias:     tt.fields.Alias,
				Slug:      tt.fields.Slug,
				URL:       tt.fields.URL,
				CustomURL: tt.fields.CustomURL,
				Template:  tt.fields.Template,
				Position:  tt.fields.Position,
				Decorate:  tt.fields.Decorate,
			}
			assert.Equal(t, tt.want, p.IsError(), "IsError() should return the expected boolean value")
		})
	}
}

func TestPage_IsHybrid(t *testing.T) {
	type fields struct {
		ID        ID
		SiteID    ID
		Site      *Site
		ParentID  *ID
		Parent    *Page
		Children  []*Page
		Created   time.Time
		Updated   time.Time
		Status    Status
		MetaTags  *MetaTags
		Metadata  map[string]any
		Header    map[string][]string
		Name      string
		Title     string
		Pattern   string
		Alias     string
		Slug      string
		URL       string
		CustomURL string
		Template  string
		Position  int
		Decorate  bool
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{name: "Error", fields: fields{Pattern: PageError5xx}, want: false},
		{name: "Error (foo)", fields: fields{Pattern: PageErrorPrefix + "foo"}, want: false},
		{name: "Dynamic", fields: fields{Pattern: "/test/{foo}"}, want: true},
		{name: "No Dynamic", fields: fields{Pattern: "/test/foo"}, want: true},
		{name: "CMS", fields: fields{Pattern: PageCMS}, want: false},
		{name: "Internal create", fields: fields{Pattern: PageInternalCreate}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Page{
				ID:        tt.fields.ID,
				SiteID:    tt.fields.SiteID,
				Site:      tt.fields.Site,
				ParentID:  tt.fields.ParentID,
				Parent:    tt.fields.Parent,
				Children:  tt.fields.Children,
				Created:   tt.fields.Created,
				Updated:   tt.fields.Updated,
				Status:    tt.fields.Status,
				MetaTags:  tt.fields.MetaTags,
				Metadata:  tt.fields.Metadata,
				Header:    tt.fields.Header,
				Name:      tt.fields.Name,
				Title:     tt.fields.Title,
				Pattern:   tt.fields.Pattern,
				Alias:     tt.fields.Alias,
				Slug:      tt.fields.Slug,
				URL:       tt.fields.URL,
				CustomURL: tt.fields.CustomURL,
				Template:  tt.fields.Template,
				Position:  tt.fields.Position,
				Decorate:  tt.fields.Decorate,
			}
			assert.Equal(t, tt.want, p.IsHybrid(), "IsHybrid() should return the expected boolean value")
		})
	}
}

func TestPage_IsInternal(t *testing.T) {
	type fields struct {
		ID        ID
		SiteID    ID
		Site      *Site
		ParentID  *ID
		Parent    *Page
		Children  []*Page
		Created   time.Time
		Updated   time.Time
		Status    Status
		MetaTags  *MetaTags
		Metadata  map[string]any
		Header    map[string][]string
		Name      string
		Title     string
		Pattern   string
		Alias     string
		Slug      string
		URL       string
		CustomURL string
		Template  string
		Position  int
		Decorate  bool
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{name: "Error", fields: fields{Pattern: PageError5xx}, want: true},
		{name: "Error (foo)", fields: fields{Pattern: PageErrorPrefix + "foo"}, want: true},
		{name: "Dynamic", fields: fields{Pattern: "/test/{foo}"}, want: false},
		{name: "No Dynamic", fields: fields{Pattern: "/test/foo"}, want: false},
		{name: "CMS", fields: fields{Pattern: PageCMS}, want: false},
		{name: "Internal create", fields: fields{Pattern: PageInternalCreate}, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Page{
				ID:        tt.fields.ID,
				SiteID:    tt.fields.SiteID,
				Site:      tt.fields.Site,
				ParentID:  tt.fields.ParentID,
				Parent:    tt.fields.Parent,
				Children:  tt.fields.Children,
				Created:   tt.fields.Created,
				Updated:   tt.fields.Updated,
				Status:    tt.fields.Status,
				MetaTags:  tt.fields.MetaTags,
				Metadata:  tt.fields.Metadata,
				Header:    tt.fields.Header,
				Name:      tt.fields.Name,
				Title:     tt.fields.Title,
				Pattern:   tt.fields.Pattern,
				Alias:     tt.fields.Alias,
				Slug:      tt.fields.Slug,
				URL:       tt.fields.URL,
				CustomURL: tt.fields.CustomURL,
				Template:  tt.fields.Template,
				Position:  tt.fields.Position,
				Decorate:  tt.fields.Decorate,
			}
			assert.Equal(t, tt.want, p.IsInternal(), "IsInternal() should return the expected boolean value")
		})
	}
}

func TestPage_SetAlias(t *testing.T) {
	type fields struct {
		ID        ID
		SiteID    ID
		Site      *Site
		ParentID  *ID
		Parent    *Page
		Children  []*Page
		Created   time.Time
		Updated   time.Time
		Status    Status
		MetaTags  *MetaTags
		Metadata  map[string]any
		Header    map[string][]string
		Name      string
		Title     string
		Pattern   string
		Alias     string
		Slug      string
		URL       string
		CustomURL string
		Template  string
		Position  int
		Decorate  bool
	}
	type args struct {
		alias string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{name: "foo", args: args{alias: "foo"}, want: PageAliasPrefix + "foo"},
		{name: PageAliasPrefix + "boo", args: args{alias: PageAliasPrefix + "boo"}, want: PageAliasPrefix + "boo"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Page{
				ID:        tt.fields.ID,
				SiteID:    tt.fields.SiteID,
				Site:      tt.fields.Site,
				ParentID:  tt.fields.ParentID,
				Parent:    tt.fields.Parent,
				Children:  tt.fields.Children,
				Created:   tt.fields.Created,
				Updated:   tt.fields.Updated,
				Status:    tt.fields.Status,
				MetaTags:  tt.fields.MetaTags,
				Metadata:  tt.fields.Metadata,
				Header:    tt.fields.Header,
				Name:      tt.fields.Name,
				Title:     tt.fields.Title,
				Pattern:   tt.fields.Pattern,
				Alias:     tt.fields.Alias,
				Slug:      tt.fields.Slug,
				URL:       tt.fields.URL,
				CustomURL: tt.fields.CustomURL,
				Template:  tt.fields.Template,
				Position:  tt.fields.Position,
				Decorate:  tt.fields.Decorate,
			}
			p.SetAlias(tt.args.alias)
			assert.Equal(t, tt.want, p.Alias, "Alias should be correctly set")
		})
	}
}

func TestPage_Copy(t *testing.T) {
	parentID := ID("parent-id")
	childID := ID("child-id")
	siteID := ID("site-id")

	originalTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)

	// Create parent page
	parent := &Page{
		ID:       parentID,
		Name:     "Parent Page",
		Pattern:  PageCMS,
		URL:      "/parent",
		Created:  originalTime,
		Updated:  originalTime,
		Metadata: map[string]any{"key": "parent-value"},
		Header:   map[string][]string{"X-Parent": {"value1", "value2"}},
		MetaTags: NewMetaTags("utf-8"),
	}

	// Create child page with all fields populated
	child := &Page{
		ID:        childID,
		SiteID:    siteID,
		ParentID:  &parentID,
		Parent:    parent,
		Children:  []*Page{},
		Created:   originalTime,
		Updated:   originalTime,
		Status:    Published,
		MetaTags:  NewMetaTags("iso-8859-1"),
		Metadata:  map[string]any{"key1": "value1", "key2": 42, "key3": true},
		Header:    map[string][]string{"Content-Type": {"text/html"}, "X-Custom": {"custom-value"}},
		Name:      "Test Page",
		Title:     "Test Title",
		Pattern:   "/test/{slug}",
		Alias:     "_page_alias_test",
		Slug:      "test-page",
		URL:       "/test/page",
		CustomURL: "/custom/test",
		Template:  "test-template",
		Position:  5,
		Decorate:  true,
	}

	// Add grandchild for testing nested children
	grandchild := &Page{
		ID:       ID("grandchild-id"),
		Name:     "Grandchild Page",
		Pattern:  PageCMS,
		Created:  originalTime,
		Updated:  originalTime,
		Metadata: map[string]any{"nested": "deep"},
	}
	child.Children = append(child.Children, grandchild)

	tests := []struct {
		name     string
		page     *Page
		validate func(*testing.T, *Page, *Page)
	}{
		{
			name: "Complete page with all fields",
			page: child,
			validate: func(t *testing.T, original, copied *Page) {
				// Verify basic fields are copied but not the same reference
				assert.NotSame(t, original, copied, "Copied page should be a new instance")
				assert.Equal(t, original.ID, copied.ID, "ID should be copied")
				assert.Equal(t, original.SiteID, copied.SiteID, "SiteID should be copied")
				assert.Equal(t, original.Name, copied.Name, "Name should be copied")
				assert.Equal(t, original.Title, copied.Title, "Title should be copied")
				assert.Equal(t, original.Pattern, copied.Pattern, "Pattern should be copied")
				assert.Equal(t, original.Alias, copied.Alias, "Alias should be copied")
				assert.Equal(t, original.Slug, copied.Slug, "Slug should be copied")
				assert.Equal(t, original.URL, copied.URL, "URL should be copied")
				assert.Equal(t, original.CustomURL, copied.CustomURL, "CustomURL should be copied")
				assert.Equal(t, original.Template, copied.Template, "Template should be copied")
				assert.Equal(t, original.Position, copied.Position, "Position should be copied")
				assert.Equal(t, original.Decorate, copied.Decorate, "Decorate should be copied")
				assert.Equal(t, original.Status, copied.Status, "Status should be copied")

				// Verify ParentID is properly cloned
				assert.NotSame(t, original.ParentID, copied.ParentID, "ParentID should be a new pointer")
				assert.Equal(t, *original.ParentID, *copied.ParentID, "ParentID value should be copied")

				// Verify complex objects are cloned but not the same reference
				assert.NotEqual(t, fmt.Sprintf("%p", original.Metadata), fmt.Sprintf("%p", copied.Metadata), "Metadata should be cloned")
				assert.Equal(t, original.Metadata, copied.Metadata, "Metadata content should be copied")

				assert.NotEqual(t, fmt.Sprintf("%p", original.Header), fmt.Sprintf("%p", copied.Header), "Header should be cloned")
				assert.Equal(t, original.Header, copied.Header, "Header content should be copied")

				assert.NotSame(t, original.MetaTags, copied.MetaTags, "MetaTags should be cloned")
				assert.Equal(t, original.MetaTags.Charset, copied.MetaTags.Charset, "MetaTags charset should be copied")

				// Verify nested objects are cloned
				assert.NotSame(t, original.Parent, copied.Parent, "Parent should be cloned")
				assert.Equal(t, original.Parent.Name, copied.Parent.Name, "Parent content should be copied")

				// Verify children are properly copied
				if original.Children != nil && copied.Children != nil {
					assert.NotEqual(t, fmt.Sprintf("%p", original.Children), fmt.Sprintf("%p", copied.Children), "Children slice should be new")
					assert.Len(t, copied.Children, len(original.Children), "Children count should be preserved")
					if len(original.Children) > 0 {
						assert.NotSame(t, original.Children[0], copied.Children[0], "Child should be cloned")
						assert.Equal(t, original.Children[0].Name, copied.Children[0].Name, "Child content should be copied")
					}
				}
			},
		},
		{
			name: "Page with nil fields",
			page: &Page{
				ID:       ID("minimal"),
				Name:     "Minimal",
				Pattern:  PageCMS,
				ParentID: nil,
				Parent:   nil,
				Site:     nil,
				MetaTags: nil,
				Header:   nil,
				Children: []*Page{},
			},
			validate: func(t *testing.T, original, copied *Page) {
				assert.Nil(t, copied.ParentID, "ParentID should remain nil")
				assert.Nil(t, copied.Parent, "Parent should remain nil")
				assert.Nil(t, copied.Site, "Site should remain nil")
				assert.Nil(t, copied.MetaTags, "MetaTags should remain nil")
				assert.Nil(t, copied.Header, "Header should remain nil")
				assert.NotNil(t, copied.Children, "Children should be initialized")
				assert.Empty(t, copied.Children, "Children should be empty")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			copied := tt.page.Copy()
			tt.validate(t, tt.page, copied)
		})
	}
}

func TestPage_FixURL_EdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		setup func() *Page
		want  string
	}{
		{
			name: "CMS page with leading slashes in CustomURL",
			setup: func() *Page {
				return &Page{
					Pattern:   PageCMS,
					CustomURL: "///custom///url///",
				}
			},
			want: "/custom///url///",
		},
		{
			name: "CMS page with leading slashes in Slug",
			setup: func() *Page {
				return &Page{
					Pattern: PageCMS,
					Slug:    "///slug///path///",
					Name:    "Test Name",
				}
			},
			want: "/slug///path///",
		},
		{
			name: "Deep nesting with parent without trailing slash",
			setup: func() *Page {
				parent := &Page{
					Pattern: PageCMS,
					URL:     "/parent",
				}
				return &Page{
					Pattern:   PageCMS,
					Parent:    parent,
					CustomURL: "child",
				}
			},
			want: "/parent/child",
		},
		{
			name: "Deep nesting with parent with trailing slash",
			setup: func() *Page {
				parent := &Page{
					Pattern: PageCMS,
					URL:     "/parent/",
				}
				return &Page{
					Pattern:   PageCMS,
					Parent:    parent,
					CustomURL: "child",
				}
			},
			want: "/parent/child",
		},
		{
			name: "Nested structure with multiple levels",
			setup: func() *Page {
				grandparent := &Page{
					Pattern: PageCMS,
					URL:     "/gp",
				}
				parent := &Page{
					Pattern: PageCMS,
					URL:     "/gp/parent",
					Parent:  grandparent,
				}
				child := &Page{
					Pattern:  PageCMS,
					Parent:   parent,
					Name:     "Child Page",
					ParentID: &parent.ID,
				}
				// Set up the relationships
				parent.ParentID = &grandparent.ID
				grandparent.Children = []*Page{parent}
				parent.Children = []*Page{child}
				return child
			},
			want: "/gp/parent/child-page",
		},
		{
			name: "FixURL propagates to all children",
			setup: func() *Page {
				parent := &Page{
					Pattern: PageCMS,
					Name:    "Parent",
					URL:     "/initial",
				}
				child1 := &Page{
					Pattern: PageCMS,
					Name:    "Child1",
					Parent:  parent,
				}
				child2 := &Page{
					Pattern: PageCMS,
					Name:    "Child2",
					Parent:  parent,
				}
				grandchild := &Page{
					Pattern: PageCMS,
					Name:    "Grandchild",
					Parent:  child1,
				}
				parent.Children = []*Page{child1, child2}
				child1.Children = []*Page{grandchild}
				return parent
			},
			want: "/parent",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			page := tt.setup()
			page.FixURL()
			assert.Equal(t, tt.want, page.URL)

			// For the nested structure test, verify children URLs are also fixed
			if tt.name == "FixURL propagates to all children" {
				assert.Equal(t, "/parent/child1", page.Children[0].URL)
				assert.Equal(t, "/parent/child2", page.Children[1].URL)
				assert.Equal(t, "/parent/child1/grandchild", page.Children[0].Children[0].URL)
			}
		})
	}
}

func TestPage_AbsURL_EdgeCases(t *testing.T) {
	tests := []struct {
		name string
		page *Page
		args []any
		want string
	}{
		{
			name: "CMS page with relative site URL containing trailing slash",
			page: func() *Page {
				site := NewSite()
				site.Host = "example.com"
				site.Scheme = "https"
				return &Page{
					Pattern: PageCMS,
					URL:     "/path",
					Site:    site,
				}
			}(),
			want: "https://example.com/path",
		},
		{
			name: "Empty key in args should be ignored",
			page: func() *Page {
				return &Page{
					Pattern: PageCMS,
					URL:     "/path",
					Site:    NewSite(),
				}
			}(),
			args: []any{"", "value", "key", "value2"},
			want: "https://localhost/path?key=value2",
		},
		{
			name: "Mixed dynamic path and query parameters",
			page: func() *Page {
				return &Page{
					Pattern: "/api/v1/users/{userID}/posts/{postID}",
					Site:    NewSite(),
				}
			}(),
			args: []any{"{userID}", "123", "{postID}", "456", "format", "json"},
			want: "https://localhost/api/v1/users/123/posts/456?format=json",
		},
		{
			name: "Dynamic pattern with partial parameter replacement",
			page: func() *Page {
				return &Page{
					Pattern: "/search/{category}",
					Site:    NewSite(),
				}
			}(),
			args: []any{"{category}", "electronics", "sort", "price"},
			want: "https://localhost/search/electronics?sort=price",
		},
		{
			name: "URL encoding in query parameters",
			page: func() *Page {
				return &Page{
					Pattern: PageCMS,
					URL:     "/search",
					Site:    NewSite(),
				}
			}(),
			args: []any{"q", "hello world&more", "filter", "color=red"},
			want: "", // We'll check that the URL contains the right parameters in any order
		},
		{
			name: "Path without leading slash gets slash added",
			page: func() *Page {
				return &Page{
					Pattern: "no-leading-slash",
					Site:    NewSite(),
				}
			}(),
			want: "https://localhost/no-leading-slash",
		},
		{
			name: "Root path handling",
			page: func() *Page {
				site := NewSite()
				site.Scheme = "http"
				site.Host = "test.com"
				return &Page{
					Pattern: PageCMS,
					URL:     "/",
					Site:    site,
				}
			}(),
			want: "http://test.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.page.AbsURL(tt.args...)
			if tt.name == "URL encoding in query parameters" {
				// For this test, we need to check that parameters are properly encoded
				assert.Contains(t, got, "https://localhost/search?")
				assert.Contains(t, got, "q=hello+world%26more")
				assert.Contains(t, got, "filter=color%3Dred")
			} else {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestPage_Constants(t *testing.T) {
	assert.Equal(t, "_page_cms", PageCMS, "PageCMS constant should match")
	assert.Equal(t, "_page_alias_", PageAliasPrefix, "PageAliasPrefix constant should match")
	assert.Equal(t, "_page_internal_", PageInternalPrefix, "PageInternalPrefix constant should match")
	assert.Equal(t, "_page_internal_create", PageInternalCreate, "PageInternalCreate constant should match")
	assert.Equal(t, "_page_internal_error_", PageErrorPrefix, "PageErrorPrefix constant should match")
	assert.Equal(t, "_page_internal_error_4xx", PageError4xx, "PageError4xx constant should match")
	assert.Equal(t, "_page_internal_error_5xx", PageError5xx, "PageError5xx constant should match")
	assert.Equal(t, '{', dynamicPatternChar, "dynamicPatternChar should be opening brace")
}
