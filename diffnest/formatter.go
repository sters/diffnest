package diffnest

import (
	"fmt"
	"strings"
)

// Constants.
const (
	valueNull = "null"
)

// Formatter interface for different output formats.
type Formatter interface {
	Format(results []*DiffResult) string
}

// UnifiedFormatter implements unified diff format.
type UnifiedFormatter struct {
	ShowOnlyDiff bool
	Verbose      bool
}

// Format formats diff results.
func (f *UnifiedFormatter) Format(results []*DiffResult) string {
	var builder strings.Builder

	for i, result := range results {
		if i > 0 {
			builder.WriteString("---\n")
		}

		f.formatDiff(&builder, result, "")
	}

	return builder.String()
}

func (f *UnifiedFormatter) formatDiff(builder *strings.Builder, diff *DiffResult, indent string) {
	// Skip unchanged items if ShowOnlyDiff is true
	if f.ShowOnlyDiff && diff.Status == StatusSame {
		return
	}

	pathStr := strings.Join(diff.Path, ".")
	if pathStr != "" {
		pathStr = " " + pathStr
	}

	switch diff.Status {
	case StatusSame:
		if len(diff.Children) > 0 {
			// Show children for containers
			for _, child := range diff.Children {
				f.formatDiff(builder, child, indent)
			}
		} else {
			// Show value for primitives
			builder.WriteString(fmt.Sprintf("  %s%s: %s\n", indent, pathStr, f.formatValue(diff.From)))
		}

	case StatusModified:
		if len(diff.Children) > 0 {
			// Check if this is a multiline string diff
			if diff.From != nil && diff.From.Type == TypeString && diff.To != nil && diff.To.Type == TypeString {
				// Multiline string with line-by-line diff
				if !f.ShowOnlyDiff {
					builder.WriteString(fmt.Sprintf("  %s%s:\n", indent, pathStr))
				}
				for _, child := range diff.Children {
					if !f.ShowOnlyDiff || child.Status != StatusSame {
						f.formatLineDiff(builder, child, indent+"  ")
					}
				}
			} else {
				// Show children for containers
				for _, child := range diff.Children {
					f.formatDiff(builder, child, indent)
				}
			}
		} else {
			// Show old and new values
			builder.WriteString(fmt.Sprintf("- %s%s: %s\n", indent, pathStr, f.formatValue(diff.From)))
			builder.WriteString(fmt.Sprintf("+ %s%s: %s\n", indent, pathStr, f.formatValue(diff.To)))
		}

	case StatusDeleted:
		f.formatDeleted(builder, diff, indent)

	case StatusAdded:
		f.formatAdded(builder, diff, indent)
	}
}

func (f *UnifiedFormatter) formatDeleted(builder *strings.Builder, diff *DiffResult, indent string) {
	pathStr := strings.Join(diff.Path, ".")
	if pathStr != "" {
		pathStr = " " + pathStr
	}

	if diff.From == nil {
		builder.WriteString(fmt.Sprintf("- %s%s\n", indent, pathStr))

		return
	}

	switch diff.From.Type {
	case TypeObject:
		builder.WriteString(fmt.Sprintf("- %s%s:\n", indent, pathStr))
		f.formatStructure(builder, diff.From, indent+"  ", "- ")

	case TypeArray:
		builder.WriteString(fmt.Sprintf("- %s%s:\n", indent, pathStr))
		f.formatStructure(builder, diff.From, indent+"  ", "- ")

	default:
		builder.WriteString(fmt.Sprintf("- %s%s: %s\n", indent, pathStr, f.formatValue(diff.From)))
	}
}

func (f *UnifiedFormatter) formatAdded(builder *strings.Builder, diff *DiffResult, indent string) {
	pathStr := strings.Join(diff.Path, ".")
	if pathStr != "" {
		pathStr = " " + pathStr
	}

	if diff.To == nil {
		builder.WriteString(fmt.Sprintf("+ %s%s\n", indent, pathStr))

		return
	}

	switch diff.To.Type {
	case TypeObject:
		builder.WriteString(fmt.Sprintf("+ %s%s:\n", indent, pathStr))
		f.formatStructure(builder, diff.To, indent+"  ", "+ ")

	case TypeArray:
		builder.WriteString(fmt.Sprintf("+ %s%s:\n", indent, pathStr))
		f.formatStructure(builder, diff.To, indent+"  ", "+ ")

	default:
		builder.WriteString(fmt.Sprintf("+ %s%s: %s\n", indent, pathStr, f.formatValue(diff.To)))
	}
}

