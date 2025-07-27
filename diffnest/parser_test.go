package diffnest

import (
	"reflect"
	"strings"
	"testing"
)

func TestDetectFormatFromFilename(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		expected string
	}{
		{
			name:     "JSON file",
			filename: "test.json",
			expected: FormatJSON,
		},
		{
			name:     "JSON file with path",
			filename: "/path/to/file.json",
			expected: FormatJSON,
		},
		{
			name:     "YAML file .yaml",
			filename: "test.yaml",
			expected: FormatYAML,
		},
		{
			name:     "YAML file .yml",
			filename: "test.yml",
			expected: FormatYAML,
		},
		{
			name:     "TOML file",
			filename: "test.toml",
			expected: FormatTOML,
		},
		{
			name:     "Unknown extension defaults to YAML",
			filename: "test.txt",
			expected: FormatYAML,
		},
		{
			name:     "No extension defaults to YAML",
			filename: "test",
			expected: FormatYAML,
		},
		{
			name:     "Case insensitive",
			filename: "test.JSON",
			expected: FormatJSON,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetectFormatFromFilename(tt.filename)
			if result != tt.expected {
				t.Errorf("DetectFormatFromFilename() = %v, want %v", result, tt.expected)
			}
		})
	}
}


func TestJSONParser_Parse(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []*StructuredData
		wantErr bool
	}{
		{
			name:  "Simple JSON object",
			input: `{"name": "John", "age": 30}`,
			want: []*StructuredData{
				{
					Type: TypeObject,
					Children: map[string]*StructuredData{
						"name": {
							Type:  TypeString,
							Value: "John",
							Meta:  &Metadata{Format: FormatJSON},
						},
						"age": {
							Type:  TypeNumber,
							Value: float64(30),
							Meta:  &Metadata{Format: FormatJSON},
						},
					},
					Meta: &Metadata{Format: FormatJSON},
				},
			},
		},
		{
			name:  "JSON array",
			input: `["apple", "banana", "cherry"]`,
			want: []*StructuredData{
				{
					Type: TypeArray,
					Elements: []*StructuredData{
						{Type: TypeString, Value: "apple", Meta: &Metadata{Format: FormatJSON}},
						{Type: TypeString, Value: "banana", Meta: &Metadata{Format: FormatJSON}},
						{Type: TypeString, Value: "cherry", Meta: &Metadata{Format: FormatJSON}},
					},
					Meta: &Metadata{Format: FormatJSON},
				},
			},
		},
		{
			name:  "Multiple JSON documents",
			input: `{"a": 1}` + "\n" + `{"b": 2}`,
			want: []*StructuredData{
				{
					Type: TypeObject,
					Children: map[string]*StructuredData{
						"a": {Type: TypeNumber, Value: float64(1), Meta: &Metadata{Format: FormatJSON}},
					},
					Meta: &Metadata{Format: FormatJSON},
				},
				{
					Type: TypeObject,
					Children: map[string]*StructuredData{
						"b": {Type: TypeNumber, Value: float64(2), Meta: &Metadata{Format: FormatJSON}},
					},
					Meta: &Metadata{Format: FormatJSON},
				},
			},
		},
		{
			name:  "Null value",
			input: `{"value": null}`,
			want: []*StructuredData{
				{
					Type: TypeObject,
					Children: map[string]*StructuredData{
						"value": {Type: TypeNull, Meta: &Metadata{Format: FormatJSON}},
					},
					Meta: &Metadata{Format: FormatJSON},
				},
			},
		},
		{
			name:  "Boolean values",
			input: `{"active": true, "deleted": false}`,
			want: []*StructuredData{
				{
					Type: TypeObject,
					Children: map[string]*StructuredData{
						"active":  {Type: TypeBool, Value: true, Meta: &Metadata{Format: FormatJSON}},
						"deleted": {Type: TypeBool, Value: false, Meta: &Metadata{Format: FormatJSON}},
					},
					Meta: &Metadata{Format: FormatJSON},
				},
			},
		},
		{
			name:    "Invalid JSON",
			input:   `{"invalid": `,
			wantErr: true,
		},
	}

	parser := &JSONParser{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parser.Parse(strings.NewReader(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("JSONParser.Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !equalStructuredDataSlice(got, tt.want) {
				t.Errorf("JSONParser.Parse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestYAMLParser_Parse(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []*StructuredData
		wantErr bool
	}{
		{
			name: "Simple YAML object",
			input: `name: John
age: 30`,
			want: []*StructuredData{
				{
					Type: TypeObject,
					Children: map[string]*StructuredData{
						"name": {Type: TypeString, Value: "John", Meta: &Metadata{Format: FormatYAML}},
						"age":  {Type: TypeNumber, Value: 30, Meta: &Metadata{Format: FormatYAML}},
					},
					Meta: &Metadata{Format: FormatYAML},
				},
			},
		},
		{
			name: "YAML array",
			input: `fruits:
  - apple
  - banana
  - cherry`,
			want: []*StructuredData{
				{
					Type: TypeObject,
					Children: map[string]*StructuredData{
						"fruits": {
							Type: TypeArray,
							Elements: []*StructuredData{
								{Type: TypeString, Value: "apple", Meta: &Metadata{Format: FormatYAML}},
								{Type: TypeString, Value: "banana", Meta: &Metadata{Format: FormatYAML}},
								{Type: TypeString, Value: "cherry", Meta: &Metadata{Format: FormatYAML}},
							},
							Meta: &Metadata{Format: FormatYAML},
						},
					},
					Meta: &Metadata{Format: FormatYAML},
				},
			},
		},
		{
			name: "Multiple YAML documents",
			input: `doc: 1
---
doc: 2`,
			want: []*StructuredData{
				{
					Type: TypeObject,
					Children: map[string]*StructuredData{
						"doc": {Type: TypeNumber, Value: 1, Meta: &Metadata{Format: FormatYAML}},
					},
					Meta: &Metadata{Format: FormatYAML},
				},
				{
					Type: TypeObject,
					Children: map[string]*StructuredData{
						"doc": {Type: TypeNumber, Value: 2, Meta: &Metadata{Format: FormatYAML}},
					},
					Meta: &Metadata{Format: FormatYAML},
				},
			},
		},
		{
			name:    "Invalid YAML",
			input:   `invalid: [`,
			wantErr: true,
		},
	}

	parser := &YAMLParser{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parser.Parse(strings.NewReader(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("YAMLParser.Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if len(got) != len(tt.want) {
					t.Errorf("YAMLParser.Parse() returned %d items, want %d", len(got), len(tt.want))
				} else {
					for i := range got {
						if !equalStructuredData(got[i], tt.want[i]) {
							t.Logf("Debug: got[%d].Type=%v, want[%d].Type=%v", i, got[i].Type, i, tt.want[i].Type)
							if got[i].Type == TypeObject && tt.want[i].Type == TypeObject {
								t.Logf("Debug: got children keys: %v", getKeys(got[i].Children))
								t.Logf("Debug: want children keys: %v", getKeys(tt.want[i].Children))
								for k := range got[i].Children {
									gotChild := got[i].Children[k]
									wantChild := tt.want[i].Children[k]
									t.Logf("Debug: key=%s", k)
									t.Logf("  got: Type=%v, Value=%v (%T)", gotChild.Type, gotChild.Value, gotChild.Value)
									t.Logf("  want: Type=%v, Value=%v (%T)", wantChild.Type, wantChild.Value, wantChild.Value)
									if gotChild.Type == TypeNumber && wantChild.Type == TypeNumber {
										t.Logf("  toFloat64(got)=%v, toFloat64(want)=%v", toFloat64(gotChild.Value), toFloat64(wantChild.Value))
									}
									t.Logf("  equal=%v", equalStructuredData(gotChild, wantChild))
								}
							}
							t.Errorf("YAMLParser.Parse() item[%d] differs", i)
						}
					}
				}
			}
		})
	}
}

func TestYAMLParser_MultilineStringFormats(t *testing.T) {
	// Test for issue similar to https://github.com/sters/yaml-diff/issues/29
	tests := []struct {
		name   string
		yaml1  string
		yaml2  string
		expect string // "same" or "different"
	}{
		{
			name: "Multiline strings with different syntax should be same",
			yaml1: `value: |-
  foo
  bar
  baz
  special
    multiline`,
			yaml2: `value: "foo\nbar\nbaz\nspecial\n  multiline"`,
			expect: "same",
		},
		{
			name: "Multiline strings with backslash continuation",
			yaml1: `value: |-
  foo
  bar
  baz
  special
    multiline`,
			yaml2: `value: "foo\nbar\nbaz\n\
special\n\
\  multiline"`,
			expect: "same",
		},
		{
			name: "Different multiline content should be different",
			yaml1: `value: |-
  foo
  bar`,
			yaml2: `value: |-
  foo
  baz`,
			expect: "different",
		},
		{
			name: "Complex multiline with quotes and special chars",
			yaml1: `description: |-
  This is a "test" string
  with 'quotes' and special chars: !@#$%
  and multiple
    indented
      lines`,
			yaml2: `description: "This is a \"test\" string\nwith 'quotes' and special chars: !@#$%\nand multiple\n  indented\n    lines"`,
			expect: "same",
		},
	}

	parser := &YAMLParser{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse both YAML strings
			docs1, err1 := parser.Parse(strings.NewReader(tt.yaml1))
			if err1 != nil {
				t.Fatalf("Failed to parse yaml1: %v", err1)
			}
			
			docs2, err2 := parser.Parse(strings.NewReader(tt.yaml2))
			if err2 != nil {
				t.Fatalf("Failed to parse yaml2: %v", err2)
			}
			
			// Both should parse to single documents
			if len(docs1) != 1 || len(docs2) != 1 {
				t.Fatalf("Expected single document, got %d and %d", len(docs1), len(docs2))
			}
			
			// Compare the parsed values
			engine := &DiffEngine{options: DiffOptions{}}
			result := engine.Compare(docs1[0], docs2[0])
			
			if tt.expect == "same" && result.Status != StatusSame {
				// Log the actual values for debugging
				if docs1[0].Type == TypeObject && docs2[0].Type == TypeObject {
					val1 := docs1[0].Children["value"]
					val2 := docs2[0].Children["value"]
					if val1 != nil && val2 != nil {
						t.Logf("yaml1 value: %q (type: %T)", val1.Value, val1.Value)
						t.Logf("yaml2 value: %q (type: %T)", val2.Value, val2.Value)
					}
				}
				t.Errorf("Expected same but got different. Status: %v", result.Status)
			} else if tt.expect == "different" && result.Status == StatusSame {
				t.Errorf("Expected different but got same")
			}
		})
	}
}

