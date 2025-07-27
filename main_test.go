package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Test file generators
func createJSONDiffFiles(t *testing.T, tempDir string) (string, string) {
	json1 := filepath.Join(tempDir, "test1.json")
	json2 := filepath.Join(tempDir, "test2.json")
	
	if err := os.WriteFile(json1, []byte(`{
		"name": "John",
		"age": 30,
		"city": "Tokyo"
	}`), 0644); err != nil {
		t.Fatal(err)
	}
	
	if err := os.WriteFile(json2, []byte(`{
		"name": "John",
		"age": 31,
		"city": "Osaka"
	}`), 0644); err != nil {
		t.Fatal(err)
	}
	
	return json1, json2
}

func createYAMLDiffFiles(t *testing.T, tempDir string) (string, string) {
	yaml1 := filepath.Join(tempDir, "test1.yaml")
	yaml2 := filepath.Join(tempDir, "test2.yaml")
	
	if err := os.WriteFile(yaml1, []byte(`name: Alice
age: 25
hobbies:
  - reading
  - gaming`), 0644); err != nil {
		t.Fatal(err)
	}
	
	if err := os.WriteFile(yaml2, []byte(`name: Alice
age: 26
hobbies:
  - reading
  - swimming
  - gaming`), 0644); err != nil {
		t.Fatal(err)
	}
	
	return yaml1, yaml2
}

