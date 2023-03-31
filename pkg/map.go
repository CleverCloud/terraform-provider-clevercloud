package pkg

func Map[T, U any](items []T, fn func(T) U) []U {
	r := make([]U, len(items))

	for i, item := range items {
		r[i] = fn(item)
	}

	return r
}

// Filter items in a list
// return true => keep item
func Filter[T any](items []T, fn func(T) bool) []T {
	r := []T{}

	for _, item := range items {
		if fn(item) {
			r = append(r, item)
		}
	}

	return r
}

// Test if any item match a criteria
func HasSome[T any](items []T, fn func(item T) bool) bool {
	for _, item := range items {
		if fn(item) {
			return true
		}
	}

	return false
}

// return the first element mathing a criteria
func First[T any](items []T, fn func(item T) bool) *T {
	for _, item := range items {
		if fn(item) {
			return &item
		}
	}

	return nil
}

func Reduce[T, U any](items []T, acc U, fn func(U, T) U) U {
	for _, item := range items {
		acc = fn(acc, item)
	}

	return acc
}

// Compute a new aggregated map
func Merge[T comparable, U any](m1, m2 map[T]U) map[T]U {
	m := map[T]U{}

	for k, v := range m1 {
		m[k] = v
	}
	for k, v := range m2 {
		m[k] = v
	}

	return m
}
