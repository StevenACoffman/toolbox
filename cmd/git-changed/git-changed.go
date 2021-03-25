// 2>/dev/null;/usr/bin/env go run "$0" "$@" ; exit "$?"
package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/utils/merkletrie"
)

// Example how to resolve a revision into its commit counterpart
func main() {
	dir, err := os.Getwd()
	CheckIfError(err)

	repo, err := git.PlainOpen(dir)
	for err != nil {
		isRoot := dir == "/"
		dir = filepath.Dir(dir)
		repo, err = git.PlainOpen(dir)
		if isRoot {
			CheckIfError(err)
		}
	}

	changedFiles, err := uncommittedFilePaths(repo)
	CheckIfError(err)

	// If have no uncommitted changes, then assume you meant
	// to look at last diff
	// If you passed one or more git SHA1 arguments, that is
	// pretty clear signal of intent for that!
	if len(changedFiles) == 0 && len(os.Args) == 1 {
		commitChangedFiles, err := commitFileChanges(repo)
		CheckIfError(err)
		changedFiles = append(changedFiles, commitChangedFiles...)
	}

	for _, filePath := range changedFiles {
		fmt.Println(filePath)
	}
}

func commitFileChanges(repo *git.Repository) ([]string, error) {
	var hash plumbing.Hash
	if len(os.Args) < 3 {
		headRef, err := repo.Head()
		if err != nil {
			return nil, err
		}
		// ... retrieving the head commit object
		hash = headRef.Hash()
	} else {
		arg2 := os.Args[2] // optional descendent sha
		hash = plumbing.NewHash(arg2)
	}

	commit, err := repo.CommitObject(hash)
	if err != nil {
		return nil, err
	}

	var prevCommit *object.Commit
	if len(os.Args) < 2 {
		parent, argErr := commit.Parent(0)
		if argErr != nil {
			return nil, argErr
		}
		prevCommit = parent
	} else {
		prevSha := os.Args[1] // prevSha
		prevHash := plumbing.NewHash(prevSha)

		parent, commitErr := repo.CommitObject(prevHash)
		if commitErr != nil {
			return nil, commitErr
		}
		prevCommit = parent
	}

	//  fmt.Println(
	//   	"Previous SHA1:" + prevCommit.Hash.String() + " Current SHA1:" +
	//   commit.Hash.String(),
	//  )

	//  isAncestor, err := commit.IsAncestor(prevCommit)
	//  if err != nil {
	//  	return nil, err
	//  }

	// fmt.Printf("Is the prevCommit an ancestor of commit? : %v\n", isAncestor)

	commitChangedFiles, err := changesBetweenCommits(commit, prevCommit)
	if err != nil {
		return nil, err
	}
	return commitChangedFiles, nil
}

func changesBetweenCommits(
	commit *object.Commit,
	prevCommit *object.Commit,
) ([]string, error) {
	var changedFiles []string
	currentTree, err := commit.Tree()
	if err != nil {
		return nil, err
	}

	prevTree, err := prevCommit.Tree()
	if err != nil {
		return nil, err
	}
	//
	//patch, err := currentTree.Patch(prevTree)
	//if err != nil {
	//	return nil, err
	//}
	//// fmt.Println("----- Patch Stats ------")
	//
	//for _, fileStat := range patch.Stats() {
	//	// fmt.Println(fileStat.Name)
	//	changedFiles = append(changedFiles, fileStat.Name)
	//}

	changes, err := currentTree.Diff(prevTree)
	if err != nil {
		return nil, err
	}

	// fmt.Println("----- Changes -----")
	for _, change := range changes {
		// Ignore deleted files
		action, err := change.Action()
		if err != nil {
			return nil, err
		}
		if action == merkletrie.Delete {
			// fmt.Println("Skipping delete")
			continue
		}

		// Get list of involved files
		name := getChangeName(change)
		// fmt.Println(name)
		changedFiles = append(changedFiles, name)
	}

	changedFiles = unique(changedFiles)
	return changedFiles, nil
}

func uncommittedFilePaths(repo *git.Repository) ([]string, error) {
	var uncommittedFilePaths []string
	w, treeErr := repo.Worktree()
	CheckIfError(treeErr)

	cfg, err := parseGitConfig()
	if err != nil {
		return nil, fmt.Errorf("could not read ~/.gitconfig: %w", err)
	}

	excludesfile := getExcludesFile(cfg)
	if excludesfile == "" {
		return nil, fmt.Errorf(
			"could not get core.excludesfile from ~/.gitconfig",
		)
	}

	ps, err := parseExcludesFile(excludesfile)
	if err != nil {
		return nil, err
	}
	w.Excludes = append(ps, w.Excludes...)

	status, err := w.Status()
	if err != nil {
		return nil, err
	}

	for filePath := range status {
		uncommittedFilePaths = append(uncommittedFilePaths, filePath)
	}
	return uncommittedFilePaths, err
}

func getChangeName(change *object.Change) string {
	empty := object.ChangeEntry{}
	if change.From != empty {
		return change.From.Name
	}

	return change.To.Name
}

// CheckIfError should be used to naively panics if an error is not nil.
func CheckIfError(err error) {
	if err == nil {
		return
	}

	fmt.Printf("\x1b[31;1m%s\x1b[0m\n", fmt.Sprintf("error: %+v", err))
	os.Exit(1)
}

func parseGitConfig() (*config.Config, error) {
	cfg := config.NewConfig()

	usr, err := user.Current()
	CheckIfError(err)

	b, err := ioutil.ReadFile(usr.HomeDir + "/.gitconfig")
	if err != nil {
		return nil, err
	}

	if unmarshalErr := cfg.Unmarshal(b); unmarshalErr != nil {
		return nil, unmarshalErr
	}

	return cfg, err
}

func getExcludesFile(cfg *config.Config) string {
	for _, sec := range cfg.Raw.Sections {
		if sec.Name == "core" {
			for _, opt := range sec.Options {
				if opt.Key == "excludesfile" {
					return opt.Value
				}
			}
		}
	}
	return ""
}

func parseExcludesFile(excludesfile string) ([]gitignore.Pattern, error) {
	excludesfile, err := expandTilde(excludesfile)
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadFile(excludesfile)
	if err != nil {
		return nil, err
	}

	var ps []gitignore.Pattern
	for _, s := range strings.Split(string(data), "\n") {
		if !strings.HasPrefix(s, "#") && len(strings.TrimSpace(s)) > 0 {
			ps = append(ps, gitignore.ParsePattern(s, nil))
		}
	}

	return ps, nil
}

// "~/.gitignore" -> "/home/tyru/.gitignore"
func expandTilde(path string) (string, error) {
	if !strings.HasPrefix(path, "~") {
		return path, nil
	}
	var paths []string
	u, err := user.Current()
	if err != nil {
		return "", err
	}
	for _, p := range strings.Split(path, string(filepath.Separator)) {
		if p == "~" {
			paths = append(paths, u.HomeDir)
		} else {
			paths = append(paths, p)
		}
	}
	return "/" + filepath.Join(paths...), nil
}

func unique(sSlice []string) []string {
	keys := make(map[string]bool)
	var list []string
	for _, entry := range sSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}
