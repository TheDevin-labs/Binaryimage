package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/TheDevin-labs/binary-image/internal/compiler"
)

func main() {
	arch := flag.String("arch", "", "target architecture: amd64, arm64 (default: host)")
	out := flag.String("o", "output.sbimg", "output .sbimg file path")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: binary-image compile [flags] <source>")
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}

	err := compiler.Compile(compiler.Options{
		Source: flag.Arg(0),
		Output: *out,
		Arch:   *arch,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
