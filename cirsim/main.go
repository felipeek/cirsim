package main

import (
	"os"

	"github.com/felipeek/cirsim/internal"
)

func main() {
	internal.ParserInit(os.Args[1])
	internal.ParserInit("./res/example.spice")
}
