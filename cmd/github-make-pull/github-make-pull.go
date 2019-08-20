///usr/bin/env go run "$0" "$@" ; exit "$?"
package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"gopkg.in/src-d/go-git.v4"
	"io/ioutil"
	"os"
	"path/filepath"
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
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	repo, err := git.PlainOpen(dir)
	if err != nil {
		panic(err)
	}

	h, err := repo.Head()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		panic(err)
	}
	head := h.Name().Short() //current branch name
	fmt.Println("Current Branch" + head)

	list, err := repo.Remotes()
	if err != nil {
		panic(err)
	}
	repoName := ""
	organizationName := ""
	for _, r := range list {
		rc := r.Config()
		if rc.Name == "origin" {
			segments := strings.Split(rc.URLs[0], "/")
			basename := segments[len(segments)-1]
			fmt.Println("BaseName:" + basename)
			repoName = strings.TrimSuffix(basename, filepath.Ext(basename))
			remoteUrlChunks := strings.Split(segments[len(segments)-2], ":")
			organizationName = remoteUrlChunks[len(remoteUrlChunks)-1]
		}
	}
	fmt.Println("Current Org/Repo" + organizationName + "/" + repoName)

	title := "Snappier title"

	args := getArgs()
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: github-make-pull <title>")
		panic(errors.New("Usage: github-make-pull <title>"))
	} else {
		title = args[0]
	}

	stat, err := os.Stdin.Stat()
	if err != nil {
		panic(err)
	}
	// os.ModeNamedPipe
	if stat.Mode()&os.ModeNamedPipe != 0 {
		fmt.Println("The command is intended to work with pipes.")
		fmt.Println("Usage: github-make-pull <title>")
		return
	}
	bytes, _ := ioutil.ReadAll(os.Stdin)
	prDescription := string(bytes)

	input := &github.NewPullRequest{Title: github.String(title), Head: github.String(head), Body: github.String(prDescription)}
	pull, response, err := client.PullRequests.Create(context.Background(), organizationName, repoName, input)
	if err != nil {
		fmt.Errorf("PullRequests.Create returned error: %v", err)
	}
	fmt.Print(pull)
	fmt.Print(response.String())

}

func getEnvOrDie(key string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	panic("No github personal access token in env " + key)
}
