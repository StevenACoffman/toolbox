///usr/bin/env go run "$0" "$@" ; exit "$?"
package main

import (
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/utils/merkletrie"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)


// Laundry always includes uncommitted files:
// 	Unmodified         StatusCode = ' '
//	Untracked          StatusCode = '?'
//	Modified           StatusCode = 'M'
//	Added              StatusCode = 'A'
//	Deleted            StatusCode = 'D'
//	Renamed            StatusCode = 'R'
//	Copied             StatusCode = 'C'
//	UpdatedButUnmerged StatusCode = 'U'
//
func main() {
	fmt.Println("git-changed ", len(os.Args))


	var hash plumbing.Hash

	dir, err := os.Getwd()
	CheckIfError(err)

	repo, err := git.PlainOpen(dir)
	CheckIfError(err)

	w, treeErr := repo.Worktree()
	CheckIfError(treeErr)

	cfg, err := parseGitConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not read ~/.gitconfig: %+v\n", err)
		return
	}

	excludesfile := getExcludesFile(cfg)
	if excludesfile == "" {
		fmt.Fprintln(os.Stderr, "Could not get core.excludesfile from ~/.gitconfig")
		return
	}

	ps, err := parseExcludesFile(excludesfile)
	CheckIfError(err)
	w.Excludes = append(ps, w.Excludes...)

	status, statusErr := w.Status()
	CheckIfError(statusErr)

	for filePath, fileStatus := range status {

		fmt.Print(filePath, string(fileStatus.Staging), string(fileStatus.Worktree), "\n")
	}

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

	commit, err := repo.CommitObject(hash)
	CheckIfError(err)

	var prevCommit *object.Commit
	if len(os.Args) < 2 {
		parent, parErr := commit.Parent(0)
		CheckIfError(parErr)
		prevCommit = parent
	} else {
		prevSha := os.Args[1] //prevSha
		prevHash := plumbing.NewHash(prevSha)

		parent, err := repo.CommitObject(prevHash)
		CheckIfError(err)
		prevCommit = parent
	}


	fmt.Println("Previous SHA1:"+ prevCommit.Hash.String() + " Current SHA1:"+ commit.Hash.String())

	isAncestor, err := commit.IsAncestor(prevCommit)
	CheckIfError(err)

	fmt.Printf("Is the prevCommit an ancestor of commit? : %v\n",isAncestor)

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
	}
}

func getChangeName(change *object.Change) string {
	var empty = object.ChangeEntry{}
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

// Info should be used to describe the example commands that are about to run.
func Info(format string, args ...interface{}) {
	fmt.Printf("\x1b[34;1m%s\x1b[0m\n", fmt.Sprintf(format, args...))
}

// Warning should be used to display a warning
func Warning(format string, args ...interface{}) {
	fmt.Printf("\x1b[36;1m%s\x1b[0m\n", fmt.Sprintf(format, args...))
}

func excludeIgnoredChanges(w *git.Worktree, changes merkletrie.Changes) merkletrie.Changes {
	patterns, err := gitignore.ReadPatterns(w.Filesystem, nil)
	if err != nil {
		return changes
	}

	patterns = append(patterns, w.Excludes...)

	if len(patterns) == 0 {
		return changes
	}

	m := gitignore.NewMatcher(patterns)

	var res merkletrie.Changes
	for _, ch := range changes {
		var path []string
		for _, n := range ch.To {
			path = append(path, n.Name())
		}
		if len(path) == 0 {
			for _, n := range ch.From {
				path = append(path, n.Name())
			}
		}
		if len(path) != 0 {
			isDir := (len(ch.To) > 0 && ch.To.IsDir()) || (len(ch.From) > 0 && ch.From.IsDir())
			if m.Match(path, isDir) {
				continue
			}
		}
		res = append(res, ch)
	}
	return res
}


func parseGitConfig() (*config.Config, error) {
	cfg := config.NewConfig()

	usr, err := user.Current()
	CheckIfError(err)

	b, err := ioutil.ReadFile(usr.HomeDir + "/.gitconfig")
	if err != nil {
		return nil, err
	}

	if err := cfg.Unmarshal(b); err != nil {
		return nil, err
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