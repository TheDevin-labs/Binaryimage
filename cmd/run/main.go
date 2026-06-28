package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/The-DevinLabs/binaryimage/internal/runner"
	"github.com/The-DevinLabs/binaryimage/pkg/sbimg"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: binary-image run <file.sbimg>")
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}

	img, err := sbimg.Read(flag.Arg(0))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("loading: %s  arch=%s  size=%d bytes\n",
		flag.Arg(0), img.Header.Arch, img.Header.Size)

	result, err := runner.Execute(img)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if result.Fault != nil {
		fmt.Fprintf(os.Stderr, "terminated: hardware exception — %s\n", result.Fault.Error())
		os.Exit(2)
	}

	fmt.Println("execution complete")
}
