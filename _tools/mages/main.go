///usr/bin/env go run github.com/magefile/mage "$@" ; exit "$?"

// This is a stub file so you can either execute `./main.go <target>` or `go run main.go <target>`
package main

import (
"os"
"github.com/magefile/mage/mage"
)

func main() { os.Exit(mage.Main()) }
