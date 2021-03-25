package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func main() {
	echo := exec.Command("echo", "message")
	flags := getFlags()

	commitFlags := append(flags, "-F", "-")
	gitArgs := append([]string{"commit"}, commitFlags...)
	//Get the flags to this program, and pass through to git
	// for git command
	// insert flags after "commit" arg, before "-F"
	// go run main.go -s -S -a should work
	commit := exec.Command("git", gitArgs...)

	// Get echo's stdout and attach it to commit's stdin.
	pipe, _ := echo.StdoutPipe()
	defer pipe.Close()

	commit.Stdin = pipe

	// Run echo first.
	echo.Start()

	// Run and get the output of commit.
	res, _ := commit.Output()

	fmt.Println(string(res))
}

// flags but no args
func getFlags() []string {
	var args []string
	for _, arg := range os.Args[1:] {
		if strings.HasPrefix(arg, "--") || strings.HasPrefix(arg, "-") {
			args = append(args, arg)
		}
	}
	return args
}