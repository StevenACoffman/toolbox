///usr/bin/env go run "$0" "$@" ; exit "$?"
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type JIRAResponse struct {
	Key    string `json:"key"`
	Fields struct {
		Description string `json:"description"`
		Summary     string `json:"summary"`
	} `json:"fields"`
}


type Jiration struct {
	re   *regexp.Regexp
	repl interface{}
}

func main() {
	//jiraUserId := getEnv("JIRA_LOGIN", "login")
	//jiraPassword := getEnv("JIRA_PASSWORD", "password")
	//jiraBaseURL := getEnvOrDie("JIRA_BASE_URL", "https://jira.jstor.org")
	//jiraAPIURI := getEnv("JIRA_API_URI", "/rest/api/2/issue/")
	jiraUserId := getEnvOrDie("JIRA_LOGIN")
	jiraPassword := getEnvOrDie("JIRA_PASSWORD")
	jiraBaseURL := getEnvOrDie("JIRA_BASE_URL")
	jiraAPIURI := getEnv("JIRA_API_URI", "/rest/api/2/issue/")


	repoName, organizationName, ticket := checkGitStatus()
	fmt.Println(repoName, organizationName)

	flag.CommandLine.Parse(getFlags())

	url := jiraBaseURL+jiraAPIURI+ticket

	jiraClient, req := BuildRequest(url, jiraUserId, jiraPassword)

	jiraResponse := GetJiraResponse(jiraClient, req)

	title := fmt.Sprintf("%s - %s", jiraResponse.Key, jiraResponse.Fields.Summary)

	prDescription := fmt.Sprintf("Resolves [%s|%s%s%s]\n\n%s\n", jiraResponse.Key, jiraBaseURL,"/browse/", jiraResponse.Key,jiraResponse.Fields.Description)
	prDescription = JiraToMD(prDescription)
	ctx := context.Background()
	token := getEnvOrDie("GO_JIRA_PULL_REQUEST_AUTH_TOKEN")

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)
	input := &github.NewPullRequest{Title: github.String(title), Head: github.String(ticket), Body: github.String(prDescription), Base: github.String("master")}
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
		openBrowser(*pull.HTMLURL)
	} else {
		fmt.Println("Pull request was not created")
	}
}

// no args please
func getFlags() []string {
	var args []string
	for _, arg := range os.Args[1:] {
		if strings.HasPrefix(arg, "--") || strings.HasPrefix(arg, "-"){
			args = append(args,arg)
		}
	}
	return args
}


// no flags please, also I don't need the program name
func getArgs() []string {
	var args []string
	for _, arg := range os.Args[1:] {
		if strings.HasPrefix(arg, "--") || strings.HasPrefix(arg, "-"){
			continue
		}
		args = append(args, arg)
	}
	return args
}

