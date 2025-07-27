package diffnest

import (
	"fmt"
	"sort"
	"strings"
)

// DiffOptions contains options for diff computation.
type DiffOptions struct {
	IgnoreEmptyFields bool
	IgnoreZeroValues  bool
	ArrayDiffStrategy ArrayDiffStrategy
}

// ArrayDiffStrategy defines how to compare arrays.
type ArrayDiffStrategy int

const (
	ArrayStrategyIndex ArrayDiffStrategy = iota // Compare by index
	ArrayStrategyValue                          // Find best matching
)

// DiffEngine computes differences between structures.
type DiffEngine struct {
	options DiffOptions
}

// NewDiffEngine creates a new diff engine.
func NewDiffEngine(options DiffOptions) *DiffEngine {
	return &DiffEngine{options: options}
}

// Compare compares two structured data.
func (e *DiffEngine) Compare(a, b *StructuredData) *DiffResult {
	return e.compareWithPath(a, b, []string{})
}

func (e *DiffEngine) compareWithPath(a, b *StructuredData, path []string) *DiffResult {
	// Handle nil cases
	if a == nil && b == nil {
		return &DiffResult{
			Status: StatusSame,
			Path:   path,
		}
	}

	if a == nil {
		return &DiffResult{
			Status: StatusAdded,
			Path:   path,
			To:     b,
			Meta:   &DiffMeta{DiffCount: e.calculateSize(b)},
		}
	}

	if b == nil {
		return &DiffResult{
			Status: StatusDeleted,
			Path:   path,
			From:   a,
			Meta:   &DiffMeta{DiffCount: e.calculateSize(a)},
		}
	}

	// Type mismatch
	if a.Type != b.Type {
		return &DiffResult{
			Status: StatusModified,
			Path:   path,
			From:   a,
			To:     b,
			Meta:   &DiffMeta{DiffCount: e.calculateSize(a) + e.calculateSize(b)},
		}
	}

	// Compare based on type
	switch a.Type {
	case TypeNull:
		return &DiffResult{
			Status: StatusSame,
			Path:   path,
			From:   a,
			To:     b,
		}

	case TypeBool:
		if a.Value == b.Value {
			return &DiffResult{
				Status: StatusSame,
				Path:   path,
				From:   a,
				To:     b,
			}
		}

		return &DiffResult{
			Status: StatusModified,
			Path:   path,
			From:   a,
			To:     b,
			Meta:   &DiffMeta{DiffCount: 1},
		}

	case TypeString:
		// Check if we should do line-by-line diff for multiline strings
		if e.shouldDoLineDiff(a, b) {
			return e.compareMultilineStrings(a, b, path)
		}
		
		// Default string comparison
		if a.Value == b.Value {
			return &DiffResult{
				Status: StatusSame,
				Path:   path,
				From:   a,
				To:     b,
			}
		}

		return &DiffResult{
			Status: StatusModified,
			Path:   path,
			From:   a,
			To:     b,
			Meta:   &DiffMeta{DiffCount: 1},
		}

	case TypeNumber:
		// Compare numbers with type conversion
		if e.equalNumbers(a.Value, b.Value) {
			return &DiffResult{
				Status: StatusSame,
				Path:   path,
				From:   a,
				To:     b,
			}
		}

		return &DiffResult{
			Status: StatusModified,
			Path:   path,
			From:   a,
			To:     b,
			Meta:   &DiffMeta{DiffCount: 1},
		}

	case TypeArray:
		return e.compareArrays(a, b, path)

	case TypeObject:
		return e.compareObjects(a, b, path)

	default:
		return &DiffResult{
			Status: StatusModified,
			Path:   path,
			From:   a,
			To:     b,
			Meta:   &DiffMeta{DiffCount: 1},
		}
	}
}

func (e *DiffEngine) compareArrays(a, b *StructuredData, path []string) *DiffResult {
	if e.options.ArrayDiffStrategy == ArrayStrategyValue {
		return e.compareArraysByValue(a, b, path)
	}

	return e.compareArraysByIndex(a, b, path)
}

