///usr/bin/env go run "$0" "$@" ; exit "$?"
package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/format/gitignore"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"os/user"

)

/*
 https://developer.github.com/v3/pulls/#create-a-pull-request
type NewPullRequest struct {
	Title               *string `json:"title,omitempty"`
	Head                *string `json:"head,omitempty"`
	Base                *string `json:"base,omitempty"`
	Body                *string `json:"body,omitempty"`
	Issue               *int    `json:"issue,omitempty"`
	MaintainerCanModify *bool   `json:"maintainer_can_modify,omitempty"`
}
*/

// no flags please
func getArgs() []string {
	var args []string
	for _, arg := range os.Args[1:] {
		if strings.HasPrefix(arg, "--") || strings.HasPrefix(arg, "-") {
			continue
		}
		args = append(args, arg)
	}
	return args
}

func checkGitStatus() (repoName, organizationName, headBranchName string) {

	dir, err := os.Getwd()
	CheckIfError(err)

	repo, err := git.PlainOpen(dir)
	CheckIfError(err)

	headRef, err := repo.Head()
	CheckIfError(err)

	headBranchName = headRef.Name().Short()
	if headBranchName == "master" {
		fmt.Fprintln(os.Stderr, "You are on master so not making a pull request")
		os.Exit(1)
	}
	fmt.Println("Current Branch" + headBranchName)

	// ... retrieving the commit object
	headCommit, err := repo.CommitObject(headRef.Hash())
	CheckIfError(err)

	revision := "origin/"+headBranchName

	revHash, err := repo.ResolveRevision(plumbing.Revision(revision))
	CheckIfError(err)
	revCommit, err := repo.CommitObject(*revHash)

	CheckIfError(err)

	isAncestor, err := headCommit.IsAncestor(revCommit)
	CheckIfError(err)

	if !isAncestor {
		fmt.Fprintf(os.Stderr, "Did you forget to push? Your HEAD is not an ancestor of %s so not making a pull request\n", revision)
		os.Exit(1)
	}

	list, err := repo.Remotes()
	CheckIfError(err)

	for _, r := range list {
		rc := r.Config()
		if rc.Name == "origin" {
			segments := strings.Split(rc.URLs[0], "/")
			basename := segments[len(segments)-1]

			repoName = strings.TrimSuffix(basename, filepath.Ext(basename))
			remoteUrlChunks := strings.Split(segments[len(segments)-2], ":")
			organizationName = remoteUrlChunks[len(remoteUrlChunks)-1]
		}
	}

	w, err := repo.Worktree()
	CheckIfError(err)
	// We cannot trust verification of the current status of the worktree using the method Status.
	// Any globally ignored files will

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

	status, err := w.Status()
	CheckIfError(err)

	if !status.IsClean() {
		fmt.Fprintln(os.Stderr, "Did you forget to git commit or git add -A? You have modified or untracked files so not making a pull request")
		os.Exit(1)
	}

	fmt.Println("Current Org/Repo: %s/%s branch %s", organizationName,repoName, headBranchName)
	return
}

func main() {
	ctx := context.Background()
	token := getEnvOrDie("GO_JIRA_PULL_REQUEST_AUTH_TOKEN")

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	// list all repositories for the authenticated user
	//repos, _, _ := client.Repositories.List(ctx, "StevenACoffman", nil)
	//fmt.Println(repos)


	repoName, organizationName, headBranchName := checkGitStatus()

	title := "Snappier title"

	args := getArgs()
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: github-make-pull <title>")
		panic(errors.New("Usage: github-make-pull <title>"))
	} else {
		title = args[0]
	}
	fmt.Println("title:" + title)
	stat, err := os.Stdin.Stat()
	if err != nil {
		panic(err)
	}
	// os.ModeNamedPipe
	if (stat.Mode() & os.ModeNamedPipe) == 0 {
		fmt.Println("The command is intended to work with pipes.")
		fmt.Println("Usage: github-make-pull <title>")
		return
	}
	stdInBytes, _ := ioutil.ReadAll(os.Stdin)
	prDescription := string(stdInBytes)
	fmt.Println("PR Description:" + prDescription)

	fmt.Println("title:" + title)
	fmt.Println("head:" + headBranchName)
	fmt.Println("body:" + prDescription)
	input := &github.NewPullRequest{Title: github.String(title), Head: github.String(headBranchName), Body: github.String(prDescription), Base: github.String("master")}
	pull, response, err := client.PullRequests.Create(context.Background(), organizationName, repoName, input)

	if err != nil {
		fmt.Errorf("PullRequests.Create returned error: %v", err)
		fmt.Printf("%v\n", response.Status)
		fmt.Printf("%v\n", response.StatusCode)
		buf := new(bytes.Buffer)
		buf.ReadFrom(response.Body)
		bodyString := buf.String()
		fmt.Printf("%v\n", bodyString)
		fmt.Printf("%v\n", response.String())
	}
	if pull != nil {
		fmt.Printf("Created Pull Request Successfully. Opening browser for %v\n", pull.HTMLURL)
		openbrowser(*pull.HTMLURL)
	} else {
		fmt.Println("Pull request was not created")
	}

}

func CheckIfError(err error) {
	if err == nil {
		return
	}

	fmt.Printf("\x1b[31;1m%s\x1b[0m\n", fmt.Sprintf("error: %s", err))
	os.Exit(1)
}

func getEnvOrDie(key string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	panic("No github personal access token in env " + key)
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


func openbrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Fatal(err)
	}

}