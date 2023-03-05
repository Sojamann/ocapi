package sliceops

import "sync"

type T2[T any, V any] struct {
	a T
	b V
}

func Combine[T any, V any](items []T, fn func(T) V) []T2[T, V] {
	res := make([]T2[T, V], len(items))
	for i, item := range items {
		res[i] = T2[T, V]{item, fn(item)}
	}

	return res
}

func CombineAsync[T any, V any](items []T, fn func(T) V) []T2[T, V] {
	res := make([]T2[T, V], len(items))

	var wg sync.WaitGroup

	wg.Add(len(items))

	for i, item := range items {
		go func(_i int, _item T) {
			defer wg.Done()
			res[_i] = T2[T, V]{_item, fn(_item)}
		}(i, item)
	}

	wg.Wait()

	return res
}
