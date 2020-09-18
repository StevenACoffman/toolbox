package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"
)

func main() {
	filename := "/tmp/healthy"
	touchFile(filename)
}

func touchFile(filename string) {
	loc, _ := time.LoadLocation("UTC")
	now := time.Now().In(loc)

	// get last modified time
	file, err := os.Stat(filename)

	if os.IsNotExist(err) {
		// write timestamp to file so you can compare initial creation to last modification
		const stampF = "20060102150405"
		err = ioutil.WriteFile(filename, []byte(now.Format(stampF)), 0644)
		file, err = os.Stat(filename)
	}

	if err != nil {
		fmt.Println(err)
	}

	modTime := file.ModTime()
	fmt.Println("Last modified time : ", modTime)

	// change both atime and mtime to current
	err = os.Chtimes(filename, now, now)

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Changed the file time : ", now.Format(time.RFC3339))
}
