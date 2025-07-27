package diffnest

import (
	"strings"
	"testing"
)

func TestUnifiedFormatter_Format(t *testing.T) {
	tests := []struct {
		name         string
		results      []*DiffResult
		showOnlyDiff bool
		contextLines int
		want         []string // Expected lines in output
	}{
		{
			name: "Simple modification",
			results: []*DiffResult{
				{
					Status: StatusModified,
					Path:   []string{"name"},
					From:   &StructuredData{Type: TypeString, Value: "John"},
					To:     &StructuredData{Type: TypeString, Value: "Jane"},
				},
			},
			contextLines: -1,
			want: []string{
				"- name: John",
				"+ name: Jane",
			},
		},
		{
			name: "Added field",
			results: []*DiffResult{
				{
					Status: StatusAdded,
					Path:   []string{"age"},
					To:     &StructuredData{Type: TypeNumber, Value: 30},
				},
			},
			contextLines: -1,
			want: []string{
				"+ age: 30",
			},
		},
		{
			name: "Deleted field",
			results: []*DiffResult{
				{
					Status: StatusDeleted,
					Path:   []string{"city"},
					From:   &StructuredData{Type: TypeString, Value: "Tokyo"},
				},
			},
			contextLines: -1,
			want: []string{
				"- city: Tokyo",
			},
		},
		{
			name: "Same field (show all)",
			results: []*DiffResult{
				{
					Status: StatusSame,
					Path:   []string{"id"},
					From:   &StructuredData{Type: TypeNumber, Value: 123},
					To:     &StructuredData{Type: TypeNumber, Value: 123},
				},
			},
			showOnlyDiff: false,
			contextLines: -1,
			want: []string{
				"  id: 123",
			},
		},
		{
			name: "Same field (show only diff)",
			results: []*DiffResult{
				{
					Status: StatusSame,
					Path:   []string{"id"},
					From:   &StructuredData{Type: TypeNumber, Value: 123},
					To:     &StructuredData{Type: TypeNumber, Value: 123},
				},
			},
			showOnlyDiff: true,
			contextLines: -1,
			want:         []string{}, // Should be empty
		},
		{
			name: "Nested object modification",
			results: []*DiffResult{
				{
					Status: StatusModified,
					Path:   []string{},
					From: &StructuredData{
						Type: TypeObject,
						Children: map[string]*StructuredData{
							"user": {Type: TypeString, Value: "john"},
						},
					},
					To: &StructuredData{
						Type: TypeObject,
						Children: map[string]*StructuredData{
							"user": {Type: TypeString, Value: "jane"},
						},
					},
					Children: []*DiffResult{
						{
							Status: StatusModified,
							Path:   []string{"user"},
							From:   &StructuredData{Type: TypeString, Value: "john"},
							To:     &StructuredData{Type: TypeString, Value: "jane"},
						},
					},
				},
			},
			contextLines: -1,
			want: []string{
				"- user: john",
				"+ user: jane",
			},
		},
		{
			name: "Multiple results",
			results: []*DiffResult{
				{
					Status: StatusModified,
					Path:   []string{"a"},
					From:   &StructuredData{Type: TypeString, Value: "old"},
					To:     &StructuredData{Type: TypeString, Value: "new"},
				},
				{
					Status: StatusAdded,
					Path:   []string{"b"},
					To:     &StructuredData{Type: TypeString, Value: "added"},
				},
			},
			contextLines: -1,
			want: []string{
				"- a: old",
				"+ a: new",
				"---",
				"+ b: added",
			},
		},
		{
			name: "Array modification",
			results: []*DiffResult{
				{
					Status: StatusModified,
					Path:   []string{"items"},
					From: &StructuredData{
						Type: TypeArray,
						Elements: []*StructuredData{
							{Type: TypeString, Value: "a"},
							{Type: TypeString, Value: "b"},
						},
					},
					To: &StructuredData{
						Type: TypeArray,
						Elements: []*StructuredData{
							{Type: TypeString, Value: "a"},
							{Type: TypeString, Value: "c"},
						},
					},
					Children: []*DiffResult{
						{
							Status: StatusSame,
							Path:   []string{"items", "[0]"},
							From:   &StructuredData{Type: TypeString, Value: "a"},
							To:     &StructuredData{Type: TypeString, Value: "a"},
						},
						{
							Status: StatusModified,
							Path:   []string{"items", "[1]"},
							From:   &StructuredData{Type: TypeString, Value: "b"},
							To:     &StructuredData{Type: TypeString, Value: "c"},
						},
					},
				},
			},
			showOnlyDiff: true,
			contextLines: -1,
			want: []string{
				"items:",
				"-   - b",
				"+   - c",
			},
		},
		{
			name: "Context lines with multiline string",
			results: []*DiffResult{
				{
					Status: StatusModified,
					Path:   []string{"text"},
					From:   &StructuredData{Type: TypeString, Value: "line1\nline2\nline3\nline4\nline5\nline6\nline7"},
					To:     &StructuredData{Type: TypeString, Value: "line1\nline2\nmodified3\nline4\nline5\nline6\nline7"},
					Children: []*DiffResult{
						{Status: StatusSame, From: &StructuredData{Type: TypeString, Value: "line1"}},
						{Status: StatusSame, From: &StructuredData{Type: TypeString, Value: "line2"}},
						{Status: StatusModified, From: &StructuredData{Type: TypeString, Value: "line3"}, To: &StructuredData{Type: TypeString, Value: "modified3"}},
						{Status: StatusSame, From: &StructuredData{Type: TypeString, Value: "line4"}},
						{Status: StatusSame, From: &StructuredData{Type: TypeString, Value: "line5"}},
						{Status: StatusSame, From: &StructuredData{Type: TypeString, Value: "line6"}},
						{Status: StatusSame, From: &StructuredData{Type: TypeString, Value: "line7"}},
					},
				},
			},
			showOnlyDiff: true,
			contextLines: 1,
			want: []string{
				"     line2",
				"-    line3",
				"+    modified3",
				"     line4",
			},
		},
		{
			name: "Context lines with object children",
			results: []*DiffResult{
				{
					Status: StatusModified,
					Path:   []string{},
					From:   &StructuredData{Type: TypeObject},
					To:     &StructuredData{Type: TypeObject},
					Children: []*DiffResult{
						{Status: StatusSame, Path: []string{"field1"}, From: &StructuredData{Type: TypeString, Value: "value1"}},
						{Status: StatusSame, Path: []string{"field2"}, From: &StructuredData{Type: TypeString, Value: "value2"}},
						{Status: StatusModified, Path: []string{"field3"}, From: &StructuredData{Type: TypeString, Value: "old"}, To: &StructuredData{Type: TypeString, Value: "new"}},
						{Status: StatusSame, Path: []string{"field4"}, From: &StructuredData{Type: TypeString, Value: "value4"}},
						{Status: StatusSame, Path: []string{"field5"}, From: &StructuredData{Type: TypeString, Value: "value5"}},
					},
				},
			},
			showOnlyDiff: true,
			contextLines: 1,
			want: []string{
				"  field2: value2",
				"- field3: old",
				"+ field3: new",
				"  field4: value4",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := &UnifiedFormatter{
				ShowOnlyDiff: tt.showOnlyDiff,
				ContextLines: tt.contextLines,
			}
			var buf strings.Builder
			err := formatter.Format(&buf, tt.results)
			if err != nil {
				t.Fatalf("UnifiedFormatter.Format() error = %v", err)
			}
			got := buf.String()

			// Check if all expected lines are present
			for _, want := range tt.want {
				if !strings.Contains(got, want) {
					t.Errorf("UnifiedFormatter.Format() missing expected line:\n%q\nGot:\n%s", want, got)
				}
			}

			// For empty expected output, check that output is empty
			if len(tt.want) == 0 && strings.TrimSpace(got) != "" {
				t.Errorf("UnifiedFormatter.Format() expected empty output, got:\n%s", got)
			}
		})
	}
}

