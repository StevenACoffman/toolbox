// +build mage
///usr/bin/env go run github.com/magefile/mage "$@" ; exit "$?"
package gons

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"github.com/magefile/mage/target"
)

var (
	// ModuleName if not set this is determined from the go.mod file
	ModuleName = moduleName()
	// Version of go to use. If not set it defaults to the latest version of Go
	Version = ""
	// CoverArgs to supply to go test.
	CoverArgs = "-html=coverage.out -o coverage.html"
	// TestArgs to supply to go test.
	TestArgs = "-v -race -coverprofile=coverage.out -covermode=atomic ./..."
)

//TODO: warning or error instead of just empty return
func moduleName() string {
	d, err := os.Getwd()
	if err != nil {
		return ""
	}
	f, err := os.Open(filepath.Join(d, "go.mod"))
	if err != nil {
		return ""
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	var m string
	for s.Scan() {
		m = s.Text()
		p := strings.SplitN(m, " ", 2)
		if len(p) == 2 && p[0] == "module" {
			return p[1]
		}
	}
	return ""
}

// Go namespace
type Go mg.Namespace

var (
	goCover = sh.RunCmd("go", "tool", "cover")
	goList  = sh.OutCmd("go", "list", "-json", "./...")
)

// Test runs `go test` with default args set from `TestArgs`. Will always output
// `go test` output.
func (g Go) Test(ctx context.Context) error {
	mg.CtxDeps(ctx, g.CheckVersion)
	return sh.RunV("go", append([]string{"test"}, strings.Split(TestArgs, " ")...)...)
}

func goFiles() ([]string, error) {
	out, err := goList()
	if err != nil {
		return nil, err
	}
	type glp struct {
		Dir        string `json:"Dir"`
		ImportPath string `json:"ImportPath"`
		Name       string `json:"Name"`
		Doc        string `json:"Doc"`
		Target     string `json:"Target"`
		Root       string `json:"Root"`
		Module     struct {
			Path      string `json:"Path"`
			Main      bool   `json:"Main"`
			Dir       string `json:"Dir"`
			GoMod     string `json:"GoMod"`
			GoVersion string `json:"GoVersion"`
		} `json:"Module"`
		Match   []string `json:"Match"`
		GoFiles []string `json:"GoFiles"`
		Imports []string `json:"Imports"`
		Deps    []string `json:"Deps"`
	}
	b := strings.NewReader(out)
	d := json.NewDecoder(b)
	var goFiles []string
	for d.More() {
		var t glp
		if err := d.Decode(&t); err != nil {
			return goFiles, err
		}
		for _, f := range t.GoFiles {
			goFiles = append(goFiles, filepath.Join(t.Dir, f))
		}
	}
	return goFiles, nil
}

// Find any kubernetes objects files
func FindK8S() error {
	count := 0
	fileList := []string{}
	err := filepath.Walk(".", func(path string, f os.FileInfo, err error) error {
		if strings.HasSuffix(path, "yaml") || strings.HasSuffix(path, "yml") {

			b, err := ioutil.ReadFile(path)
			if err != nil {
				panic(err)
			}
			s := string(b)

			if strings.Contains(s, "spec:") && strings.Contains(s, "kind:") {
				fileList = append(fileList, path)
				count++
			}

		}
		return nil
	})

	for _, file := range fileList {
		fmt.Println(file)
	}

	fmt.Printf("\nFound %v Kubernetes object files\n", count)

	return err
}

// CheckVersion checks that the version of go being used is the version specified or the latest version
func (g Go) CheckVersion(ctx context.Context) error {
	ver := Version
	if ver == "" {
		var err error
		ver, err = latestVersion()
		if err != nil {
			return err
		}
	}
	cv, err := sh.Output("go", "version")
	if err != nil {
		return err
	}
	scv := strings.Split(cv, " ")
	if len(scv) != 4 {
		return fmt.Errorf("unknown `go version` string: %q", cv)
	}
	ver, err = expandVersion(ver)
	if err != nil {
		return err
	}
	if ver != scv[2] {
		return fmt.Errorf("current version (%s) is not the same as specified/latest version (%s)", scv[2], ver)
	}
	fmt.Printf("current go version (%s) matches specified/latest version (%s)\n", scv[2], ver)
	return nil
}

func expandVersion(ver string) (string, error) {
	p := strings.Split(ver, ".")
	if len(p) == 3 && p[2] != "x" { // already expanded, except if the patch is ".x"
		if p[2] == "0" { // if the patch is ".0", though it's the original version
			return fmt.Sprintf("%s.%s", p[0], p[1]), nil
		}
		return ver, nil
	}
	if len(p) == 3 && p[2] == "x" { // we want any version in the series, so truncate off so the reduction can happen below
		ver = fmt.Sprintf("%s.%s", p[0], p[1])
	}
	vers, err := versions()
	if err != nil {
		return "", err
	}
	// filter out impossible versions via a prefix match
	var match []string
	for _, v := range vers {
		if strings.HasPrefix(v, ver) {
			match = append(match, v)
		}
	}
	var majv, minv, pv int
	for _, v := range match {
		v = strings.TrimPrefix(v, "go")
		p := strings.Split(v, ".")
		t, err := strconv.Atoi(p[0])
		if err != nil {
			return "", err
		}
		if t > majv {
			majv = t
			minv = 0
			pv = 0
		}
		if strings.Contains(p[1], "beta") || strings.Contains(p[1], "rc") {
			// TODO: Fix beta / rc handling
		} else {
			t, err = strconv.Atoi(p[1])
			if err != nil {
				return "", err
			}
			if t > minv {
				minv = t
				pv = 0
			}
		}
		if len(p) >= 3 {
			t, err := strconv.Atoi(p[2])
			if err != nil {
				return "", err
			}
			if t > pv {
				pv = t
			}
		}
	}
	ver = fmt.Sprintf("go%d.%d", majv, minv)
	if pv > 0 {
		ver += fmt.Sprintf(".%d", pv)
	}
	return ver, nil
}

func latestVersion() (string, error) {
	r, err := http.Get("https://golang.org/VERSION?m=text")
	if err != nil {
		return "", err
	}
	d, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return "", err
	}
	return string(d), nil
}

// Cover runs go tool cover with default args set from `CoverArgs`
func (g Go) Cover(ctx context.Context) error {
	mg.CtxDeps(ctx, g.CheckVersion)
	gf, err := goFiles()
	if err != nil {
		return err
	}
	if need, _ := target.Path("coverage.out", gf...); need {
		mg.Deps(g.Test)
	}
	return goCover(strings.Split(CoverArgs, " ")...)
}

// Coverage opens the coverage output in your browser (runs "go tool cover -html=coverage.out")
func (g Go) Coverage(ctx context.Context) error {
	mg.CtxDeps(ctx, g.CheckVersion, g.Cover)
	gf, err := goFiles()
	if err != nil {
		return err
	}
	need, _ := target.Path("coverage.out", gf...)
	if need {
		mg.Deps(g.Test)
	}
	return goCover("-html=coverage.out")
}