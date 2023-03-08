package sliceops

import (
	"math"
)

type TaskFn[T any] func(chan<- T, chan<- error)

type Task[T any] struct {
	total        int
	itemChan     chan T
	errChan      chan error
	progressChan chan int
	doneChan     chan interface{}

	collectedItems []T
	errs           []error
}

// In order to not leak go routines one should either
// send into itemChan or into errChan because we
// read exactly N items.
func NewTask[T any](n int, fn TaskFn[T]) *Task[T] {
	t := Task[T]{
		total:          n,
		itemChan:       make(chan T),
		errChan:        make(chan error),
		progressChan:   make(chan int),
		doneChan:       make(chan interface{}),
		collectedItems: make([]T, 0, n),
		errs:           make([]error, 0),
	}
	go fn(t.itemChan, t.errChan)
	go func() {
		defer close(t.errChan)
		defer close(t.itemChan)
		defer close(t.progressChan)
		defer close(t.doneChan)

		// read up to n times from the channels
		for i := 0; i < n; i++ {
			select {
			case err := <-t.errChan:
				t.errs = append(t.errs, err)
			case item := <-t.itemChan:
				t.collectedItems = append(t.collectedItems, item)
				t.progressChan <- int(math.Ceil(float64(i) / float64(n) * 100))
			}
		}
	}()
	return &t
}

func (t *Task[T]) Progress() <-chan int {
	return t.progressChan
}

func (t *Task[T]) Wait() {
	<-t.doneChan
}

func (t *Task[T]) Completed() bool {
	select {
	case _, ok := <-t.doneChan:
		return !ok
	default:
		return false
	}
}

func (t *Task[T]) Items() []T {
	if !t.Completed() {
		panic("Cannot call Items of non completed task")
	}

	return t.collectedItems
}

func (t *Task[T]) Errors() []error {
	if !t.Completed() {
		panic("Cannot call Err of non completed task")
	}

	return t.errs
}