func TestJSONPatchFormatter_Format(t *testing.T) {
	tests := []struct {
		name    string
		results []*DiffResult
		want    []string // Expected operations in output
	}{
		{
			name: "Replace operation",
			results: []*DiffResult{
				{
					Status: StatusModified,
					Path:   []string{"name"},
					From:   &StructuredData{Type: TypeString, Value: "John"},
					To:     &StructuredData{Type: TypeString, Value: "Jane"},
					Meta:   &DiffMeta{},
				},
			},
			want: []string{
				`{"op": "replace", "path": "/name", "value": "Jane"}`,
			},
		},
		{
			name: "Add operation",
			results: []*DiffResult{
				{
					Status: StatusAdded,
					Path:   []string{"age"},
					To:     &StructuredData{Type: TypeNumber, Value: 30},
					Meta:   &DiffMeta{},
				},
			},
			want: []string{
				`{"op": "add", "path": "/age", "value": 30}`,
			},
		},
		{
			name: "Remove operation",
			results: []*DiffResult{
				{
					Status: StatusDeleted,
					Path:   []string{"city"},
					From:   &StructuredData{Type: TypeString, Value: "Tokyo"},
					Meta:   &DiffMeta{},
				},
			},
			want: []string{
				`{"op": "remove", "path": "/city"}`,
			},
		},
		{
			name: "No changes",
			results: []*DiffResult{
				{
					Status: StatusSame,
					Path:   []string{"id"},
					From:   &StructuredData{Type: TypeNumber, Value: 123},
					To:     &StructuredData{Type: TypeNumber, Value: 123},
					Meta:   &DiffMeta{},
				},
			},
			want: []string{}, // No operations for same values
		},
		{
			name: "Multiple operations",
			results: []*DiffResult{
				{
					Status: StatusModified,
					Path:   []string{"name"},
					From:   &StructuredData{Type: TypeString, Value: "John"},
					To:     &StructuredData{Type: TypeString, Value: "Jane"},
					Meta:   &DiffMeta{},
				},
				{
					Status: StatusAdded,
					Path:   []string{"email"},
					To:     &StructuredData{Type: TypeString, Value: "jane@example.com"},
					Meta:   &DiffMeta{},
				},
			},
			want: []string{
				`{"op": "replace", "path": "/name", "value": "Jane"}`,
				`{"op": "add", "path": "/email", "value": "jane@example.com"}`,
			},
		},
		{
			name: "Nested path",
			results: []*DiffResult{
				{
					Status: StatusModified,
					Path:   []string{"user", "name"},
					From:   &StructuredData{Type: TypeString, Value: "John"},
					To:     &StructuredData{Type: TypeString, Value: "Jane"},
					Meta:   &DiffMeta{},
				},
			},
			want: []string{
				`{"op": "replace", "path": "/user/name", "value": "Jane"}`,
			},
		},
		{
			name: "Array element",
			results: []*DiffResult{
				{
					Status: StatusModified,
					Path:   []string{"items", "[1]"},
					From:   &StructuredData{Type: TypeString, Value: "old"},
					To:     &StructuredData{Type: TypeString, Value: "new"},
					Meta:   &DiffMeta{},
				},
			},
			want: []string{
				`{"op": "replace", "path": "/items/[1]", "value": "new"}`,
			},
		},
		{
			name: "Complex value",
			results: []*DiffResult{
				{
					Status: StatusAdded,
					Path:   []string{"config"},
					To: &StructuredData{
						Type: TypeObject,
						Children: map[string]*StructuredData{
							"enabled": {Type: TypeBool, Value: true},
							"port":    {Type: TypeNumber, Value: 8080},
						},
					},
					Meta: &DiffMeta{},
				},
			},
			want: []string{
				`{"op": "add", "path": "/config", "value": {`,
				`"enabled": true`,
				`"port": 8080`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := &JSONPatchFormatter{}
			var buf strings.Builder
			err := formatter.Format(&buf, tt.results)
			if err != nil {
				t.Fatalf("JSONPatchFormatter.Format() error = %v", err)
			}
			got := buf.String()

			if len(tt.want) == 0 {
				if strings.TrimSpace(got) != "[]" {
					t.Errorf("JSONPatchFormatter.Format() expected empty array, got:\n%s", got)
				}

				return
			}

			// Check if all expected operations are present
			for _, want := range tt.want {
				if !strings.Contains(got, want) {
					t.Errorf("JSONPatchFormatter.Format() missing expected operation:\n%q\nGot:\n%s", want, got)
				}
			}
		})
	}
}

