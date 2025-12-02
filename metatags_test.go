package pages

import (
	"testing"

	"github.com/stretchr/testify/assert"
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

			assert.Equal(t, tt.want.Charset, got.Charset, "NewMetaTags().Charset should match")
			assert.Equal(t, tt.want.Name, got.Name, "NewMetaTags().Name should match")
			assert.Equal(t, tt.want.Property, got.Property, "NewMetaTags().Property should match")
			assert.Equal(t, tt.want.HTTPEquiv, got.HTTPEquiv, "NewMetaTags().HTTPEquiv should match")
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
	assert.Equal(t, 1, len(base.Name["author"]), "With() should not modify original MetaTags - author length")
	assert.Equal(t, "Original Author", base.Name["author"][0], "With() should not modify original MetaTags - author value")

	assert.Equal(t, 1, len(newTags.Name["author"]), "With() should not modify author in original - length")
	assert.Equal(t, "Original Author", newTags.Name["author"][0], "With() should not modify author in original - value")

	// Verify description is added
	assert.Equal(t, 1, len(newTags.Name["description"]), "With() should add description from other - length")
	assert.Equal(t, "Additional Description", newTags.Name["description"][0], "With() should add description from other - value")

	// Verify og:title is preserved and description is added
	assert.Equal(t, 1, len(newTags.Property["og:title"]), "With() should preserve og:title from original - length")
	assert.Equal(t, "Original Title", newTags.Property["og:title"][0], "With() should preserve og:title from original - value")
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

	assert.Equal(t, "ISO-8859-1", mt.Charset, "Set().Charset should match expected value")
	assert.Equal(t, other.Name, mt.Name, "Set().Name should match expected value")
	assert.Equal(t, other.Property, mt.Property, "Set().Property should match expected value")
	assert.Equal(t, other.HTTPEquiv, mt.HTTPEquiv, "Set().HTTPEquiv should match expected value")
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
	assert.Equal(t, "UTF-8", mt.Charset, "Set(nil).Charset should remain unchanged")
	assert.Equal(t, 1, len(mt.Name), "Set(nil) should not modify Name map length")
	assert.Equal(t, "Original", mt.Name["description"][0], "Set(nil) should not modify Name description value")
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
	assert.Equal(t, expectedName, mt.Name, "Append().Name should match expected value")

	expectedProperty := map[string][]string{
		"og:title":       {"Original Title"},
		"og:description": {"New Description"},
	}
	assert.Equal(t, expectedProperty, mt.Property, "Append().Property should match expected value")

	expectedHTTPEquiv := map[string][]string{
		"refresh": {"30"},
	}
	assert.Equal(t, expectedHTTPEquiv, mt.HTTPEquiv, "Append().HTTPEquiv should match expected value")
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
	assert.Equal(t, "UTF-8", mt.Charset, "Append(nil).Charset should remain unchanged")
	assert.Equal(t, 1, len(mt.Name), "Append(nil) should not modify Name map length")
	assert.Equal(t, "Original", mt.Name["description"][0], "Append(nil) should not modify Name description value")
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

	assert.Equal(t, expected, mt.Name, "SetName() should match expected value")

	// Test overwriting existing value
	mt.SetName("description", "New description")

	expected = map[string][]string{
		"description": {"New description"},
		"keywords":    {"test", "meta"},
	}

	assert.Equal(t, expected, mt.Name, "SetName() overwrite should match expected value")
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

	assert.Equal(t, expected, mt.Name, "AppendName() should match expected value")

	// Test appending to non-existent key
	mt.AppendName("author", "John Doe")

	expected = map[string][]string{
		"keywords": {"test", "meta", "tags"},
		"author":   {"John Doe"},
	}

	assert.Equal(t, expected, mt.Name, "AppendName() to new key should match expected value")
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

	assert.Equal(t, expected, mt.Property, "SetProperty() should match expected value")

	// Test overwriting existing value
	mt.SetProperty("og:title", "New Title")

	expected = map[string][]string{
		"og:title":       {"New Title"},
		"og:description": {"Test description", "Additional"},
	}

	assert.Equal(t, expected, mt.Property, "SetProperty() overwrite should match expected value")
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

	assert.Equal(t, expected, mt.Property, "AppendProperty() should match expected value")

	// Test appending to non-existent key
	mt.AppendProperty("og:image", "image.jpg")

	expected = map[string][]string{
		"og:title": {"Original Title", "Additional Title"},
		"og:image": {"image.jpg"},
	}

	assert.Equal(t, expected, mt.Property, "AppendProperty() to new key should match expected value")
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

	assert.Equal(t, expected, mt.HTTPEquiv, "SetHTTPEquiv() should match expected value")

	// Test overwriting existing value
	mt.SetHTTPEquiv("refresh", "60")

	expected = map[string][]string{
		"refresh":      {"60"},
		"content-type": {"text/html; charset=UTF-8"},
	}

	assert.Equal(t, expected, mt.HTTPEquiv, "SetHTTPEquiv() overwrite should match expected value")
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

	assert.Equal(t, expected, mt.HTTPEquiv, "AppendHTTPEquiv() should match expected value")

	// Test appending to non-existent key
	mt.AppendHTTPEquiv("x-frame-options", "DENY")

	expected = map[string][]string{
		"cache-control":   {"no-cache", "no-store"},
		"x-frame-options": {"DENY"},
	}

	assert.Equal(t, expected, mt.HTTPEquiv, "AppendHTTPEquiv() to new key should match expected value")
}

func TestMetaTags_DefaultCharset(t *testing.T) {
	mt := NewMetaTags(DefaultCharset)

	assert.Equal(t, DefaultCharset, mt.Charset, "DefaultCharset should match expected value")
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
	assert.NotNil(t, mt.Name, "Name map should be initialized")
	assert.NotNil(t, mt.Property, "Property map should be initialized")
	assert.NotNil(t, mt.HTTPEquiv, "HTTPEquiv map should be initialized")
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
	assert.Equal(t, "original", newMt.Name["test"][0], "Set() should clone data")
	assert.Equal(t, "original title", newMt.Property["og:title"][0], "Set() should clone property data")
	assert.Equal(t, "30", newMt.HTTPEquiv["refresh"][0], "Set() should clone HTTPEquiv data")
}
