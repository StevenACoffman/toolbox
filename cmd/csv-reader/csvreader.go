package main

import (
    "encoding/csv"
    "errors"
    "fmt"
    "log"
    "os"
    "path/filepath"
    "strings"
)

func main() {
    argsWithoutProg := getArgs()
    if len(argsWithoutProg) == 0 {
        fmt.Fprintln(os.Stderr,"Usage: csvreader <filename>")
        panic(errors.New("Usage: csvreader <filename>"))
    }
    workDir, _ := os.Getwd()
    records := readCsvFile(filepath.Join(workDir, argsWithoutProg[0]))
    fmt.Println(records)
}

// no flags please
func getArgs() []string {
    var args []string
    for _, arg := range os.Args[1:] {
        if strings.HasPrefix(arg, "--") || strings.HasPrefix(arg, "-"){
            continue
        }
        args = append(args,arg)
    }
    return args
}

func readCsvFile(filePath string) [][]string {
    f, err := os.Open(filePath)
    if err != nil {
        log.Fatal("Unable to read input file " + filePath, err)
    }
    defer f.Close()

    csvReader := csv.NewReader(f)
    records, err := csvReader.ReadAll()
    if err != nil {
        log.Fatal("Unable to parse file as CSV for " + filePath, err)
    }

    return records
}
