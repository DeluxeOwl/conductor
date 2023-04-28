package conductor

import (
	"context"
	"fmt"
	"os"
	"syscall"
	"testing"
	"time"
)

func TestSend_tagged(t *testing.T) {
	c := Tagged[string]()

	catchAll := setupListener(c, true)
	first := setupListener(c, true, "first")
	second := setupListener(c, true, "second")

	go Send[string](c)("ciao")

	var allReceived, firstReceived, secondReceived bool
testLoop:
	for {
		select {
		case err := <-catchAll:
			if err != nil {
				t.Fatal(err)
			}
			allReceived = true
		case err := <-first:
			if err != nil {
				t.Fatal(err)
			}
			firstReceived = true
		case err := <-second:
			if err != nil {
				t.Fatal(err)
			}
			secondReceived = true
		case <-time.After(successTimeout):
			break testLoop
		}
	}

	if !allReceived {
		t.Fatal("Catch all command listener did not receive")
	}

	if !firstReceived {
		t.Fatal("First command listener did not receive")
	}

	if !secondReceived {
		t.Fatal("Second command listener did not receive")
	}
}

func TestSend_tagged_specifyingTag(t *testing.T) {
	c := Tagged[string]()

	catchAll := setupListener(c, true)
	first := setupListener(c, true, "first")
	second := setupListener(c, false, "second")

	go Send(c, "first")("ciao")

	var allReceived, firstReceived bool
testLoop:
	for {
		select {
		case err := <-catchAll:
			if err != nil {
				t.Fatal(err)
			}
			allReceived = true
		case err := <-first:
			if err != nil {
				t.Fatal(err)
			}
			firstReceived = true
		case err := <-second:
			t.Fatalf("second received: %s", err)
		case <-time.After(successTimeout):
			break testLoop
		}
	}

	if !allReceived {
		t.Fatal("Catch all command listener did not receive")
	}

	if !firstReceived {
		t.Fatal("First command listener did not receive")
	}
}

func TestNotify_tagged_notifyAll(t *testing.T) {
	c := Tagged[string]()

	started := make(chan struct{})
	go func() {
		started <- struct{}{}
		Notify[string](c)("ciao", syscall.SIGTSTP)
	}()

	catchAll := setupListener(c, true)
	first := setupListener(c, true, "first")
	second := setupListener(c, true, "second")

	<-started
	// XXX: this is not elegant, but it avoids deadlocks.
	time.Sleep(50 * time.Millisecond)

	p, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Fatal(err)
	}

	if err := p.Signal(syscall.SIGTSTP); err != nil {
		t.Fatal(err)
	}

	var allReceived, firstReceived, secondReceived bool
testLoop:
	for {
		select {
		case err := <-catchAll:
			if err != nil {
				t.Fatal(err)
			}
			allReceived = true
		case err := <-first:
			if err != nil {
				t.Fatal(err)
			}
			firstReceived = true
		case err := <-second:
			if err != nil {
				t.Fatal(err)
			}
			secondReceived = true
		case <-time.After(successTimeout):
			break testLoop
		}
	}

	if !allReceived {
		t.Fatal("Catch all command listener did not receive")
	}

	if !firstReceived {
		t.Fatal("First command listener did not receive")
	}

	if !secondReceived {
		t.Fatal("Second command listener did not receive")
	}
}

func TestNotify_tagged_notifyTagged(t *testing.T) {
	c := Tagged[string]()

	started := make(chan struct{})
	go func() {
		started <- struct{}{}
		Notify(c, "first")("ciao", syscall.SIGUSR2)
	}()

	<-started

	catchAll := setupListener(c, true)
	first := setupListener(c, true, "first")
	second := setupListener(c, false, "second")

	p, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Fatal(err)
	}

	if err := p.Signal(syscall.SIGUSR2); err != nil {
		t.Fatal(err)
	}

	var allReceived, firstReceived bool
testLoop:
	for {
		select {
		case err := <-catchAll:
			if err != nil {
				t.Fatal(err)
			}
			allReceived = true
		case err := <-first:
			if err != nil {
				t.Fatal(err)
			}
			firstReceived = true
		case err := <-second:
			t.Fatalf("second received: %s", err)
		case <-time.After(successTimeout):
			break testLoop
		}
	}

	if !allReceived {
		t.Fatal("Catch all command listener did not receive")
	}

	if !firstReceived {
		t.Fatal("First command listener did not receive")
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
	case <-time.After(failureTimeout):
		t.Fatalf("Timeout")
	}
}

func ExampleSend_tagged() {
	tagged := TaggedFromContext[string](context.Background())

	lisFirst := WithTag(tagged, "first").Cmd()
	lisSecond := WithTag(tagged, "second").Cmd()
	lisThird := WithTag(tagged, "third").Cmd()

	tagged, cancel := WithCancel(tagged)
	feedback := make(chan struct{}, 3)

	go func() {
		Send(tagged, "first")("ciao")
		Send(tagged, "second")("miao")
		Send(tagged, "third")("bau")
		<-feedback
		<-feedback
		<-feedback
		cancel()
	}()

loop:
	for {
		select {
		case cmd := <-lisFirst:
			fmt.Println(cmd)
			feedback <- struct{}{}
		case cmd := <-lisSecond:
			fmt.Println(cmd)
			feedback <- struct{}{}
		case cmd := <-lisThird:
			fmt.Println(cmd)
			feedback <- struct{}{}
		case <-tagged.Done():
			break loop
		}
	}
	// Unordered output:
	// ciao
	// miao
	// bau
}
