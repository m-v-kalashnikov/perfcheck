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
	defer in.Close()

	if err := os.MkdirAll(filepath.Dir(*dst), 0o755); err != nil {
		exitErr(fmt.Errorf("create destination dir: %w", err))
	}

	out, err := os.Create(*dst)
	if err != nil {
		exitErr(fmt.Errorf("create destination: %w", err))
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		exitErr(fmt.Errorf("copy data: %w", err))
	}
}

func exitErr(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
