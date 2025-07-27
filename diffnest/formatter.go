package diffnest

import (
	"fmt"
	"io"
	"strings"
)

// Constants.
const (
	valueNull = "null"
)

// Formatter interface for different output formats.
type Formatter interface {
	Format(w io.Writer, results []*DiffResult) error
}

// UnifiedFormatter implements unified diff format.
type UnifiedFormatter struct {
	ShowOnlyDiff bool
	Verbose      bool
	ContextLines int
}

// Format formats diff results.
func (f *UnifiedFormatter) Format(w io.Writer, results []*DiffResult) error {
	for i, result := range results {
		if i > 0 {
			if _, err := fmt.Fprint(w, "---\n"); err != nil {
				return fmt.Errorf("write separator: %w", err)
			}
		}

		// Apply context filtering at the top level if needed
		if f.ShowOnlyDiff && f.ContextLines >= 0 && len(result.Children) > 0 {
			if err := f.formatWithContext(w, result, ""); err != nil {
				return err
			}
		} else {
			if err := f.formatDiff(w, result, ""); err != nil {
				return err
			}
		}
	}

	return nil
}

// formatWithContext formats a diff result with context lines applied.
func (f *UnifiedFormatter) formatWithContext(w io.Writer, diff *DiffResult, indent string) error {
	if len(diff.Children) == 0 {
		return f.formatDiff(w, diff, indent)
	}

	pathStr := strings.Join(diff.Path, ".")
	if pathStr != "" {
		pathStr = " " + pathStr
	}

	// Check if this is a multiline string diff
	if diff.From != nil && diff.From.Type == TypeString && diff.To != nil && diff.To.Type == TypeString {
		// Multiline string with line-by-line diff
		if !f.ShowOnlyDiff {
			if _, err := fmt.Fprintf(w, "  %s%s:\n", indent, pathStr); err != nil {
				return fmt.Errorf("write multiline header: %w", err)
			}
		}
		return f.formatMultilineWithContext(w, diff.Children, indent+"  ")
	}

	// Object or array with children
	return f.formatChildrenWithContext(w, diff.Children, indent)
}

func (f *UnifiedFormatter) formatDiff(w io.Writer, diff *DiffResult, indent string) error {
	// Skip unchanged items if ShowOnlyDiff is true and not using context lines
	if f.ShowOnlyDiff && diff.Status == StatusSame && f.ContextLines < 0 {
		return nil
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
				if err := f.formatDiff(w, child, indent); err != nil {
					return err
				}
			}
		} else {
			// Show value for primitives
			if _, err := fmt.Fprintf(w, "  %s%s: %s\n", indent, pathStr, f.formatValue(diff.From)); err != nil {
				return fmt.Errorf("write same value: %w", err)
			}
		}

	case StatusModified:
		if len(diff.Children) > 0 {
			// Check if this is a multiline string diff
			if diff.From != nil && diff.From.Type == TypeString && diff.To != nil && diff.To.Type == TypeString {
				// Multiline string with line-by-line diff
				if !f.ShowOnlyDiff {
					if _, err := fmt.Fprintf(w, "  %s%s:\n", indent, pathStr); err != nil {
						return fmt.Errorf("write multiline header: %w", err)
					}
				}
				if f.ShowOnlyDiff && f.ContextLines >= 0 {
					// Format with context lines
					if err := f.formatMultilineWithContext(w, diff.Children, indent+"  "); err != nil {
						return err
					}
				} else {
					// Original behavior
					for _, child := range diff.Children {
						if !f.ShowOnlyDiff || child.Status != StatusSame {
							if err := f.formatLineDiff(w, child, indent+"  "); err != nil {
								return err
							}
						}
					}
				}
			} else {
				// Show children for containers
				if f.ShowOnlyDiff && f.ContextLines >= 0 {
					// Format with context for object/array children
					if err := f.formatChildrenWithContext(w, diff.Children, indent); err != nil {
						return err
					}
				} else {
					for _, child := range diff.Children {
						if err := f.formatDiff(w, child, indent); err != nil {
							return err
						}
					}
				}
			}
		} else {
			// Show old and new values
			if _, err := fmt.Fprintf(w, "- %s%s: %s\n", indent, pathStr, f.formatValue(diff.From)); err != nil {
				return fmt.Errorf("write deleted value: %w", err)
			}
			if _, err := fmt.Fprintf(w, "+ %s%s: %s\n", indent, pathStr, f.formatValue(diff.To)); err != nil {
				return fmt.Errorf("write added value: %w", err)
			}
		}

	case StatusDeleted:
		if err := f.formatDeleted(w, diff, indent); err != nil {
			return err
		}

	case StatusAdded:
		if err := f.formatAdded(w, diff, indent); err != nil {
			return err
		}
	}

	return nil
}

