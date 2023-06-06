package main

import (
	"fmt"
	"time"

	"git.sr.ht/~blallo/conductor"
)

func init() {
	conductor.SetLogFile("/tmp/wrong.log")
}

func run(c conductor.Conductor[int], inst int) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	// You should do this
	// lis := c.Cmd()

	for {
		select {
		// You should do this
		// case cmd := <-lis:
		case cmd := <-c.Cmd():
			fmt.Println(inst, "received:", cmd)
		case <-ticker.C:
			fmt.Println(inst, "tick")
		}
	}
}

func main() {
	c := conductor.Simple[int]()

	for i := 0; i < 5; i++ {
		go run(c, i)
	}

	time.Sleep(10 * time.Second)
	conductor.Send[int](c)(0)
	time.Sleep(500 * time.Millisecond)
}
