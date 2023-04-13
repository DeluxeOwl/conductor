package conductor

import (
	"context"
	"time"
)

// WithCancel mimics what [context.WithCancel] does, but for a [Conductor]. It returns
// a *copy* of the given conductor, that behaves as a context subject to a cancel function
// returned as second value of the output.
func WithCancel[T any](conductor Conductor[T]) (Conductor[T], context.CancelFunc) {
	ctx, cancel := context.WithCancel(conductor)
	return newConductorWithCtx(conductor, ctx), cancel
}

// WithDeadline mimics what [context.WithCancel] does, but for a [Conductor]. It returns
// a *copy* of the given conductor, that behaves as a context subject to a cancel function
// returned as second value of the output, and that will be cancelled at the given deadline.
func WithDeadline[T any](conductor Conductor[T], deadline time.Time) (Conductor[T], context.CancelFunc) {
	ctx, cancel := context.WithDeadline(conductor, deadline)
	return newConductorWithCtx(conductor, ctx), cancel
}

// WithTimeout mimics what [context.WithCancel] does, but for a [Conductor]. It returns
// a *copy* of the given conductor, that behaves as a context subject to a cancel function
// returned as second value of the output, and that will be cancelled after the given interval.
func WithTimeout[T any](conductor Conductor[T], interval time.Duration) (Conductor[T], context.CancelFunc) {
	ctx, cancel := context.WithTimeout(conductor, interval)
	return newConductorWithCtx(conductor, ctx), cancel
}

func newConductorWithCtx[T any](conductor Conductor[T], ctx context.Context) Conductor[T] {
	switch c := any(conductor).(type) {
	case *simple[T]:
		return &simple[T]{
			listeners: c.listeners,
			ctx:       ctx,
		}
	case *tagged[T]:
		listeners := make(map[any]*simple[T])
		for k, v := range c.tagged {
			newV := &simple[T]{
				listeners: v.listeners,
				ctx:       ctx,
			}
			listeners[k] = newV
		}
		return &tagged[T]{
			tagged: listeners,
			ctx:    ctx,
		}
	default:
		panic("unsupported conductor")
	}
}
