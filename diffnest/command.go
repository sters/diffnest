package diffnest

import (
	"errors"
	"flag"
	"fmt"
	"io"
)

const arrayStrategyIndex = "index"

var ErrInvalidArgs = errors.New("expected 2 files")

// Command represents the CLI command configuration.
type Command struct {
	// Flags
	ShowOnlyDiff     bool
	IgnoreZeroValues bool
	IgnoreEmpty      bool
	ArrayStrategy    string
	OutputFormat     string
	Format1          string
	Format2          string
	Verbose          bool
	Help             bool

	// Arguments
	File1 string
	File2 string

	// FlagSet for parsing
	flags *flag.FlagSet
}

// NewCommand creates a new Command instance.
func NewCommand(name string, errorHandling flag.ErrorHandling) *Command {
	cmd := &Command{
		flags: flag.NewFlagSet(name, errorHandling),
	}

	cmd.flags.BoolVar(&cmd.ShowOnlyDiff, "diff-only", false, "Show only differences")
	cmd.flags.BoolVar(&cmd.IgnoreZeroValues, "ignore-zero-values", false, "Treat zero values (0, false, \"\", [], {}) as null")
	cmd.flags.BoolVar(&cmd.IgnoreEmpty, "ignore-empty", false, "Ignore empty fields")
	cmd.flags.StringVar(&cmd.ArrayStrategy, "array-strategy", "value", "Array comparison strategy: 'index' or 'value'")
	cmd.flags.StringVar(&cmd.OutputFormat, "format", "unified", "Output format: 'unified' or 'json-patch'")
	cmd.flags.StringVar(&cmd.Format1, "format1", "", "Format for first file: 'json', 'yaml', or auto-detect from filename")
	cmd.flags.StringVar(&cmd.Format2, "format2", "", "Format for second file: 'json', 'yaml', or auto-detect from filename")
	cmd.flags.BoolVar(&cmd.Verbose, "v", false, "Verbose output")
	cmd.flags.BoolVar(&cmd.Help, "h", false, "Show help")

	return cmd
}

// SetOutput sets the output destination for error messages.
func (c *Command) SetOutput(w io.Writer) {
	c.flags.SetOutput(w)
}

// Parse parses command line arguments.
func (c *Command) Parse(args []string) error {
	if err := c.flags.Parse(args); err != nil {
		return fmt.Errorf("parse flags: %w", err)
	}

	if c.flags.NArg() != 2 && !c.Help {
		return fmt.Errorf("%w, got %d", ErrInvalidArgs, c.flags.NArg())
	}

	if c.flags.NArg() >= 2 {
		c.File1 = c.flags.Arg(0)
		c.File2 = c.flags.Arg(1)
	}

	return nil
}

// Usage prints usage information.
func (c *Command) Usage(w io.Writer) {
	fmt.Fprintf(w, "Usage: diffnest [options] <file1> <file2>\n")
	fmt.Fprintf(w, "\nOptions:\n")
	c.flags.SetOutput(w)
	c.flags.PrintDefaults()
	fmt.Fprintf(w, "\nExample:\n")
	fmt.Fprintf(w, "  diffnest file1.json file2.json\n")
	fmt.Fprintf(w, "  diffnest file1.yaml file2.yaml\n")
	fmt.Fprintf(w, "  diffnest file1.json file2.yaml  # Compare different formats\n")
	fmt.Fprintf(w, "  cat file1.json | diffnest - file2.json\n")
	fmt.Fprintf(w, "  diffnest --format1 json - file2.yaml  # Force JSON format for stdin\n")
}

// GetDiffOptions returns DiffOptions based on command flags.
func (c *Command) GetDiffOptions() DiffOptions {
	opts := DiffOptions{
		IgnoreZeroValues:  c.IgnoreZeroValues,
		IgnoreEmptyFields: c.IgnoreEmpty,
	}

	if c.ArrayStrategy == arrayStrategyIndex {
		opts.ArrayDiffStrategy = ArrayStrategyIndex
	} else {
		opts.ArrayDiffStrategy = ArrayStrategyValue
	}

	return opts
}

// GetFormatter returns the appropriate formatter based on command flags.
func (c *Command) GetFormatter() Formatter {
	switch c.OutputFormat {
	case "json-patch":
		return &JSONPatchFormatter{}
	default:
		return &UnifiedFormatter{
			ShowOnlyDiff: c.ShowOnlyDiff,
			Verbose:      c.Verbose,
		}
	}
}

// GetFormat1 returns the format for file1, auto-detecting if necessary.
func (c *Command) GetFormat1() string {
	if c.Format1 != "" {
		return c.Format1
	}
	if c.File1 == "-" {
		return FormatYAML
	}

	return DetectFormatFromFilename(c.File1)
}

// GetFormat2 returns the format for file2, auto-detecting if necessary.
func (c *Command) GetFormat2() string {
	if c.Format2 != "" {
		return c.Format2
	}
	if c.File2 == "-" {
		return FormatYAML
	}

	return DetectFormatFromFilename(c.File2)
}
