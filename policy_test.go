package conductor

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func Test_constantPolicy(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	c := Simple[string]().
		WithContext(ctx).
		WithContextPolicy(ConstantPolicy("ciao"))

	lis1 := c.Cmd()
	lis2 := c.Cmd()

	go cancel()

	var first, second int
testLoop:
	for {
		select {
		case cmd := <-lis1:
			if cmd != "ciao" {
				t.Fatalf("unexpected: %s", cmd)
			}
			first++
		case cmd := <-lis2:
			if cmd != "ciao" {
				t.Fatalf("unexpected: %s", cmd)
			}
			second++
		case <-time.After(successTimeout):
			break testLoop
		}
	}

	if first != 1 || second != 1 {
		t.Fatalf("Some command listener did not receive: %d, %d", first, second)
	}
}

func Test_setPolicy(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	c := Tagged[string]().
		WithContext(ctx).
		WithContextPolicy(SetPolicy(map[any]string{
			"zeroth": "ciao",
			"first":  "ciao",
		}))

	zeroth := setupListener(c, true, "zeroth")
	first := setupListener(c, true, "first")
	second := setupListener(c, false, "second")

	go cancel()

	var zerothReceived, firstReceived bool
testLoop:
	for {
		select {
		case err := <-zeroth:
			if err != nil {
				t.Fatal(err)
			}
			zerothReceived = true
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

	if !zerothReceived {
		t.Fatal("Catch all command listener did not receive")
	}

	if !firstReceived {
		t.Fatal("First command listener did not receive")
	}
}

func ExampleConstantPolicy() {
	simple, cancel := WithCancel(Simple[string]())
	simple.WithContextPolicy(ConstantPolicy("ciao"))

	lis := simple.Cmd()

	go cancel()

	cmd := <-lis
	fmt.Println(cmd)
	// Output: ciao
}

func ExampleSetPolicy() {
	tagged, cancel := WithCancel(Tagged[string]())
	tagged.WithContextPolicy(SetPolicy(map[any]string{
		"first":  "ciao",
		"second": "miao",
		"third":  "bau",
	}))

	lisFirst := WithTag(tagged, "first").Cmd()
	lisSecond := WithTag(tagged, "second").Cmd()
	lisThird := WithTag(tagged, "third").Cmd()

	go cancel()

loop:
	for {
		select {
		case cmd := <-lisFirst:
			fmt.Println(cmd)
		case cmd := <-lisSecond:
			fmt.Println(cmd)
		case cmd := <-lisThird:
			fmt.Println(cmd)
		case <-time.After(successTimeout):
			break loop
		}
	}
	// Unordered output:
	// ciao
	// miao
	// bau
}
