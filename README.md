# diffnest

[![go](https://github.com/sters/diffnest/workflows/Go/badge.svg)](https://github.com/sters/diffnest/actions?query=workflow%3AGo)
[![coverage](docs/coverage.svg)](https://github.com/sters/diffnest)
[![go-report](https://goreportcard.com/badge/github.com/sters/diffnest)](https://goreportcard.com/report/github.com/sters/diffnest)
[![Go Reference](https://pkg.go.dev/badge/github.com/sters/diffnest.svg)](https://pkg.go.dev/github.com/sters/diffnest)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A powerful cross-format diff tool that compares JSON, YAML, and other structured data files with an intuitive unified diff output.

## Features

- **Cross-format comparison**: Compare files in different formats (JSON vs YAML)
- **Multiple document support**: Handle multiple documents in a single file
- **Smart array comparison**: Compare arrays by index or by value matching
- **Multiple output formats**: Unified diff (default) or JSON Patch (RFC 6902)
- **Detailed multiline string diffs**: Line-by-line comparison for multiline strings
- **Flexible options**: Ignore zero values, show only differences, and more
- **Standard input support**: Compare files from stdin using `-`

## Install

### Using Go

```shell
go install github.com/sters/diffnest@latest
```

### Download Binary

Download pre-built binaries from [Releases](https://github.com/sters/diffnest/releases).

## Usage

### Basic Usage

```shell
# Compare two JSON files
diffnest file1.json file2.json

# Compare JSON and YAML files
diffnest config.json config.yaml

# Compare from stdin
cat file1.json | diffnest - file2.json
```

### Options

```
-diff-only              Show only differences (hide unchanged fields)
-ignore-zero-values     Treat zero values (0, false, "", [], {}) as null
-ignore-empty          Ignore empty fields
-array-strategy        Array comparison strategy: 'index' or 'value' (default: value)
-format                Output format: 'unified' or 'json-patch' (default: unified)
-format1               Format for first file: 'json', 'yaml', or auto-detect
-format2               Format for second file: 'json', 'yaml', or auto-detect
-v                     Verbose output
-h                     Show help
```

### Examples

#### Show only differences
```shell
diffnest -diff-only actual.json expected.json
```

#### Compare arrays by index (ordered comparison)
```shell
diffnest -array-strategy index list1.yaml list2.yaml
```

#### Output as JSON Patch
```shell
diffnest -format json-patch old.json new.json
```

#### Ignore zero values
```shell
diffnest -ignore-zero-values sparse1.json sparse2.json
```

## Output Examples

### Unified Diff Format (Default)

```diff
-  name: oldValue
+  name: newValue
   unchanged: sameValue
+  added: newField
-  removed: deletedField
```

### JSON Patch Format

```json
[
  {"op": "replace", "path": "/name", "value": "newValue"},
  {"op": "add", "path": "/added", "value": "newField"},
  {"op": "remove", "path": "/removed"}
]
```

## Advanced Features

### Cross-Format Comparison

diffnest can seamlessly compare files in different formats:

```shell
# Compare JSON config with YAML config
diffnest app.json app.yaml
```

### Multiple Document Support

YAML files with multiple documents (separated by `---`) are fully supported:

```yaml
---
doc1: value1
---
doc2: value2
```

### Smart Array Comparison

Choose between two array comparison strategies:

- **`value`** (default): Finds the best matching between array elements
- **`index`**: Compares arrays by position

```shell
# Smart matching (reordered elements are considered equal)
diffnest -array-strategy value items1.json items2.json

# Strict ordering (position matters)
diffnest -array-strategy index items1.json items2.json
```

### Multiline String Comparison

Multiline strings are compared line-by-line for better readability:

```diff
  config:
-    line 1: old text
+    line 1: new text
     line 2: unchanged
+    line 3: added line
```

## Exit Codes

- `0`: No differences found
- `1`: Differences found or error occurred

## Library Usage

diffnest can also be used as a Go library:

```go
import "github.com/sters/diffnest/diffnest"

// Compare two data structures
opts := diffnest.DiffOptions{
    ArrayDiffStrategy: diffnest.ArrayStrategyValue,
}
results := diffnest.Compare(data1, data2, opts)

// Format the results
formatter := &diffnest.UnifiedFormatter{
    ShowOnlyDiff: true,
}
formatter.Format(os.Stdout, results)
```
