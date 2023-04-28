package conductor

import (
	"testing"
	"time"
)

func TestWithCancel(t *testing.T) {
	c, cancel := WithCancel(Simple[string]())

	go cancel()

	select {
	case <-c.Done():
		// Happy path
	case <-time.After(50 * time.Millisecond):
		t.Fatal("Timeout")
	}
}

func TestWithDeadline(t *testing.T) {
	const delta = 25 * time.Millisecond
	now := time.Now()

	c, cancel := WithDeadline(Simple[string](), now.Add(delta))

	go cancel()

	select {
	case <-c.Done():
		// Happy path
	case <-time.After(2 * delta):
		t.Fatal("Timeout")
	}
}

func TestWithTimeout(t *testing.T) {
	const delta = 25 * time.Millisecond

	c, cancel := WithTimeout(Simple[string](), delta)

	go cancel()

	select {
	case <-c.Done():
		// Happy path
	case <-time.After(2 * delta):
		t.Fatal("Timeout")
	}
}
