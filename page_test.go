package pages

import (
	"testing"
	"time"
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
			if got := p.AbsURL(tt.args.args...); got != tt.want {
				t.Errorf("AbsURL() = %v, want %v", got, tt.want)
			}
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
			if got := p.URL; got != tt.want {
				t.Errorf("URL = %v, want %v", got, tt.want)
			}
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
			if got := p.IsCMS(); got != tt.want {
				t.Errorf("IsCMS() = %v, want %v", got, tt.want)
			}
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
			if got := p.IsDynamic(); got != tt.want {
				t.Errorf("IsDynamic() = %v, want %v", got, tt.want)
			}
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
			if got := p.IsError(); got != tt.want {
				t.Errorf("IsError() = %v, want %v", got, tt.want)
			}
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
			if got := p.IsHybrid(); got != tt.want {
				t.Errorf("IsHybrid() = %v, want %v", got, tt.want)
			}
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
			if got := p.IsInternal(); got != tt.want {
				t.Errorf("IsInternal() = %v, want %v", got, tt.want)
			}
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
			if got := p.Alias; got != tt.want {
				t.Errorf("Alias = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPage_String(t *testing.T) {
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
		{name: "n/a", want: "n/a"},
		{name: "foo", fields: fields{Name: "foo"}, want: "foo"},
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
			if got := p.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}
