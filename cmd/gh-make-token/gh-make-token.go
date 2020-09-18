package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/crypto/ssh/terminal"
)

var defaultConfigsFile = filepath.Join(os.Getenv("HOME"), ".config", "hub")

const (
	headerOTP = "X-GitHub-OTP"
)

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
	args := getArgs()

	note := "Demonstration Personal Access Token"

	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: gh-make-token <token-description>")
		panic(errors.New("Usage: gh-make-token <token-description>"))
	} else {
		note = args[0]
	}

	username, password, otp := credentials()
	fmt.Printf("Username: %s, Password: %s, OTP: %s\n", username, strings.Repeat("*", len(password)), otp)
	makePersonalAccessToken(username, password, otp, note)
}

func credentials() (string, string, string) {
	fmt.Print("Enter Username: ")
	username := GetUser()
	password := GetPassword(username)
	otp := GetOTP()

	return strings.TrimSpace(username), strings.TrimSpace(password), strings.TrimSpace(otp)
}

func Check(err error) {
	if err != nil {
		log.Fatal(err)
		// panic(err)
		os.Exit(1)
	}
}

func scanLine() string {
	var line string
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		line = scanner.Text()
	}
	Check(scanner.Err())

	return line
}

func GetUser() (user string) {
	user = os.Getenv("GITHUB_USER")
	if user != "" {
		return
	}

	fmt.Printf("Github username: ")
	user = scanLine()

	return
}

func GetPassword(user string) (pass string) {
	pass = os.Getenv("GITHUB_PASSWORD")
	if pass != "" {
		return
	}

	fmt.Printf("password for %s (never stored): ", user)
	bytePassword, err := terminal.ReadPassword(0)
	Check(err)
	pass = string(bytePassword)

	return
}

func GetOTP() string {
	fmt.Print("two-factor authentication code: ")
	return scanLine()
}

func configsFile() string {
	configsFile := os.Getenv("GH_CONFIG")
	if configsFile == "" {
		configsFile = defaultConfigsFile
	}

	return configsFile
}

type AuthorizationEntry struct {
	Token string `json:"token"`
}

func makePersonalAccessToken(username, password, twoFactorCode, note string) {
	body := strings.NewReader(`{"scopes":["repo"],"note":"` + note + `"}`)
	req, err := http.NewRequest("POST", "https://api.github.com/authorizations", body)
	if err != nil {
		// handle err
	}
	req.SetBasicAuth(username, password)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	// req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	if twoFactorCode != "" {
		req.Header.Set("X-GitHub-OTP", twoFactorCode)
	}

	githubClient := http.Client{
		Timeout: time.Second * 15, // Maximum of 15 secs
	}

	resp, err := githubClient.Do(req)
	Check(err)

	if resp.StatusCode == http.StatusUnauthorized && strings.HasPrefix(resp.Header.Get(headerOTP), "required") {
		fmt.Errorf(" status code: %s, headerOTP: %s", resp.StatusCode, resp.Header.Get(headerOTP))
	}
	respBody, readErr := ioutil.ReadAll(resp.Body)
	if readErr != nil {
		log.Fatal(readErr)
		panic(readErr)
	}
	auth := &AuthorizationEntry{}

	json.Unmarshal(respBody, &auth)
	fmt.Printf("Authorization Token: %s", auth.Token)
	defer resp.Body.Close()
}
