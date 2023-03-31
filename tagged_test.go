package conductor

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestSend_tagged(t *testing.T) {
	c := Tagged[string]()

	sync1 := make(chan struct{})
	sync2 := make(chan struct{})
	errCh := make(chan error)

	go func() {
		sync1 <- struct{}{}
		var err error
		cmd := <-WithTag[string](c, "first").Cmd()
		if cmd != "ciao" {
			err = fmt.Errorf("[first] unexpected: %s", cmd)
		}
		errCh <- err
	}()

	go func() {
		sync2 <- struct{}{}
		var err error
		cmd := <-WithTag[string](c, "second").Cmd()
		if cmd != "ciao" {
			err = fmt.Errorf("[second] unexpected: %s", cmd)
		}
		errCh <- err
	}()

	go func() {
		<-sync1
		<-sync2
		WithSend[string](c)("ciao")
	}()

	counter := 0
testLoop:
	for {
		select {
		case err := <-errCh:
			if err != nil {
				t.Fatal(err)
			}
			counter++
			if counter == 2 {
				break testLoop
			}
		case <-time.After(50 * time.Millisecond):
			t.Fatalf("Timeout")
		}
	}
}

func TestTaggedSend(t *testing.T) {
	c := Tagged[string]()

	go WithTaggedSend[string](c, "first")("ciao")

	first := make(chan error)
	go func() {
		var err error
		cmd := <-WithTag[string](c, "first").Cmd()
		if cmd != "ciao" {
			err = fmt.Errorf("[first] unexpected: %s", cmd)
		}
		first <- err
	}()

	second := make(chan error)
	go func() {
		var err error
		cmd := <-WithTag[string](c, "second").Cmd()
		if cmd != "ciao" {
			err = fmt.Errorf("[second] unexpected: %s", cmd)
		}
		second <- err
	}()

testLoop:
	for {
		select {
		case err := <-first:
			if err != nil {
				t.Fatal(err)
			}
		case err := <-second:
			t.Fatalf("second received: %s", err)
		case <-time.After(50 * time.Millisecond):
			break testLoop
		}
	}
}

func TestTaggedAsContext(t *testing.T) {
	ctx := context.TODO()
	c := TaggedFromContext[string](ctx)

	expiring, cancel := context.WithTimeout(c, 10*time.Millisecond)
	defer cancel()

	select {
	case <-expiring.Done():
		// success
	case <-time.After(50 * time.Millisecond):
		t.Fatalf("Timeout")
	}
}
