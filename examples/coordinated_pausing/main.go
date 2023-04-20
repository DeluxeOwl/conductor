package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"syscall"
	"time"

	"git.sr.ht/~blallo/conductor"
)

const (
	workers    = 4
	timeFormat = "15:04:05.999"
)

type Action int

const (
	ActionPause Action = iota
	ActionUnpause
	ActionStop
)

func (a Action) IsStop() bool {
	return a == ActionStop
}

func (a Action) String() string {
	switch a {
	case ActionStop:
		return "STOP"
	case ActionPause:
		return "PAUSE"
	case ActionUnpause:
		return "UNPAUSE"
	default:
		panic("Impossible")
	}
}

func Worker(c conductor.Conductor[Action], out chan<- time.Time, d time.Duration) {
	currentAction := ActionUnpause

	ticker := time.NewTicker(d)
	defer ticker.Stop()

	lis := c.Cmd()
	for {
		select {
		case action := <-lis:
			switch action {
			case ActionStop:
				return
			default:
				currentAction = action
			}
		case t := <-ticker.C:
			if currentAction == ActionUnpause {
				out <- t
			}
		}
	}
}

type collected struct {
	idx  int
	time time.Time
}

func printLine(c conductor.Conductor[Action], incoming ...chan time.Time) {
	current := make([]any, len(incoming))
	collect := make(chan collected)
	var formats []string

	for i, ch := range incoming {
		current[i] = ""
		formats = append(formats, fmt.Sprintf("%d: %%s", i+1))
		go func(i int, ch <-chan time.Time) {
			lis := c.Cmd()
			for {
				select {
				case cmd := <-lis:
					if cmd.IsStop() {
						return
					}
				case t := <-ch:
					collect <- collected{idx: i, time: t}
				}
			}
		}(i, ch)
	}

	formatStr := strings.Join(formats, " |")

	lis := c.Cmd()
	for {
		select {
		case cmd := <-lis:
			fmt.Print("\033[2K\r")
			fmt.Printf("Status: %s", cmd)
			if cmd.IsStop() {
				fmt.Println("    =>   Stopping...")
				return
			}
		case data := <-collect:
			current[data.idx] = data.time.Format(timeFormat) // fmt.Sprintf("\033[42m%s\033[0m", data.time.Format(timeFormat))
			fmt.Print("\033[2K\r")
			fmt.Printf(formatStr, current...)
		}
	}
}

func funkyTime(max int) time.Duration {
	return time.Duration(float64(max)*(1/float64(max)+rand.Float64())) * time.Second
}

func init() {
	rand.Seed(time.Now().UnixNano())
	var logFile = "/tmp/conductor.log"
	conductor.LogFilePath = &logFile
}

func main() {
	c := conductor.SimpleFromContext[Action](context.Background())

	go conductor.Notify(c)(ActionStop, os.Interrupt)
	go conductor.Notify(c)(ActionPause, syscall.SIGUSR1)
	go conductor.Notify(c)(ActionUnpause, syscall.SIGUSR2)

	collectors := make([]chan time.Time, workers)
	deltas := make([]any, workers)
	deltasFmt := make([]string, workers)

	for i := 0; i < workers; i++ {
		d := funkyTime(workers)
		collector := make(chan time.Time)
		go Worker(c, collector, d)
		collectors[i] = collector
		deltas[i] = d
		deltasFmt[i] = "%s"
	}

	fmt.Printf(strings.Join(deltasFmt, "              |"), deltas...)
	fmt.Println("")
	printLine(c, collectors...)
}
