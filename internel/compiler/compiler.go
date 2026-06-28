package compiler

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/TheDevin-Labs/binaryimage/pkg/sbimg"
)

type Options struct {
	Source string
	Output string
	Arch   string
}

func Compile(opts Options) error {
	arch, err := resolveArch(opts.Arch)
	if err != nil {
		return err
	}

	tmp, err := os.MkdirTemp("", "sbimg-build-*")
	if err != nil {
		return fmt.Errorf("compiler: failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmp)

	elfPath := filepath.Join(tmp, "out.elf")
	rawPath := filepath.Join(tmp, "out.bin")

	goarch := opts.Arch
	if goarch == "" {
		goarch = runtime.GOARCH
	}

	build := exec.Command("go", "build",
		"-ldflags", "-s -w",
		"-o", elfPath,
		opts.Source,
	)
	build.Env = append(os.Environ(),
		"GOOS=linux",
		"GOARCH="+goarch,
		"CGO_ENABLED=0",
	)
	build.Stdout = os.Stdout
	build.Stderr = os.Stderr

	if err := build.Run(); err != nil {
		return fmt.Errorf("compiler: go build failed: %w", err)
	}

	objcopy := exec.Command("objcopy", "-O", "binary",
		"--only-section=.text",
		elfPath, rawPath,
	)
	objcopy.Stdout = os.Stdout
	objcopy.Stderr = os.Stderr

	if err := objcopy.Run(); err != nil {
		return fmt.Errorf("compiler: objcopy failed: %w", err)
	}

	code, err := os.ReadFile(rawPath)
	if err != nil {
		return fmt.Errorf("compiler: failed to read raw binary: %w", err)
	}

	out := opts.Output
	if out == "" {
		out = "output.sbimg"
	}

	if err := sbimg.Write(out, arch, code); err != nil {
		return fmt.Errorf("compiler: failed to write sbimg: %w", err)
	}

	info, _ := os.Stat(out)
	fmt.Printf("compiled: %s (%d bytes of machine code)\n", out, info.Size())
	return nil
}

func resolveArch(s string) (sbimg.Arch, error) {
	switch s {
	case "amd64", "x86_64", "":
		if s == "" && runtime.GOARCH == "arm64" {
			return sbimg.ArchARM64, nil
		}
		return sbimg.ArchAMD64, nil
	case "arm64", "aarch64":
		return sbimg.ArchARM64, nil
	default:
		return 0, fmt.Errorf("compiler: unsupported arch %q", s)
	}
}
