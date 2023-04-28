package conductor

import (
	"context"
	"fmt"
	"os"
	"syscall"
	"testing"
	"time"
)

func TestSend_simple(t *testing.T) {
	c := Simple[string]()

	start := make(chan struct{})

	go func() {
		<-start
		Send(c)("ciao")
	}()

	select {
	case start <- struct{}{}:
		// XXX start the Send in the goroutine scheduled above
	case cmd := <-c.Cmd():
		if cmd != "ciao" {
			t.Fatalf("Unexpected cmd: %s", cmd)
		}
	case <-time.After(failureTimeout):
		t.Fatal("Timeout")
	}
}

func TestNotify_simple(t *testing.T) {
	c := Simple[string]()

	go Notify(c)("ciao", syscall.SIGUSR1)

    // XXX: flakey test
    var attempts int
loop:
	for ;attempts < 5; attempts++ {
		p, err := os.FindProcess(os.Getpid())
		if err != nil {
			t.Fatalf("Could not find the process: %s", err)
		}

		if err := p.Signal(syscall.SIGUSR1); err != nil {
			t.Fatalf("Failed to send signal: %s", err)
		}

		select {
		case cmd := <-c.Cmd():
			if cmd != "ciao" {
				t.Fatalf("Unexpected cmd: %s", cmd)
			}
			break loop
		case <-time.After(failureTimeout):
		}
	}

    if attempts == 5 {
        t.Fatal("Failed to notify 5 times in a row")
    }
}

func TestAsContext(t *testing.T) {
	ctx := context.TODO()
	c := SimpleFromContext[string](ctx)

	expiring, cancel := context.WithTimeout(c, 10*time.Millisecond)
	defer cancel()

	select {
	case <-expiring.Done():
		// success
	case <-time.After(failureTimeout):
		t.Fatalf("Timeout")
	}
}

func ExampleSend_simple() {
	simple := SimpleFromContext[string](context.Background())

	lis := simple.Cmd()

	go Send(simple)("ciao")

	cmd := <-lis
	fmt.Println(cmd)
	// Output: ciao
}
