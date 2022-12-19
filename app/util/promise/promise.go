package promise

import "sync"

type Promise[T any] interface {
	Resolve(value T)
	OnProvided(func(T))
}

type promiseImpl[T any] struct {
	m           *sync.Mutex
	value       *T
	subscribers []func(T)
}

func New[T any]() Promise[T] {
	return promiseImpl[T]{m: &sync.Mutex{}, value: nil, subscribers: make([]func(T), 0)}
}

func (p promiseImpl[T]) Resolve(value T) {
	p.m.Lock()
	defer p.m.Unlock()
	p.value = &value
	wg := sync.WaitGroup{}
	wg.Add(len(p.subscribers))
	for _, subscriber := range p.subscribers {
		subscriber := subscriber
		go func() {
			subscriber(value)
			wg.Done()
		}()
	}
	wg.Wait()
}

func (p promiseImpl[T]) OnProvided(f func(T)) {
	p.m.Lock()
	defer p.m.Unlock()
	if p.value != nil {
		f(*p.value)
		return
	}
	p.subscribers = append(p.subscribers, f)
}
