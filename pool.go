package gpool

import (
	"errors"
	"time"
)

const defaultTimeout = 1 * time.Second

var ErrPoolTimeout = errors.New("pool timeout: no free slot")

// Pool is a simple buffered thread-safe object pool for managing objects of a single type.
type Pool[T any] struct {
	Size    int
	NewF    func() T
	Timeout time.Duration
	pool    chan T
}

// New creates a new object pool with the specified size, object creation function, and optional timeout.
// If no timeout is provided, the default timeout value will be used.
func New[T any](size int, newF func() T, args ...time.Duration) *Pool[T] {
	var timeout time.Duration
	if len(args) > 0 {
		timeout = args[0]
	}

	return (&Pool[T]{Size: size, NewF: newF, Timeout: timeout}).Init()
}

// Init initializes the object pool by creating the specified number of objects and adding them to the pool.
// If no timeout is set, the default timeout value will be used.
func (s *Pool[T]) Init() *Pool[T] {
	if s.Timeout == 0 {
		s.Timeout = defaultTimeout
	}

	s.pool = make(chan T, s.Size)
	for i := 0; i < cap(s.pool); i++ {
		s.pool <- s.NewF()
	}
	return s
}

// Acquire acquires an object from the pool. If there are no available objects within the timeout period,
// it returns an error indicating a pool timeout.
func (s *Pool[T]) Acquire() (obj T, err error) {
	select {
	case obj = <-s.pool:
		return
	case <-time.After(s.Timeout):
		err = ErrPoolTimeout
		return
	}
}

// Release releases an object back to the pool.
func (s *Pool[T]) Release(x T) {
	s.pool <- x
}

// Close closes the pool and releases all of its resources.
func (s *Pool[T]) Close() {
	close(s.pool)

}
