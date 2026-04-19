package main

func Map[T, R any](arr []T, fn func(T, int) R) []R {
	res := make([]R, 0, len(arr))
	for i, v := range arr {
		res = append(res, fn(v, i))
	}
	return res
}
