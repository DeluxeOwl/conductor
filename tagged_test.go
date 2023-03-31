package conductor

import (
	"context"
	"testing"
	"time"
)

func TestTaggedSend_all(t *testing.T) {
	c := Tagged[string]()

	go c.Send("ciao")

	select {
	case cmd := <-c.Cmd():
		if cmd != "ciao" {
			t.Fatalf("Unexpected cmd: %s", cmd)
		}
	case <-time.After(50 * time.Millisecond):
		t.Fatalf("Timeout")
	}
}

func TestTaggedAsContext(t *testing.T) {
	ctx := context.TODO()
	c := FromContext[string](ctx)

	expiring, cancel := context.WithTimeout(c, 10*time.Millisecond)
	defer cancel()

	select {
	case <-expiring.Done():
		// success
	case <-time.After(50 * time.Millisecond):
		t.Fatalf("Timeout")
	}
}

func setupChild() chan error {
    errCh := make(chan error)

    go func(){

    }()
}
