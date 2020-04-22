///usr/bin/env go run "$0" "$@" ; exit "$?"
package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/build"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"golang.org/x/tools/go/buildutil"
)

var test = flag.Bool("t", false, "include test .go files")
var name = flag.Bool("n", false, "only show failing file names")
var match = flag.String("match", "", "filepath to exact match")

func init() {
	flag.Var((*buildutil.TagsFlag)(&build.Default.BuildTags), "tags", buildutil.TagsFlagDoc)
}

func main() {
	flag.Parse()

	dirs := flag.Args()
	if len(dirs) == 0 {
		dirs = []string{"."}
	}

	var wg sync.WaitGroup
	goFiles := make([]string, 0)
	for _, dir := range dirs {
		err := filepath.Walk(dir, func(path string, fi os.FileInfo, err error) error {
			if fi.Mode().IsDir() {
				if name := fi.Name(); path != dir && (name[0] == '_' || name[0] == '.') {
					return filepath.SkipDir
				}

				wg.Add(1)
				go func() {
					defer wg.Done()
					goFiles = append(goFiles, listGoFiles(path)...)
				}()
			}
			return nil
		})
		if err != nil {
			log.Fatal(err)
		}
	}
	wg.Wait()
	for _, filePath := range goFiles {
		cmd := exec.Command("golangci-lint", "run", filePath)
		var errOut bytes.Buffer
		cmd.Stderr = &errOut
		var out bytes.Buffer
		cmd.Stdout = &out
		if err := cmd.Run(); err != nil {
			if exitError, ok := err.(*exec.ExitError); ok {
				if exitError.ExitCode() != 0 {
					outString := out.String()
					errString := errOut.String()
					fmt.Printf("Lint failed for %v\n", filePath)
					if ! *name {
						if outString != "" {
							fmt.Println(outString)
						}
						if errString != "" {
							fmt.Println(errString)
						}
					}
				}
			}
		}
	}
}

var outputMu sync.Mutex

func filesToPaths(dir string, files []string) []string {
	goFiles := make([]string, 0)
	outputMu.Lock()
	defer outputMu.Unlock()
	for _, file := range files {
		path := filepath.Join(dir, file)
		if *match == "" || strings.Contains(path, *match) {
			goFiles = append(goFiles, path)
		}
	}
	return goFiles
}

func listGoFiles(dir string) []string {
	pkg, err := build.ImportDir(dir, 0)
	if err != nil {
		if _, ok := err.(*build.NoGoError); !ok {
			log.Fatalf("ImportDir %s: %s", dir, err)
		}
	}
	goFiles := make([]string, 0)

	goFiles = append(goFiles, filesToPaths(dir, pkg.GoFiles)...)
	if *test {
		goFiles = append(goFiles, filesToPaths(dir, pkg.TestGoFiles)...)
		goFiles = append(goFiles, filesToPaths(dir, pkg.XTestGoFiles)...)
	}
	return goFiles
}
