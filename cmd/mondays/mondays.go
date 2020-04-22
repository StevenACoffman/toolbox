package main

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

func main() {
	now := time.Now()
	argsWithoutProgram := os.Args[1:]
	if len(argsWithoutProgram) > 0 {
		weeks, err := strconv.Atoi(argsWithoutProgram[0])
		if err == nil {
			now = now.AddDate(0, 0, weeks*7)
		}
	}
	last := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	for ; last.Weekday() != time.Monday; last = last.AddDate(0, 0, -1) {}
	next := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, 1)
	for ; next.Weekday() != time.Monday; next = next.AddDate(0, 0, 1) {}
	//backup
	next = next.AddDate(0, 0, -1)
	prefix := "./tools/devshell.py --prod --script districts/tools/district_xls_report.py --begin \""
	middle := "\" --end \""
	suffix := os.ExpandEnv("\" $HOME/DistrictReports 'Clark County'")
	fmt.Printf("%s%s%s%s%s\n", prefix, last.Format("2006-01-02"), middle, next.Format("2006-01-02"), suffix)
}
