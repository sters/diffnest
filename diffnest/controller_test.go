package diffnest

import (
	"strings"
	"testing"
)

func TestController_Run(t *testing.T) {
	tests := []struct {
		name            string
		content1        string
		content2        string
		format1         string
		format2         string
		diffOpts        DiffOptions
		formatter       Formatter
		wantErr         bool
		wantDifferences bool
		contains        []string
	}{
		// Basic tests
		{
			name:            "Same JSON files",
			content1:        `{"name": "test", "value": 42}`,
			content2:        `{"name": "test", "value": 42}`,
			format1:         FormatJSON,
			format2:         FormatJSON,
			diffOpts:        DiffOptions{},
			formatter:       &UnifiedFormatter{},
			wantErr:         false,
			wantDifferences: false,
			contains:        []string{"  name: test", "  value: 42"},
		},
		{
			name:            "Different JSON files",
			content1:        `{"name": "test1", "value": 42}`,
			content2:        `{"name": "test2", "value": 43}`,
			format1:         FormatJSON,
			format2:         FormatJSON,
			diffOpts:        DiffOptions{},
			formatter:       &UnifiedFormatter{},
			wantErr:         false,
			wantDifferences: true,
			contains: []string{
				"- name: test1",
				"+ name: test2",
				"- value: 42",
				"+ value: 43",
			},
		},
		{
			name:     "Cross-format comparison",
			content1: `{"name": "test", "value": 42}`,
			content2: `name: test
value: 42`,
			format1:         FormatJSON,
			format2:         FormatYAML,
			diffOpts:        DiffOptions{},
			formatter:       &UnifiedFormatter{},
			wantErr:         false,
			wantDifferences: false,
			contains:        []string{"  name: test", "  value: 42"},
		},
		{
			name:      "Invalid JSON",
			content1:  `{"invalid": `,
			content2:  `{"valid": true}`,
			format1:   FormatJSON,
			format2:   FormatJSON,
			diffOpts:  DiffOptions{},
			formatter: &UnifiedFormatter{},
			wantErr:   true,
		},
		{
			name: "Multiline string diff",
			content1: `data: |
  line1
  line2
  line3`,
			content2: `data: |
  line1
  modified
  line3`,
			format1:         FormatYAML,
			format2:         FormatYAML,
			diffOpts:        DiffOptions{},
			formatter:       &UnifiedFormatter{},
			wantErr:         false,
			wantDifferences: true,
			contains: []string{
				"data:",
				"-    line2",
				"+    modified",
			},
		},
		// Test cases from main_test.go
		{
			name: "JSON with nested differences",
			content1: `{
				"name": "John",
				"age": 30,
				"city": "Tokyo"
			}`,
			content2: `{
				"name": "John",
				"age": 31,
				"city": "Osaka"
			}`,
			format1:         FormatJSON,
			format2:         FormatJSON,
			diffOpts:        DiffOptions{},
			formatter:       &UnifiedFormatter{},
			wantErr:         false,
			wantDifferences: true,
			contains: []string{
				"- age: 30",
				"+ age: 31",
				"- city: Tokyo",
				"+ city: Osaka",
				"  name: John",
			},
		},
		{
			name: "JSON with diff-only option",
			content1: `{
				"name": "John",
				"age": 30,
				"city": "Tokyo"
			}`,
			content2: `{
				"name": "John",
				"age": 31,
				"city": "Osaka"
			}`,
			format1:         FormatJSON,
			format2:         FormatJSON,
			diffOpts:        DiffOptions{},
			formatter:       &UnifiedFormatter{ShowOnlyDiff: true},
			wantErr:         false,
			wantDifferences: true,
			contains: []string{
				"- age: 30",
				"+ age: 31",
				"- city: Tokyo",
				"+ city: Osaka",
			},
		},
		{
			name: "YAML with array differences",
			content1: `name: Alice
age: 25
hobbies:
  - reading
  - gaming`,
			content2: `name: Alice
age: 26
hobbies:
  - reading
  - swimming
  - gaming`,
			format1:         FormatYAML,
			format2:         FormatYAML,
			diffOpts:        DiffOptions{},
			formatter:       &UnifiedFormatter{},
			wantErr:         false,
			wantDifferences: true,
			contains: []string{
				"- age: 25",
				"+ age: 26",
				"hobbies:",
				"+   - swimming",
			},
		},
		{
			name: "Multiple JSON documents",
			content1: `{"id": 1, "name": "Alice", "active": true}
{"id": 2, "name": "Bob", "active": false}
{"id": 3, "name": "Charlie", "active": true}`,
			content2: `{"id": 1, "name": "Alice", "active": false}
{"id": 2, "name": "Robert", "active": true}
{"id": 4, "name": "David", "active": true}`,
			format1:         FormatJSON,
			format2:         FormatJSON,
			diffOpts:        DiffOptions{},
			formatter:       &UnifiedFormatter{},
			wantErr:         false,
			wantDifferences: true,
			contains: []string{
				"- active: true",
				"+ active: false",
				"- name: Charlie",
				"+ name: Robert",
				"- name: Bob",
				"+ name: David",
			},
		},
		{
			name: "Multiple YAML documents",
			content1: `id: 1
name: Alice
department: Engineering
---
id: 2
name: Bob
department: Sales
---
id: 3
name: Charlie
department: Marketing`,
			content2: `id: 1
name: Alice
department: Marketing
---
id: 2
name: Bob
department: Engineering
---
id: 3
name: Charles
department: Marketing`,
			format1:         FormatYAML,
			format2:         FormatYAML,
			diffOpts:        DiffOptions{},
			formatter:       &UnifiedFormatter{},
			wantErr:         false,
			wantDifferences: true,
			contains: []string{
				"- department: Engineering",
				"+ department: Marketing",
				"- department: Sales",
				"+ department: Engineering",
				"- name: Charlie",
				"+ name: Charles",
			},
		},
		{
			name: "Array differences with index strategy",
			content1: `{
				"items": ["apple", "banana", "cherry"],
				"tags": ["fruit", "healthy", "organic"]
			}`,
			content2: `{
				"items": ["banana", "cherry", "date", "apple"],
				"tags": ["fruit", "organic", "fresh", "healthy"]
			}`,
			format1:         FormatJSON,
			format2:         FormatJSON,
			diffOpts:        DiffOptions{ArrayDiffStrategy: ArrayStrategyIndex},
			formatter:       &UnifiedFormatter{},
			wantErr:         false,
			wantDifferences: true,
			contains: []string{
				"items:",
				"-   - apple",
				"+   - banana",
				"-   - banana",
				"+   - cherry",
				"-   - cherry",
				"+   - date",
				"+   - apple",
				"tags:",
				"-   - healthy",
				"+   - organic",
				"-   - organic",
				"+   - fresh",
				"+   - healthy",
			},
		},
		{
			name: "Array differences with value strategy",
			content1: `{
				"items": ["apple", "banana", "cherry"],
				"tags": ["fruit", "healthy", "organic"]
			}`,
			content2: `{
				"items": ["banana", "cherry", "date", "apple"],
				"tags": ["fruit", "organic", "fresh", "healthy"]
			}`,
			format1:         FormatJSON,
			format2:         FormatJSON,
			diffOpts:        DiffOptions{ArrayDiffStrategy: ArrayStrategyValue},
			formatter:       &UnifiedFormatter{},
			wantErr:         false,
			wantDifferences: true,
			contains: []string{
				"items:",
				"+   - date",
				"tags:",
				"+   - fresh",
			},
		},
		{
			name: "YAML multiline string formats (issue #29)",
			content1: `value: |-
  foo
  bar
  baz
  special
   multiline`,
			content2:        `value: "foo\nbar\nbaz\nspecial\n  multiline"`,
			format1:         FormatYAML,
			format2:         FormatYAML,
			diffOpts:        DiffOptions{},
			formatter:       &UnifiedFormatter{},
			wantErr:         false,
			wantDifferences: true,
			contains: []string{
				"value:",
			},
		},
		{
			name: "YAML multiline config diff (issue #52)",
			content1: `data:
  config: |
   logging.a: false
   logging.b: false`,
			content2: `data:
  config: |
   logging.a: false
   logging.c: false`,
			format1:         FormatYAML,
			format2:         FormatYAML,
			diffOpts:        DiffOptions{},
			formatter:       &UnifiedFormatter{},
			wantErr:         false,
			wantDifferences: true,
			contains: []string{
				"data:",
				"data.config:",
				"logging.a: false",
				"-      logging.b: false",
				"+      logging.c: false",
			},
		},
		{
			name:            "JSON Patch formatter",
			content1:        `{"name": "test", "value": 42}`,
			content2:        `{"name": "test2", "value": 43}`,
			format1:         FormatJSON,
			format2:         FormatJSON,
			diffOpts:        DiffOptions{},
			formatter:       &JSONPatchFormatter{},
			wantErr:         false,
			wantDifferences: true,
			contains: []string{
				`{"op": "replace", "path": "/name", "value": "test2"}`,
				`{"op": "replace", "path": "/value", "value": 43}`,
			},
		},
		{
			name: "Ignore zero values",
			content1: `{
				"name": "test",
				"count": 0,
				"active": false,
				"tags": []
			}`,
			content2: `{
				"name": "test2",
				"count": 5,
				"active": true,
				"tags": ["new"]
			}`,
			format1:         FormatJSON,
			format2:         FormatJSON,
			diffOpts:        DiffOptions{IgnoreZeroValues: true},
			formatter:       &UnifiedFormatter{},
			wantErr:         false,
			wantDifferences: true,
			contains: []string{
				"- name: test",
				"+ name: test2",
				// count, active, and tags should be treated as added since they were zero values
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader1 := strings.NewReader(tt.content1)
			reader2 := strings.NewReader(tt.content2)
			var output strings.Builder

			controller := NewController(
				reader1,
				reader2,
				tt.format1,
				tt.format2,
				tt.diffOpts,
				tt.formatter,
				&output,
			)

			hasDifferences, err := controller.Run()

			// Check error
			if (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if tt.wantErr {
				return // Skip other checks if we expected an error
			}

			// Check differences found
			if hasDifferences != tt.wantDifferences {
				t.Errorf("Run() hasDifferences = %v, want %v", hasDifferences, tt.wantDifferences)
			}

			// Check output contains expected strings
			outputStr := output.String()
			for _, want := range tt.contains {
				if !strings.Contains(outputStr, want) {
					t.Errorf("Output missing expected string %q\nGot output:\n%s", want, outputStr)
				}
			}
		})
	}
}

func TestHasDifferences(t *testing.T) {
	tests := []struct {
		name    string
		results []*DiffResult
		want    bool
	}{
		{
			name:    "No results",
			results: []*DiffResult{},
			want:    false,
		},
		{
			name: "All same",
			results: []*DiffResult{
				{Status: StatusSame},
				{Status: StatusSame},
			},
			want: false,
		},
		{
			name: "Has modifications",
			results: []*DiffResult{
				{Status: StatusSame},
				{Status: StatusModified},
				{Status: StatusSame},
			},
			want: true,
		},
		{
			name: "Has additions",
			results: []*DiffResult{
				{Status: StatusAdded},
			},
			want: true,
		},
		{
			name: "Has deletions",
			results: []*DiffResult{
				{Status: StatusDeleted},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HasDifferences(tt.results); got != tt.want {
				t.Errorf("HasDifferences() = %v, want %v", got, tt.want)
			}
		})
	}
}