func getEnvOrDie(key string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	panic("You must set your environment variable " + key)
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
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
func checkIfError(err error) {
	if err == nil {
		return
	}

	fmt.Printf("\x1b[31;1m%s\x1b[0m\n", fmt.Sprintf("error: %s", err))
	os.Exit(1)
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

func BuildRequest(url string, jiraUserId string, jiraPassword string) (http.Client, *http.Request) {
	jiraClient := http.Client{
		Timeout: time.Second * 15, // Maximum of 15 secs
	}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	req.SetBasicAuth(jiraUserId, jiraPassword)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	return jiraClient, req
}

func GetJiraResponse(jiraClient http.Client, req *http.Request) JIRAResponse {
	jiraResponse := JIRAResponse{}
	res, getErr := jiraClient.Do(req)
	if getErr != nil {
		log.Fatal(getErr)
		panic(getErr)
	}

	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		log.Fatal(readErr)
		panic(readErr)
	}

	jsonErr := json.Unmarshal(body, &jiraResponse)
	if jsonErr != nil {
		log.Fatal(jsonErr)
		panic(getErr)
	}
	return jiraResponse
}

func JiraToMD(str string) string {
	jirations := []Jiration{
		{ // UnOrdered Lists
			re: regexp.MustCompile(`(?m)^[ \t]*(\*+)\s+`),
			repl: func(groups []string) string {
				_, stars := groups[0], groups[1]
				return strings.Repeat("  ", len(stars)-1) + "* "
			},
		},
		{ //Ordered Lists
			re: regexp.MustCompile(`(?m)^[ \t]*(#+)\s+`),
			repl: func(groups []string) string {
				_, nums := groups[0], groups[1]
				return strings.Repeat("  ", len(nums)-1) + "1. "
			},
		},
		{ //Headers 1-6
			re: regexp.MustCompile(`(?m)^h([0-6])\.(.*)$`),
			repl: func(groups []string) string {
				_, level, content := groups[0], groups[1], groups[2]
				i, _ := strconv.Atoi(level)
				return strings.Repeat("#", i) + content
			},
		},
		{ // Bold
			re:   regexp.MustCompile(`\*(\S.*)\*`),
			repl: "**$1**",
		},
		{ // Italic
			re:   regexp.MustCompile(`\_(\S.*)\_`),
			repl: "*$1*",
		},
		{ // Monospaced text
			re:   regexp.MustCompile(`\{\{([^}]+)\}\}`),
			repl: "`$1`",
		},
		{ // Citations (buggy)
			re:   regexp.MustCompile(`\?\?((?:.[^?]|[^?].)+)\?\?`),
			repl: "<cite>$1</cite>",
		},
		{ // Inserts
			re:   regexp.MustCompile(`\+([^+]*)\+`),
			repl: "<ins>$1</ins>",
		},
		{ // Superscript
			re:   regexp.MustCompile(`\^([^^]*)\^`),
			repl: "<sup>$1</sup>",
		},
		{ // Subscript
			re:   regexp.MustCompile(`~([^~]*)~`),
			repl: "<sub>$1</sub>",
		},
		{ // Strikethrough
			re:   regexp.MustCompile(`(\s+)-(\S+.*?\S)-(\s+)`),
			repl: "$1~~$2~~$3",
		},
		{ // Code Block
			re:   regexp.MustCompile(`\{code(:([a-z]+))?([:|]?(title|borderStyle|borderColor|borderWidth|bgColor|titleBGColor)=.+?)*\}`),
			repl: "```$2",
		},
		{ // Code Block End
			re:   regexp.MustCompile(`{code}`),
			repl: "```",
		},
		{ // Pre-formatted text
			re:   regexp.MustCompile(`{noformat}`),
			repl: "```",
		},
		{ // Un-named Links
			re:   regexp.MustCompile(`(?U)\[([^|]+)\]`),
			repl: "<$1>",
		},
		{ // Images
			re:   regexp.MustCompile(`!(.+)!`),
			repl: "![]($1)",
		},
		{ // Named Links
			re:   regexp.MustCompile(`\[(.+?)\|(.+)\]`),
			repl: "[$1]($2)",
		},
		{ // Single Paragraph Blockquote
			re:   regexp.MustCompile(`(?m)^bq\.\s+`),
			repl: "> ",
		},
		{ // Remove color: unsupported in md
			re:   regexp.MustCompile(`(?m)\{color:[^}]+\}(.*)\{color\}`),
			repl: "$1",
		},
		{ // panel into table
			re:   regexp.MustCompile(`(?m)\{panel:title=([^}]*)\}\n?(.*?)\n?\{panel\}`),
			repl: "\n| $1 |\n| --- |\n| $2 |",
		},
		{ //table header
			re: regexp.MustCompile(`(?m)^[ \t]*((?:\|\|.*?)+\|\|)[ \t]*$`),
			repl: func(groups []string) string {
				_, headers := groups[0], groups[1]
				reBarred := regexp.MustCompile(`\|\|`)

				singleBarred := reBarred.ReplaceAllString(headers, "|")
				fillerRe := regexp.MustCompile(`\|[^|]+`)
				return "\n" + singleBarred + "\n" + fillerRe.ReplaceAllString(singleBarred, "| --- ")
			},
		},
		{ // remove leading-space of table headers and rows
			re:   regexp.MustCompile(`(?m)^[ \t]*\|`),
			repl: "|",
		},
	}
	for _, jiration := range jirations {
		switch v := jiration.repl.(type) {
		case string:
			str = jiration.re.ReplaceAllString(str, v)
		case func([]string) string:
			str = ReplaceAllStringSubmatchFunc(jiration.re, str, v)
		default:
			fmt.Printf("I don't know about type %T!\n", v)
		}
	}
	return str
}

// https://gist.github.com/elliotchance/d419395aa776d632d897
func ReplaceAllStringSubmatchFunc(re *regexp.Regexp, str string, repl func([]string) string) string {
	result := ""
	lastIndex := 0

	for _, v := range re.FindAllSubmatchIndex([]byte(str), -1) {
		groups := []string{}
		for i := 0; i < len(v); i += 2 {
			groups = append(groups, str[v[i]:v[i+1]])
		}

		result += str[lastIndex:v[0]] + repl(groups)
		lastIndex = v[1]
	}

	return result + str[lastIndex:]
}

func openBrowser(url string) {
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