func TestParseWithFormat(t *testing.T) {
	tests := []struct {
		name    string
		content string
		format  string
		wantLen int
		wantErr bool
	}{
		{
			name:    "JSON with JSON format",
			content: `{"test": "value"}`,
			format:  FormatJSON,
			wantLen: 1,
		},
		{
			name:    "YAML with YAML format",
			content: `test: value`,
			format:  FormatYAML,
			wantLen: 1,
		},
		{
			name:    "JSON with YAML format",
			content: `{"test": "value"}`,
			format:  FormatYAML,
			wantLen: 1, // YAML can parse JSON
		},
		{
			name:    "YAML with JSON format",
			content: `test: value`,
			format:  FormatJSON,
			wantErr: true, // JSON cannot parse YAML
		},
		{
			name:    "TOML format (not implemented)",
			content: `test = "value"`,
			format:  FormatTOML,
			wantErr: true,
		},
		{
			name:    "Unknown format",
			content: `test`,
			format:  "unknown",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseWithFormat(strings.NewReader(tt.content), tt.format)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseWithFormat() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(got) != tt.wantLen {
				t.Errorf("ParseWithFormat() returned %d documents, want %d", len(got), tt.wantLen)
			}
		})
	}
}


// Helper function to compare StructuredData slices
func equalStructuredDataSlice(a, b []*StructuredData) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !equalStructuredData(a[i], b[i]) {
			return false
		}
	}
	return true
}

// Helper function to compare StructuredData
func equalStructuredData(a, b *StructuredData) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	if a.Type != b.Type {
		return false
	}
	
	// Handle integer/float comparison for numbers
	if a.Type == TypeNumber && b.Type == TypeNumber {
		// Convert both to float64 for comparison
		aVal := toFloat64(a.Value)
		bVal := toFloat64(b.Value)
		if aVal != bVal {
			return false
		}
	} else if !reflect.DeepEqual(a.Value, b.Value) {
		return false
	}
	
	// Compare Children
	if len(a.Children) != len(b.Children) {
		return false
	}
	for k, v := range a.Children {
		if bv, ok := b.Children[k]; !ok || !equalStructuredData(v, bv) {
			return false
		}
	}
	
	// Compare Elements
	if len(a.Elements) != len(b.Elements) {
		return false
	}
	for i := range a.Elements {
		if !equalStructuredData(a.Elements[i], b.Elements[i]) {
			return false
		}
	}
	
	// Note: We intentionally don't compare Meta field as it contains pointers
	// and format-specific information that might differ between instances
	
	return true
}


func getKeys(m map[string]*StructuredData) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}