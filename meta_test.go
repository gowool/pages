package pages

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewMeta(t *testing.T) {
	tests := []struct {
		name    string
		input   map[string]any
		wantLen int
	}{
		{
			name:    "Nil input returns empty meta",
			input:   nil,
			wantLen: 0,
		},
		{
			name:    "Empty map returns empty meta",
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
			meta := NewMeta(tt.input)
			assert.NotNil(t, meta)
			assert.Len(t, meta, tt.wantLen)

			if tt.input != nil {
				for k, v := range tt.input {
					assert.Equal(t, v, meta[k])
				}
			}
		})
	}
}

func TestMeta_Set(t *testing.T) {
	meta := NewMeta(nil)

	meta.Set("stringKey", "value")
	assert.Equal(t, "value", meta["stringKey"])

	meta.Set("intKey", 123)
	assert.Equal(t, 123, meta["intKey"])

	meta.Set("nilKey", nil)
	assert.Nil(t, meta["nilKey"])
}

func TestMeta_Get(t *testing.T) {
	meta := NewMeta(map[string]any{"key1": "value1", "key2": 123})

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
			got := meta.Get(tt.key)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMeta_GetOK(t *testing.T) {
	meta := NewMeta(map[string]any{"key1": "value1", "key2": 123})

	value, ok := meta.GetOK("key1")
	assert.Equal(t, "value1", value)
	assert.True(t, ok)

	value, ok = meta.GetOK("nonExisting")
	assert.Nil(t, value)
	assert.False(t, ok)

	value, ok = meta.GetOK("key2")
	assert.Equal(t, 123, value)
	assert.True(t, ok)
}

func TestMeta_Has(t *testing.T) {
	meta := NewMeta(map[string]any{"key1": "value1", "key2": 123})

	assert.True(t, meta.Has("key1"))
	assert.True(t, meta.Has("key2"))
	assert.False(t, meta.Has("nonExisting"))

	meta.Set("nilValue", nil)
	assert.True(t, meta.Has("nilValue"))
}

func TestMeta_Delete(t *testing.T) {
	meta := NewMeta(map[string]any{"key1": "value1", "key2": 123})

	assert.True(t, meta.Has("key1"))
	meta.Delete("key1")
	assert.False(t, meta.Has("key1"))

	assert.True(t, meta.Has("key2"))
	assert.True(t, meta.Has("key2"))

	meta.Delete("nonExisting")
	assert.Len(t, meta, 1)
}

func TestMeta_Str(t *testing.T) {
	meta := NewMeta(nil)

	meta.Set("string", "hello")
	assert.Equal(t, "hello", meta.Str("string"))

	meta.Set("int", 123)
	assert.Equal(t, "123", meta.Str("int"))

	meta.Set("bool", true)
	assert.Equal(t, "true", meta.Str("bool"))

	assert.Equal(t, "", meta.Str("nonExisting"))

	meta.Set("nil", nil)
	assert.Equal(t, "", meta.Str("nil"))
}

func TestMeta_Bool(t *testing.T) {
	meta := NewMeta(nil)

	meta.Set("true", true)
	assert.True(t, meta.Bool("true"))

	meta.Set("false", false)
	assert.False(t, meta.Bool("false"))

	meta.Set("stringTrue", "true")
	assert.True(t, meta.Bool("stringTrue"))

	meta.Set("int1", 1)
	assert.True(t, meta.Bool("int1"))

	meta.Set("int0", 0)
	assert.False(t, meta.Bool("int0"))

	assert.False(t, meta.Bool("nonExisting"))

	meta.Set("nil", nil)
	assert.False(t, meta.Bool("nil"))
}

func TestMeta_Time(t *testing.T) {
	meta := NewMeta(nil)

	now := time.Now().UTC().Truncate(time.Second)
	meta.Set("time", now)
	result := meta.Time("time")
	assert.Equal(t, now, result)

	meta.Set("timeStr", now.Format(time.RFC3339))
	result = meta.Time("timeStr")
	assert.Equal(t, now.Unix(), result.Unix())

	assert.True(t, meta.Time("nonExisting").IsZero())

	meta.Set("nil", nil)
	assert.True(t, meta.Time("nil").IsZero())
}

func TestMeta_Duration(t *testing.T) {
	meta := NewMeta(nil)

	expected := time.Hour * 2
	meta.Set("duration", expected)
	result := meta.Duration("duration")
	assert.Equal(t, expected, result)

	meta.Set("durationStr", "2h")
	result = meta.Duration("durationStr")
	assert.Equal(t, expected, result)

	assert.Equal(t, time.Duration(0), meta.Duration("nonExisting"))

	meta.Set("nil", nil)
	assert.Equal(t, time.Duration(0), meta.Duration("nil"))
}

func TestMeta_Int(t *testing.T) {
	meta := NewMeta(nil)

	meta.Set("int", 123)
	assert.Equal(t, 123, meta.Int("int"))

	meta.Set("intStr", "456")
	assert.Equal(t, 456, meta.Int("intStr"))

	meta.Set("float", 78.9)
	assert.Equal(t, 78, meta.Int("float"))

	assert.Equal(t, 0, meta.Int("nonExisting"))

	meta.Set("nil", nil)
	assert.Equal(t, 0, meta.Int("nil"))
}

func TestMeta_Int8(t *testing.T) {
	meta := NewMeta(nil)

	meta.Set("int8", int8(127))
	assert.Equal(t, int8(127), meta.Int8("int8"))

	meta.Set("int8Str", "100")
	assert.Equal(t, int8(100), meta.Int8("int8Str"))

	assert.Equal(t, int8(0), meta.Int8("nonExisting"))
}

func TestMeta_Int16(t *testing.T) {
	meta := NewMeta(nil)

	meta.Set("int16", int16(32767))
	assert.Equal(t, int16(32767), meta.Int16("int16"))

	meta.Set("int16Str", "1000")
	assert.Equal(t, int16(1000), meta.Int16("int16Str"))

	assert.Equal(t, int16(0), meta.Int16("nonExisting"))
}

func TestMeta_Int32(t *testing.T) {
	meta := NewMeta(nil)

	meta.Set("int32", int32(2147483647))
	assert.Equal(t, int32(2147483647), meta.Int32("int32"))

	meta.Set("int32Str", "100000")
	assert.Equal(t, int32(100000), meta.Int32("int32Str"))

	assert.Equal(t, int32(0), meta.Int32("nonExisting"))
}

func TestMeta_Int64(t *testing.T) {
	meta := NewMeta(nil)

	meta.Set("int64", int64(9223372036854775807))
	assert.Equal(t, int64(9223372036854775807), meta.Int64("int64"))

	meta.Set("int64Str", "1000000")
	assert.Equal(t, int64(1000000), meta.Int64("int64Str"))

	assert.Equal(t, int64(0), meta.Int64("nonExisting"))
}

func TestMeta_Uint(t *testing.T) {
	meta := NewMeta(nil)

	meta.Set("uint", uint(123))
	assert.Equal(t, uint(123), meta.Uint("uint"))

	meta.Set("uintStr", "456")
	assert.Equal(t, uint(456), meta.Uint("uintStr"))

	assert.Equal(t, uint(0), meta.Uint("nonExisting"))
}

func TestMeta_Uint8(t *testing.T) {
	meta := NewMeta(nil)

	meta.Set("uint8", uint8(255))
	assert.Equal(t, uint8(255), meta.Uint8("uint8"))

	meta.Set("uint8Str", "100")
	assert.Equal(t, uint8(100), meta.Uint8("uint8Str"))

	assert.Equal(t, uint8(0), meta.Uint8("nonExisting"))
}

func TestMeta_Uint16(t *testing.T) {
	meta := NewMeta(nil)

	meta.Set("uint16", uint16(65535))
	assert.Equal(t, uint16(65535), meta.Uint16("uint16"))

	meta.Set("uint16Str", "1000")
	assert.Equal(t, uint16(1000), meta.Uint16("uint16Str"))

	assert.Equal(t, uint16(0), meta.Uint16("nonExisting"))
}

func TestMeta_Uint32(t *testing.T) {
	meta := NewMeta(nil)

	meta.Set("uint32", uint32(4294967295))
	assert.Equal(t, uint32(4294967295), meta.Uint32("uint32"))

	meta.Set("uint32Str", "100000")
	assert.Equal(t, uint32(100000), meta.Uint32("uint32Str"))

	assert.Equal(t, uint32(0), meta.Uint32("nonExisting"))
}

func TestMeta_Uint64(t *testing.T) {
	meta := NewMeta(nil)

	meta.Set("uint64", uint64(18446744073709551615))
	assert.Equal(t, uint64(18446744073709551615), meta.Uint64("uint64"))

	meta.Set("uint64Str", "1000000")
	assert.Equal(t, uint64(1000000), meta.Uint64("uint64Str"))

	assert.Equal(t, uint64(0), meta.Uint64("nonExisting"))
}

func TestMeta_Float32(t *testing.T) {
	meta := NewMeta(nil)

	meta.Set("float32", float32(3.14))
	assert.Equal(t, float32(3.14), meta.Float32("float32"))

	meta.Set("float32Str", "2.71")
	assert.Equal(t, float32(2.71), meta.Float32("float32Str"))

	assert.Equal(t, float32(0), meta.Float32("nonExisting"))

	meta.Set("nil", nil)
	assert.Equal(t, float32(0), meta.Float32("nil"))
}

func TestMeta_Float64(t *testing.T) {
	meta := NewMeta(nil)

	meta.Set("float64", 3.14159265359)
	assert.Equal(t, 3.14159265359, meta.Float64("float64"))

	meta.Set("float64Str", "2.718281828")
	assert.Equal(t, 2.718281828, meta.Float64("float64Str"))

	assert.Equal(t, float64(0), meta.Float64("nonExisting"))

	meta.Set("nil", nil)
	assert.Equal(t, float64(0), meta.Float64("nil"))
}

func TestMeta_Slice(t *testing.T) {
	meta := NewMeta(nil)

	expected := []any{"a", "b", 123}
	meta.Set("slice", expected)
	result := meta.Slice("slice")
	assert.Equal(t, expected, result)

	assert.Nil(t, meta.Slice("nonExisting"))

	meta.Set("nil", nil)
	assert.Nil(t, meta.Slice("nil"))
}

func TestMeta_BoolSlice(t *testing.T) {
	meta := NewMeta(nil)

	expected := []bool{true, false, true}
	meta.Set("boolSlice", expected)
	result := meta.BoolSlice("boolSlice")
	assert.Equal(t, expected, result)

	assert.Nil(t, meta.BoolSlice("nonExisting"))

	meta.Set("nil", nil)
	assert.Nil(t, meta.BoolSlice("nil"))
}

func TestMeta_StrSlice(t *testing.T) {
	meta := NewMeta(nil)

	expected := []string{"a", "b", "c"}
	meta.Set("strSlice", expected)
	result := meta.StrSlice("strSlice")
	assert.Equal(t, expected, result)

	assert.Nil(t, meta.StrSlice("nonExisting"))

	meta.Set("nil", nil)
	assert.Nil(t, meta.StrSlice("nil"))
}

func TestMeta_IntSlice(t *testing.T) {
	meta := NewMeta(nil)

	expected := []int{1, 2, 3}
	meta.Set("intSlice", expected)
	result := meta.IntSlice("intSlice")
	assert.Equal(t, expected, result)

	assert.Nil(t, meta.IntSlice("nonExisting"))

	meta.Set("nil", nil)
	assert.Nil(t, meta.IntSlice("nil"))
}

func TestMeta_Integration(t *testing.T) {
	meta := NewMeta(map[string]any{
		"name":   "Test",
		"age":    30,
		"active": true,
		"score":  95.5,
		"tags":   []string{"golang", "testing"},
	})

	assert.Equal(t, "Test", meta.Str("name"))
	assert.Equal(t, 30, meta.Int("age"))
	assert.True(t, meta.Bool("active"))
	assert.Equal(t, 95.5, meta.Float64("score"))
	assert.Equal(t, []string{"golang", "testing"}, meta.StrSlice("tags"))

	meta.Set("name", "Updated")
	assert.Equal(t, "Updated", meta.Str("name"))

	assert.True(t, meta.Has("age"))
	meta.Delete("age")
	assert.False(t, meta.Has("age"))
}
