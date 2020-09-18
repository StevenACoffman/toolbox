///usr/bin/env go run "$0" "$@" ; exit "$?"
package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
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

// Laundry always includes uncommitted files.
// Laundry with one git SHA1 argument will include the modified files between
// that commit and it's parent
// Laundry with two git SHA1 arguments will include the modified files between
// those commits.
// Laundry filters these to Go files, and for each executes:
// gofumpt -s -w -l <file>
// gofumports -w -l <file>
// golines -m 80 --shorten-comments -w <file>
// golangci-lint run --fix <file>
func main() {
	fmt.Println("git-changed ", len(os.Args))

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
		if strings.HasSuffix(filePath, ".go") {
			cmd := exec.Command("gofumpt", "-s", "-w", "-l", filePath)
			execCommandOnFilePath(cmd, "gofumpt", filePath)
			cmd = exec.Command("gofumports", "-w", "-l", filePath)
			execCommandOnFilePath(cmd, "gofumports", filePath)
			cmd = exec.Command(
				"golines",
				"-m",
				"80",
				"--shorten-comments",
				"-w",
				filePath,
			)
			execCommandOnFilePath(cmd, "golines", filePath)

			cmd = exec.Command("golangci-lint", "run", "--fix", filePath)
			execCommandOnFilePath(cmd, "golangci-lint", filePath)
		}
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
		if err != nil {
			return nil, err
		}
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
		parent, err := commit.Parent(0)
		if err != nil {
			return nil, err
		}
		prevCommit = parent
	} else {
		prevSha := os.Args[1] // prevSha
		prevHash := plumbing.NewHash(prevSha)

		parent, err := repo.CommitObject(prevHash)
		if err != nil {
			return nil, err
		}
		prevCommit = parent
	}

	//fmt.Println(
	// 	"Previous SHA1:" + prevCommit.Hash.String() + " Current SHA1:" +
	// commit.Hash.String(),
	//)

	isAncestor, err := commit.IsAncestor(prevCommit)
	if err != nil {
		return nil, err
	}

	fmt.Printf("Is the prevCommit an ancestor of commit? : %v\n", isAncestor)

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

	patch, err := currentTree.Patch(prevTree)
	if err != nil {
		return nil, err
	}
	// fmt.Println("----- Patch Stats ------")

	for _, fileStat := range patch.Stats() {
		// fmt.Println(fileStat.Name)
		changedFiles = append(changedFiles, fileStat.Name)
	}

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

func execCommandOnFilePath(cmd *exec.Cmd, commandName, filePath string) {
	var errout bytes.Buffer
	cmd.Stderr = &errout
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			if exitError.ExitCode() != 0 {
				outString := out.String()
				errString := errout.String()
				fmt.Printf("%s failed for %v\n", commandName, filePath)
				if outString != "" {
					fmt.Println(outString)
				}
				if errString != "" {
					fmt.Fprintln(os.Stderr, errString)
				}
			}
		}
	}
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

// Info should be used to describe the example commands that are about to run.
func Info(format string, args ...interface{}) {
	fmt.Printf("\x1b[34;1m%s\x1b[0m\n", fmt.Sprintf(format, args...))
}

// Warning should be used to display a warning
func Warning(format string, args ...interface{}) {
	fmt.Printf("\x1b[36;1m%s\x1b[0m\n", fmt.Sprintf(format, args...))
}

func excludeIgnoredChanges(
	w *git.Worktree,
	changes merkletrie.Changes,
) merkletrie.Changes {
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
			isDir := (len(ch.To) > 0 && ch.To.IsDir()) ||
				(len(ch.From) > 0 && ch.From.IsDir())
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
