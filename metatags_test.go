package pages

import (
	"reflect"
	"testing"
)

func TestNewMetaTags(t *testing.T) {
	tests := []struct {
		name    string
		charset string
		other   []*MetaTags
		want    *MetaTags
	}{
		{
			name:    "Create with charset only",
			charset: "UTF-8",
			other:   nil,
			want: &MetaTags{
				Charset:   "UTF-8",
				Name:      make(map[string][]string),
				Property:  make(map[string][]string),
				HTTPEquiv: make(map[string][]string),
			},
		},
		{
			name:    "Create with charset and empty other slice",
			charset: "ISO-8859-1",
			other:   []*MetaTags{},
			want: &MetaTags{
				Charset:   "ISO-8859-1",
				Name:      make(map[string][]string),
				Property:  make(map[string][]string),
				HTTPEquiv: make(map[string][]string),
			},
		},
		{
			name:    "Create with charset and one other",
			charset: "UTF-8",
			other: []*MetaTags{
				{
					Charset: "UTF-8", // The charset from other[0] will override the base charset
					Name: map[string][]string{
						"description": {"Test description"},
						"keywords":    {"test", "meta"},
					},
					Property: map[string][]string{
						"og:title": {"Test Title"},
					},
					HTTPEquiv: map[string][]string{
						"refresh": {"30"},
					},
				},
			},
			want: &MetaTags{
				Charset: "UTF-8",
				Name: map[string][]string{
					"description": {"Test description"},
					"keywords":    {"test", "meta"},
				},
				Property: map[string][]string{
					"og:title": {"Test Title"},
				},
				HTTPEquiv: map[string][]string{
					"refresh": {"30"},
				},
			},
		},
		{
			name:    "Create with charset and multiple others",
			charset: "UTF-8",
			other: []*MetaTags{
				{
					Charset: "UTF-8", // The charset from other[0] will override the base charset
					Name: map[string][]string{
						"description": {"First description"},
					},
					Property: map[string][]string{
						"og:title": {"First Title"},
					},
					HTTPEquiv: map[string][]string{
						"cache-control": {"no-cache"}, // Add at least one entry to initialize the map
					},
				},
				{
					Name: map[string][]string{
						"keywords": {"test", "meta"},
					},
					Property: map[string][]string{
						"og:description": {"Second description"},
					},
					HTTPEquiv: map[string][]string{
						"refresh": {"30"},
					},
				},
			},
			want: &MetaTags{
				Charset: "UTF-8",
				Name: map[string][]string{
					"description": {"First description"},
					"keywords":    {"test", "meta"},
				},
				Property: map[string][]string{
					"og:title":       {"First Title"},
					"og:description": {"Second description"},
				},
				HTTPEquiv: map[string][]string{
					"cache-control": {"no-cache"},
					"refresh":       {"30"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewMetaTags(tt.charset, tt.other...)

			if got.Charset != tt.want.Charset {
				t.Errorf("NewMetaTags().Charset = %v, want %v", got.Charset, tt.want.Charset)
			}

			if !reflect.DeepEqual(got.Name, tt.want.Name) {
				t.Errorf("NewMetaTags().Name = %v, want %v", got.Name, tt.want.Name)
			}

			if !reflect.DeepEqual(got.Property, tt.want.Property) {
				t.Errorf("NewMetaTags().Property = %v, want %v", got.Property, tt.want.Property)
			}

			if !reflect.DeepEqual(got.HTTPEquiv, tt.want.HTTPEquiv) {
				t.Errorf("NewMetaTags().HTTPEquiv = %v, want %v", got.HTTPEquiv, tt.want.HTTPEquiv)
			}
		})
	}
}

func TestMetaTags_With(t *testing.T) {
	// Test With functionality - create a new MetaTags with existing ones
	base := &MetaTags{
		Charset: "UTF-8",
		Name: map[string][]string{
			"author": {"Original Author"},
		},
		Property: map[string][]string{
			"og:title": {"Original Title"},
		},
		HTTPEquiv: map[string][]string{
			"refresh": {"30"},
		},
	}

	// Create new MetaTags using With method
	newTags := base.With(
		&MetaTags{
			Name: map[string][]string{
				"description": {"Additional Description"},
			},
		},
	)

	// Verify original is not modified
	if len(base.Name["author"]) != 1 || base.Name["author"][0] != "Original Author" {
		t.Errorf("With() should not modify original MetaTags")
	}

	if len(newTags.Name["author"]) != 1 || newTags.Name["author"][0] != "Original Author" {
		t.Errorf("With() should not modify author in original")
	}

	// Verify description is added
	if len(newTags.Name["description"]) != 1 || newTags.Name["description"][0] != "Additional Description" {
		t.Errorf("With() should add description from other")
	}

	// Verify og:title is preserved and description is added
	if len(newTags.Property["og:title"]) != 1 || newTags.Property["og:title"][0] != "Original Title" {
		t.Errorf("With() should preserve og:title from original")
	}
}

func TestMetaTags_Set(t *testing.T) {
	mt := &MetaTags{}

	other := &MetaTags{
		Charset: "ISO-8859-1",
		Name: map[string][]string{
			"description": {"Test description"},
			"keywords":    {"test", "meta"},
		},
		Property: map[string][]string{
			"og:title": {"Test Title"},
		},
		HTTPEquiv: map[string][]string{
			"refresh": {"30"},
		},
	}

	mt.Set(other)

	if mt.Charset != "ISO-8859-1" {
		t.Errorf("Set().Charset = %v, want %v", mt.Charset, "ISO-8859-1")
	}

	if !reflect.DeepEqual(mt.Name, other.Name) {
		t.Errorf("Set().Name = %v, want %v", mt.Name, other.Name)
	}

	if !reflect.DeepEqual(mt.Property, other.Property) {
		t.Errorf("Set().Property = %v, want %v", mt.Property, other.Property)
	}

	if !reflect.DeepEqual(mt.HTTPEquiv, other.HTTPEquiv) {
		t.Errorf("Set().HTTPEquiv = %v, want %v", mt.HTTPEquiv, other.HTTPEquiv)
	}
}

func TestMetaTags_Set_Nil(t *testing.T) {
	mt := &MetaTags{
		Charset: "UTF-8",
		Name: map[string][]string{
			"description": {"Original"},
		},
		Property:  make(map[string][]string),
		HTTPEquiv: make(map[string][]string),
	}

	mt.Set(nil)

	// Should remain unchanged
	if mt.Charset != "UTF-8" {
		t.Errorf("Set(nil).Charset = %v, want %v", mt.Charset, "UTF-8")
	}

	if len(mt.Name) != 1 || mt.Name["description"][0] != "Original" {
		t.Errorf("Set(nil) modified original data unexpectedly")
	}
}

func TestMetaTags_Append(t *testing.T) {
	mt := &MetaTags{
		Charset: "UTF-8",
		Name: map[string][]string{
			"description": {"Original description"},
			"keywords":    {"original"},
		},
		Property: map[string][]string{
			"og:title": {"Original Title"},
		},
		HTTPEquiv: map[string][]string{}, // Initialize empty map
	}

	other := &MetaTags{
		Name: map[string][]string{
			"description": {"Additional description"},
			"author":      {"New Author"},
		},
		Property: map[string][]string{
			"og:description": {"New Description"},
		},
		HTTPEquiv: map[string][]string{
			"refresh": {"30"},
		},
	}

	mt.Append(other)

	expectedName := map[string][]string{
		"description": {"Original description", "Additional description"},
		"keywords":    {"original"},
		"author":      {"New Author"},
	}
	if !reflect.DeepEqual(mt.Name, expectedName) {
		t.Errorf("Append().Name = %v, want %v", mt.Name, expectedName)
	}

	expectedProperty := map[string][]string{
		"og:title":       {"Original Title"},
		"og:description": {"New Description"},
	}
	if !reflect.DeepEqual(mt.Property, expectedProperty) {
		t.Errorf("Append().Property = %v, want %v", mt.Property, expectedProperty)
	}

	expectedHTTPEquiv := map[string][]string{
		"refresh": {"30"},
	}
	if !reflect.DeepEqual(mt.HTTPEquiv, expectedHTTPEquiv) {
		t.Errorf("Append().HTTPEquiv = %v, want %v", mt.HTTPEquiv, expectedHTTPEquiv)
	}
}

func TestMetaTags_Append_Nil(t *testing.T) {
	mt := &MetaTags{
		Charset: "UTF-8",
		Name: map[string][]string{
			"description": {"Original"},
		},
		Property:  make(map[string][]string),
		HTTPEquiv: make(map[string][]string),
	}

	mt.Append(nil)

	// Should remain unchanged
	if mt.Charset != "UTF-8" {
		t.Errorf("Append(nil).Charset = %v, want %v", mt.Charset, "UTF-8")
	}

	if len(mt.Name) != 1 || mt.Name["description"][0] != "Original" {
		t.Errorf("Append(nil) modified original data unexpectedly")
	}
}

func TestMetaTags_SetName(t *testing.T) {
	mt := &MetaTags{
		Name:      make(map[string][]string),
		Property:  make(map[string][]string),
		HTTPEquiv: make(map[string][]string),
	}

	mt.SetName("description", "Test description")
	mt.SetName("keywords", "test", "meta")

	expected := map[string][]string{
		"description": {"Test description"},
		"keywords":    {"test", "meta"},
	}

	if !reflect.DeepEqual(mt.Name, expected) {
		t.Errorf("SetName() = %v, want %v", mt.Name, expected)
	}

	// Test overwriting existing value
	mt.SetName("description", "New description")

	expected = map[string][]string{
		"description": {"New description"},
		"keywords":    {"test", "meta"},
	}

	if !reflect.DeepEqual(mt.Name, expected) {
		t.Errorf("SetName() overwrite = %v, want %v", mt.Name, expected)
	}
}

func TestMetaTags_AppendName(t *testing.T) {
	mt := &MetaTags{
		Name:      make(map[string][]string),
		Property:  make(map[string][]string),
		HTTPEquiv: make(map[string][]string),
	}

	mt.SetName("keywords", "test")
	mt.AppendName("keywords", "meta", "tags")

	expected := map[string][]string{
		"keywords": {"test", "meta", "tags"},
	}

	if !reflect.DeepEqual(mt.Name, expected) {
		t.Errorf("AppendName() = %v, want %v", mt.Name, expected)
	}

	// Test appending to non-existent key
	mt.AppendName("author", "John Doe")

	expected = map[string][]string{
		"keywords": {"test", "meta", "tags"},
		"author":   {"John Doe"},
	}

	if !reflect.DeepEqual(mt.Name, expected) {
		t.Errorf("AppendName() to new key = %v, want %v", mt.Name, expected)
	}
}

func TestMetaTags_SetProperty(t *testing.T) {
	mt := &MetaTags{
		Name:      make(map[string][]string),
		Property:  make(map[string][]string),
		HTTPEquiv: make(map[string][]string),
	}

	mt.SetProperty("og:title", "Test Title")
	mt.SetProperty("og:description", "Test description", "Additional")

	expected := map[string][]string{
		"og:title":       {"Test Title"},
		"og:description": {"Test description", "Additional"},
	}

	if !reflect.DeepEqual(mt.Property, expected) {
		t.Errorf("SetProperty() = %v, want %v", mt.Property, expected)
	}

	// Test overwriting existing value
	mt.SetProperty("og:title", "New Title")

	expected = map[string][]string{
		"og:title":       {"New Title"},
		"og:description": {"Test description", "Additional"},
	}

	if !reflect.DeepEqual(mt.Property, expected) {
		t.Errorf("SetProperty() overwrite = %v, want %v", mt.Property, expected)
	}
}

func TestMetaTags_AppendProperty(t *testing.T) {
	mt := &MetaTags{
		Name:      make(map[string][]string),
		Property:  make(map[string][]string),
		HTTPEquiv: make(map[string][]string),
	}

	mt.SetProperty("og:title", "Original Title")
	mt.AppendProperty("og:title", "Additional Title")

	expected := map[string][]string{
		"og:title": {"Original Title", "Additional Title"},
	}

	if !reflect.DeepEqual(mt.Property, expected) {
		t.Errorf("AppendProperty() = %v, want %v", mt.Property, expected)
	}

	// Test appending to non-existent key
	mt.AppendProperty("og:image", "image.jpg")

	expected = map[string][]string{
		"og:title": {"Original Title", "Additional Title"},
		"og:image": {"image.jpg"},
	}

	if !reflect.DeepEqual(mt.Property, expected) {
		t.Errorf("AppendProperty() to new key = %v, want %v", mt.Property, expected)
	}
}

func TestMetaTags_SetHTTPEquiv(t *testing.T) {
	mt := &MetaTags{
		Name:      make(map[string][]string),
		Property:  make(map[string][]string),
		HTTPEquiv: make(map[string][]string),
	}

	mt.SetHTTPEquiv("refresh", "30")
	mt.SetHTTPEquiv("content-type", "text/html; charset=UTF-8")

	expected := map[string][]string{
		"refresh":      {"30"},
		"content-type": {"text/html; charset=UTF-8"},
	}

	if !reflect.DeepEqual(mt.HTTPEquiv, expected) {
		t.Errorf("SetHTTPEquiv() = %v, want %v", mt.HTTPEquiv, expected)
	}

	// Test overwriting existing value
	mt.SetHTTPEquiv("refresh", "60")

	expected = map[string][]string{
		"refresh":      {"60"},
		"content-type": {"text/html; charset=UTF-8"},
	}

	if !reflect.DeepEqual(mt.HTTPEquiv, expected) {
		t.Errorf("SetHTTPEquiv() overwrite = %v, want %v", mt.HTTPEquiv, expected)
	}
}

func TestMetaTags_AppendHTTPEquiv(t *testing.T) {
	mt := &MetaTags{
		Name:      make(map[string][]string),
		Property:  make(map[string][]string),
		HTTPEquiv: make(map[string][]string),
	}

	mt.SetHTTPEquiv("cache-control", "no-cache")
	mt.AppendHTTPEquiv("cache-control", "no-store")

	expected := map[string][]string{
		"cache-control": {"no-cache", "no-store"},
	}

	if !reflect.DeepEqual(mt.HTTPEquiv, expected) {
		t.Errorf("AppendHTTPEquiv() = %v, want %v", mt.HTTPEquiv, expected)
	}

	// Test appending to non-existent key
	mt.AppendHTTPEquiv("x-frame-options", "DENY")

	expected = map[string][]string{
		"cache-control":   {"no-cache", "no-store"},
		"x-frame-options": {"DENY"},
	}

	if !reflect.DeepEqual(mt.HTTPEquiv, expected) {
		t.Errorf("AppendHTTPEquiv() to new key = %v, want %v", mt.HTTPEquiv, expected)
	}
}

func TestMetaTags_DefaultCharset(t *testing.T) {
	mt := NewMetaTags(DefaultCharset)

	if mt.Charset != DefaultCharset {
		t.Errorf("DefaultCharset = %v, want %v", mt.Charset, DefaultCharset)
	}
}

func TestMetaTags_EmptyMethods(t *testing.T) {
	mt := &MetaTags{
		Name:      make(map[string][]string),
		Property:  make(map[string][]string),
		HTTPEquiv: make(map[string][]string),
	}

	// Test methods with empty content
	mt.SetName("test")
	mt.AppendName("test")
	mt.SetProperty("test")
	mt.AppendProperty("test")
	mt.SetHTTPEquiv("test")
	mt.AppendHTTPEquiv("test")

	// Should not panic and maps should be initialized
	if mt.Name == nil {
		t.Error("Name map should be initialized")
	}
	if mt.Property == nil {
		t.Error("Property map should be initialized")
	}
	if mt.HTTPEquiv == nil {
		t.Error("HTTPEquiv map should be initialized")
	}
}

func TestMetaTags_CloneBehavior(t *testing.T) {
	original := &MetaTags{
		Name: map[string][]string{
			"test": {"original"},
		},
		Property: map[string][]string{
			"og:title": {"original title"},
		},
		HTTPEquiv: map[string][]string{
			"refresh": {"30"},
		},
	}

	newMt := &MetaTags{
		Name:      make(map[string][]string),
		Property:  make(map[string][]string),
		HTTPEquiv: make(map[string][]string),
	}
	newMt.Set(original)

	// Modify original after copying
	original.Name["test"] = []string{"modified"}
	original.Property["og:title"] = []string{"modified title"}
	original.HTTPEquiv["refresh"] = []string{"60"}

	// New instance should not be affected
	if newMt.Name["test"][0] != "original" {
		t.Errorf("Set() should clone data, got %v, want %v", newMt.Name["test"][0], "original")
	}

	if newMt.Property["og:title"][0] != "original title" {
		t.Errorf("Set() should clone property data, got %v, want %v", newMt.Property["og:title"][0], "original title")
	}

	if newMt.HTTPEquiv["refresh"][0] != "30" {
		t.Errorf("Set() should clone HTTPEquiv data, got %v, want %v", newMt.HTTPEquiv["refresh"][0], "30")
	}
}
