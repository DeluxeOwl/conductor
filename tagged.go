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
	tag    *string
}

func Tagged[T any]() *tagged[T] {
	return &tagged[T]{
		tagged: make(map[string]*simple[T]),
		ctx:    context.TODO(),
	}
}

func WithTag[T any](conductor Conductor[T], tag string) Conductor[T] {
	c, ok := any(conductor).(*tagged[T])
	if !ok {
		panic("not a conductor.Tagged")
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.tag = &tag
	return c
}

func WithTags[T any](conductor Conductor[T], tags ...string) Conductor[T] {
	t, ok := any(conductor).(*tagged[T])
	if !ok {
		panic("not a conductor.Tagged")
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	cmds := make([]chan T, 0, len(tags))
	for _, tag := range tags {
		if c, ok := t.tagged[tag]; ok {
			lis := make(chan T)
			cmds = append(cmds, lis)
			c.listeners = append(c.listeners, lis)
		}
	}
	s := Simple[T]()
	go func() {
		for {
			cmd := <-s.Cmd()
			for _, cmdCh := range cmds {
				cmdCh <- cmd
			}
		}
	}()
	return s
}

func (t *tagged[T]) Cmd() <-chan T {
	if t.tag == nil {
		panic("tag not set")
	}
	t.mu.RLock()
	if c, ok := t.tagged[*t.tag]; ok {
		defer t.mu.RUnlock()
		return c.Cmd()
	}
	t.mu.RUnlock()
	t.mu.Lock()
	defer t.mu.Unlock()
	c := FromContext[T](t.ctx)
	t.tagged[*t.tag] = c
	t.tag = nil
	return c.Cmd()
}

func (t *tagged[T]) send(cmd T, tags []string) {
	t.mu.RLock()
	for _, tag := range tags {
		if c, ok := t.tagged[tag]; ok {
			c.send(cmd)
		}
	}
	t.mu.RUnlock()
}

func (t *tagged[T]) broadcast(cmd T) {
	t.mu.RLock()
	for _, c := range t.tagged {
		c.send(cmd)
	}
	t.mu.RUnlock()
}

func (t *tagged[T]) notifyAll(cmd T, signals ...os.Signal) {
	ch := make(chan os.Signal)
	signal.Notify(ch, signals...)
	go func() {
		for {
			<-ch
			t.broadcast(cmd)
		}
	}()
}

func (t *tagged[T]) notifyTagged(cmd T, tags []string, signals []os.Signal) {
	ch := make(chan os.Signal)
	signal.Notify(ch, signals...)
	go func() {
		for {
			<-ch
			t.send(cmd, tags)
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

func TaggedFromContext[T any](parent context.Context) Conductor[T] {
	t := Tagged[T]()
	t.ctx = parent

	return t
}

func TaggedFromSimple[T any](s Conductor[T]) Conductor[T] {
	c, ok := any(s).(*simple[T])
	if !ok {
		panic("not a conductor.Simple")
	}
	return &tagged[T]{
		tagged: map[string]*simple[T]{
			"": c,
		},
		ctx: c.ctx,
	}
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
				lis.send(cmd)
			}
		}
	}()

	return c
}