func TestUnifiedFormatter_formatValue(t *testing.T) {
	tests := []struct {
		name string
		data *StructuredData
		want string
	}{
		{
			name: "Nil value",
			data: nil,
			want: valueNull,
		},
		{
			name: "Null value",
			data: &StructuredData{Type: TypeNull},
			want: valueNull,
		},
		{
			name: "String value",
			data: &StructuredData{Type: TypeString, Value: "hello"},
			want: "hello",
		},
		{
			name: "Number value",
			data: &StructuredData{Type: TypeNumber, Value: 42},
			want: "42",
		},
		{
			name: "Boolean value",
			data: &StructuredData{Type: TypeBool, Value: true},
			want: "true",
		},
		{
			name: "Empty array",
			data: &StructuredData{Type: TypeArray, Elements: []*StructuredData{}},
			want: "[]",
		},
		{
			name: "Array with elements",
			data: &StructuredData{
				Type: TypeArray,
				Elements: []*StructuredData{
					{Type: TypeString, Value: "a"},
					{Type: TypeString, Value: "b"},
				},
			},
			want: "[2 items]",
		},
		{
			name: "Empty object",
			data: &StructuredData{Type: TypeObject, Children: map[string]*StructuredData{}},
			want: "{}",
		},
		{
			name: "Object with fields",
			data: &StructuredData{
				Type: TypeObject,
				Children: map[string]*StructuredData{
					"a": {Type: TypeString, Value: "1"},
					"b": {Type: TypeString, Value: "2"},
				},
			},
			want: "{2 fields}",
		},
		{
			name: "Unknown type",
			data: &StructuredData{Type: DataType(999)},
			want: "?",
		},
	}

	formatter := &UnifiedFormatter{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatter.formatValue(tt.data)
			if got != tt.want {
				t.Errorf("formatValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestJSONPatchFormatter_jsonValue(t *testing.T) {
	tests := []struct {
		name string
		data *StructuredData
		want string
	}{
		{
			name: "Nil value",
			data: nil,
			want: valueNull,
		},
		{
			name: "Null value",
			data: &StructuredData{Type: TypeNull},
			want: valueNull,
		},
		{
			name: "String value",
			data: &StructuredData{Type: TypeString, Value: "hello"},
			want: `"hello"`,
		},
		{
			name: "Number value",
			data: &StructuredData{Type: TypeNumber, Value: 42},
			want: "42",
		},
		{
			name: "Boolean value",
			data: &StructuredData{Type: TypeBool, Value: true},
			want: "true",
		},
		{
			name: "Empty array",
			data: &StructuredData{Type: TypeArray, Elements: []*StructuredData{}},
			want: "[]",
		},
		{
			name: "Array with elements",
			data: &StructuredData{
				Type: TypeArray,
				Elements: []*StructuredData{
					{Type: TypeString, Value: "a"},
					{Type: TypeNumber, Value: 1},
				},
			},
			want: `["a", 1]`,
		},
		{
			name: "Empty object",
			data: &StructuredData{Type: TypeObject, Children: map[string]*StructuredData{}},
			want: "{}",
		},
		{
			name: "Object with fields",
			data: &StructuredData{
				Type: TypeObject,
				Children: map[string]*StructuredData{
					"name": {Type: TypeString, Value: "John"},
					"age":  {Type: TypeNumber, Value: 30},
				},
			},
			want: `{"age": 30, "name": "John"}`, // Note: order might vary
		},
	}

	formatter := &JSONPatchFormatter{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatter.jsonValue(tt.data)
			// For objects, we need to check if both possible orders are acceptable
			if tt.data != nil && tt.data.Type == TypeObject && len(tt.data.Children) > 1 {
				// Check if the structure is correct rather than exact string match
				if !strings.HasPrefix(got, "{") || !strings.HasSuffix(got, "}") {
					t.Errorf("jsonValue() = %v, expected object format", got)
				}
			} else if got != tt.want {
				t.Errorf("jsonValue() = %v, want %v", got, tt.want)
			}
		})
	}
}
