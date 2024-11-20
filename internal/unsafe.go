package internal

import "unsafe"

// Bytes converts string to byte slice without a memory allocation.
// For more details, see https://github.com/golang/go/issues/53003#issuecomment-1140276077.
func Bytes(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

// String converts byte slice to string without a memory allocation.
// For more details, see https://github.com/golang/go/issues/53003#issuecomment-1140276077.
func String(b []byte) string {
	return unsafe.String(unsafe.SliceData(b), len(b))
}
