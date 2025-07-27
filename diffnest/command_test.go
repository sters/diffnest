package diffnest

import (
	"bytes"
	"flag"
	"strings"
	"testing"
)

func TestCommand_Parse(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
		check   func(t *testing.T, cmd *Command)
	}{
		{
			name:    "Valid two files",
			args:    []string{"file1.json", "file2.json"},
			wantErr: false,
			check: func(t *testing.T, cmd *Command) {
				t.Helper()
				if cmd.File1 != "file1.json" {
					t.Errorf("File1 = %v, want file1.json", cmd.File1)
				}
				if cmd.File2 != "file2.json" {
					t.Errorf("File2 = %v, want file2.json", cmd.File2)
				}
			},
		},
		{
			name:    "With flags",
			args:    []string{"-diff-only", "-v", "file1.yaml", "file2.yaml"},
			wantErr: false,
			check: func(t *testing.T, cmd *Command) {
				t.Helper()
				if !cmd.ShowOnlyDiff {
					t.Error("ShowOnlyDiff should be true")
				}
				if !cmd.Verbose {
					t.Error("Verbose should be true")
				}
				if cmd.File1 != "file1.yaml" {
					t.Errorf("File1 = %v, want file1.yaml", cmd.File1)
				}
				if cmd.File2 != "file2.yaml" {
					t.Errorf("File2 = %v, want file2.yaml", cmd.File2)
				}
			},
		},
		{
			name:    "Help flag",
			args:    []string{"-h"},
			wantErr: false,
			check: func(t *testing.T, cmd *Command) {
				t.Helper()
				if !cmd.Help {
					t.Error("Help should be true")
				}
			},
		},
		{
			name:    "Missing files",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "Only one file",
			args:    []string{"file1.json"},
			wantErr: true,
		},
		{
			name:    "Array strategy",
			args:    []string{"-array-strategy", "index", "f1", "f2"},
			wantErr: false,
			check: func(t *testing.T, cmd *Command) {
				t.Helper()
				if cmd.ArrayStrategy != "index" {
					t.Errorf("ArrayStrategy = %v, want index", cmd.ArrayStrategy)
				}
			},
		},
		{
			name:    "Format flags",
			args:    []string{"-format1", "json", "-format2", "yaml", "f1", "f2"},
			wantErr: false,
			check: func(t *testing.T, cmd *Command) {
				t.Helper()
				if cmd.Format1 != "json" {
					t.Errorf("Format1 = %v, want json", cmd.Format1)
				}
				if cmd.Format2 != "yaml" {
					t.Errorf("Format2 = %v, want yaml", cmd.Format2)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewCommand("test", flag.ContinueOnError)
			var buf bytes.Buffer
			cmd.SetOutput(&buf)

			err := cmd.Parse(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if !tt.wantErr && tt.check != nil {
				tt.check(t, cmd)
			}
		})
	}
}

func TestCommand_GetDiffOptions(t *testing.T) {
	tests := []struct {
		name  string
		setup func(cmd *Command)
		check func(t *testing.T, opts DiffOptions)
	}{
		{
			name:  "Default options",
			setup: func(_ *Command) {},
			check: func(t *testing.T, opts DiffOptions) {
				t.Helper()
				if opts.IgnoreZeroValues {
					t.Error("IgnoreZeroValues should be false by default")
				}
				if opts.IgnoreEmptyFields {
					t.Error("IgnoreEmptyFields should be false by default")
				}
				if opts.ArrayDiffStrategy != ArrayStrategyValue {
					t.Error("ArrayDiffStrategy should be Value by default")
				}
			},
		},
		{
			name: "With ignore flags",
			setup: func(cmd *Command) {
				cmd.IgnoreZeroValues = true
				cmd.IgnoreEmpty = true
			},
			check: func(t *testing.T, opts DiffOptions) {
				t.Helper()
				if !opts.IgnoreZeroValues {
					t.Error("IgnoreZeroValues should be true")
				}
				if !opts.IgnoreEmptyFields {
					t.Error("IgnoreEmptyFields should be true")
				}
			},
		},
		{
			name: "Index array strategy",
			setup: func(cmd *Command) {
				cmd.ArrayStrategy = "index"
			},
			check: func(t *testing.T, opts DiffOptions) {
				t.Helper()
				if opts.ArrayDiffStrategy != ArrayStrategyIndex {
					t.Error("ArrayDiffStrategy should be Index")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewCommand("test", flag.ContinueOnError)
			tt.setup(cmd)
			opts := cmd.GetDiffOptions()
			tt.check(t, opts)
		})
	}
}

func TestCommand_GetFormatter(t *testing.T) {
	tests := []struct {
		name     string
		format   string
		diffOnly bool
		verbose  bool
		wantType string
	}{
		{
			name:     "Default unified formatter",
			format:   "unified",
			wantType: "*diffnest.UnifiedFormatter",
		},
		{
			name:     "JSON patch formatter",
			format:   "json-patch",
			wantType: "*diffnest.JSONPatchFormatter",
		},
		{
			name:     "Unified with options",
			format:   "unified",
			diffOnly: true,
			verbose:  true,
			wantType: "*diffnest.UnifiedFormatter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewCommand("test", flag.ContinueOnError)
			cmd.OutputFormat = tt.format
			cmd.ShowOnlyDiff = tt.diffOnly
			cmd.Verbose = tt.verbose

			formatter := cmd.GetFormatter()

			switch formatter := formatter.(type) {
			case *UnifiedFormatter:
				if tt.wantType != "*diffnest.UnifiedFormatter" {
					t.Errorf("Got UnifiedFormatter, want %s", tt.wantType)
				}
				if tt.diffOnly {
					uf := formatter
					if !uf.ShowOnlyDiff {
						t.Error("ShowOnlyDiff should be true")
					}
				}
			case *JSONPatchFormatter:
				if tt.wantType != "*diffnest.JSONPatchFormatter" {
					t.Errorf("Got JSONPatchFormatter, want %s", tt.wantType)
				}
			default:
				t.Errorf("Unknown formatter type")
			}
		})
	}
}

func TestCommand_GetFormat(t *testing.T) {
	tests := []struct {
		name        string
		format1     string
		format2     string
		file1       string
		file2       string
		wantFormat1 string
		wantFormat2 string
	}{
		{
			name:        "Auto-detect from filename",
			file1:       "test.json",
			file2:       "test.yaml",
			wantFormat1: FormatJSON,
			wantFormat2: FormatYAML,
		},
		{
			name:        "Explicit formats",
			format1:     FormatJSON,
			format2:     FormatYAML,
			file1:       "test.txt",
			file2:       "test.txt",
			wantFormat1: FormatJSON,
			wantFormat2: FormatYAML,
		},
		{
			name:        "Stdin defaults to YAML",
			file1:       "-",
			file2:       "-",
			wantFormat1: FormatYAML,
			wantFormat2: FormatYAML,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewCommand("test", flag.ContinueOnError)
			cmd.Format1 = tt.format1
			cmd.Format2 = tt.format2
			cmd.File1 = tt.file1
			cmd.File2 = tt.file2

			if got := cmd.GetFormat1(); got != tt.wantFormat1 {
				t.Errorf("GetFormat1() = %v, want %v", got, tt.wantFormat1)
			}
			if got := cmd.GetFormat2(); got != tt.wantFormat2 {
				t.Errorf("GetFormat2() = %v, want %v", got, tt.wantFormat2)
			}
		})
	}
}

func TestCommand_Usage(t *testing.T) {
	cmd := NewCommand("test", flag.ContinueOnError)
	var buf bytes.Buffer

	cmd.Usage(&buf)

	output := buf.String()
	expectedStrings := []string{
		"Usage: diffnest",
		"Options:",
		"-diff-only",
		"-array-strategy",
		"Example:",
		"file1.json file2.json",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Usage output missing %q", expected)
		}
	}
}
