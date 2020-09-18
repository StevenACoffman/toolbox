package main

import (
	"fmt"
	"os"
	"time"
)

func main() {
	filename := "/tmp/healthy"
	var threshold float64 = 30
	healthy := livenessProbe(filename, threshold)

	if !healthy {
		os.Exit(1)
	}
}

func livenessProbe(filename string, threshold float64) bool {
	// if you aren't using UTC you are wrong.
	loc, _ := time.LoadLocation("UTC")
	now := time.Now().In(loc)
	// get last modified time
	file, err := os.Stat(filename)
	if err != nil {
		fmt.Println(err)
		// if file does not exist, or is unreadable, then it should fail liveness probe
		return false
	}

	modTime := file.ModTime()
	elapsed := now.Sub(modTime).Seconds()

	fmt.Println("Last modified time : ", modTime)
	fmt.Println("Seconds since last file modification : ", elapsed)
	return elapsed < threshold
}
