package internal

import "slices"

func Map[T any, R any](collection []T, iteratee func(item T) R) []R {
	if len(collection) == 0 {
		return nil
	}

	result := make([]R, len(collection))

	for i, item := range collection {
		result[i] = iteratee(item)
	}

	return result
}

func Filter[T any](collection []T, callback func(item T) bool) []T {
	return slices.DeleteFunc(slices.Clone(collection), func(t T) bool {
		return !callback(t)
	})
}

func FilterMap[T any, R any](collection []T, callback func(item T) (R, bool)) []R {
	var result []R

	for _, item := range collection {
		if r, ok := callback(item); ok {
			result = append(result, r)
		}
	}

	return result
}

func Unique[S ~[]E, E comparable](s S) S {
	for i := len(s) - 1; i > 0; i-- {
		if slices.Index(s, s[i]) != i {
			s = slices.Delete(s, i, i+1)
		}
	}
	return s
}
