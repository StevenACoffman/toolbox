package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/go-git/go-git/v5"
)

func GetCurrentCommitFromRepository(repository *git.Repository) (string, error) {
	headRef, err := repository.Head()
	if err != nil {
		return "", err
	}
	headSha := headRef.Hash().String()

	commit, commitErr := repository.CommitObject(headRef.Hash())
	if commitErr != nil {
		return headSha, commitErr
	}
	when := commit.Author.When
	format := "060201-1504-"
	datetimeString := when.Format(format)

	return datetimeString + headSha[0:12], nil
}

func main() {
	CheckArgs("<repoPath> <rev> <path>")
	repoPath := os.Args[1]

	repo, err := git.PlainOpen(repoPath)
	CheckIfError(err)
	appEngineFormat, formatErr := GetCurrentCommitFromRepository(repo)
	CheckIfError(formatErr)
	fmt.Println(appEngineFormat)

}

// CheckIfError should be used to naively panics if an error is not nil.
func CheckIfError(err error) {
	if err == nil {
		return
	}

	fmt.Printf("\x1b[31;1m%s\x1b[0m\n", fmt.Sprintf("error: %s", err))
	os.Exit(1)
}

// CheckArgs should be used to ensure the right command line arguments are
// passed before executing an example.
func CheckArgs(arg ...string) {
	if len(os.Args) < len(arg)+1 {
		Warning("Usage: %s %s", os.Args[0], strings.Join(arg, " "))
		os.Exit(1)
	}
}

// Warning should be used to display a warning
func Warning(format string, args ...interface{}) {
	fmt.Printf("\x1b[36;1m%s\x1b[0m\n", fmt.Sprintf(format, args...))
}
