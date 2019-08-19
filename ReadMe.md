# Toolbox

Random command line tools I use day to day. Currently porting from various shell, Python, and JavaScript scripts to Go for
portability.

One-off, throw-away scripts I tend to run wherever like this:

```
///usr/bin/env go run "$0" "$@" ; exit "$?"

package main

func main() {
    println("Hello world!")
}
```
The things that get moved here are something I use regularly, and don't want to always recompiple, but do want to revision control.

## Tools