func createIdenticalJSONFiles(t *testing.T, tempDir string) (string, string) {
	same1 := filepath.Join(tempDir, "same1.json")
	same2 := filepath.Join(tempDir, "same2.json")
	
	sameContent := `{"id": 123, "status": "active"}`
	if err := os.WriteFile(same1, []byte(sameContent), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(same2, []byte(sameContent), 0644); err != nil {
		t.Fatal(err)
	}
	
	return same1, same2
}

func createMultipleJSONFiles(t *testing.T, tempDir string) (string, string) {
	multi1 := filepath.Join(tempDir, "multi1.json")
	multi2 := filepath.Join(tempDir, "multi2.json")
	
	// Multiple JSON documents separated by newlines
	content1 := `{"id": 1, "name": "Alice", "active": true}
{"id": 2, "name": "Bob", "active": false}
{"id": 3, "name": "Charlie", "active": true}`
	
	content2 := `{"id": 1, "name": "Alice", "active": false}
{"id": 2, "name": "Robert", "active": true}
{"id": 4, "name": "David", "active": true}`
	
	if err := os.WriteFile(multi1, []byte(content1), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(multi2, []byte(content2), 0644); err != nil {
		t.Fatal(err)
	}
	
	return multi1, multi2
}

func createMultipleYAMLFiles(t *testing.T, tempDir string) (string, string) {
	multi1 := filepath.Join(tempDir, "multi1.yaml")
	multi2 := filepath.Join(tempDir, "multi2.yaml")
	
	// Multiple YAML documents separated by ---
	content1 := `id: 1
name: Alice
department: Engineering
---
id: 2
name: Bob
department: Sales
---
id: 3
name: Charlie
department: Marketing`
	
	content2 := `id: 1
name: Alice
department: Marketing
---
id: 2
name: Bob
department: Engineering
---
id: 3
name: Charles
department: Marketing`
	
	if err := os.WriteFile(multi1, []byte(content1), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(multi2, []byte(content2), 0644); err != nil {
		t.Fatal(err)
	}
	
	return multi1, multi2
}

func createNestedMultipleJSONFiles(t *testing.T, tempDir string) (string, string) {
	nested1 := filepath.Join(tempDir, "nested1.json")
	nested2 := filepath.Join(tempDir, "nested2.json")
	
	// Multiple JSON documents with nested structures
	content1 := `{"user": {"id": 1, "profile": {"name": "Alice", "age": 30}}, "settings": {"theme": "dark"}}
{"user": {"id": 2, "profile": {"name": "Bob", "age": 25}}, "settings": {"theme": "light"}}`
	
	content2 := `{"user": {"id": 1, "profile": {"name": "Alice", "age": 31}}, "settings": {"theme": "light"}}
{"user": {"id": 2, "profile": {"name": "Bob", "age": 25}}, "settings": {"theme": "dark", "lang": "en"}}`
	
	if err := os.WriteFile(nested1, []byte(content1), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(nested2, []byte(content2), 0644); err != nil {
		t.Fatal(err)
	}
	
	return nested1, nested2
}

func createMixedMultipleYAMLFiles(t *testing.T, tempDir string) (string, string) {
	mixed1 := filepath.Join(tempDir, "mixed1.yaml")
	mixed2 := filepath.Join(tempDir, "mixed2.yaml")
	
	// Multiple YAML documents with mixed content
	content1 := `# Configuration file
config:
  database:
    host: localhost
    port: 5432
  cache:
    enabled: true
    ttl: 3600
---
# Service definition
service:
  name: api-gateway
  version: 1.2.3
  endpoints:
    - /users
    - /products`
	
	content2 := `# Configuration file
config:
  database:
    host: db.example.com
    port: 5432
  cache:
    enabled: false
    ttl: 7200
  monitoring:
    enabled: true
---
# Service definition
service:
  name: api-gateway
  version: 1.3.0
  endpoints:
    - /users
    - /products
    - /orders`
	
	if err := os.WriteFile(mixed1, []byte(content1), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(mixed2, []byte(content2), 0644); err != nil {
		t.Fatal(err)
	}
	
	return mixed1, mixed2
}

func createArrayDiffJSONFiles(t *testing.T, tempDir string) (string, string) {
	array1 := filepath.Join(tempDir, "array1.json")
	array2 := filepath.Join(tempDir, "array2.json")
	
	// JSON files with array differences
	content1 := `{
		"items": ["apple", "banana", "cherry"],
		"tags": ["fruit", "healthy", "organic"]
	}`
	
	content2 := `{
		"items": ["banana", "cherry", "date", "apple"],
		"tags": ["fruit", "organic", "fresh", "healthy"]
	}`
	
	if err := os.WriteFile(array1, []byte(content1), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(array2, []byte(content2), 0644); err != nil {
		t.Fatal(err)
	}
	
	return array1, array2
}

func createCrossFormatFiles(t *testing.T, tempDir string) (string, string) {
	json1 := filepath.Join(tempDir, "data.json")
	yaml1 := filepath.Join(tempDir, "data.yaml")
	
	// Same content in different formats
	jsonContent := `{
		"name": "Alice",
		"age": 25,
		"hobbies": ["reading", "gaming"],
		"active": true
	}`
	
	yamlContent := `name: Alice
age: 25
hobbies:
  - reading
  - gaming
active: true`
	
	if err := os.WriteFile(json1, []byte(jsonContent), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(yaml1, []byte(yamlContent), 0644); err != nil {
		t.Fatal(err)
	}
	
	return json1, yaml1
}

func createDifferentCrossFormatFiles(t *testing.T, tempDir string) (string, string) {
	json1 := filepath.Join(tempDir, "config.json")
	yaml1 := filepath.Join(tempDir, "config.yaml")
	
	// Different content in different formats
	jsonContent := `{
		"server": {
			"host": "localhost",
			"port": 8080
		},
		"debug": false
	}`
	
	yamlContent := `server:
  host: example.com
  port: 9000
debug: true
logging:
  level: info`
	
	if err := os.WriteFile(json1, []byte(jsonContent), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(yaml1, []byte(yamlContent), 0644); err != nil {
		t.Fatal(err)
	}
	
	return json1, yaml1
}

func createYAMLMultilineFiles(t *testing.T, tempDir string) (string, string) {
	yaml1 := filepath.Join(tempDir, "multiline1.yaml")
	yaml2 := filepath.Join(tempDir, "multiline2.yaml")
	
	// Test case from https://github.com/sters/yaml-diff/issues/29
	content1 := `value: |-
  foo
  bar
  baz
  special
    multiline`
	
	content2 := `value: "foo\nbar\nbaz\n\
special\n\
\  multiline"`
	
	if err := os.WriteFile(yaml1, []byte(content1), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(yaml2, []byte(content2), 0644); err != nil {
		t.Fatal(err)
	}
	
	return yaml1, yaml2
}

func createYAMLMultilineConfigFiles(t *testing.T, tempDir string) (string, string) {
	yaml1 := filepath.Join(tempDir, "config1.yaml")
	yaml2 := filepath.Join(tempDir, "config2.yaml")
	
	// Test case from https://github.com/sters/yaml-diff/issues/52
	content1 := `data:
  config: |
    logging.a: false
    logging.b: false`
	
	content2 := `data:
  config: |
    logging.a: false
    logging.c: false`
	
	if err := os.WriteFile(yaml1, []byte(content1), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(yaml2, []byte(content2), 0644); err != nil {
		t.Fatal(err)
	}
	
	return yaml1, yaml2
}

func TestIntegration(t *testing.T) {
	// Create temporary test files
	tempDir := t.TempDir()
	
	// Generate test files for each case
	json1, json2 := createJSONDiffFiles(t, tempDir)
	yaml1, yaml2 := createYAMLDiffFiles(t, tempDir)
	same1, same2 := createIdenticalJSONFiles(t, tempDir)
	multiJSON1, multiJSON2 := createMultipleJSONFiles(t, tempDir)
	multiYAML1, multiYAML2 := createMultipleYAMLFiles(t, tempDir)
	nestedJSON1, nestedJSON2 := createNestedMultipleJSONFiles(t, tempDir)
	mixedYAML1, mixedYAML2 := createMixedMultipleYAMLFiles(t, tempDir)
	arrayJSON1, arrayJSON2 := createArrayDiffJSONFiles(t, tempDir)
	crossFormatSame1, crossFormatSame2 := createCrossFormatFiles(t, tempDir)
	crossFormatDiff1, crossFormatDiff2 := createDifferentCrossFormatFiles(t, tempDir)
	yamlMultiline1, yamlMultiline2 := createYAMLMultilineFiles(t, tempDir)
	yamlConfig1, yamlConfig2 := createYAMLMultilineConfigFiles(t, tempDir)
	
	tests := []struct {
		name     string
		file1    string
		file2    string
		args     []string
		wantExit int
		wantOut  []string // Expected strings in output
	}{
		{
			name:     "Test case 1: JSON files with differences",
			file1:    json1,
			file2:    json2,
			wantExit: 1,
			wantOut: []string{
				"-  age: 30",
				"+  age: 31",
				"-  city: Tokyo",
				"+  city: Osaka",
				"   name: John",
			},
		},
		{
			name:     "Test case 1a: JSON files with diff-only",
			file1:    json1,
			file2:    json2,
			args:     []string{"-diff-only"},
			wantExit: 1,
			wantOut: []string{
				"-  age: 30",
				"+  age: 31",
				"-  city: Tokyo",
				"+  city: Osaka",
			},
		},
		{
			name:     "Test case 2: YAML files with differences",
			file1:    yaml1,
			file2:    yaml2,
			wantExit: 1,
			wantOut: []string{
				"-  age: 25",
				"+  age: 26",
				"hobbies",
			},
		},
		{
			name:     "Test case 3: Identical files",
			file1:    same1,
			file2:    same2,
			wantExit: 0,
			wantOut: []string{
				"   id: 123",
				"   status: active",
			},
		},
		{
			name:     "Test case 3a: Identical files with diff-only",
			file1:    same1,
			file2:    same2,
			args:     []string{"-diff-only"},
			wantExit: 0,
			wantOut:  []string{}, // No output expected
		},
		{
			name:     "Test case 4: Multiple JSON documents",
			file1:    multiJSON1,
			file2:    multiJSON2,
			wantExit: 1,
			wantOut: []string{
				"-  active: true",
				"+  active: false",
				"-  name: Charlie",
				"+  name: Robert",
				"-  name: Bob",
				"+  name: David",
			},
		},
		{
			name:     "Test case 4a: Multiple JSON documents with diff-only",
			file1:    multiJSON1,
			file2:    multiJSON2,
			args:     []string{"-diff-only"},
			wantExit: 1,
			wantOut: []string{
				"-  active: true",
				"+  active: false",
				"-  name: Charlie",
				"+  name: Robert",
				"-  name: Bob",
				"+  name: David",
			},
		},
		{
			name:     "Test case 5: Multiple YAML documents",
			file1:    multiYAML1,
			file2:    multiYAML2,
			wantExit: 1,
			wantOut: []string{
				"-  department: Engineering",
				"+  department: Marketing",
				"-  department: Sales",
				"+  department: Engineering",
				"-  name: Charlie",
				"+  name: Charles",
			},
		},
		{
			name:     "Test case 6: Multiple nested JSON documents",
			file1:    nestedJSON1,
			file2:    nestedJSON2,
			wantExit: 1,
			wantOut: []string{
				"-  user.profile.age: 30",
				"+  user.profile.age: 31",
				"-  settings.theme: dark",
				"+  settings.theme: light",
				"+  settings.lang: en",
			},
		},
		{
			name:     "Test case 6a: Multiple nested JSON with array strategy",
			file1:    nestedJSON1,
			file2:    nestedJSON2,
			args:     []string{"-array-strategy", "value"},
			wantExit: 1,
			wantOut: []string{
				"-  user.profile.age: 30",
				"+  user.profile.age: 31",
				"-  settings.theme: dark",
				"+  settings.theme: light",
				"+  settings.lang: en",
			},
		},
		{
			name:     "Test case 7: Mixed multiple YAML documents",
			file1:    mixedYAML1,
			file2:    mixedYAML2,
			wantExit: 1,
			wantOut: []string{
				"-  config.database.host: localhost",
				"+  config.database.host: db.example.com",
				"-  config.cache.enabled: true",
				"+  config.cache.enabled: false",
				"-  config.cache.ttl: 3600",
				"+  config.cache.ttl: 7200",
				"+  config.monitoring",
				"-  service.version: 1.2.3",
				"+  service.version: 1.3.0",
				"+  service.endpoints.[2]: /orders",
			},
		},
		{
			name:     "Test case 7a: Mixed YAML with diff-only",
			file1:    mixedYAML1,
			file2:    mixedYAML2,
			args:     []string{"-diff-only"},
			wantExit: 1,
			wantOut: []string{
				"-  config.database.host: localhost",
				"+  config.database.host: db.example.com",
				"-  config.cache.enabled: true",
				"+  config.cache.enabled: false",
				"-  config.cache.ttl: 3600",
				"+  config.cache.ttl: 7200",
				"+  config.monitoring",
				"-  service.version: 1.2.3",
				"+  service.version: 1.3.0",
				"+  service.endpoints.[2]: /orders",
			},
		},
		{
			name:     "Test case 8: Array differences with index strategy",
			file1:    arrayJSON1,
			file2:    arrayJSON2,
			args:     []string{"-array-strategy", "index"},
			wantExit: 1,
			wantOut: []string{
				"-  items.[0]: apple",
				"+  items.[0]: banana",
				"-  items.[1]: banana",
				"+  items.[1]: cherry",
				"-  items.[2]: cherry",
				"+  items.[2]: date",
				"+  items.[3]: apple",
				"-  tags.[1]: healthy",
				"+  tags.[1]: organic",
				"-  tags.[2]: organic",
				"+  tags.[2]: fresh",
				"+  tags.[3]: healthy",
			},
		},
		{
			name:     "Test case 8a: Array differences with value strategy",
			file1:    arrayJSON1,
			file2:    arrayJSON2,
			args:     []string{"-array-strategy", "value"},
			wantExit: 1,
			wantOut: []string{
				"+  items.[2]: date",
				"+  tags.[2]: fresh",
			},
		},
		{
			name:     "Test case 9: Ignore zero values",
			file1:    json1,
			file2:    json2,
			args:     []string{"-ignore-zero-values"},
			wantExit: 1,
			wantOut: []string{
				"-  age: 30",
				"+  age: 31",
				"-  city: Tokyo",
				"+  city: Osaka",
			},
		},
		{
			name:     "Test case 10: Cross-format comparison (same content)",
			file1:    crossFormatSame1,
			file2:    crossFormatSame2,
			wantExit: 0, // Exit 0 because numeric type differences are now handled
			wantOut: []string{
				"   name: Alice",
				"   age: 25",
				"   hobbies",
				"   active: true",
			},
		},
		{
			name:     "Test case 11: Cross-format comparison (different content)",
			file1:    crossFormatDiff1,
			file2:    crossFormatDiff2,
			wantExit: 1,
			wantOut: []string{
				"-  server.host: localhost",
				"+  server.host: example.com",
				"-  server.port: 8080",
				"+  server.port: 9000",
				"-  debug: false",
				"+  debug: true",
				"+  logging",
			},
		},
		{
			name:     "Test case 12: Force format with cross-format files",
			file1:    crossFormatDiff1,
			file2:    crossFormatDiff2,
			args:     []string{"-format1", "json", "-format2", "yaml"},
			wantExit: 1,
			wantOut: []string{
				"-  server.host: localhost",
				"+  server.host: example.com",
				"-  server.port: 8080",
				"+  server.port: 9000",
				"-  debug: false",
				"+  debug: true",
				"+  logging",
			},
		},
		{
			name:     "Test case 13: YAML multiline string formats (issue #29)",
			file1:    yamlMultiline1,
			file2:    yamlMultiline2,
			wantExit: 0, // Should be identical despite different YAML string syntax
			wantOut: []string{
				"   value: foo\nbar\nbaz\nspecial\n  multiline",
			},
		},
		{
			name:     "Test case 14: YAML multiline config diff (issue #52)",
			file1:    yamlConfig1,
			file2:    yamlConfig2,
			wantExit: 1,
			wantOut: []string{
				// Now shows line-by-line diff
				"   data.config:",
				"     logging.a: false",
				"-    logging.b: false",
				"+    logging.c: false",
			},
		},
		{
			name:     "Test case 14a: YAML multiline config diff with diff-only",
			file1:    yamlConfig1,
			file2:    yamlConfig2,
			args:     []string{"-diff-only"},
			wantExit: 1,
			wantOut: []string{
				// Only shows changed lines
				"-    logging.b: false",
				"+    logging.c: false",
			},
		},
	}
	
	// Run actual tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Prepare command arguments
			args := append([]string{}, tt.args...)
			args = append(args, tt.file1, tt.file2)
			
			// Capture stdout and stderr
			var stdout, stderr bytes.Buffer
			
			// Run main function
			exitCode := run(args, &stdout, &stderr)
			
			// Get output
			outputStr := stdout.String()
			
			// Debug output
			if t.Failed() || (exitCode != tt.wantExit) {
				t.Logf("Test: %s", tt.name)
				t.Logf("args: %v", args)
				t.Logf("exitCode: %d (want %d)", exitCode, tt.wantExit)
				t.Logf("stdout:\n%s", outputStr)
				t.Logf("stderr:\n%s", stderr.String())
			}
			
			// Check exit code
			if exitCode != tt.wantExit {
				t.Errorf("Exit code = %d, want %d", exitCode, tt.wantExit)
			}
			
			// Check output contains expected strings
			for _, want := range tt.wantOut {
				if !strings.Contains(outputStr, want) {
					t.Errorf("Output missing expected string %q\nGot output:\n%s", want, outputStr)
				}
			}
			
			// For diff-only tests with no differences, output should be empty
			if tt.wantExit == 0 && contains(tt.args, "-diff-only") && len(tt.wantOut) == 0 {
				if strings.TrimSpace(outputStr) != "" {
					t.Errorf("Expected empty output for diff-only with no differences, got:\n%s", outputStr)
				}
			}
		})
	}
}

func contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

func TestReadFile(t *testing.T) {
	// Create a temporary file
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	testContent := "Hello, World!"
	
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatal(err)
	}
	
	tests := []struct {
		name     string
		filename string
		want     string
		wantErr  bool
	}{
		{
			name:     "Read existing file",
			filename: testFile,
			want:     testContent,
			wantErr:  false,
		},
		{
			name:     "Read non-existent file",
			filename: filepath.Join(tempDir, "nonexistent.txt"),
			want:     "",
			wantErr:  true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := readFile(tt.filename)
			if (err != nil) != tt.wantErr {
				t.Errorf("readFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("readFile() = %v, want %v", got, tt.want)
			}
			if err != nil && !strings.Contains(err.Error(), "failed to open file") && !strings.Contains(err.Error(), "failed to read file") {
				t.Errorf("readFile() error message should contain 'failed to open file' or 'failed to read file', got %v", err)
			}
		})
	}
}

func TestReadFileStdin(t *testing.T) {
	// Test reading from stdin (-)
	// This is difficult to test directly without mocking os.Stdin
	// For now, we'll just verify the function handles the "-" case
	
	// Save original stdin
	oldStdin := os.Stdin
	defer func() {
		os.Stdin = oldStdin
	}()
	
	// Create a pipe to simulate stdin
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdin = r
	
	testData := "stdin test data"
	go func() {
		w.Write([]byte(testData))
		w.Close()
	}()
	
	got, err := readFile("-")
	if err != nil {
		t.Errorf("readFile(\"-\") unexpected error: %v", err)
	}
	if got != testData {
		t.Errorf("readFile(\"-\") = %v, want %v", got, testData)
	}
}