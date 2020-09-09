///usr/bin/env go run "$0" "$@" ; exit "$?"
package main

import (
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

// Example how to resolve a revision into its commit counterpart
func main() {
	dir, err := os.Getwd()
	CheckIfError(err)

	repo, err := git.PlainOpen(dir)
	CheckIfError(err)

	revision := "origin/master"

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

	fmt.Printf("Is the HEAD an IsAncestor of origin/master? : %v\n", isAncestor)

	w, err := repo.Worktree()
	CheckIfError(err)

	cfg, err := parseGitConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Could not read ~/.gitconfig: "+err.Error())
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
	// gp, err := gitignore.LoadGlobalPatterns(homeFS)
	// CheckIfError(err)
	// fmt.Println(gp)
	// w.Excludes = append(gp, w.Excludes...)
	// sp, err := gitignore.LoadSystemPatterns(homeFS)
	// fmt.Println(sp)
	// CheckIfError(err)
	// w.Excludes = append(sp, w.Excludes...)

	status, err := w.Status()
	CheckIfError(err)
	fmt.Println(status)

	fmt.Println(status.IsClean())

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