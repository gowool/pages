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

var replacer = strings.NewReplacer("/", " ", "-", " ", "_", " ")

func ToTitle(str string) string {
	str = replacer.Replace(str)
	str = strings.TrimSpace(str)
	str = strings.Join(strings.Fields(str), " ")
	return strings.ToTitle(str)
}
