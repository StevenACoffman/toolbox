package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"go/parser"
	"go/scanner"
	"go/token"
	"hash/adler32"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

var (
	fileSet    = token.NewFileSet() // per process FileSet
	exitCode   = 0
	parserMode parser.Mode
	write      = flag.Bool("w", false, "write result to (source) file instead of stdout")
)

func main() {
	rootPathBytes, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		panic(err)
	}
	rootPath := strings.TrimSpace(string(rootPathBytes))
	fmt.Println(rootPath)
	// aStringToHash := []byte("foo bar baz٪☃🍣")
	// adler32Int := Checksum(aStringToHash)
	// fmt.Println("Adler32 String is ", adler32Int)
	// strconv.FormatUint(uint64(adler32Int, 16))

	// tag := "// sync-start:format-key-path 2056707582 appengine/backup_model.py"
	// me := parseSyncStart("// sync-start:",tag)
	// fmt.Printf("// sync-start:%s %d %s\n", me.MarkerID, me.TargetChecksum, me.TargetFile)

	codeSyncMain()
}

func parseSyncStart(tagStart, tag string) *MarkerEdge {
	formatString := fmt.Sprintf("%s%%s %%d %%s", tagStart)
	me := MarkerEdge{}
	me.TargetFileToSourceTargetChecksum = make(map[string]uint32)
	var sourceTargetChecksum uint32
	var targetFile string
	fmt.Sscanf(strings.TrimSpace(tag), formatString, &me.MarkerID, &sourceTargetChecksum, &targetFile)

	me.TargetFileToSourceTargetChecksum[targetFile] = sourceTargetChecksum
	if me.MarkerID == "" {
		fmt.Fprintln(os.Stderr, tag)
		return nil
	}
	return &me
}

func parseSyncEnd(tagEnd, tag string) string {
	formatString := fmt.Sprintf("%s%%s", tagEnd)
	var targetId string
	fmt.Sscanf(strings.TrimSpace(tag), formatString, &targetId)
	return targetId
}

func usage() {
	fmt.Fprintf(os.Stderr, "usage: gofumpt [flags] [path ...]\n")
	flag.PrintDefaults()
}

func isGoFile(f os.FileInfo) bool {
	// ignore non-Go files
	name := f.Name()
	return !f.IsDir() && !strings.HasPrefix(name, ".") && strings.HasSuffix(name, ".go")
}

func report(err error) {
	scanner.PrintError(os.Stderr, err)
	exitCode = 2
}

func codeSyncMain() {
	flag.Usage = usage
	flag.Parse()

	// Print the gofumpt version if the user asks for it.
	if *showVersion {
		printVersion()
		return
	}

	if flag.NArg() == 0 {
		if *write {
			fmt.Fprintln(os.Stderr, "error: cannot use -w with standard lines")
			exitCode = 2
			return
		}
		if _, err := processFile("<standard lines>", os.Stdin, os.Stdout, true); err != nil {
			report(err)
		}
		return
	}

	pathToIDToMarker := make(map[string]map[string]*MarkerEdge)
	markerWalker := MarkerWalker{PathToIDToMarker: pathToIDToMarker}

	for i := 0; i < flag.NArg(); i++ {
		path := flag.Arg(i)
		switch dir, err := os.Stat(path); {
		case err != nil:
			report(err)
		case dir.IsDir():
			markerWalker.walkDir(path)
		default:
			targetToMarker, err := processFile(path, nil, os.Stdout, false)
			if err != nil {
				report(err)
			}
			markerWalker.PathToIDToMarker[path] = targetToMarker
		}
	}
	markerWalker.resolve()
}

type MarkerWalker struct {
	// filepath To target ID To MarkerEdge
	PathToIDToMarker map[string]map[string]*MarkerEdge
}

func (mw *MarkerWalker) resolve() {
	exactMatches := 0
	differences := 0
	for path, sourceIDToMarker := range mw.PathToIDToMarker {
		for id, sourceMarker := range sourceIDToMarker {
			for targetFile, sourceTargetChecksum := range sourceMarker.TargetFileToSourceTargetChecksum {
				targetIDToMarker, ok := mw.PathToIDToMarker[targetFile]
				// if target file has not yet been processed
				// try to collect all the file's source tagged block info
				if !ok {
					var err error
					targetIDToMarker, err = processFile(targetFile, nil, os.Stdout, false)
					if err != nil || targetIDToMarker == nil {
						report(err)
						continue
					}
				}
				targetSourceMarker, markerExists := targetIDToMarker[id]

				if markerExists {
					if targetSourceMarker.SourceBlockChecksum == sourceTargetChecksum {
						exactMatches++
					} else {
						differences++
						// Checksum(append([]byte(sourceMarker.TargetBlock), newlineBytes...))
						calcSum := int64(sourceTargetChecksum) - int64(targetSourceMarker.SourceBlockChecksum)
						fmt.Println(
							"Difference found in", sourceMarker.SourceFile,
							"target:", sourceMarker.MarkerID,
							" ptcs:", sourceTargetChecksum,
							" atcs:", targetSourceMarker.SourceBlockChecksum,
							"calc:", calcSum)
						fmt.Println("E:", strings.TrimSpace(sourceMarker.SourceDeclaration))
						fmt.Print(targetSourceMarker.SourceBlock)
						fmt.Println("<END>")
					}
				} else {
					fmt.Println("in:", path, " marker ", targetIDToMarker, " not found")
				}
			}
		}
	}
	fmt.Println("Exact Matches:", exactMatches)
	fmt.Println("Differences:", differences)
}

