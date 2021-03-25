package main

import (
	"bufio"
	"bytes"
	"fmt"
	"hash/adler32"
	"os"
)

func main() {
	//const someCode = "does a thing";
	//console.log(someCode);
	//

	var block []byte

	vss := "// sync-start:" // javascript, go
	vssBytes := []byte(vss)
	vse := "// sync-end:"
	vseBytes := []byte(vse)

	scanner := bufio.NewScanner(bufio.NewReader(os.Stdin))

	scanner.Split(ScanLines)

	for scanner.Scan() {
		currentBytes := scanner.Bytes()

		if !bytes.Contains(currentBytes, vssBytes) && !bytes.Contains(currentBytes, vseBytes) {
			block = append(block, currentBytes...)
		}
	}
	salted := append( []byte("\n"), block...)
	adler32Int := adler32.Checksum(salted)
	fmt.Println("Adler32 String is ", adler32Int)

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