func (e *DiffEngine) compareArraysByIndex(a, b *StructuredData, path []string) *DiffResult {
	result := &DiffResult{
		Status:   StatusSame,
		Path:     path,
		From:     a,
		To:       b,
		Children: []*DiffResult{},
		Meta:     &DiffMeta{DiffCount: 0},
	}

	maxLen := len(a.Elements)
	if len(b.Elements) > maxLen {
		maxLen = len(b.Elements)
	}

	for i := range maxLen {
		var elemA, elemB *StructuredData

		if i < len(a.Elements) {
			elemA = a.Elements[i]
		}
		if i < len(b.Elements) {
			elemB = b.Elements[i]
		}

		childPath := append(append([]string{}, path...), fmt.Sprintf("[%d]", i))
		childDiff := e.compareWithPath(elemA, elemB, childPath)

		if childDiff.Status != StatusSame {
			result.Status = StatusModified
			if childDiff.Meta != nil {
				result.Meta.DiffCount += childDiff.Meta.DiffCount
			}
		}

		result.Children = append(result.Children, childDiff)
	}

	return result
}

func (e *DiffEngine) compareArraysByValue(a, b *StructuredData, path []string) *DiffResult {
	result := &DiffResult{
		Status:   StatusSame,
		Path:     path,
		From:     a,
		To:       b,
		Children: []*DiffResult{},
		Meta:     &DiffMeta{DiffCount: 0},
	}

	// Build all possible comparisons
	type match struct {
		indexA int
		indexB int
		diff   *DiffResult
		cost   int
	}

	var matches []match

	// Compare all pairs
	for i, elemA := range a.Elements {
		for j, elemB := range b.Elements {
			childDiff := e.compareWithPath(elemA, elemB, []string{})
			cost := 0
			if childDiff.Meta != nil {
				cost = childDiff.Meta.DiffCount
			}
			matches = append(matches, match{
				indexA: i,
				indexB: j,
				diff:   childDiff,
				cost:   cost,
			})
		}
	}

	// Sort by cost (best matches first)
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].cost < matches[j].cost
	})

	// Find optimal matching
	usedA := make(map[int]bool)
	usedB := make(map[int]bool)
	var finalMatches []match

	for _, m := range matches {
		if !usedA[m.indexA] && !usedB[m.indexB] {
			usedA[m.indexA] = true
			usedB[m.indexB] = true
			finalMatches = append(finalMatches, m)
		}
	}

	// Add unmatched elements from A
	for i := range a.Elements {
		if !usedA[i] {
			childPath := append(append([]string{}, path...), fmt.Sprintf("[%d]", i))
			childDiff := e.compareWithPath(a.Elements[i], nil, childPath)
			result.Children = append(result.Children, childDiff)
			result.Status = StatusModified
			if childDiff.Meta != nil {
				result.Meta.DiffCount += childDiff.Meta.DiffCount
			}
		}
	}

	// Add unmatched elements from B
	for j := range b.Elements {
		if !usedB[j] {
			childPath := append(append([]string{}, path...), fmt.Sprintf("[%d]", j))
			childDiff := e.compareWithPath(nil, b.Elements[j], childPath)
			result.Children = append(result.Children, childDiff)
			result.Status = StatusModified
			if childDiff.Meta != nil {
				result.Meta.DiffCount += childDiff.Meta.DiffCount
			}
		}
	}

	// Add matched elements
	for _, m := range finalMatches {
		childPath := append(append([]string{}, path...), fmt.Sprintf("[%d]", m.indexA))
		m.diff.Path = childPath
		result.Children = append(result.Children, m.diff)
		if m.diff.Status != StatusSame {
			result.Status = StatusModified
			if m.diff.Meta != nil {
				result.Meta.DiffCount += m.diff.Meta.DiffCount
			}
		}
	}

	return result
}

