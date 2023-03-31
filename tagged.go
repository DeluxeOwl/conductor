package conductor

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"time"
)

type tagged[T any] struct {
	tagged map[string]*simple[T]
	mu     sync.RWMutex
	ctx    context.Context
}

func Tagged[T any]() *tagged[T] {
	return &tagged[T]{
		tagged: make(map[string]*simple[T]),
		ctx:    context.TODO(),
	}
}

func (t *tagged[T]) Cmd(tag string) <-chan T {
	t.mu.RLock()
	if c, ok := t.tagged[tag]; ok {
		defer t.mu.RUnlock()
		return c.Cmd()
	}
	t.mu.RUnlock()
	t.mu.Lock()
	defer t.mu.Unlock()
	c := FromContext[T](t.ctx)
	t.tagged[tag] = c
	return c.Cmd()
}

func (t *tagged[T]) Send(cmd T, tags ...string) {
	t.mu.RLock()
	if len(tags) == 0 {
		t.broadcast(cmd)
	} else {
		t.send(cmd, tags)
	}
	t.mu.RUnlock()
}

func (t *tagged[T]) send(cmd T, tags []string) {
	for _, tag := range tags {
		if c, ok := t.tagged[tag]; ok {
			c.Send(cmd)
		}
	}
}

func (t *tagged[T]) broadcast(cmd T) {
	for _, c := range t.tagged {
		c.Send(cmd)
	}
}

func (t *tagged[T]) Notify(cmd T, tag string, signals ...os.Signal) {
	ch := make(chan os.Signal)
	signal.Notify(ch, signals...)
	go func() {
		for {
			<-ch
			t.Send(cmd, tag)
		}
	}()
}

func (t *tagged[T]) Deadline() (time.Time, bool) {
	return t.ctx.Deadline()
}

func (t *tagged[T]) Done() <-chan struct{} {
	return t.ctx.Done()
}

func (t *tagged[T]) Err() error {
	return t.ctx.Err()
}

func (t *tagged[T]) Value(key any) any {
	return t.ctx.Value(key)
}

var _ context.Context = &tagged[struct{}]{}

func TaggedFromContext[T any](parent context.Context) *tagged[T] {
	t := Tagged[T]()
	t.ctx = parent

	return t
}

func (t *tagged[T]) WithContext(ctx context.Context) *tagged[T] {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.ctx = ctx
	for _, tagged := range t.tagged {
		tagged.WithContext(ctx)
	}
	return t
}

func (c *tagged[T]) WithContextPolicy(policy TaggedPolicy[T]) *tagged[T] {
	go func() {
		<-c.ctx.Done()
		c.mu.Lock()
		defer c.mu.Unlock()
		for tag, lis := range c.tagged {
			if cmd, ok := policy.Decide(tag); ok {
				lis.Send(cmd)
			}
		}
	}()

	return c
}
