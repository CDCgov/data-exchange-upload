package main

import (
	"flag"
	"fmt"
	"strings"
)

var inputFiles CSVFiles

type CSVFiles []string

func (ids CSVFiles) String() string {
	return strings.Join(inputFiles, ",")
}

func (ids CSVFiles) Set(value string) error {
	inputFiles = strings.Split(value, ",")
	return nil
}

func init() {
	flag.Var(&inputFiles, "csvFiles", "file1.csv,file2.csv")
	flag.Parse()
}

func main() {
	// First, load upload IDs from CSV
	fmt.Printf("files: %v", inputFiles)
}
