package internal

import "hash/fnv"

func ToInt64(s string) int64 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(s))
	return int64(h.Sum32())
}
