package pkg

func MergeMap[K comparable, V any](maps ...map[K]V) map[K]V {
	r := map[K]V{}

	for _, m := range maps {
		for k, v := range m {
			r[k] = v
		}
	}

	return r
}

func ReduceMap[T comparable, U, V any](m map[T]U, fn func(V, T, U) V) V {
	var acc V

	for t, u := range m {
		acc = fn(acc, t, u)
	}

	return acc
}

func MapList[T any, U any](items []T, fn func(T) U) []U {
	n := make([]U, len(items))

	for i, item := range items {
		n[i] = fn(item)
	}

	return n
}

func ReduceList[T any, U any](items []T, initialAcc U, fn func(U, T) U) U {

	for _, item := range items {
		initialAcc = fn(initialAcc, item)
	}

	return initialAcc
}