func (e *DiffEngine) compareObjects(a, b *StructuredData, path []string) *DiffResult {
	result := &DiffResult{
		Status:   StatusSame,
		Path:     path,
		From:     a,
		To:       b,
		Children: []*DiffResult{},
		Meta:     &DiffMeta{DiffCount: 0},
	}

	// Collect all keys
	allKeys := make(map[string]bool)
	for k := range a.Children {
		allKeys[k] = true
	}
	for k := range b.Children {
		allKeys[k] = true
	}

	// Sort keys for consistent output
	keys := make([]string, 0, len(allKeys))
	for k := range allKeys {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Compare each key
	for _, key := range keys {
		childA, hasA := a.Children[key]
		childB, hasB := b.Children[key]

		// Apply ignore options
		if e.shouldIgnore(childA, hasA) && e.shouldIgnore(childB, hasB) {
			continue
		}

		childPath := append(append([]string{}, path...), key)

		if !hasA {
			childA = nil
		}
		if !hasB {
			childB = nil
		}

		childDiff := e.compareWithPath(childA, childB, childPath)

		if childDiff.Status != StatusSame {
			result.Status = StatusModified
			if childDiff.Meta != nil {
				result.Meta.DiffCount += childDiff.Meta.DiffCount
			}
		}

		result.Children = append(result.Children, childDiff)
	}

	return result
}

func (e *DiffEngine) equalNumbers(a, b any) bool {
	// Convert both values to float64 for comparison
	aFloat := toFloat64(a)
	bFloat := toFloat64(b)
	
	// Check if the conversion was successful for both
	aInt, aIsInt := toInt64(a)
	bInt, bIsInt := toInt64(b)
	
	// If both are integers, compare as integers
	if aIsInt && bIsInt {
		return aInt == bInt
	}
	
	// Otherwise compare as floats
	return aFloat == bFloat
}

// toFloat64 converts various numeric types to float64
func toFloat64(v any) float64 {
	switch val := v.(type) {
	case int:
		return float64(val)
	case int8:
		return float64(val)
	case int16:
		return float64(val)
	case int32:
		return float64(val)
	case int64:
		return float64(val)
	case uint:
		return float64(val)
	case uint8:
		return float64(val)
	case uint16:
		return float64(val)
	case uint32:
		return float64(val)
	case uint64:
		return float64(val)
	case float32:
		return float64(val)
	case float64:
		return val
	default:
		return 0
	}
}

// toInt64 converts various integer types to int64
func toInt64(v any) (int64, bool) {
	switch val := v.(type) {
	case int:
		return int64(val), true
	case int8:
		return int64(val), true
	case int16:
		return int64(val), true
	case int32:
		return int64(val), true
	case int64:
		return val, true
	case uint:
		return int64(val), true
	case uint8:
		return int64(val), true
	case uint16:
		return int64(val), true
	case uint32:
		return int64(val), true
	case uint64:
		if val <= 9223372036854775807 { // max int64
			return int64(val), true
		}
		return 0, false
	case float32:
		if float64(val) == float64(int64(val)) {
			return int64(val), true
		}
		return 0, false
	case float64:
		if val == float64(int64(val)) {
			return int64(val), true
		}
		return 0, false
	default:
		return 0, false
	}
}

func (e *DiffEngine) shouldIgnore(data *StructuredData, exists bool) bool {
	if !exists && e.options.IgnoreEmptyFields {
		return true
	}

	if !e.options.IgnoreZeroValues {
		return false
	}

	if data == nil {
		return true
	}

	switch data.Type {
	case TypeNull:
		return true
	case TypeBool:
		return data.Value == false
	case TypeNumber:
		return data.Value == 0 || data.Value == 0.0
	case TypeString:
		return data.Value == ""
	case TypeArray:
		return len(data.Elements) == 0
	case TypeObject:
		return len(data.Children) == 0
	}

	return false
}

func (e *DiffEngine) calculateSize(data *StructuredData) int {
	if data == nil {
		return 0
	}

	switch data.Type {
	case TypeNull, TypeBool, TypeNumber, TypeString:
		return 1
	case TypeArray:
		size := 0
		for _, elem := range data.Elements {
			size += e.calculateSize(elem)
		}

		return size
	case TypeObject:
		size := 0
		for _, child := range data.Children {
			size += e.calculateSize(child)
		}

		return size
	}

	return 1
}

// Compare compares multiple documents and finds optimal pairings.
func Compare(docsA, docsB []*StructuredData, options DiffOptions) []*DiffResult {
	engine := NewDiffEngine(options)

	// If single documents, compare directly
	if len(docsA) == 1 && len(docsB) == 1 {
		return []*DiffResult{engine.Compare(docsA[0], docsB[0])}
	}

	// Find optimal pairings for multiple documents
	type pairing struct {
		indexA int
		indexB int
		diff   *DiffResult
		cost   int
	}

	pairings := make([]pairing, 0, len(docsA)*len(docsB)+len(docsA)+len(docsB))

	// Compare all pairs
	for i, docA := range docsA {
		for j, docB := range docsB {
			diff := engine.Compare(docA, docB)
			cost := 0
			if diff.Meta != nil {
				cost = diff.Meta.DiffCount
			}
			pairings = append(pairings, pairing{
				indexA: i,
				indexB: j,
				diff:   diff,
				cost:   cost,
			})
		}
	}

	// Add pairings with nil
	for i := range docsA {
		diff := engine.Compare(docsA[i], nil)
		cost := 0
		if diff.Meta != nil {
			cost = diff.Meta.DiffCount
		}
		pairings = append(pairings, pairing{
			indexA: i,
			indexB: -1,
			diff:   diff,
			cost:   cost,
		})
	}

	for j := range docsB {
		diff := engine.Compare(nil, docsB[j])
		cost := 0
		if diff.Meta != nil {
			cost = diff.Meta.DiffCount
		}
		pairings = append(pairings, pairing{
			indexA: -1,
			indexB: j,
			diff:   diff,
			cost:   cost,
		})
	}

	// Sort by cost
	sort.Slice(pairings, func(i, j int) bool {
		return pairings[i].cost < pairings[j].cost
	})

	// Find optimal matching
	usedA := make(map[int]bool)
	usedB := make(map[int]bool)
	var results []*DiffResult

	for _, p := range pairings {
		aUsed := p.indexA >= 0 && usedA[p.indexA]
		bUsed := p.indexB >= 0 && usedB[p.indexB]

		if !aUsed && !bUsed {
			if p.indexA >= 0 {
				usedA[p.indexA] = true
			}
			if p.indexB >= 0 {
				usedB[p.indexB] = true
			}
			results = append(results, p.diff)
		}
	}

	return results
}

// shouldDoLineDiff determines if we should do line-by-line diff for multiline strings
func (e *DiffEngine) shouldDoLineDiff(a, b *StructuredData) bool {
	// Both must be strings
	if a.Type != TypeString || b.Type != TypeString {
		return false
	}

	aStr, ok := a.Value.(string)
	if !ok {
		return false
	}
	bStr, ok := b.Value.(string)
	if !ok {
		return false
	}

	// Check if either string is multiline
	aMultiline := strings.Contains(aStr, "\n")
	bMultiline := strings.Contains(bStr, "\n")
	
	// For multiline strings, do line-by-line diff
	// This helps with configuration files and similar content
	return aMultiline || bMultiline
}

// compareMultilineStrings compares multiline strings line by line
func (e *DiffEngine) compareMultilineStrings(a, b *StructuredData, path []string) *DiffResult {
	aStr := a.Value.(string)
	bStr := b.Value.(string)

	aLines := strings.Split(aStr, "\n")
	bLines := strings.Split(bStr, "\n")

	// Create a virtual array structure for line-by-line comparison
	aArray := &StructuredData{
		Type:     TypeArray,
		Elements: make([]*StructuredData, len(aLines)),
		Meta:     a.Meta,
	}
	bArray := &StructuredData{
		Type:     TypeArray,
		Elements: make([]*StructuredData, len(bLines)),
		Meta:     b.Meta,
	}

	for i, line := range aLines {
		aArray.Elements[i] = &StructuredData{
			Type:  TypeString,
			Value: line,
			Meta:  a.Meta,
		}
	}
	for i, line := range bLines {
		bArray.Elements[i] = &StructuredData{
			Type:  TypeString,
			Value: line,
			Meta:  b.Meta,
		}
	}

	// Use array comparison with index strategy to preserve line order
	oldStrategy := e.options.ArrayDiffStrategy
	e.options.ArrayDiffStrategy = ArrayStrategyIndex
	arrayResult := e.compareArrays(aArray, bArray, path)
	e.options.ArrayDiffStrategy = oldStrategy

	// Convert back to string result
	result := &DiffResult{
		Status: arrayResult.Status,
		Path:   path,
		From:   a,
		To:     b,
		Meta:   arrayResult.Meta,
	}

	// Add line-level children if there are differences
	if arrayResult.Status != StatusSame && len(arrayResult.Children) > 0 {
		result.Children = make([]*DiffResult, 0)
		for _, child := range arrayResult.Children {
			// Convert array index path to line number
			if len(child.Path) > 0 {
				lastPath := child.Path[len(child.Path)-1]
				if strings.HasPrefix(lastPath, "[") && strings.HasSuffix(lastPath, "]") {
					// Extract line number from [n] format
					lineNum := lastPath[1:len(lastPath)-1]
					child.Path[len(child.Path)-1] = "line " + lineNum
				}
			}
			result.Children = append(result.Children, child)
		}
	}

	return result
}