func (f *UnifiedFormatter) formatDeleted(w io.Writer, diff *DiffResult, indent string) error {
	return f.formatAddedOrDeleted(w, diff.From, diff.Path, indent, "- ")
}

func (f *UnifiedFormatter) formatAdded(w io.Writer, diff *DiffResult, indent string) error {
	return f.formatAddedOrDeleted(w, diff.To, diff.Path, indent, "+ ")
}

func (f *UnifiedFormatter) formatAddedOrDeleted(w io.Writer, data *StructuredData, path []string, indent, prefix string) error {
	pathStr := strings.Join(path, ".")
	if pathStr != "" {
		pathStr = " " + pathStr
	}

	if data == nil {
		if _, err := fmt.Fprintf(w, "%s%s%s\n", prefix, indent, pathStr); err != nil {
			return fmt.Errorf("write path: %w", err)
		}

		return nil
	}

	switch data.Type {
	case TypeObject, TypeArray:
		if _, err := fmt.Fprintf(w, "%s%s%s:\n", prefix, indent, pathStr); err != nil {
			return fmt.Errorf("write structure header: %w", err)
		}
		if err := f.formatStructure(w, data, indent+"  ", prefix); err != nil {
			return err
		}

	default:
		if _, err := fmt.Fprintf(w, "%s%s%s: %s\n", prefix, indent, pathStr, f.formatValue(data)); err != nil {
			return fmt.Errorf("write value: %w", err)
		}
	}

	return nil
}

