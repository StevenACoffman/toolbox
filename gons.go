// +build ignore

package main

import (
	"os"

	"github.com/magefile/mage/mage"
)

// sync-start:correct 770446101 a.js
func main() { os.Exit(mage.Main()) }

// sync-end:correct
