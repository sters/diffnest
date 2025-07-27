package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/sters/diffnest/diffnest"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	cmd := diffnest.NewCommand("diffnest", flag.ContinueOnError)
	cmd.SetOutput(stderr)

	if err := cmd.Parse(args); err != nil {
		if !errors.Is(err, flag.ErrHelp) {
			fmt.Fprintf(stderr, "Error: %v\n", err)
		}

		return 1
	}

	if cmd.Help {
		cmd.Usage(stderr)

		return 0
	}

	reader1, err := openFile(cmd.File1)
	if err != nil {
		fmt.Fprintf(stderr, "Error opening first file: %v\n", err)

		return 1
	}
	defer closeReader(reader1)

	reader2, err := openFile(cmd.File2)
	if err != nil {
		fmt.Fprintf(stderr, "Error opening second file: %v\n", err)

		return 1
	}
	defer closeReader(reader2)

	controller := diffnest.NewController(
		reader1,
		reader2,
		cmd.GetFormat1(),
		cmd.GetFormat2(),
		cmd.GetDiffOptions(),
		cmd.GetFormatter(),
		stdout,
	)

	hasDifferences, err := controller.Run()
	if err != nil {
		fmt.Fprintf(stderr, "%v\n", err)

		return 1
	}

	if hasDifferences {
		return 1
	}

	return 0
}

func openFile(filename string) (io.ReadCloser, error) {
	if filename == "-" {
		return io.NopCloser(os.Stdin), nil
	}

	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", filename, err)
	}

	return file, nil
}

func closeReader(r io.ReadCloser) {
	if r != nil && r != io.NopCloser(os.Stdin) {
		r.Close()
	}
}