var ValidSyncStart = []string{
	"# sync-start:",    // python, graphql
	"// sync-start:",   // javascript, go
	"/// sync-start:",  // rust
	"{/* sync-start:",  // jsx
	"{{/* sync-start:", // go templates
}

var ValidSyncEnd = []string{
	"# sync-end:",
	"// sync-end:",
	"/// sync-end:",
	"{/* sync-end:",
	"{{/* sync-end:",
}

func (mw *MarkerWalker) visitFile(path string, f os.FileInfo, err error) error {
	if f.IsDir() && (f.Name() == ".git" ||
		f.Name() == "node_modules" ||
		f.Name() == "bin" ||
		f.Name() == "generated") {
		return filepath.SkipDir
	}
	//fmt.Println(path)
	//
	if err == nil && isGoFile(f) {
		targetToMarker, processErr := processFile(path, nil, os.Stdout, false)
		err = processErr
		mw.PathToIDToMarker[path] = targetToMarker
	}
	// Don't complain if a file was deleted in the meantime
	if err != nil && !os.IsNotExist(err) {
		report(err)
	}
	return nil
}

func (mw *MarkerWalker) walkDir(path string) {
	filepath.Walk(path, mw.visitFile)
}

const chmodSupported = runtime.GOOS != "windows"

// backupFile writes data to a new file named filename<number> with permissions perm,
// with <number randomly chosen such that the file name is unique. backupFile returns
// the chosen file name.
func backupFile(filename string, data []byte, perm os.FileMode) (string, error) {
	// create backup file
	f, err := ioutil.TempFile(filepath.Dir(filename), filepath.Base(filename))
	if err != nil {
		return "", err
	}
	bakname := f.Name()
	if chmodSupported {
		err = f.Chmod(perm)
		if err != nil {
			f.Close()
			os.Remove(bakname)
			return bakname, err
		}
	}

	// write data to backup file
	_, err = f.Write(data)
	if err1 := f.Close(); err == nil {
		err = err1
	}

	return bakname, err
}

var Empty struct{}

// If in == nil, the source is the contents of the file with the given filename.
func processFile(filename string, in io.Reader, out io.Writer, stdin bool) (map[string]*MarkerEdge, error) {
	// var perm os.FileMode = 0o644
	if in == nil {
		f, err := os.Open(filename)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		//fi, err := f.Stat()
		//if err != nil {
		//	return nil, err
		//}
		in = f
		// perm = fi.Mode().Perm()
	}

	// Splits on newlines by default.
	scanner := bufio.NewScanner(in)
	// set scanner to 1 MB buffer
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)
	scanner.Split(ScanLines)
	idToMarker := make(map[string]*MarkerEdge)
	targetToSourceBlock := make(map[string][]byte)
	currentTargets := make(map[string]struct{})
	line := 1

	// https://golang.org/pkg/bufio/#Scanner.Scan
	for scanner.Scan() {
		currentBytes := scanner.Bytes()
		var latestMarker string
		containsSyncTag := false
		for i, vss := range ValidSyncStart {
			vssBytes := []byte(vss)
			if bytes.Contains(currentBytes, vssBytes) {
				fmt.Fprintln(os.Stderr, "Got here!")
				syncStart := string(currentBytes)
				fmt.Fprintln(os.Stderr, "START:", syncStart)
				marker := parseSyncStart(vss, syncStart)
				if marker != nil {
					containsSyncTag = true
					marker.SourceLine = line
					marker.SourceDeclaration = syncStart
					marker.SourceCommentStart = vss
					marker.SourceCommentEnd = ValidSyncEnd[i]
					marker.SourceFile = filename
					// fmt.Printf("%s %s ", filename, perm.String())
					// fmt.Printf("line: %+v\n",*marker)
					idToMarker[marker.MarkerID] = marker
					latestMarker = marker.MarkerID
					// do not keep looking for sync tags
					break
				}
			}
		}

		for _, vse := range ValidSyncEnd {
			vseBytes := []byte(vse)
			if bytes.Contains(currentBytes, vseBytes) {
				targetID := parseSyncEnd(vse, string(currentBytes))
				marker, ok := idToMarker[targetID]
				if targetID != "" && ok {
					containsSyncTag = true
					block := targetToSourceBlock[marker.MarkerID]
					marker.SourceBlockChecksum = Checksum(block)
					fmt.Fprintln(os.Stderr, "end", targetID, marker.SourceBlockChecksum)
					marker.SourceBlock = string(block)

					// fmt.Println("STARTBLOCK:",string(block), "ENDBLOCK")
					// fmt.Println("ADLER32:",marker.SourceChecksum)
					delete(currentTargets, marker.MarkerID)
					// no longer needed, so free memory
					delete(targetToSourceBlock, marker.MarkerID)
					// hex is more compact but sadly not the spec
					// fmt.Println("Adler32 String is ", strconv.FormatUint(uint64(marker.SourceChecksum), 16))
					// do not keep looking for sync tag ends
					break
				} else {
					fmt.Fprintln(os.Stderr, "We should not be here")
				}
			}
		}
		if !containsSyncTag {
			for currentTarget := range currentTargets {
				block := targetToSourceBlock[currentTarget]
				targetToSourceBlock[currentTarget] = append(block, currentBytes...)
			}
		}
		// if the current bytes contains a start tag, we need to delay adding the target
		// until after we added the block
		// this avoids including the sync tag in the source to sync
		if latestMarker != "" {
			currentTargets[latestMarker] = Empty
		}

		line++
	}

	if err := scanner.Err(); err != nil {
		// Handle the error
		fmt.Printf("Got an error %+v\n", err)
	}
	fmt.Fprintln(os.Stderr, "idToMarker", idToMarker)
	return idToMarker, nil
}

