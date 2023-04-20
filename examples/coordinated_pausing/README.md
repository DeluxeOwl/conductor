# Example 1: simple conductor used to coordinate many workers

This example shows how to use a conductor created with
`conductor.SimpleFromContext` to coordinate multiple workers at once.

Start the example

```
$ go run main.go
```

In another terminal, send signal to the running program

```
$ kill -SIGUSR1 <PID_OF_GO_RUN>  # To pause the workers
$ kill -SIGUSR2 <PID_OF_GO_RUN>  # To unpause the workers
```

even easier with `pkill`

```
$ pkill -SIGUSR1 main
$ pkill -SIGUSR2 main
```

`ctrl+c` in the original terminal to clean exit all the workers.
