package pages

import (
	"maps"
	"time"

	"github.com/spf13/cast"
)

type Metadata map[string]any

func NewMetadata(m map[string]any) Metadata {
	if m == nil {
		return make(Metadata)
	}

	metadata := make(Metadata, len(m))
	maps.Copy(metadata, m)

	return metadata
}

func (m Metadata) Set(key string, value any) {
	m[key] = value
}

func (m Metadata) Get(key string) any {
	return m[key]
}

func (m Metadata) GetOK(key string) (any, bool) {
	v, ok := m[key]
	return v, ok
}

func (m Metadata) Has(key string) bool {
	_, ok := m[key]
	return ok
}

func (m Metadata) Delete(key string) {
	delete(m, key)
}

func (m Metadata) Str(key string) string {
	return cast.ToString(m.Get(key))
}

func (m Metadata) Bool(key string) bool {
	return cast.ToBool(m.Get(key))
}

func (m Metadata) Time(key string) time.Time {
	return cast.ToTime(m.Get(key))
}

func (m Metadata) Duration(key string) time.Duration {
	return cast.ToDuration(m.Get(key))
}

func (m Metadata) Int(key string) int {
	return cast.ToInt(m.Get(key))
}

func (m Metadata) Int8(key string) int8 {
	return cast.ToInt8(m.Get(key))
}

func (m Metadata) Int16(key string) int16 {
	return cast.ToInt16(m.Get(key))
}

func (m Metadata) Int32(key string) int32 {
	return cast.ToInt32(m.Get(key))
}

func (m Metadata) Int64(key string) int64 {
	return cast.ToInt64(m.Get(key))
}

func (m Metadata) Uint(key string) uint {
	return cast.ToUint(m.Get(key))
}

func (m Metadata) Uint8(key string) uint8 {
	return cast.ToUint8(m.Get(key))
}

func (m Metadata) Uint16(key string) uint16 {
	return cast.ToUint16(m.Get(key))
}

func (m Metadata) Uint32(key string) uint32 {
	return cast.ToUint32(m.Get(key))
}

func (m Metadata) Uint64(key string) uint64 {
	return cast.ToUint64(m.Get(key))
}

func (m Metadata) Float32(key string) float32 {
	return cast.ToFloat32(m.Get(key))
}

func (m Metadata) Float64(key string) float64 {
	return cast.ToFloat64(m.Get(key))
}

func (m Metadata) Slice(key string) []any {
	return cast.ToSlice(m.Get(key))
}

func (m Metadata) BoolSlice(key string) []bool {
	return cast.ToBoolSlice(m.Get(key))
}

func (m Metadata) StrSlice(key string) []string {
	return cast.ToStringSlice(m.Get(key))
}

func (m Metadata) IntSlice(key string) []int {
	return cast.ToIntSlice(m.Get(key))
}
