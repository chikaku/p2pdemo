package main

func Must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

func Map[T any, R any](v []T, f func(T) R) []R {
	result := make([]R, len(v))
	for i, v := range v {
		result[i] = f(v)
	}
	return result
}
