# Example 3: avoid to walk the stack

This example serves the only purpose to show how inefficient might be to use
the `Cmd` call directly in a `case` statement.

To run, in a terminal

```
$ go run main.go
```

You can then follow the log at `/tmp/wrong.log`:

```
$ tail -f /tmp/wrong.log
```

Try to follow the suggestion in the code and see what happens.
