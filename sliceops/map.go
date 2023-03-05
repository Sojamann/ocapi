package sliceops

import "sync"

func Map[T any, V any](items []T, fn func(T) V) []V {
	res := make([]V, len(items))
	for i, item := range items {
		res[i] = fn(item)
	}

	return res
}

func MapAsync[T any, V any](items []T, fn func(T) V) []V {
	res := make([]V, len(items))

	var wg sync.WaitGroup

	wg.Add(len(items))

	for i, item := range items {
		go func(_i int, _item T) {
			defer wg.Done()
			res[_i] = fn(_item)
		}(i, item)
	}

	wg.Wait()

	return res
}
