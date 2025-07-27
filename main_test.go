package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLI(t *testing.T) {
	tempDir := t.TempDir()
	json1 := filepath.Join(tempDir, "test1.json")
	json2 := filepath.Join(tempDir, "test2.json")
	yaml1 := filepath.Join(tempDir, "test1.yaml")

	if err := os.WriteFile(json1, []byte(`{"name": "test", "value": 42, "enabled": true}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(json2, []byte(`{"name": "test2", "value": 43, "enabled": true}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(yaml1, []byte("name: test\nvalue: 42\nenabled: true"), 0o644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name     string
		args     []string
		wantExit int
		wantOut  string
		wantErr  string
	}{
		{
			name:     "Help flag",
			args:     []string{"-h"},
			wantExit: 0,
			wantErr:  "Usage: diffnest",
		},
		{
			name:     "Missing files",
			args:     []string{},
			wantExit: 1,
			wantErr:  "expected 2 files",
		},
		{
			name:     "Only one file",
			args:     []string{json1},
			wantExit: 1,
			wantErr:  "expected 2 files",
		},
		{
			name:     "Non-existent file",
			args:     []string{json1, "nonexistent.json"},
			wantExit: 1,
			wantErr:  "Error opening second file",
		},
		{
			name:     "Different files",
			args:     []string{json1, json2},
			wantExit: 1,
			wantOut:  "- name: test",
		},
		{
			name:     "Same content different format",
			args:     []string{"-show-all", json1, yaml1},
			wantExit: 0,
			wantOut:  "name: test",
		},
		{
			name:     "JSON patch format",
			args:     []string{"-format", "json-patch", json1, json2},
			wantExit: 1,
			wantOut:  `{"op": "replace"`,
		},
		{
			name:     "Show all flag",
			args:     []string{"-show-all", json1, json2},
			wantExit: 1,
			wantOut:  "  enabled: true",
		},
		{
			name:     "Verbose flag",
			args:     []string{"-v", json1, json2},
			wantExit: 1,
			wantOut:  "- name: test",
		},
		{
			name:     "Force formats",
			args:     []string{"-show-all", "-format1", "json", "-format2", "yaml", json1, yaml1},
			wantExit: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer

			exitCode := run(tt.args, &stdout, &stderr)

			if exitCode != tt.wantExit {
				t.Errorf("Exit code = %d, want %d", exitCode, tt.wantExit)
			}

			if tt.wantOut != "" && !strings.Contains(stdout.String(), tt.wantOut) {
				t.Errorf("stdout missing expected string %q\nGot:\n%s", tt.wantOut, stdout.String())
			}

			if tt.wantErr != "" && !strings.Contains(stderr.String(), tt.wantErr) {
				t.Errorf("stderr missing expected string %q\nGot:\n%s", tt.wantErr, stderr.String())
			}
		})
	}
}
