package diffnest

import (
	"strings"
	"testing"
)

func TestDiffEngine_Compare(t *testing.T) {
	tests := []struct {
		name   string
		a      *StructuredData
		b      *StructuredData
		opts   DiffOptions
		status DiffStatus
	}{
		{
			name:   "Same null values",
			a:      &StructuredData{Type: TypeNull},
			b:      &StructuredData{Type: TypeNull},
			status: StatusSame,
		},
		{
			name:   "Same string values",
			a:      &StructuredData{Type: TypeString, Value: "hello"},
			b:      &StructuredData{Type: TypeString, Value: "hello"},
			status: StatusSame,
		},
		{
			name:   "Different string values",
			a:      &StructuredData{Type: TypeString, Value: "hello"},
			b:      &StructuredData{Type: TypeString, Value: "world"},
			status: StatusModified,
		},
		{
			name:   "Same number values",
			a:      &StructuredData{Type: TypeNumber, Value: 42},
			b:      &StructuredData{Type: TypeNumber, Value: 42.0},
			status: StatusSame,
		},
		{
			name:   "Different number values",
			a:      &StructuredData{Type: TypeNumber, Value: 42},
			b:      &StructuredData{Type: TypeNumber, Value: 43},
			status: StatusModified,
		},
		{
			name:   "Same boolean values",
			a:      &StructuredData{Type: TypeBool, Value: true},
			b:      &StructuredData{Type: TypeBool, Value: true},
			status: StatusSame,
		},
		{
			name:   "Different boolean values",
			a:      &StructuredData{Type: TypeBool, Value: true},
			b:      &StructuredData{Type: TypeBool, Value: false},
			status: StatusModified,
		},
		{
			name:   "Type mismatch",
			a:      &StructuredData{Type: TypeString, Value: "42"},
			b:      &StructuredData{Type: TypeNumber, Value: 42},
			status: StatusModified,
		},
		{
			name:   "Added value",
			a:      nil,
			b:      &StructuredData{Type: TypeString, Value: "new"},
			status: StatusAdded,
		},
		{
			name:   "Deleted value",
			a:      &StructuredData{Type: TypeString, Value: "old"},
			b:      nil,
			status: StatusDeleted,
		},
		{
			name:   "Both nil",
			a:      nil,
			b:      nil,
			status: StatusSame,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewDiffEngine(tt.opts)
			result := engine.Compare(tt.a, tt.b)
			if result.Status != tt.status {
				t.Errorf("Compare() status = %v, want %v", result.Status, tt.status)
			}
		})
	}
}

func TestDiffEngine_CompareArrays(t *testing.T) {
	tests := []struct {
		name     string
		a        *StructuredData
		b        *StructuredData
		strategy ArrayDiffStrategy
		status   DiffStatus
	}{
		{
			name: "Same arrays by index",
			a: &StructuredData{
				Type: TypeArray,
				Elements: []*StructuredData{
					{Type: TypeString, Value: "a"},
					{Type: TypeString, Value: "b"},
				},
			},
			b: &StructuredData{
				Type: TypeArray,
				Elements: []*StructuredData{
					{Type: TypeString, Value: "a"},
					{Type: TypeString, Value: "b"},
				},
			},
			strategy: ArrayStrategyIndex,
			status:   StatusSame,
		},
		{
			name: "Different arrays by index",
			a: &StructuredData{
				Type: TypeArray,
				Elements: []*StructuredData{
					{Type: TypeString, Value: "a"},
					{Type: TypeString, Value: "b"},
				},
			},
			b: &StructuredData{
				Type: TypeArray,
				Elements: []*StructuredData{
					{Type: TypeString, Value: "a"},
					{Type: TypeString, Value: "c"},
				},
			},
			strategy: ArrayStrategyIndex,
			status:   StatusModified,
		},
		{
			name: "Arrays with different lengths",
			a: &StructuredData{
				Type: TypeArray,
				Elements: []*StructuredData{
					{Type: TypeString, Value: "a"},
				},
			},
			b: &StructuredData{
				Type: TypeArray,
				Elements: []*StructuredData{
					{Type: TypeString, Value: "a"},
					{Type: TypeString, Value: "b"},
				},
			},
			strategy: ArrayStrategyIndex,
			status:   StatusModified,
		},
		{
			name: "Same arrays by value (reordered)",
			a: &StructuredData{
				Type: TypeArray,
				Elements: []*StructuredData{
					{Type: TypeString, Value: "a"},
					{Type: TypeString, Value: "b"},
					{Type: TypeString, Value: "c"},
				},
			},
			b: &StructuredData{
				Type: TypeArray,
				Elements: []*StructuredData{
					{Type: TypeString, Value: "c"},
					{Type: TypeString, Value: "a"},
					{Type: TypeString, Value: "b"},
				},
			},
			strategy: ArrayStrategyValue,
			status:   StatusSame,
		},
		{
			name: "Different arrays by value",
			a: &StructuredData{
				Type: TypeArray,
				Elements: []*StructuredData{
					{Type: TypeString, Value: "a"},
					{Type: TypeString, Value: "b"},
				},
			},
			b: &StructuredData{
				Type: TypeArray,
				Elements: []*StructuredData{
					{Type: TypeString, Value: "a"},
					{Type: TypeString, Value: "c"},
				},
			},
			strategy: ArrayStrategyValue,
			status:   StatusModified,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewDiffEngine(DiffOptions{ArrayDiffStrategy: tt.strategy})
			result := engine.Compare(tt.a, tt.b)
			if result.Status != tt.status {
				t.Errorf("Compare() status = %v, want %v", result.Status, tt.status)
			}
		})
	}
}

