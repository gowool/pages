package pages

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewMetadata(t *testing.T) {
	tests := []struct {
		name    string
		input   map[string]any
		wantLen int
	}{
		{
			name:    "Nil input returns empty metadata",
			input:   nil,
			wantLen: 0,
		},
		{
			name:    "Empty map returns empty metadata",
			input:   map[string]any{},
			wantLen: 0,
		},
		{
			name:    "Map with values copies correctly",
			input:   map[string]any{"key1": "value1", "key2": 123},
			wantLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metadata := NewMetadata(tt.input)
			assert.NotNil(t, metadata)
			assert.Len(t, metadata, tt.wantLen)

			if tt.input != nil {
				for k, v := range tt.input {
					assert.Equal(t, v, metadata[k])
				}
			}
		})
	}
}

func TestMetadata_Set(t *testing.T) {
	metadata := NewMetadata(nil)

	metadata.Set("stringKey", "value")
	assert.Equal(t, "value", metadata["stringKey"])

	metadata.Set("intKey", 123)
	assert.Equal(t, 123, metadata["intKey"])

	metadata.Set("nilKey", nil)
	assert.Nil(t, metadata["nilKey"])
}

func TestMetadata_Get(t *testing.T) {
	metadata := NewMetadata(map[string]any{"key1": "value1", "key2": 123})

	tests := []struct {
		name string
		key  string
		want any
	}{
		{
			name: "Existing string key",
			key:  "key1",
			want: "value1",
		},
		{
			name: "Existing int key",
			key:  "key2",
			want: 123,
		},
		{
			name: "Non-existing key returns nil",
			key:  "nonExisting",
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := metadata.Get(tt.key)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMetadata_GetOK(t *testing.T) {
	metadata := NewMetadata(map[string]any{"key1": "value1", "key2": 123})

	value, ok := metadata.GetOK("key1")
	assert.Equal(t, "value1", value)
	assert.True(t, ok)

	value, ok = metadata.GetOK("nonExisting")
	assert.Nil(t, value)
	assert.False(t, ok)

	value, ok = metadata.GetOK("key2")
	assert.Equal(t, 123, value)
	assert.True(t, ok)
}

func TestMetadata_Has(t *testing.T) {
	metadata := NewMetadata(map[string]any{"key1": "value1", "key2": 123})

	assert.True(t, metadata.Has("key1"))
	assert.True(t, metadata.Has("key2"))
	assert.False(t, metadata.Has("nonExisting"))

	metadata.Set("nilValue", nil)
	assert.True(t, metadata.Has("nilValue"))
}

func TestMetadata_Delete(t *testing.T) {
	metadata := NewMetadata(map[string]any{"key1": "value1", "key2": 123})

	assert.True(t, metadata.Has("key1"))
	metadata.Delete("key1")
	assert.False(t, metadata.Has("key1"))

	assert.True(t, metadata.Has("key2"))
	assert.True(t, metadata.Has("key2"))

	metadata.Delete("nonExisting")
	assert.Len(t, metadata, 1)
}

func TestMetadata_Str(t *testing.T) {
	metadata := NewMetadata(nil)

	metadata.Set("string", "hello")
	assert.Equal(t, "hello", metadata.Str("string"))

	metadata.Set("int", 123)
	assert.Equal(t, "123", metadata.Str("int"))

	metadata.Set("bool", true)
	assert.Equal(t, "true", metadata.Str("bool"))

	assert.Equal(t, "", metadata.Str("nonExisting"))

	metadata.Set("nil", nil)
	assert.Equal(t, "", metadata.Str("nil"))
}

func TestMetadata_Bool(t *testing.T) {
	metadata := NewMetadata(nil)

	metadata.Set("true", true)
	assert.True(t, metadata.Bool("true"))

	metadata.Set("false", false)
	assert.False(t, metadata.Bool("false"))

	metadata.Set("stringTrue", "true")
	assert.True(t, metadata.Bool("stringTrue"))

	metadata.Set("int1", 1)
	assert.True(t, metadata.Bool("int1"))

	metadata.Set("int0", 0)
	assert.False(t, metadata.Bool("int0"))

	assert.False(t, metadata.Bool("nonExisting"))

	metadata.Set("nil", nil)
	assert.False(t, metadata.Bool("nil"))
}

func TestMetadata_Time(t *testing.T) {
	metadata := NewMetadata(nil)

	now := time.Now().UTC().Truncate(time.Second)
	metadata.Set("time", now)
	result := metadata.Time("time")
	assert.Equal(t, now, result)

	metadata.Set("timeStr", now.Format(time.RFC3339))
	result = metadata.Time("timeStr")
	assert.Equal(t, now.Unix(), result.Unix())

	assert.True(t, metadata.Time("nonExisting").IsZero())

	metadata.Set("nil", nil)
	assert.True(t, metadata.Time("nil").IsZero())
}

func TestMetadata_Duration(t *testing.T) {
	metadata := NewMetadata(nil)

	expected := time.Hour * 2
	metadata.Set("duration", expected)
	result := metadata.Duration("duration")
	assert.Equal(t, expected, result)

	metadata.Set("durationStr", "2h")
	result = metadata.Duration("durationStr")
	assert.Equal(t, expected, result)

	assert.Equal(t, time.Duration(0), metadata.Duration("nonExisting"))

	metadata.Set("nil", nil)
	assert.Equal(t, time.Duration(0), metadata.Duration("nil"))
}

func TestMetadata_Int(t *testing.T) {
	metadata := NewMetadata(nil)

	metadata.Set("int", 123)
	assert.Equal(t, 123, metadata.Int("int"))

	metadata.Set("intStr", "456")
	assert.Equal(t, 456, metadata.Int("intStr"))

	metadata.Set("float", 78.9)
	assert.Equal(t, 78, metadata.Int("float"))

	assert.Equal(t, 0, metadata.Int("nonExisting"))

	metadata.Set("nil", nil)
	assert.Equal(t, 0, metadata.Int("nil"))
}

func TestMetadata_Int8(t *testing.T) {
	metadata := NewMetadata(nil)

	metadata.Set("int8", int8(127))
	assert.Equal(t, int8(127), metadata.Int8("int8"))

	metadata.Set("int8Str", "100")
	assert.Equal(t, int8(100), metadata.Int8("int8Str"))

	assert.Equal(t, int8(0), metadata.Int8("nonExisting"))
}

func TestMetadata_Int16(t *testing.T) {
	metadata := NewMetadata(nil)

	metadata.Set("int16", int16(32767))
	assert.Equal(t, int16(32767), metadata.Int16("int16"))

	metadata.Set("int16Str", "1000")
	assert.Equal(t, int16(1000), metadata.Int16("int16Str"))

	assert.Equal(t, int16(0), metadata.Int16("nonExisting"))
}

func TestMetadata_Int32(t *testing.T) {
	metadata := NewMetadata(nil)

	metadata.Set("int32", int32(2147483647))
	assert.Equal(t, int32(2147483647), metadata.Int32("int32"))

	metadata.Set("int32Str", "100000")
	assert.Equal(t, int32(100000), metadata.Int32("int32Str"))

	assert.Equal(t, int32(0), metadata.Int32("nonExisting"))
}

func TestMetadata_Int64(t *testing.T) {
	metadata := NewMetadata(nil)

	metadata.Set("int64", int64(9223372036854775807))
	assert.Equal(t, int64(9223372036854775807), metadata.Int64("int64"))

	metadata.Set("int64Str", "1000000")
	assert.Equal(t, int64(1000000), metadata.Int64("int64Str"))

	assert.Equal(t, int64(0), metadata.Int64("nonExisting"))
}

func TestMetadata_Uint(t *testing.T) {
	metadata := NewMetadata(nil)

	metadata.Set("uint", uint(123))
	assert.Equal(t, uint(123), metadata.Uint("uint"))

	metadata.Set("uintStr", "456")
	assert.Equal(t, uint(456), metadata.Uint("uintStr"))

	assert.Equal(t, uint(0), metadata.Uint("nonExisting"))
}

func TestMetadata_Uint8(t *testing.T) {
	metadata := NewMetadata(nil)

	metadata.Set("uint8", uint8(255))
	assert.Equal(t, uint8(255), metadata.Uint8("uint8"))

	metadata.Set("uint8Str", "100")
	assert.Equal(t, uint8(100), metadata.Uint8("uint8Str"))

	assert.Equal(t, uint8(0), metadata.Uint8("nonExisting"))
}

func TestMetadata_Uint16(t *testing.T) {
	metadata := NewMetadata(nil)

	metadata.Set("uint16", uint16(65535))
	assert.Equal(t, uint16(65535), metadata.Uint16("uint16"))

	metadata.Set("uint16Str", "1000")
	assert.Equal(t, uint16(1000), metadata.Uint16("uint16Str"))

	assert.Equal(t, uint16(0), metadata.Uint16("nonExisting"))
}

func TestMetadata_Uint32(t *testing.T) {
	metadata := NewMetadata(nil)

	metadata.Set("uint32", uint32(4294967295))
	assert.Equal(t, uint32(4294967295), metadata.Uint32("uint32"))

	metadata.Set("uint32Str", "100000")
	assert.Equal(t, uint32(100000), metadata.Uint32("uint32Str"))

	assert.Equal(t, uint32(0), metadata.Uint32("nonExisting"))
}

func TestMetadata_Uint64(t *testing.T) {
	metadata := NewMetadata(nil)

	metadata.Set("uint64", uint64(18446744073709551615))
	assert.Equal(t, uint64(18446744073709551615), metadata.Uint64("uint64"))

	metadata.Set("uint64Str", "1000000")
	assert.Equal(t, uint64(1000000), metadata.Uint64("uint64Str"))

	assert.Equal(t, uint64(0), metadata.Uint64("nonExisting"))
}

func TestMetadata_Float32(t *testing.T) {
	metadata := NewMetadata(nil)

	metadata.Set("float32", float32(3.14))
	assert.Equal(t, float32(3.14), metadata.Float32("float32"))

	metadata.Set("float32Str", "2.71")
	assert.Equal(t, float32(2.71), metadata.Float32("float32Str"))

	assert.Equal(t, float32(0), metadata.Float32("nonExisting"))

	metadata.Set("nil", nil)
	assert.Equal(t, float32(0), metadata.Float32("nil"))
}

func TestMetadata_Float64(t *testing.T) {
	metadata := NewMetadata(nil)

	metadata.Set("float64", 3.14159265359)
	assert.Equal(t, 3.14159265359, metadata.Float64("float64"))

	metadata.Set("float64Str", "2.718281828")
	assert.Equal(t, 2.718281828, metadata.Float64("float64Str"))

	assert.Equal(t, float64(0), metadata.Float64("nonExisting"))

	metadata.Set("nil", nil)
	assert.Equal(t, float64(0), metadata.Float64("nil"))
}

func TestMetadata_Slice(t *testing.T) {
	metadata := NewMetadata(nil)

	expected := []any{"a", "b", 123}
	metadata.Set("slice", expected)
	result := metadata.Slice("slice")
	assert.Equal(t, expected, result)

	assert.Nil(t, metadata.Slice("nonExisting"))

	metadata.Set("nil", nil)
	assert.Nil(t, metadata.Slice("nil"))
}

func TestMetadata_BoolSlice(t *testing.T) {
	metadata := NewMetadata(nil)

	expected := []bool{true, false, true}
	metadata.Set("boolSlice", expected)
	result := metadata.BoolSlice("boolSlice")
	assert.Equal(t, expected, result)

	assert.Nil(t, metadata.BoolSlice("nonExisting"))

	metadata.Set("nil", nil)
	assert.Nil(t, metadata.BoolSlice("nil"))
}

func TestMetadata_StrSlice(t *testing.T) {
	metadata := NewMetadata(nil)

	expected := []string{"a", "b", "c"}
	metadata.Set("strSlice", expected)
	result := metadata.StrSlice("strSlice")
	assert.Equal(t, expected, result)

	assert.Nil(t, metadata.StrSlice("nonExisting"))

	metadata.Set("nil", nil)
	assert.Nil(t, metadata.StrSlice("nil"))
}

func TestMetadata_IntSlice(t *testing.T) {
	metadata := NewMetadata(nil)

	expected := []int{1, 2, 3}
	metadata.Set("intSlice", expected)
	result := metadata.IntSlice("intSlice")
	assert.Equal(t, expected, result)

	assert.Nil(t, metadata.IntSlice("nonExisting"))

	metadata.Set("nil", nil)
	assert.Nil(t, metadata.IntSlice("nil"))
}

func TestMetadata_Integration(t *testing.T) {
	metadata := NewMetadata(map[string]any{
		"name":   "Test",
		"age":    30,
		"active": true,
		"score":  95.5,
		"tags":   []string{"golang", "testing"},
	})

	assert.Equal(t, "Test", metadata.Str("name"))
	assert.Equal(t, 30, metadata.Int("age"))
	assert.True(t, metadata.Bool("active"))
	assert.Equal(t, 95.5, metadata.Float64("score"))
	assert.Equal(t, []string{"golang", "testing"}, metadata.StrSlice("tags"))

	metadata.Set("name", "Updated")
	assert.Equal(t, "Updated", metadata.Str("name"))

	assert.True(t, metadata.Has("age"))
	metadata.Delete("age")
	assert.False(t, metadata.Has("age"))
}
