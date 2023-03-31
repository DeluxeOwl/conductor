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

func Simple[T any]() *simple[T] {
	return &simple[T]{
		ctx: context.TODO(),
	}
}

func (c *simple[T]) Cmd() <-chan T {
	lis := make(chan T)
	c.mu.Lock()
	c.listeners = append(c.listeners, lis)
	c.mu.Unlock()
	return lis
}

func (c *simple[T]) Send(cmd T) {
	c.mu.Lock()
	for _, c := range c.listeners {
		c <- cmd
	}
	c.mu.Unlock()
}

func (c *simple[T]) Notify(cmd T, signals ...os.Signal) {
	ch := make(chan os.Signal)
	signal.Notify(ch, signals...)
	go func() {
		for {
			<-ch
			c.Send(cmd)
		}
	}()
}

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

var _ context.Context = &simple[struct{}]{}

func FromContext[T any](parent context.Context) *simple[T] {
	c := Simple[T]()
	c.ctx = parent

	return c
}

func (c *simple[T]) WithContext(ctx context.Context) *simple[T] {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.ctx = ctx
	return c
}

func (c *simple[T]) WithContextPolicy(policy Policy[T]) *simple[T] {
	go func() {
		<-c.ctx.Done()
		if cmd, ok := policy.Decide(); ok {
			c.Send(cmd)
		}
	}()

	return c
}
