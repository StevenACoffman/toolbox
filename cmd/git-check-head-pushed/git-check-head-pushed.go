///usr/bin/env go run "$0" "$@" ; exit "$?"
package main

import (
	"fmt"
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

// Example how to resolve a revision into its commit counterpart
func main() {
	dir, err := os.Getwd()
	CheckIfError(err)

	repo, err := git.PlainOpen(dir)
	CheckIfError(err)

	revision := "origin/master"

	// Resolve revision into a sha1 commit, only some revisions are resolved
	// look at the doc to get more details
	Info("git rev-parse %s", revision)

	revHash, err := repo.ResolveRevision(plumbing.Revision(revision))
	CheckIfError(err)
	revCommit, err := repo.CommitObject(*revHash)

	CheckIfError(err)

	headRef, err := repo.Head()
	CheckIfError(err)
	// ... retrieving the commit object
	headCommit, err := repo.CommitObject(headRef.Hash())
	CheckIfError(err)

	isAncestor, err := headCommit.IsAncestor(revCommit)

	CheckIfError(err)

	fmt.Printf("Is the HEAD an IsAncestor of origin/master? : %v\n",isAncestor)
}

// CheckIfError should be used to naively panics if an error is not nil.
func CheckIfError(err error) {
	if err == nil {
		return
	}

	fmt.Printf("\x1b[31;1m%s\x1b[0m\n", fmt.Sprintf("error: %s", err))
	os.Exit(1)
}

// Info should be used to describe the example commands that are about to run.
func Info(format string, args ...interface{}) {
	fmt.Printf("\x1b[34;1m%s\x1b[0m\n", fmt.Sprintf(format, args...))
}