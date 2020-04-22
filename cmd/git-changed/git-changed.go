///usr/bin/env go run "$0" "$@" ; exit "$?"
package main

import (
	"fmt"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/utils/merkletrie"
	"os"
	"strings"
)


// Example how to resolve a revision into its commit counterpart
func main() {
	fmt.Println("git ")
	CheckArgs("<revision1>")

	var hash plumbing.Hash

	prevSha := os.Args[1] //prevSha

	dir, err := os.Getwd()
	CheckIfError(err)

	repo, err := git.PlainOpen(dir)
	CheckIfError(err)

	if len(os.Args) < 3 {
		headRef, err := repo.Head()
		CheckIfError(err)
		// ... retrieving the head commit object
		hash = headRef.Hash()
		CheckIfError(err)
	} else {
		arg2 := os.Args[2] //optional descendent sha
		hash = plumbing.NewHash(arg2)
	}

	prevHash := plumbing.NewHash(prevSha)

	prevCommit, err := repo.CommitObject(prevHash)
	CheckIfError(err)

	commit, err := repo.CommitObject(hash)
	CheckIfError(err)

	fmt.Println("You are not crazy"+ prevCommit.Hash.String() + " "+ commit.Hash.String())

	isAncestor, err := commit.IsAncestor(prevCommit)
	CheckIfError(err)

	fmt.Printf("Is the prevCommit an ancestor of commit? : %v %v\n",isAncestor)

	currentTree, err := commit.Tree()
	CheckIfError(err)

	prevTree, err := prevCommit.Tree()
	CheckIfError(err)

	patch, err := currentTree.Patch(prevTree)
	CheckIfError(err)
	fmt.Println("----- Patch Stats ------")

	var changedFiles []string
	for _, fileStat := range patch.Stats() {
		fmt.Println(fileStat.Name)
		changedFiles = append(changedFiles,fileStat.Name)
	}

	changes, err := currentTree.Diff(prevTree)
	CheckIfError(err)

	fmt.Println("----- Changes -----")
	for _, change := range changes {


		// Ignore deleted files
		action, err := change.Action()
		CheckIfError(err)
		if action == merkletrie.Delete {
			//fmt.Println("Skipping delete")
			continue
		}

		// Get list of involved files
		name := getChangeName(change)
		fmt.Println(name)

		//_, to, err := change.Files()
		//CheckIfError(err)
		//
		////Ignore binary files
		//bin, err := to.IsBinary()
		//if bin || err != nil {
		//	fmt.Println("Skipping Binary" + to.Name)
		//	continue
		//}
		//to.Type()
		//fmt.Println(to.Name)

		//for _, re := range config.WhiteList.files {
		//	if re.FindString(to.Name) != "" {
		//		log.Debugf("skipping whitelisted file (matched regex '%s'): %s", re.String(), to.Name)
		//		return nil
		//	}
		//}

	}
}

func getChangeName(change *object.Change) string {
		var empty = object.ChangeEntry{}
		if change.From != empty {
			return change.From.Name
		}
		return change.To.Name
}


// CheckArgs should be used to ensure the right command line arguments are
// passed before executing an example.
func CheckArgs(arg ...string) {
	if len(os.Args) < len(arg)+1 {
		Warning("Usage: %s %s", os.Args[0], strings.Join(arg, " "))
		os.Exit(1)
	}
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

// Warning should be used to display a warning
func Warning(format string, args ...interface{}) {
	fmt.Printf("\x1b[36;1m%s\x1b[0m\n", fmt.Sprintf(format, args...))
}