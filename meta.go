package pages

import (
	"maps"
	"time"

	"github.com/spf13/cast"
)

type Meta map[string]any

func NewMeta(m map[string]any) Meta {
	if m == nil {
		return make(Meta)
	}

	return maps.Clone(m)
}

func (m Meta) Set(key string, value any) {
	m[key] = value
}

func (m Meta) Get(key string) any {
	return m[key]
}

func (m Meta) GetOK(key string) (any, bool) {
	v, ok := m[key]
	return v, ok
}

func (m Meta) Has(key string) bool {
	_, ok := m[key]
	return ok
}

func (m Meta) Delete(key string) {
	delete(m, key)
}

func (m Meta) Str(key string) string {
	return cast.ToString(m.Get(key))
}

func (m Meta) Bool(key string) bool {
	return cast.ToBool(m.Get(key))
}

func (m Meta) Time(key string) time.Time {
	return cast.ToTime(m.Get(key))
}

func (m Meta) Duration(key string) time.Duration {
	return cast.ToDuration(m.Get(key))
}

func (m Meta) Int(key string) int {
	return cast.ToInt(m.Get(key))
}

func (m Meta) Int8(key string) int8 {
	return cast.ToInt8(m.Get(key))
}

func (m Meta) Int16(key string) int16 {
	return cast.ToInt16(m.Get(key))
}

func (m Meta) Int32(key string) int32 {
	return cast.ToInt32(m.Get(key))
}

func (m Meta) Int64(key string) int64 {
	return cast.ToInt64(m.Get(key))
}

func (m Meta) Uint(key string) uint {
	return cast.ToUint(m.Get(key))
}

func (m Meta) Uint8(key string) uint8 {
	return cast.ToUint8(m.Get(key))
}

func (m Meta) Uint16(key string) uint16 {
	return cast.ToUint16(m.Get(key))
}

func (m Meta) Uint32(key string) uint32 {
	return cast.ToUint32(m.Get(key))
}

func (m Meta) Uint64(key string) uint64 {
	return cast.ToUint64(m.Get(key))
}

func (m Meta) Float32(key string) float32 {
	return cast.ToFloat32(m.Get(key))
}

func (m Meta) Float64(key string) float64 {
	return cast.ToFloat64(m.Get(key))
}

func (m Meta) Slice(key string) []any {
	return cast.ToSlice(m.Get(key))
}

func (m Meta) BoolSlice(key string) []bool {
	return cast.ToBoolSlice(m.Get(key))
}

func (m Meta) StrSlice(key string) []string {
	return cast.ToStringSlice(m.Get(key))
}

func (m Meta) IntSlice(key string) []int {
	return cast.ToIntSlice(m.Get(key))
}
