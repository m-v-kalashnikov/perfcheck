package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func main() {
	src := flag.String("src", "", "path to source rule file")
	dst := flag.String("dst", "", "path to destination file relative to cwd")
	flag.Parse()

	if *src == "" || *dst == "" {
		flag.Usage()
		os.Exit(1)
	}

	in, err := os.Open(*src)
	if err != nil {
		exitErr(fmt.Errorf("open source: %w", err))
	}
	defer func() {
		if cerr := in.Close(); cerr != nil {
			exitErr(fmt.Errorf("close source: %w", cerr))
		}
	}()

	if mkErr := os.MkdirAll(filepath.Dir(*dst), 0o755); mkErr != nil {
		exitErr(fmt.Errorf("create destination dir: %w", mkErr))
	}

	out, err := os.Create(*dst)
	if err != nil {
		exitErr(fmt.Errorf("create destination: %w", err))
	}
	defer func() {
		if cerr := out.Close(); cerr != nil {
			exitErr(fmt.Errorf("close destination: %w", cerr))
		}
	}()

	if _, err := io.Copy(out, in); err != nil {
		exitErr(fmt.Errorf("copy data: %w", err))
	}
}

func exitErr(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
