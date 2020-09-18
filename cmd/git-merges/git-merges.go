package main

import (
	"fmt"
	"os"

	"github.com/go-git/go-git/v5"
	. "github.com/go-git/go-git/v5/_examples"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// Does git log --merges
func main() {
	CheckArgs("<path>")
	path := os.Args[1]

	r, err := git.PlainOpen(path)
	CheckIfError(err)

	ref, err := r.Head()
	CheckIfError(err)

	c, err := r.CommitObject(ref.Hash())
	CheckIfError(err)

	cIter := object.NewCommitPostorderIter(c, nil)

	err = cIter.ForEach(func(c *object.Commit) error {
		// In git, a merge commit is any commit with more than one parent.
		if c.NumParents() > 1 {
			fmt.Println(c)
		}
		return nil
	})
	CheckIfError(err)
}
