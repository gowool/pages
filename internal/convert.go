package internal

import (
	"strings"
	"unsafe"
)

// BytesToString converts byte slice to string without a memory allocation.
// For more details, see https://github.com/golang/go/issues/53003#issuecomment-1140276077.
func BytesToString(b []byte) string {
	return unsafe.String(unsafe.SliceData(b), len(b))
}

func ToTitle(str string) string {
	str = strings.TrimLeft(str, "/")
	str = strings.ReplaceAll(str, "/", " ")
	return strings.ToTitle(str)
}
