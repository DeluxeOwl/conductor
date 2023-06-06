package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"git.sr.ht/~blallo/conductor"
)

func init() {
	conductor.SetLogFile("/tmp/targeted_commands.log")
}

type Action int
type Color string

const (
	ActionStart Action = iota
	ActionStop
	ActionReset

	ColorRed    Color = "\033[0;31m"
	ColorGreen  Color = "\033[0;32m"
	ColorYellow Color = "\033[0;33m"
	ColorBlue   Color = "\033[0;34m"
	ColorPurple Color = "\033[0;35m"
	ColorCyan   Color = "\033[0;36m"
	ColorWhite  Color = "\033[0;37m"
	ColorReset  Color = "\033[0;0m"
)

func (a Action) String() string {
	switch a {
	case ActionStart:
		return "start"
	case ActionStop:
		return "stop"
	case ActionReset:
		return "reset"
	default:
		panic("not an action")
	}
}

func (c Color) String() string {
	switch c {
	case ColorRed:
		return "red"
	case ColorGreen:
		return "green"
	case ColorYellow:
		return "yellow"
	case ColorBlue:
		return "blue"
	case ColorPurple:
		return "purple"
	case ColorCyan:
		return "cyan"
	case ColorWhite:
		return "white"
	default:
		return "notacolor"
	}
}

func parseColor(col string) (color Color, err error) {
	switch col {
	case "red":
		color = ColorRed
	case "green":
		color = ColorGreen
	case "yellow":
		color = ColorYellow
	case "blue":
		color = ColorBlue
	case "purple":
		color = ColorPurple
	case "cyan":
		color = ColorCyan
	case "white":
		color = ColorWhite
	default:
		err = fmt.Errorf("color not supported: %s", col)
	}

	return
}

func (c Color) println(str string) {
	fmt.Fprintf(os.Stderr, "%s%s%s\n", string(c), str, string(ColorReset))
}

type Worker struct {
	color    Color
	interval time.Duration
}

func (w *Worker) Run(c conductor.Conductor[Action], instance int) {
	var counter int
	var running bool = true

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case action := <-conductor.WithTag(c, w.color.String(), instance).Cmd():
			switch action {
			case ActionStart:
				running = true
			case ActionStop:
				running = false
			case ActionReset:
				counter = 0
			}
		case <-ticker.C:
			if running {
				w.color.println(fmt.Sprintf("[%d] tick -> %d", instance, counter))
				counter++
			}
		}
	}
}

type WorkerMap struct {
	replicas map[Color]int
	workers  map[Color]*Worker
}

func (m *WorkerMap) ParseCmd(c conductor.Conductor[Action], cmd string) error {
	if cmd == "" {
		return fmt.Errorf("missing command")
	}

	comps := strings.Split(cmd, " ")
	switch comps[0] {
	case "add":
		var color Color
		var interval time.Duration
		var err error

		if len(comps) < 2 {
			color = ColorWhite

		} else {
			color, err = parseColor(comps[1])
			if err != nil {
				return err
			}
		}

		if len(comps) < 3 {
			interval = time.Second
		} else {
			interval, err = time.ParseDuration(comps[2])
			if err != nil {
				return err
			}
		}

		m.replicas[color]++
		w, ok := m.workers[color]
		if !ok {
			w = &Worker{
				color:    color,
				interval: interval,
			}
			m.workers[color] = w
		}

		go w.Run(c, m.replicas[color])

	case "start":
		if len(comps) < 2 {
			conductor.Send(c)(ActionStart)
		} else {
			color, err := parseColor(comps[1])
			if err != nil {
				return err
			}
			conductor.Send(c, color)(ActionStart)
		}

	case "stop":
		if len(comps) < 2 {
			conductor.Send(c)(ActionStop)
		} else {
			color, err := parseColor(comps[1])
			if err != nil {
				return err
			}
			conductor.Send(c, color)(ActionStop)
		}

	case "reset":
		if len(comps) < 2 {
			conductor.Send(c)(ActionReset)
		} else {
			color, err := parseColor(comps[1])
			if err != nil {
				return err
			}
			conductor.Send(c, color)(ActionReset)
		}

	}
	return nil
}

func main() {
	workers := &WorkerMap{
		replicas: make(map[Color]int),
		workers:  make(map[Color]*Worker),
	}
	c := conductor.TaggedFromContext[Action](context.Background())

	scanner := bufio.NewScanner(os.Stdin)

	fmt.Print("> ")
	for scanner.Scan() {
		cmd := scanner.Text()
		if err := workers.ParseCmd(c, cmd); err != nil {
			fmt.Printf("Failed: %s", err)
		} else {
			fmt.Print("OK")
		}
		fmt.Print("\n")
		fmt.Print("> ")
	}
}
