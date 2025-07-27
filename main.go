package main

import (
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
	flags := flag.NewFlagSet("diffnest", flag.ContinueOnError)
	flags.SetOutput(stderr)
	
	var (
		showOnlyDiff     = flags.Bool("diff-only", false, "Show only differences")
		ignoreZeroValues = flags.Bool("ignore-zero-values", false, "Treat zero values (0, false, \"\", [], {}) as null")
		ignoreEmpty      = flags.Bool("ignore-empty", false, "Ignore empty fields")
		arrayStrategy    = flags.String("array-strategy", "value", "Array comparison strategy: 'index' or 'value'")
		outputFormat     = flags.String("format", "unified", "Output format: 'unified' or 'json-patch'")
		format1          = flags.String("format1", "", "Format for first file: 'json', 'yaml', or auto-detect from filename")
		format2          = flags.String("format2", "", "Format for second file: 'json', 'yaml', or auto-detect from filename")
		verbose          = flags.Bool("v", false, "Verbose output")
		help             = flags.Bool("h", false, "Show help")
	)

	flags.Usage = func() {
		fmt.Fprintf(stderr, "Usage: diffnest [options] <file1> <file2>\n")
		fmt.Fprintf(stderr, "\nOptions:\n")
		flags.PrintDefaults()
		fmt.Fprintf(stderr, "\nExample:\n")
		fmt.Fprintf(stderr, "  diffnest file1.json file2.json\n")
		fmt.Fprintf(stderr, "  diffnest file1.yaml file2.yaml\n")
		fmt.Fprintf(stderr, "  diffnest file1.json file2.yaml  # Compare different formats\n")
		fmt.Fprintf(stderr, "  cat file1.json | diffnest - file2.json\n")
		fmt.Fprintf(stderr, "  diffnest --format1 json - file2.yaml  # Force JSON format for stdin\n")
	}

	if err := flags.Parse(args); err != nil {
		return 1
	}

	if *help || flags.NArg() != 2 {
		flags.Usage()
		return 1
	}

	file1 := flags.Arg(0)
	file2 := flags.Arg(1)

	// Determine format for file1
	fileFormat1 := *format1
	if fileFormat1 == "" {
		if file1 == "-" {
			// For stdin, default to YAML
			fileFormat1 = diffnest.FormatYAML
		} else {
			fileFormat1 = diffnest.DetectFormatFromFilename(file1)
		}
	}

	// Determine format for file2
	fileFormat2 := *format2
	if fileFormat2 == "" {
		if file2 == "-" {
			// For stdin, default to YAML
			fileFormat2 = diffnest.FormatYAML
		} else {
			fileFormat2 = diffnest.DetectFormatFromFilename(file2)
		}
	}

	// Open files
	reader1, err := openFile(file1)
	if err != nil {
		fmt.Fprintf(stderr, "Error opening first file: %v\n", err)
		return 1
	}
	defer closeReader(reader1)

	reader2, err := openFile(file2)
	if err != nil {
		fmt.Fprintf(stderr, "Error opening second file: %v\n", err)
		return 1
	}
	defer closeReader(reader2)

	// Parse data
	docs1, err := diffnest.ParseWithFormat(reader1, fileFormat1)
	if err != nil {
		fmt.Fprintf(stderr, "Error parsing first file: %v\n", err)
		return 1
	}

	docs2, err := diffnest.ParseWithFormat(reader2, fileFormat2)
	if err != nil {
		fmt.Fprintf(stderr, "Error parsing second file: %v\n", err)
		return 1
	}

	// Prepare options
	opts := diffnest.DiffOptions{
		IgnoreZeroValues:  *ignoreZeroValues,
		IgnoreEmptyFields: *ignoreEmpty,
	}

	// Set array strategy
	if *arrayStrategy == "index" {
		opts.ArrayDiffStrategy = diffnest.ArrayStrategyIndex
	} else {
		opts.ArrayDiffStrategy = diffnest.ArrayStrategyValue
	}

	// Perform diff
	results := diffnest.Compare(docs1, docs2, opts)

	// Format output
	var formatter diffnest.Formatter
	switch *outputFormat {
	case "json-patch":
		formatter = &diffnest.JSONPatchFormatter{}
	default:
		formatter = &diffnest.UnifiedFormatter{
			ShowOnlyDiff: *showOnlyDiff,
			Verbose:      *verbose,
		}
	}

	output := formatter.Format(results)
	fmt.Fprint(stdout, output)

	// Exit with non-zero status if differences found
	for _, result := range results {
		if result.Status != diffnest.StatusSame {
			return 1
		}
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

func readFile(filename string) (string, error) {
	reader, err := openFile(filename)
	if err != nil {
		return "", err
	}
	defer closeReader(reader)

	data, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %w", filename, err)
	}
	return string(data), nil
}
