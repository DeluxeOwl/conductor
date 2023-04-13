# Conductor

A generalization of the [Context][ctx]

## Motivation

The standard library [Context][ctx] is a great concept. It allows to propagate values
and allows to coordinate across API and goroutine boundaries. As much as I like it, when
dealing with application that show a lot of moving pieces, I want something more in
order to coordinate them. I want a [Conductor][cdct]. In the same spirit as the `Done`
method of a [Context][ctx], a [Conductor][cdct] `Cmd` method may be used to listen for
commands. These are represented by the ubiquitous generic parameter `T` across this
library, and may be everything, from simple `string`s to interfaces carrying a
complicated logic with them.

## How to use

There are two different [Conductor][cdct]s implemented right now, a [Simple][simple] and
a [Tagged][tagged] one.

The first is straightforward to use (we take `T` to be `string`, for the sake of
simplicity):

```go
simple := Simple[string]()

// we listen for commands

for {
    select {
        case cmd := <-simple.Cmd():
                  // react to the commands
        case <-simple.Done():
                  // a Conductor is also a Context, so we can use it the same way
                  // for example, for controlling a clean exit
    }
}

// ...in another part of the control flow we can call Send to send a commands
// to all the listeners created with Cmd

Send[string](simple)("doit")

// This will fire in the above select statement, delivering "doit" to the first case
```

The second one might sound a bit more involved, but it's hopefully just a matter of
getting used to the syntax:

```go
tagged := Tagged[string]()

// we listen on many different possible tags

for {
    select {
        case cmd := <-WithTag[string](tagged, "tag1").Cmd():
            // React to a command in the "tag1" branch
        case cmd := <-WithTag[string](tagged, "tag2").Cmd():
            // React to a command in the "tag2" branch
        case <-tagged.Done():
            // As for the Simple, also the Tagged is a Context
    }
}


// We may selectively send a command, again using the Send function

Send[string](tagged, "tag1")("doitnow")

// We may also send a broadcast command

Send[string](tagged)("allhands")
```


[ctx]: https://pkg.go.dev/context#Context
[cdct]: ./conductor.go
[simple]: ./simple.go
[tagged]: ./tagged.go


<!-- vim:set ft=markdown tw=88: -->
