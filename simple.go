package conductor

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"time"
)

type simple[T any] struct {
	listeners []chan T
	mu        sync.Mutex
	ctx       context.Context
}

/* Implement context.Context */

var _ context.Context = &simple[struct{}]{}

func (c *simple[T]) Deadline() (time.Time, bool) {
	return c.ctx.Deadline()
}

func (c *simple[T]) Done() <-chan struct{} {
	return c.ctx.Done()
}

func (c *simple[T]) Err() error {
	return c.ctx.Err()
}

func (c *simple[T]) Value(key any) any {
	return c.ctx.Value(key)
}

/* Implement Conductor[T] */

func (c *simple[T]) Cmd() <-chan T {
	lis := make(chan T)
	c.mu.Lock()
	c.listeners = append(c.listeners, lis)
	c.mu.Unlock()
	return lis
}

func (c *simple[T]) WithContext(ctx context.Context) Conductor[T] {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.ctx = ctx
	return c
}

func (c *simple[T]) WithContextPolicy(policy Policy[T]) Conductor[T] {
	go func() {
		<-c.ctx.Done()
		if cmd, ok := policy.Decide(); ok {
			c.send(cmd)
		}
	}()

	return c
}

/* Internal functions */

func (c *simple[T]) send(cmd T) {
	c.mu.Lock()
	for _, c := range c.listeners {
		c <- cmd
	}
	c.mu.Unlock()
}

func (c *simple[T]) notify(cmd T, signals ...os.Signal) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, signals...)
	for {
		select {
		case <-c.Done():
			return
		case <-ch:
			c.send(cmd)
		}
	}
}

/* Public functions */

// Simple creates a [Conductor] with a single type of listener.
func Simple[T any]() Conductor[T] {
	return &simple[T]{
		ctx: context.TODO(),
	}
}

// SimpleFromContext creates a Simple [Conductor] from a given [context.Context].
func SimpleFromContext[T any](parent context.Context) Conductor[T] {
	c := Simple[T]()
	c.(*simple[T]).ctx = parent

	return c
}
