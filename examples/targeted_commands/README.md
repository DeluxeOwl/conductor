# Example 2: tagged conductor used to selectively control groups of routines

This example shows how to use a conductor created with
`conductor.TaggedFromContext` to control groups of running tasks, grouped by
"color".

Start the example in a terminal, redirecting to a file the standard error

```
$ go run main.go 2>/tmp/coordinated.out
> 
```

In another terminal, follow the lines output in `/tmp/coordinated.out` (you
will need a terminal that supports vt100 escapes).

```
$ tail -f /tmp/coordinated.log
```

In the first terminal you will be prompted to control the execution of the
program. You can add a colored task with the syntax `add <color> [interval]`.
The `color` parameter is mandatory, the `interval` is optional and defaults to
1s. For example

```
> add red
OK
> add red 1.5s
OK
> add white
OK
> add green 3s
OK
```

You will begin to see colored lines being output on the second terminal. Now
you can control the execution of the program with the following commands

```
stop [color]
reset [color]
start [color]
```

If `color` is not specified, the command is broadcast to all colors.
