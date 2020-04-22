///usr/bin/env go run "$0" "$@" ; exit "$?"
package main

import (
	"flag"
	"fmt"
	"go/build"
	"log"
	"os"
	"path/filepath"
	"sync"

	"golang.org/x/tools/go/buildutil"
)

var test = flag.Bool("t", false, "print test .go files")

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
	for _, dir := range dirs {
		err := filepath.Walk(dir, func(path string, fi os.FileInfo, err error) error {
			if fi.Mode().IsDir() {
				if name := fi.Name(); path != dir && (name[0] == '_' || name[0] == '.') {
					return filepath.SkipDir
				}

				wg.Add(1)
				go func() {
					defer wg.Done()
					listGoFiles(path)
				}()
			}
			return nil
		})
		if err != nil {
			log.Fatal(err)
		}
	}
	wg.Wait()
}

var outputMu sync.Mutex

func printFiles(dir string, files []string) {
	outputMu.Lock()
	defer outputMu.Unlock()
	for _, file := range files {
		fmt.Println(filepath.Join(dir, file))
	}
}

func listGoFiles(dir string) {
	pkg, err := build.ImportDir(dir, 0)
	if err != nil {
		if _, ok := err.(*build.NoGoError); !ok {
			log.Fatalf("ImportDir %s: %s", dir, err)
		}
	}
	printFiles(dir, pkg.GoFiles)
	if *test {
		printFiles(dir, pkg.TestGoFiles)
		printFiles(dir, pkg.XTestGoFiles)
	}
}