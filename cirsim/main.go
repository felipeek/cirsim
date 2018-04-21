package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/felipeek/cirsim/internal"
)

func main() {
	var filePath string
	flag.StringVar(&filePath, "path", "", "Spice file path")
	flag.Parse()

	if filePath == "" {
		flag.PrintDefaults()
		os.Exit(0)
	}

	_, err := os.Stat(filePath)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Spice file does not exist")
		os.Exit(1)
	}

	internal.ParserInit(filePath)
}