func (f *UnifiedFormatter) formatStructure(w io.Writer, data *StructuredData, indent, prefix string) error {
	switch data.Type {
	case TypeObject:
		for key, child := range data.Children {
			switch child.Type {
			case TypeObject, TypeArray:
				if _, err := fmt.Fprintf(w, "%s%s%s:\n", prefix, indent, key); err != nil {
					return fmt.Errorf("write object key: %w", err)
				}
				if err := f.formatStructure(w, child, indent+"  ", prefix); err != nil {
					return err
				}
			default:
				if _, err := fmt.Fprintf(w, "%s%s%s: %s\n", prefix, indent, key, f.formatValue(child)); err != nil {
					return fmt.Errorf("write object field: %w", err)
				}
			}
		}

	case TypeArray:
		for i, elem := range data.Elements {
			switch elem.Type {
			case TypeObject, TypeArray:
				if _, err := fmt.Fprintf(w, "%s%s[%d]:\n", prefix, indent, i); err != nil {
					return fmt.Errorf("write array index: %w", err)
				}
				if err := f.formatStructure(w, elem, indent+"  ", prefix); err != nil {
					return err
				}
			default:
				if _, err := fmt.Fprintf(w, "%s%s[%d]: %s\n", prefix, indent, i, f.formatValue(elem)); err != nil {
					return fmt.Errorf("write array element: %w", err)
				}
			}
		}
	}

	return nil
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

// formatLineDiff formats a single line difference in a multiline string.
func (f *UnifiedFormatter) formatLineDiff(w io.Writer, diff *DiffResult, indent string) error {
	switch diff.Status {
	case StatusSame:
		if diff.From != nil {
			if _, err := fmt.Fprintf(w, "   %s%s\n", indent, diff.From.Value); err != nil {
				return fmt.Errorf("write same line: %w", err)
			}
		}
	case StatusDeleted:
		if diff.From != nil {
			if _, err := fmt.Fprintf(w, "-  %s%s\n", indent, diff.From.Value); err != nil {
				return fmt.Errorf("write deleted line: %w", err)
			}
		}
	case StatusAdded:
		if diff.To != nil {
			if _, err := fmt.Fprintf(w, "+  %s%s\n", indent, diff.To.Value); err != nil {
				return fmt.Errorf("write added line: %w", err)
			}
		}
	case StatusModified:
		if diff.From != nil {
			if _, err := fmt.Fprintf(w, "-  %s%s\n", indent, diff.From.Value); err != nil {
				return fmt.Errorf("write modified old line: %w", err)
			}
		}
		if diff.To != nil {
			if _, err := fmt.Fprintf(w, "+  %s%s\n", indent, diff.To.Value); err != nil {
				return fmt.Errorf("write modified new line: %w", err)
			}
		}
	}

	return nil
}

// formatMultilineWithContext formats multiline string diffs with context lines.
func (f *UnifiedFormatter) formatMultilineWithContext(w io.Writer, lines []*DiffResult, indent string) error {
	// Find all changed line indices
	var changedIndices []int
	for i, line := range lines {
		if line.Status != StatusSame {
			changedIndices = append(changedIndices, i)
		}
	}

	if len(changedIndices) == 0 {
		return nil
	}

	// Calculate which lines to show based on context
	showLine := make([]bool, len(lines))
	for _, idx := range changedIndices {
		// Always show the changed line
		showLine[idx] = true
		
		// Show context before
		for i := 1; i <= f.ContextLines && idx-i >= 0; i++ {
			showLine[idx-i] = true
		}
		
		// Show context after
		for i := 1; i <= f.ContextLines && idx+i < len(lines); i++ {
			showLine[idx+i] = true
		}
	}

	// Format the lines
	prevShown := false
	for i, line := range lines {
		if showLine[i] {
			if !prevShown && i > 0 {
				// Add separator for skipped lines
				if _, err := fmt.Fprintf(w, "   %s...\n", indent); err != nil {
					return fmt.Errorf("write separator: %w", err)
				}
			}
			if err := f.formatLineDiff(w, line, indent); err != nil {
				return err
			}
			prevShown = true
		} else {
			prevShown = false
		}
	}

	return nil
}

// formatChildrenWithContext formats object/array children with context lines.
func (f *UnifiedFormatter) formatChildrenWithContext(w io.Writer, children []*DiffResult, indent string) error {
	// Find all changed child indices
	var changedIndices []int
	for i, child := range children {
		if child.Status != StatusSame || f.hasChangedDescendants(child) {
			changedIndices = append(changedIndices, i)
		}
	}

	if len(changedIndices) == 0 {
		return nil
	}

	// Calculate which children to show based on context
	showChild := make([]bool, len(children))
	for _, idx := range changedIndices {
		// Always show the changed child
		showChild[idx] = true
		
		// Show context before
		for i := 1; i <= f.ContextLines && idx-i >= 0; i++ {
			showChild[idx-i] = true
		}
		
		// Show context after
		for i := 1; i <= f.ContextLines && idx+i < len(children); i++ {
			showChild[idx+i] = true
		}
	}

	// Format the children
	prevShown := false
	for i, child := range children {
		if showChild[i] {
			if !prevShown && i > 0 {
				// Add separator for skipped items
				if _, err := fmt.Fprintf(w, "  %s...\n", indent); err != nil {
					return fmt.Errorf("write separator: %w", err)
				}
			}
			if err := f.formatDiff(w, child, indent); err != nil {
				return err
			}
			prevShown = true
		} else {
			prevShown = false
		}
	}

	return nil
}

// hasChangedDescendants checks if a diff result has any changed descendants.
func (f *UnifiedFormatter) hasChangedDescendants(diff *DiffResult) bool {
	if diff.Status != StatusSame {
		return true
	}
	for _, child := range diff.Children {
		if f.hasChangedDescendants(child) {
			return true
		}
	}
	return false
}

// JSONPatchFormatter implements RFC 6902 JSON Patch format.
type JSONPatchFormatter struct{}

// Format formats diff results as JSON Patch.
func (f *JSONPatchFormatter) Format(w io.Writer, results []*DiffResult) error {
	var operations []string

	for _, result := range results {
		ops := f.generateOperations(result)
		operations = append(operations, ops...)
	}

	if len(operations) == 0 {
		if _, err := fmt.Fprint(w, "[]\n"); err != nil {
			return fmt.Errorf("write empty patch: %w", err)
		}

		return nil
	}

	if _, err := fmt.Fprintf(w, "[\n  %s\n]\n", strings.Join(operations, ",\n  ")); err != nil {
		return fmt.Errorf("write patch array: %w", err)
	}

	return nil
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