func TestDiffEngine_CompareObjects(t *testing.T) {
	tests := []struct {
		name   string
		a      *StructuredData
		b      *StructuredData
		opts   DiffOptions
		status DiffStatus
	}{
		{
			name: "Same objects",
			a: &StructuredData{
				Type: TypeObject,
				Children: map[string]*StructuredData{
					"name": {Type: TypeString, Value: "John"},
					"age":  {Type: TypeNumber, Value: 30},
				},
			},
			b: &StructuredData{
				Type: TypeObject,
				Children: map[string]*StructuredData{
					"name": {Type: TypeString, Value: "John"},
					"age":  {Type: TypeNumber, Value: 30},
				},
			},
			status: StatusSame,
		},
		{
			name: "Different objects",
			a: &StructuredData{
				Type: TypeObject,
				Children: map[string]*StructuredData{
					"name": {Type: TypeString, Value: "John"},
					"age":  {Type: TypeNumber, Value: 30},
				},
			},
			b: &StructuredData{
				Type: TypeObject,
				Children: map[string]*StructuredData{
					"name": {Type: TypeString, Value: "Jane"},
					"age":  {Type: TypeNumber, Value: 30},
				},
			},
			status: StatusModified,
		},
		{
			name: "Object with added field",
			a: &StructuredData{
				Type: TypeObject,
				Children: map[string]*StructuredData{
					"name": {Type: TypeString, Value: "John"},
				},
			},
			b: &StructuredData{
				Type: TypeObject,
				Children: map[string]*StructuredData{
					"name": {Type: TypeString, Value: "John"},
					"age":  {Type: TypeNumber, Value: 30},
				},
			},
			status: StatusModified,
		},
		{
			name: "Object with deleted field",
			a: &StructuredData{
				Type: TypeObject,
				Children: map[string]*StructuredData{
					"name": {Type: TypeString, Value: "John"},
					"age":  {Type: TypeNumber, Value: 30},
				},
			},
			b: &StructuredData{
				Type: TypeObject,
				Children: map[string]*StructuredData{
					"name": {Type: TypeString, Value: "John"},
				},
			},
			status: StatusModified,
		},
		{
			name: "Ignore zero values",
			a: &StructuredData{
				Type: TypeObject,
				Children: map[string]*StructuredData{
					"name":   {Type: TypeString, Value: "John"},
					"active": {Type: TypeBool, Value: false},
				},
			},
			b: &StructuredData{
				Type: TypeObject,
				Children: map[string]*StructuredData{
					"name": {Type: TypeString, Value: "John"},
				},
			},
			opts:   DiffOptions{IgnoreZeroValues: true},
			status: StatusSame,
		},
		{
			name: "Don't ignore non-empty fields",
			a: &StructuredData{
				Type: TypeObject,
				Children: map[string]*StructuredData{
					"name": {Type: TypeString, Value: "John"},
				},
			},
			b: &StructuredData{
				Type: TypeObject,
				Children: map[string]*StructuredData{
					"name":  {Type: TypeString, Value: "John"},
					"email": {Type: TypeString, Value: "john@example.com"},
				},
			},
			opts:   DiffOptions{IgnoreEmptyFields: false},
			status: StatusModified,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewDiffEngine(tt.opts)
			result := engine.Compare(tt.a, tt.b)
			if result.Status != tt.status {
				t.Errorf("Compare() status = %v, want %v", result.Status, tt.status)
			}
		})
	}
}

