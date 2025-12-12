# diffnest

[![go](https://github.com/sters/diffnest/workflows/Go/badge.svg)](https://github.com/sters/diffnest/actions?query=workflow%3AGo)
[![coverage](docs/coverage.svg)](https://github.com/sters/diffnest)
[![go-report](https://goreportcard.com/badge/github.com/sters/diffnest)](https://goreportcard.com/report/github.com/sters/diffnest)
[![Go Reference](https://pkg.go.dev/badge/github.com/sters/diffnest.svg)](https://pkg.go.dev/github.com/sters/diffnest)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A powerful cross-format diff tool that compares JSON, YAML, and other structured data files with an intuitive unified diff output. By default, diffnest shows only the differences between files, making it easy to spot changes in large configuration files.

## Features

- **Cross-format comparison**: Compare files in different formats (JSON vs YAML)
- **Multiple document support**: Handle multiple documents in a single file with optimal pairing using the Hungarian algorithm
- **Smart array comparison**: Compare arrays by index or by value matching
- **Multiple output formats**: Unified diff (default) or JSON Patch (RFC 6902)
- **Detailed multiline string diffs**: Line-by-line comparison for multiline strings
- **Context control**: Adjustable context lines around changes for better readability
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
-show-all              Show all fields including unchanged ones (default: show only differences)
-ignore-zero-values    Treat zero values (0, false, "", [], {}) as null
-ignore-empty          Ignore empty fields
-ignore-key-case       Ignore case differences in object keys
-ignore-value-case     Ignore case differences in string values
-array-strategy        Array comparison strategy: 'index' or 'value' (default: value)
-format                Output format: 'unified' or 'json-patch' (default: unified)
-format1               Format for first file: 'json', 'yaml', or auto-detect
-format2               Format for second file: 'json', 'yaml', or auto-detect
-C                     Number of context lines to show (incompatible with -show-all, default: 3)
-v                     Verbose output
-h                     Show help
```

### Examples

#### Default behavior (show only differences)
```shell
# By default, only differences are shown
diffnest actual.json expected.json
```

#### Show all fields including unchanged
```shell
diffnest -show-all actual.json expected.json
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

#### Control context lines
```shell
# Show only changed lines (no context)
diffnest -C 0 file1.json file2.json

# Show more context around changes
diffnest -C 5 file1.json file2.json
```

#### Case-insensitive comparisons
```shell
# Ignore case differences in object keys
diffnest --ignore-key-case config1.json config2.json

# Ignore case differences in string values
diffnest --ignore-value-case data1.json data2.json

# Ignore case in both keys and values
diffnest --ignore-key-case --ignore-value-case file1.json file2.json
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

#### Optimal Document Pairing with Hungarian Algorithm

When comparing files with multiple documents, diffnest uses the **Hungarian algorithm** to find the optimal pairing between documents. This ensures that similar documents are matched together, minimizing the total differences reported.

For Kubernetes-like resources, diffnest applies additional matching heuristics based on:
- `kind` (highest priority - different kinds are strongly discouraged from matching)
- `apiVersion`

This means when comparing two Kubernetes manifest files, resources of the same `kind` will be matched based on content similarity. If a resource is renamed (e.g., `metadata.name` changed), it will be shown as "modified" rather than "deleted + added", making it easier to see what actually changed.

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

### Context Lines Control

The `-C` option allows you to control how many unchanged lines are shown around changes, similar to traditional diff tools. This option is only available when showing differences (default behavior) and cannot be used with `-show-all`:

```shell
# Show minimal context (only changed lines)
diffnest -C 0 file1.json file2.json

# Invalid: cannot combine -C with -show-all
# diffnest -show-all -C 2 file1.json file2.json  # Error!
```

Example output with `-C 0`:
```diff
-  field3: old
+  field3: new
```

Example output with `-C 2`:
```diff
   field1: value1
   field2: value2
-  field3: old
+  field3: new
   field4: value4
   field5: value5
```

Example output with `-show-all`:
```diff
   field1: value1
   field2: value2
-  field3: old
+  field3: new
   field4: value4
   field5: value5
   field6: value6
   field7: value7
   ... (all fields shown)
```

For large files with scattered changes, context control helps focus on what matters:
```diff
   field2: value2
-  field3: old
+  field3: new
   field4: value4
   ...
   field98: value98
-  field99: old
+  field99: new
   field100: value100
```

**Important:** The `-C` option is incompatible with `-show-all`. When using `-show-all`, all fields are displayed regardless of context settings.

### Case-Insensitive Comparisons

diffnest provides granular control over case sensitivity:

#### Object Key Case Insensitivity

```shell
# Without --ignore-key-case (default)
diffnest file1.json file2.json
```

```diff
-  Name: John
+  name: John
```

```shell
# With --ignore-key-case
diffnest --ignore-key-case file1.json file2.json
```

```diff
(no differences - keys "Name" and "name" are considered the same)
```

#### String Value Case Insensitivity

```shell
# Without --ignore-value-case (default)
diffnest file1.json file2.json
```

```diff
-  status: ACTIVE
+  status: active
```

```shell
# With --ignore-value-case
diffnest --ignore-value-case file1.json file2.json
```

```diff
(no differences - values "ACTIVE" and "active" are considered the same)
```

## Option Compatibility

Some options are incompatible and cannot be used together:

- `-show-all` and `-C`: Context lines are only meaningful when showing differences. When using `-show-all`, all fields are displayed.

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
    ContextLines: 3,  // Number of context lines
}
formatter.Format(os.Stdout, results)
```