func (f *UnifiedFormatter) formatStructure(builder *strings.Builder, data *StructuredData, indent, prefix string) {
	switch data.Type {
	case TypeObject:
		for key, child := range data.Children {
			switch child.Type {
			case TypeObject, TypeArray:
				builder.WriteString(fmt.Sprintf("%s%s%s:\n", prefix, indent, key))
				f.formatStructure(builder, child, indent+"  ", prefix)
			default:
				builder.WriteString(fmt.Sprintf("%s%s%s: %s\n", prefix, indent, key, f.formatValue(child)))
			}
		}

	case TypeArray:
		for i, elem := range data.Elements {
			switch elem.Type {
			case TypeObject, TypeArray:
				builder.WriteString(fmt.Sprintf("%s%s[%d]:\n", prefix, indent, i))
				f.formatStructure(builder, elem, indent+"  ", prefix)
			default:
				builder.WriteString(fmt.Sprintf("%s%s[%d]: %s\n", prefix, indent, i, f.formatValue(elem)))
			}
		}
	}
}

func (f *UnifiedFormatter) formatValue(data *StructuredData) string {
	if data == nil {
		return valueNull
	}

	switch data.Type {
	case TypeNull:
		return valueNull
	case TypeBool, TypeNumber, TypeString:
		return fmt.Sprint(data.Value)
	case TypeArray:
		if len(data.Elements) == 0 {
			return "[]"
		}

		return fmt.Sprintf("[%d items]", len(data.Elements))
	case TypeObject:
		if len(data.Children) == 0 {
			return "{}"
		}

		return fmt.Sprintf("{%d fields}", len(data.Children))
	}

	return "?"
}

// formatLineDiff formats a single line difference in a multiline string
func (f *UnifiedFormatter) formatLineDiff(builder *strings.Builder, diff *DiffResult, indent string) {
	switch diff.Status {
	case StatusSame:
		if diff.From != nil {
			builder.WriteString(fmt.Sprintf("   %s%s\n", indent, diff.From.Value))
		}
	case StatusDeleted:
		if diff.From != nil {
			builder.WriteString(fmt.Sprintf("-  %s%s\n", indent, diff.From.Value))
		}
	case StatusAdded:
		if diff.To != nil {
			builder.WriteString(fmt.Sprintf("+  %s%s\n", indent, diff.To.Value))
		}
	case StatusModified:
		if diff.From != nil {
			builder.WriteString(fmt.Sprintf("-  %s%s\n", indent, diff.From.Value))
		}
		if diff.To != nil {
			builder.WriteString(fmt.Sprintf("+  %s%s\n", indent, diff.To.Value))
		}
	}
}

// JSONPatchFormatter implements RFC 6902 JSON Patch format.
type JSONPatchFormatter struct{}

// Format formats diff results as JSON Patch.
func (f *JSONPatchFormatter) Format(results []*DiffResult) string {
	var operations []string

	for _, result := range results {
		ops := f.generateOperations(result)
		operations = append(operations, ops...)
	}

	if len(operations) == 0 {
		return "[]\n"
	}

	return fmt.Sprintf("[\n  %s\n]\n", strings.Join(operations, ",\n  "))
}

func (f *JSONPatchFormatter) generateOperations(diff *DiffResult) []string {
	var ops []string

	path := "/" + strings.Join(diff.Path, "/")
	if path == "/" {
		path = ""
	}

	switch diff.Status {
	case StatusModified:
		if len(diff.Children) > 0 {
			// Generate ops for children
			for _, child := range diff.Children {
				ops = append(ops, f.generateOperations(child)...)
			}
		} else {
			// Replace operation
			op := fmt.Sprintf(`{"op": "replace", "path": "%s", "value": %s}`,
				path, f.jsonValue(diff.To))
			ops = append(ops, op)
		}

	case StatusDeleted:
		op := fmt.Sprintf(`{"op": "remove", "path": "%s"}`, path)
		ops = append(ops, op)

	case StatusAdded:
		op := fmt.Sprintf(`{"op": "add", "path": "%s", "value": %s}`,
			path, f.jsonValue(diff.To))
		ops = append(ops, op)

	case StatusSame:
		// Generate ops for modified children
		for _, child := range diff.Children {
			if child.Status != StatusSame {
				ops = append(ops, f.generateOperations(child)...)
			}
		}
	}

	return ops
}

func (f *JSONPatchFormatter) jsonValue(data *StructuredData) string {
	if data == nil {
		return valueNull
	}

	switch data.Type {
	case TypeNull:
		return valueNull
	case TypeBool:
		return fmt.Sprint(data.Value)
	case TypeNumber:
		return fmt.Sprint(data.Value)
	case TypeString:
		return fmt.Sprintf("%q", data.Value)
	case TypeArray:
		var elems []string
		for _, elem := range data.Elements {
			elems = append(elems, f.jsonValue(elem))
		}

		return fmt.Sprintf("[%s]", strings.Join(elems, ", "))
	case TypeObject:
		var fields []string
		for key, child := range data.Children {
			fields = append(fields, fmt.Sprintf("%q: %s", key, f.jsonValue(child)))
		}

		return fmt.Sprintf("{%s}", strings.Join(fields, ", "))
	}

	return valueNull
}