func ScanLines(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexByte(data, '\n'); i >= 0 {
		// We have a full newline-terminated line.
		return i + 1, data[0 : i+1], nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), data, nil
	}
	// Request more data.
	return 0, nil, nil
}

var (
	newlineBytes   = []byte("\n")
	doubleNewlines = []byte("\n\n")
	crNlBytes      = []byte("\r\n")
)

func Checksum(b []byte) uint32 {
	var salted []byte
	salted = b
	// if ! bytes.HasPrefix(b, newlineBytes) {
	// 	salted = append( newlineBytes, b...)
	// }
	salted = append(newlineBytes, salted...)

	// if !bytes.HasPrefix(salted, doubleNewlines) {
	//	salted = append( newlineBytes, salted...)
	//	//salted = bytes.TrimPrefix(salted, newlineBytes)
	// }
	// if bytes.HasSuffix(salted, doubleNewlines) {
	//	//salted = append( salted, newlineBytes...)
	//	salted = bytes.TrimSuffix(salted, newlineBytes)
	// }
	return adler32.Checksum(salted)
}

func remove(slice []string, s string) []string {
	i := indexOf(s, slice)
	return append(slice[:i], slice[i+1:]...)
}

func indexOf(element string, data []string) int {
	for k, v := range data {
		if element == v {
			return k
		}
	}
	return -1 // not found.
}

// GetStringInBetween returns empty string if no start or end string found
func GetStringInBetween(str string, start string, end string) (result string) {
	s := strings.Index(str, start)
	if s == -1 {
		return
	}
	s += len(start)
	e := strings.Index(str[s:], end)
	if e == -1 {
		return
	}
	return str[s : s+e]
}

type TargetMarker struct {
	// The marker identifier.
	MarkerID string
	// The tag path to the target file of the marker, / as path separator
	TargetFile string

	// The line number in the target file where the marker begins.
	TargetLine int

	// The actual checksum of the target content.
	TargetChecksum uint32

	// The code
	TargetBlock string
}

type MarkerEdge struct {
	// The marker identifier.
	MarkerID string

	// The line number in the source file where the marker is declared.
	SourceLine int

	// The path to the source file of the marker, / as path separator
	SourceFile string

	// The checksum that the source file has recorded for the target content.
	SourceTargetChecksum uint32

	// The checksum that the source file tagged block
	SourceBlockChecksum uint32

	// The full tag declaration of the marker target in the source file.
	SourceDeclaration string

	// The start of the tag comment that the source file uses.
	SourceCommentStart string

	// The end of the tag comment that the source file uses.
	SourceCommentEnd string

	// The end of the tag comment that the source file uses.
	SourceBlock string

	// For each target file, Data about the TargetFileToSourceTargetChecksum
	TargetFileToSourceTargetChecksum map[string]uint32
}

func (m *MarkerEdge) String() string {
	for targetFile, sourceTargetChecksum := range m.TargetFileToSourceTargetChecksum {
		return fmt.Sprintf("%s%s %d %s",
			m.SourceCommentStart,
			m.MarkerID,
			sourceTargetChecksum,
			targetFile)
	}
	return fmt.Sprintf("%s%s 0 <none>",
		m.SourceCommentStart,
		m.MarkerID)
}
