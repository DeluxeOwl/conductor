package conductor

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"time"
)

const (
	cmdBufSize = 1
)

type simple[T any] struct {
	listeners map[string]chan T
	mu        sync.RWMutex
	ctx       context.Context
	logFile   *os.File
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
	return c.cmd(2)
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

func (c *simple[T]) cmd(level int) <-chan T {
	_, file, line, ok := runtime.Caller(level)
	if !ok {
		fmt.Fprintln(c.logFile, "Cannot find caller")
		// XXX: we return a closed channel, as we are not able to properly return
		// a valid channel without leaking it for each case statement evaluation.
		ch := make(chan T)
		close(ch)
		return ch
	}
	key := fmt.Sprintf("%s:%d", file, line)

	c.mu.RLock()
	if ch, ok := c.listeners[key]; ok {
		c.mu.RUnlock()
		return ch
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()

	lis := make(chan T, cmdBufSize)
	c.listeners[key] = lis
	return lis
}

func (c *simple[T]) send(cmd T) {
	c.mu.RLock()
	for k, ch := range c.listeners {
		fmt.Fprintf(c.logFile, "Sending %v to %s listener\n", cmd, k)
		ch <- cmd
	}
	c.mu.RUnlock()
}

func (c *simple[T]) notify(cmd T, signals ...os.Signal) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, signals...)
	for {
		select {
		case <-c.Done():
			return
		case <-ch:
			go c.send(cmd)
		}
	}
}

/* Public functions */

// Simple creates a [Conductor] with a single type of listener.
func Simple[T any]() Conductor[T] {
	return &simple[T]{
		ctx:       context.TODO(),
		logFile:   initLogFile(),
		listeners: make(map[string]chan T),
	}
}

// SimpleFromContext creates a Simple [Conductor] from a given [context.Context].
func SimpleFromContext[T any](parent context.Context) Conductor[T] {
	c := Simple[T]()
	c.(*simple[T]).ctx = parent

	return c
}
