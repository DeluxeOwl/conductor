package conductor

import (
	"context"
	"os"
)

type Sender[T any] func(cmd T)

type Notifier[T any] func(cmd T, signals ...os.Signal)

type Commandable[T any] func() <-chan T

type Conductor[T any] interface {
	Cmd() <-chan T
	context.Context
}

func WithSend[T any](conductor Conductor[T]) Sender[T] {
	switch c := any(conductor).(type) {
	case *simple[T]:
		return c.send
	case *tagged[T]:
		return c.broadcast
	default:
		panic("conductor not supported")
	}
}

func WithTaggedSend[T any](conductor Conductor[T], tags ...string) Sender[T] {
	switch c := any(conductor).(type) {
	case *tagged[T]:
		return func(cmd T) {
			c.send(cmd, tags)
		}
	case *simple[T]:
		panic("simple Conductor does not support tagged send")
	default:
		panic("conductor not supported")
	}
}

func WithNotify[T any](conductor Conductor[T]) Notifier[T] {
	switch c := any(conductor).(type) {
	case *simple[T]:
		return c.notify
	case *tagged[T]:
		return c.notifyAll
	default:
		panic("conductor not supported")
	}
}

func WithTaggedNotify[T any](conductor Conductor[T], tags ...string) Notifier[T] {
	switch c := any(conductor).(type) {
	case *tagged[T]:
		return func(cmd T, signals ...os.Signal) {
			c.notifyTagged(cmd, tags, signals)
		}
	case *simple[T]:
		panic("simple Conductor does not support tagged notification")
	default:
		panic("conductor not supported")
	}
}
