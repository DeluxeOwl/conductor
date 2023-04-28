package conductor

import (
	"fmt"
	"time"
)

const (
	failureTimeout = 50 * time.Millisecond
	successTimeout = 50 * time.Millisecond
)

func setupListener(c Conductor[string], shouldReceive bool, tag ...string) chan error {
	errCh := make(chan error)

	var lisName string
	var lis <-chan string
	if len(tag) == 0 {
		lisName = "default"
		lis = c.Cmd()
	} else {
		lisName = tag[0]
		lis = WithTag(c, tag[0]).Cmd()
	}

	go func() {
		var err error
		cmd := <-lis
		if shouldReceive {
			if cmd != "ciao" {
				err = fmt.Errorf("[%s] unexpected: %s", lisName, cmd)
			}
		} else {
			err = fmt.Errorf("[%s] unexpected: %s", lisName, cmd)
		}
		errCh <- err
	}()

	return errCh
}
