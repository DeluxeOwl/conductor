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

	go Send[string](c)("ciao")

	select {
	case cmd := <-c.Cmd():
		if cmd != "ciao" {
			t.Fatalf("Unexpected cmd: %s", cmd)
		}
	case <-time.After(50 * time.Millisecond):
		t.Fatal("Timeout")
	}
}

func TestNotify_simple(t *testing.T) {
	c := Simple[string]()
	start := make(chan struct{})

	go func() {
		start <- struct{}{}
		Notify[string](c)("ciao", syscall.SIGUSR1)
	}()

	<-start

	p, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Fatal(err)
	}

	if err := p.Signal(syscall.SIGUSR1); err != nil {
		t.Fatal(err)
	}

	select {
	case cmd := <-c.Cmd():
		if cmd != "ciao" {
			t.Fatalf("Unexpected cmd: %s", cmd)
		}
	case <-time.After(50 * time.Millisecond):
		t.Fatalf("Timeout")
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
	case <-time.After(50 * time.Millisecond):
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
