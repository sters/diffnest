package diffnest

import (
	"errors"
	"flag"
	"fmt"
	"io"
)

const arrayStrategyIndex = "index"

var (
	ErrInvalidArgs         = errors.New("expected 2 files")
	ErrIncompatibleOptions = errors.New("--show-all and -C options are incompatible: context lines are only meaningful when showing only differences")
)

// Command represents the CLI command configuration.
type Command struct {
	// Flags
	ShowAll          bool
	IgnoreZeroValues bool
	IgnoreEmpty      bool
	IgnoreKeyCase    bool
	IgnoreValueCase  bool
	ArrayStrategy    string
	OutputFormat     string
	Format1          string
	Format2          string
	Verbose          bool
	Help             bool
	ContextLines     int

	// Arguments
	File1 string
	File2 string

	// FlagSet for parsing
	flags *flag.FlagSet

	// Track if context lines was explicitly set
	contextLinesSet bool
}

// NewCommand creates a new Command instance.
func NewCommand(name string, errorHandling flag.ErrorHandling) *Command {
	cmd := &Command{
		flags:        flag.NewFlagSet(name, errorHandling),
		ContextLines: 3, // Default value
	}

	cmd.flags.BoolVar(&cmd.ShowAll, "show-all", false, "Show all fields including unchanged ones")
	cmd.flags.BoolVar(&cmd.IgnoreZeroValues, "ignore-zero-values", false, "Treat zero values (0, false, \"\", [], {}) as null")
	cmd.flags.BoolVar(&cmd.IgnoreEmpty, "ignore-empty", false, "Ignore empty fields")
	cmd.flags.BoolVar(&cmd.IgnoreKeyCase, "ignore-key-case", false, "Ignore case differences in object keys")
	cmd.flags.BoolVar(&cmd.IgnoreValueCase, "ignore-value-case", false, "Ignore case differences in string values")
	cmd.flags.StringVar(&cmd.ArrayStrategy, "array-strategy", "value", "Array comparison strategy: 'index' or 'value'")
	cmd.flags.StringVar(&cmd.OutputFormat, "format", "unified", "Output format: 'unified' or 'json-patch'")
	cmd.flags.StringVar(&cmd.Format1, "format1", "", "Format for first file: 'json', 'yaml', or auto-detect from filename")
	cmd.flags.StringVar(&cmd.Format2, "format2", "", "Format for second file: 'json', 'yaml', or auto-detect from filename")
	cmd.flags.BoolVar(&cmd.Verbose, "v", false, "Verbose output")
	cmd.flags.BoolVar(&cmd.Help, "h", false, "Show help")
	cmd.flags.IntVar(&cmd.ContextLines, "C", 3, "Number of context lines to show (only for unified format)")

	return cmd
}

// SetOutput sets the output destination for error messages.
func (c *Command) SetOutput(w io.Writer) {
	c.flags.SetOutput(w)
}

// Parse parses command line arguments.
func (c *Command) Parse(args []string) error {
	// Check if -C flag was explicitly set
	for i, arg := range args {
		if arg == "-C" && i+1 < len(args) {
			c.contextLinesSet = true

			break
		}
	}

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

	// Validate incompatible options
	if c.ShowAll && c.contextLinesSet {
		return ErrIncompatibleOptions
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
		IgnoreKeyCase:     c.IgnoreKeyCase,
		IgnoreValueCase:   c.IgnoreValueCase,
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
			ShowOnlyDiff: !c.ShowAll,
			Verbose:      c.Verbose,
			ContextLines: c.ContextLines,
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
