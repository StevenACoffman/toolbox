///usr/bin/env go run "$0" "$@" ; exit "$?"
package main

import (
	"bytes"
	"context"
	"fmt"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
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
	title := ""

	args := getArgs()
	if len(args) == 1 {
		title = args[0]
	}

	prDescription := ""
	stat, err := os.Stdin.Stat()
	if err != nil {
		panic(err)
	}

	if (stat.Mode() & os.ModeNamedPipe) == 0 {
		fmt.Println("The command is intended to work with piped stdin but didn't get input. Assuming empty pull request description")
	} else {
		stdInBytes, _ := ioutil.ReadAll(os.Stdin)
		prDescription = string(stdInBytes)
		if title == "" {
			title, prDescription = extract(prDescription)
		}
	}
	fmt.Println("title:" + title)
	fmt.Println("head:" + headBranchName)
	input := &github.NewPullRequest{Title: github.String(title), Head: github.String(headBranchName), Body: github.String(prDescription), Base: github.String("master")}
	pull, response, err := client.PullRequests.Create(context.Background(), organizationName, repoName, input)

	if err != nil {
		if response.StatusCode == 422 {
			fmt.Println("Got Unprocessable Entity Error")
		}
		fmt.Errorf("PullRequests.Create returned error: %v", err)
		fmt.Printf("%v\n", response.Status)
		fmt.Printf("%v\n", response.StatusCode)
		fmt.Printf("%v\n", response.Header)
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

func checkGitStatus() (repoName, organizationName, headBranchName string) {
	dir, err := os.Getwd()
	checkIfError(err)

	repo, err := git.PlainOpen(dir)
	checkIfError(err)

	headRef, err := repo.Head()
	checkIfError(err)

	headBranchName = headRef.Name().Short()
	if headBranchName == "master" {
		fmt.Fprintln(os.Stderr, "You are on master so not making a pull request")
		os.Exit(1)
	}

	// ... retrieving the commit object
	headCommit, err := repo.CommitObject(headRef.Hash())
	checkIfError(err)
	revision := "origin/" + headBranchName

	revHash, err := repo.ResolveRevision(plumbing.Revision(revision))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Did you forget to git push --set-upstream origin %s? The current branch %s has no upstream branch.\n", headBranchName, headBranchName)
		os.Exit(1)
	}

	revCommit, err := repo.CommitObject(*revHash)
	checkIfError(err)

	isAncestor, err := headCommit.IsAncestor(revCommit)
	checkIfError(err)
	if !isAncestor {
		fmt.Fprintf(os.Stderr, "Did you forget to push? Your HEAD is not an ancestor of %s so not making a pull request\n", revision)
		os.Exit(1)
	}

	list, err := repo.Remotes()
	checkIfError(err)
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
	checkIfError(err)

	// Because it normally does not include global git config
	// We cannot trust verification of the current status of the worktree using the method Status, without this mess
	// If it doesn't work just assume everything is clean.

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
	checkIfError(err)
	w.Excludes = append(ps, w.Excludes...)

	status, err := w.Status()
	checkIfError(err)

	if !status.IsClean() {
		fmt.Fprintln(os.Stderr, "Did you forget to git commit or git add -A? You have modified or untracked files so not making a pull request")
		os.Exit(1)
	}

	fmt.Printf("Current Org/Repo: %s/%s branch: %s\n", organizationName, repoName, headBranchName)
	return
}

func extract(content string) (title, body string) {
	nl := regexp.MustCompile(`\r?\n`)
	content = nl.ReplaceAllString(content, "\n")

	parts := strings.SplitN(content, "\n\n", 2)
	if len(parts) >= 1 {
		title = strings.TrimSpace(strings.Replace(parts[0], "\n", " ", -1))
	}
	if len(parts) >= 2 {
		body = strings.TrimSpace(parts[1])
	}

	return
}

func parseGitConfig() (*config.Config, error) {
	cfg := config.NewConfig()

	usr, err := user.Current()
	checkIfError(err)

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

func checkIfError(err error) {
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

