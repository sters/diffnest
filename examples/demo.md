# diffnest Demo

This directory contains example files to demonstrate diffnest's capabilities.

## Quick Demo

### 1. Basic Comparison

Compare two JSON files:
```bash
diffnest user1.json user2.json
```

### 2. Cross-Format Comparison

Compare JSON with YAML:
```bash
diffnest config.json config.yaml
```

### 3. Show Only Differences

```bash
diffnest -diff-only user1.json user2.json
```

### 4. Array Comparison Strategies

Compare arrays by value (smart matching):
```bash
diffnest -array-strategy value items1.json items2.json
```

Compare arrays by index (position matters):
```bash
diffnest -array-strategy index items1.json items2.json
```

### 5. JSON Patch Output

```bash
diffnest -format json-patch user1.json user2.json
```

### 6. Multiline String Comparison

```bash
diffnest config-old.yaml config-new.yaml
```

## Try It Yourself

1. Clone the repository
2. Navigate to the examples directory
3. Run the commands above to see different comparison modes

## Sample Output

### Unified Diff (Default)
```diff
-  email: "user@example.com"
+  email: "user@newdomain.com"
   name: "John Doe"
+  phone: "+1-234-567-8900"
```

### JSON Patch
```json
[
  {"op": "replace", "path": "/email", "value": "user@newdomain.com"},
  {"op": "add", "path": "/phone", "value": "+1-234-567-8900"}
]
```