func TestCompare_MultipleDocuments(t *testing.T) {
	doc1a := &StructuredData{
		Type: TypeObject,
		Children: map[string]*StructuredData{
			"id": {Type: TypeNumber, Value: 1},
		},
		Meta: &Metadata{},
	}
	doc1b := &StructuredData{
		Type: TypeObject,
		Children: map[string]*StructuredData{
			"id": {Type: TypeNumber, Value: 1},
		},
		Meta: &Metadata{},
	}
	doc2a := &StructuredData{
		Type: TypeObject,
		Children: map[string]*StructuredData{
			"id": {Type: TypeNumber, Value: 2},
		},
		Meta: &Metadata{},
	}
	doc2b := &StructuredData{
		Type: TypeObject,
		Children: map[string]*StructuredData{
			"id": {Type: TypeNumber, Value: 2},
		},
		Meta: &Metadata{},
	}

	tests := []struct {
		name    string
		docsA   []*StructuredData
		docsB   []*StructuredData
		wantLen int
	}{
		{
			name:    "Single document each",
			docsA:   []*StructuredData{doc1a},
			docsB:   []*StructuredData{doc1b},
			wantLen: 1,
		},
		{
			name:    "Multiple documents with optimal pairing",
			docsA:   []*StructuredData{doc1a, doc2a},
			docsB:   []*StructuredData{doc2b, doc1b},
			wantLen: 2,
		},
		{
			name:    "Different number of documents",
			docsA:   []*StructuredData{doc1a},
			docsB:   []*StructuredData{doc1b, doc2b},
			wantLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := Compare(tt.docsA, tt.docsB, DiffOptions{})
			if len(results) != tt.wantLen {
				t.Errorf("Compare() returned %d results, want %d", len(results), tt.wantLen)
			}
		})
	}
}

func TestDiffEngine_calculateSize(t *testing.T) {
	tests := []struct {
		name string
		data *StructuredData
		want int
	}{
		{
			name: "Nil data",
			data: nil,
			want: 0,
		},
		{
			name: "Primitive value",
			data: &StructuredData{Type: TypeString, Value: "test"},
			want: 1,
		},
		{
			name: "Empty array",
			data: &StructuredData{Type: TypeArray, Elements: []*StructuredData{}},
			want: 0,
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
			want: 2,
		},
		{
			name: "Empty object",
			data: &StructuredData{Type: TypeObject, Children: map[string]*StructuredData{}},
			want: 0,
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
			want: 2,
		},
		{
			name: "Nested structure",
			data: &StructuredData{
				Type: TypeObject,
				Children: map[string]*StructuredData{
					"arr": {
						Type: TypeArray,
						Elements: []*StructuredData{
							{Type: TypeString, Value: "a"},
							{Type: TypeString, Value: "b"},
						},
					},
					"val": {Type: TypeNumber, Value: 42},
				},
			},
			want: 3,
		},
	}

	engine := NewDiffEngine(DiffOptions{})
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := engine.calculateSize(tt.data)
			if got != tt.want {
				t.Errorf("calculateSize() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDiffEngine_MultilineStringComparison(t *testing.T) {
	// Test for issue https://github.com/sters/yaml-diff/issues/52
	// Currently, multiline strings are compared as atomic values
	tests := []struct {
		name       string
		string1    string
		string2    string
		wantStatus DiffStatus
	}{
		{
			name:       "Identical multiline strings",
			string1:    "line1\nline2\nline3",
			string2:    "line1\nline2\nline3",
			wantStatus: StatusSame,
		},
		{
			name:       "Different multiline strings",
			string1:    "logging.a: false\nlogging.b: false",
			string2:    "logging.a: false\nlogging.c: false",
			wantStatus: StatusModified,
		},
		{
			name:       "One line difference in multiline",
			string1:    "line1\nline2\nline3",
			string2:    "line1\nline2-modified\nline3",
			wantStatus: StatusModified,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := &DiffEngine{}
			
			data1 := &StructuredData{
				Type: TypeObject,
				Children: map[string]*StructuredData{
					"config": {
						Type:  TypeString,
						Value: tt.string1,
						Meta:  &Metadata{Format: "yaml"},
					},
				},
				Meta: &Metadata{Format: "yaml"},
			}
			
			data2 := &StructuredData{
				Type: TypeObject,
				Children: map[string]*StructuredData{
					"config": {
						Type:  TypeString,
						Value: tt.string2,
						Meta:  &Metadata{Format: "yaml"},
					},
				},
				Meta: &Metadata{Format: "yaml"},
			}
			
			result := engine.Compare(data1, data2)
			
			if result.Status != tt.wantStatus {
				t.Errorf("Compare() status = %v, want %v", result.Status, tt.wantStatus)
			}
			
			// When strings are different, check if multiline diff is applied
			if tt.wantStatus == StatusModified {
				if len(result.Children) != 1 {
					t.Errorf("Expected 1 child diff for object comparison, got %d", len(result.Children))
				} else {
					configDiff := result.Children[0]
					if configDiff.Status != StatusModified {
						t.Errorf("Expected config field to be modified, got %v", configDiff.Status)
					}
					// For multiline strings, we now do line-by-line diff
					// so we should have children representing line differences
					if strings.Contains(tt.string1, "\n") || strings.Contains(tt.string2, "\n") {
						if len(configDiff.Children) == 0 {
							t.Errorf("Expected line-by-line diff for multiline strings")
						}
					}
				}
			}
		})
	}
}
