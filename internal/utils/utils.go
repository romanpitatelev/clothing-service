package utils

import "sync"

func Pointer[T any](v T) *T {
	return &v
}

type WorkerPool struct {
	pool   chan func()
	closed bool
	wg     sync.WaitGroup
}

func NewPool(size int) *WorkerPool {
	pool := WorkerPool{
		pool: make(chan func()),
	}

	for range size {
		pool.wg.Add(1)

		go func() {
			defer pool.wg.Done()

			for job := range pool.pool {
				job()
			}
		}()
	}

	return &pool
}

func (pool *WorkerPool) Put(job func()) {
	if pool.closed {
		return
	}

	pool.pool <- job
}

func (pool *WorkerPool) Close() {
	if pool.closed {
		return
	}

	pool.closed = true
	close(pool.pool)
	pool.wg.Wait()
}
