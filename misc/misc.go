package misc

func ToPtr[T any](t T) *T {
	return &t
}

func Filter[T comparable](list []T, fn func(T) bool) []T {
	out := make([]T, 0)
	for _, v := range list {
		if fn(v) {
			out = append(out, v)
		}
	}
	return out
}

func Extract[T any, R any](in []T, fn func(i T) R) []R {
	out := make([]R, 0)
	for _, v := range in {
		out = append(out, fn(v))
	}
	return out
}

func Contain[T comparable](in []T, target T) bool {
	return Index(in, target) != -1
}

func Index[T comparable](in []T, target T) int {
	for i, v := range in {
		if v == target {
			return i
		}
	}
	return -1
}
