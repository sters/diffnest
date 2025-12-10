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
	IgnoreKeyCase     bool
	IgnoreValueCase   bool
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
		equal := a.Value == b.Value
		if !equal && e.options.IgnoreValueCase {
			// Case-insensitive comparison
			aStr, aOk := a.Value.(string)
			bStr, bOk := b.Value.(string)
			if aOk && bOk {
				equal = strings.EqualFold(aStr, bStr)
			}
		}

		if equal {
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

	if e.options.IgnoreKeyCase {
		return e.compareObjectsIgnoreCase(a, b, path, result)
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

// compareObjectsIgnoreCase compares objects ignoring key case differences.
func (e *DiffEngine) compareObjectsIgnoreCase(a, b *StructuredData, path []string, result *DiffResult) *DiffResult {
	// Create case-insensitive key mappings
	aKeyMap := make(map[string]string) // lowercase -> original
	bKeyMap := make(map[string]string) // lowercase -> original

	for k := range a.Children {
		lowerK := strings.ToLower(k)
		aKeyMap[lowerK] = k
	}
	for k := range b.Children {
		lowerK := strings.ToLower(k)
		bKeyMap[lowerK] = k
	}

	// Collect all lowercase keys
	allLowerKeys := make(map[string]bool)
	for lowerK := range aKeyMap {
		allLowerKeys[lowerK] = true
	}
	for lowerK := range bKeyMap {
		allLowerKeys[lowerK] = true
	}

	// Sort lowercase keys for consistent output
	lowerKeys := make([]string, 0, len(allLowerKeys))
	for lowerK := range allLowerKeys {
		lowerKeys = append(lowerKeys, lowerK)
	}
	sort.Strings(lowerKeys)

	// Compare each key (case-insensitive)
	for _, lowerKey := range lowerKeys {
		originalKeyA, hasA := aKeyMap[lowerKey]
		originalKeyB, hasB := bKeyMap[lowerKey]

		var childA, childB *StructuredData
		var displayKey string

		if hasA {
			childA = a.Children[originalKeyA]
			displayKey = originalKeyA
		}
		if hasB {
			childB = b.Children[originalKeyB]
			if displayKey == "" {
				displayKey = originalKeyB
			}
		}

		// Apply ignore options
		if e.shouldIgnore(childA, hasA) && e.shouldIgnore(childB, hasB) {
			continue
		}

		childPath := append(append([]string{}, path...), displayKey)

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

// toFloat64 converts various numeric types to float64.
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

// toInt64 converts various integer types to int64.
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
		// #nosec G115 - values come from JSON/YAML parsing
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

// Compare compares multiple documents and finds optimal pairings using the Hungarian algorithm.
func Compare(docsA, docsB []*StructuredData, options DiffOptions) []*DiffResult {
	engine := NewDiffEngine(options)

	// If single documents, compare directly
	if len(docsA) == 1 && len(docsB) == 1 {
		return []*DiffResult{engine.Compare(docsA[0], docsB[0])}
	}

	// Build cost matrix for Hungarian algorithm
	// We need to handle the case where documents can be unmatched (deleted or added)
	// Matrix size: (len(docsA) + len(docsB)) x (len(docsA) + len(docsB))
	// This allows each doc from A to match with a doc from B, or be "deleted"
	// and each doc from B to be "added" if not matched
	n := len(docsA) + len(docsB)
	if n == 0 {
		return []*DiffResult{}
	}

	// Pre-compute all diff results and costs
	// diffResults[i][j] = diff between docsA[i] and docsB[j]
	diffResults := make([][]*DiffResult, len(docsA))
	costMatrix := make([][]int, n)

	for i := range n {
		costMatrix[i] = make([]int, n)
	}

	// Compute costs for matching docsA[i] with docsB[j]
	for i, docA := range docsA {
		diffResults[i] = make([]*DiffResult, len(docsB))
		for j, docB := range docsB {
			diff := engine.Compare(docA, docB)
			diffResults[i][j] = diff
			cost := 0
			if diff.Meta != nil {
				cost = diff.Meta.DiffCount
			}
			// Add penalty for mismatched Kubernetes-like resources
			// This considers apiVersion, kind, metadata.name, and metadata.namespace
			cost += calculateResourceMismatchPenalty(docA, docB)
			costMatrix[i][j] = cost
		}
	}

	// Compute costs for "deleting" docsA[i] (matching with dummy)
	deleteCosts := make([]int, len(docsA))
	deleteDiffs := make([]*DiffResult, len(docsA))
	for i, docA := range docsA {
		diff := engine.Compare(docA, nil)
		deleteDiffs[i] = diff
		if diff.Meta != nil {
			deleteCosts[i] = diff.Meta.DiffCount
		}
	}

	// Compute costs for "adding" docsB[j] (matching with dummy)
	addCosts := make([]int, len(docsB))
	addDiffs := make([]*DiffResult, len(docsB))
	for j, docB := range docsB {
		diff := engine.Compare(nil, docB)
		addDiffs[j] = diff
		if diff.Meta != nil {
			addCosts[j] = diff.Meta.DiffCount
		}
	}

	// Fill cost matrix
	// Upper-left quadrant is already filled above with kind penalty

	// Upper-right quadrant: docsA[i] matched with dummy (deletion)
	// docsA[i] can be deleted by matching with dummy slot len(docsB)+i
	for i := range len(docsA) {
		for j := len(docsB); j < n; j++ {
			if j-len(docsB) == i {
				costMatrix[i][j] = deleteCosts[i]
			} else {
				costMatrix[i][j] = 1 << 30 // Large cost to prevent invalid matching
			}
		}
	}

	// Lower-left quadrant: dummy matched with docsB[j] (addition)
	// docsB[j] can be added by matching with dummy slot len(docsA)+j
	for i := len(docsA); i < n; i++ {
		for j := range len(docsB) {
			if i-len(docsA) == j {
				costMatrix[i][j] = addCosts[j]
			} else {
				costMatrix[i][j] = 1 << 30 // Large cost to prevent invalid matching
			}
		}
	}

	// Lower-right quadrant: dummy matched with dummy (zero cost)
	for i := len(docsA); i < n; i++ {
		for j := len(docsB); j < n; j++ {
			costMatrix[i][j] = 0
		}
	}

	// Run Hungarian algorithm
	assignment := hungarianAlgorithm(costMatrix)

	// Build results from assignment
	var results []*DiffResult

	for i := range len(docsA) {
		j := assignment[i]
		if j < len(docsB) {
			// docsA[i] matched with docsB[j]
			results = append(results, diffResults[i][j])
		} else {
			// docsA[i] was deleted
			results = append(results, deleteDiffs[i])
		}
	}

	// Check for additions (docsB that weren't matched with any docsA)
	matchedB := make(map[int]bool)
	for i := range len(docsA) {
		j := assignment[i]
		if j < len(docsB) {
			matchedB[j] = true
		}
	}

	for j := range len(docsB) {
		if !matchedB[j] {
			results = append(results, addDiffs[j])
		}
	}

	return results
}

// hungarianAlgorithm implements the Hungarian algorithm for optimal assignment.
// Returns an assignment where assignment[i] is the column assigned to row i.
func hungarianAlgorithm(costMatrix [][]int) []int {
	n := len(costMatrix)
	if n == 0 {
		return []int{}
	}

	// Copy matrix to avoid modifying original
	cost := make([][]int, n)
	for i := range n {
		cost[i] = make([]int, n)
		copy(cost[i], costMatrix[i])
	}

	// u[i] = potential for row i, v[j] = potential for column j
	u := make([]int, n+1)
	v := make([]int, n+1)
	// p[j] = row assigned to column j (0 means unassigned, using 1-indexed internally)
	p := make([]int, n+1)
	// way[j] = previous column in augmenting path
	way := make([]int, n+1)

	for i := 1; i <= n; i++ {
		// Start augmenting path from row i
		p[0] = i
		j0 := 0 // Current column (0 is a virtual column)

		minv := make([]int, n+1)
		used := make([]bool, n+1)
		for j := range n + 1 {
			minv[j] = 1 << 30
			used[j] = false
		}

		// Find augmenting path
		for p[j0] != 0 {
			used[j0] = true
			i0 := p[j0]
			delta := 1 << 30
			j1 := 0

			for j := 1; j <= n; j++ {
				if !used[j] {
					// Calculate reduced cost
					cur := cost[i0-1][j-1] - u[i0] - v[j]
					if cur < minv[j] {
						minv[j] = cur
						way[j] = j0
					}
					if minv[j] < delta {
						delta = minv[j]
						j1 = j
					}
				}
			}

			// Update potentials
			for j := range n + 1 {
				if used[j] {
					u[p[j]] += delta
					v[j] -= delta
				} else {
					minv[j] -= delta
				}
			}

			j0 = j1
		}

		// Reconstruct assignment along augmenting path
		for j0 != 0 {
			j1 := way[j0]
			p[j0] = p[j1]
			j0 = j1
		}
	}

	// Build result (convert from 1-indexed to 0-indexed)
	assignment := make([]int, n)
	for j := 1; j <= n; j++ {
		if p[j] != 0 {
			assignment[p[j]-1] = j - 1
		}
	}

	return assignment
}

// calculateResourceMismatchPenalty calculates penalty for mismatched Kubernetes-like resources.
// It checks apiVersion, kind, metadata.name, and metadata.namespace.
// Returns 0 if all fields match, or a penalty value based on which fields differ.
func calculateResourceMismatchPenalty(a, b *StructuredData) int {
	if a == nil || b == nil {
		return 0
	}
	if a.Type != TypeObject || b.Type != TypeObject {
		return 0 // Not objects, no penalty
	}

	// Check if these look like Kubernetes resources (have kind field)
	_, hasKindA := a.Children["kind"]
	_, hasKindB := b.Children["kind"]
	if !hasKindA && !hasKindB {
		return 0 // Neither has kind, not K8s resources
	}

	penalty := 0

	// Check kind match (highest penalty - different resource types should never match)
	// This is the most important factor: we almost never want to match different kinds
	if !fieldValuesMatch(a, b, "kind") {
		penalty += 1 << 20 // Very high penalty for kind mismatch
	}

	// Check apiVersion match
	// Different apiVersions usually mean different resource versions
	if !fieldValuesMatch(a, b, "apiVersion") {
		penalty += 50 // Small penalty for apiVersion mismatch
	}

	// Check metadata.name match
	// Name differences should add a small penalty to prefer same-name matches
	// but not so large that it prevents matching different-named resources of the same kind
	if !metadataFieldValuesMatch(a, b, "name") {
		penalty += 10 // Very small penalty for name mismatch
	}

	// Check metadata.namespace match
	if !metadataFieldValuesMatch(a, b, "namespace") {
		penalty += 5 // Minimal penalty for namespace mismatch
	}

	return penalty
}

// fieldValuesMatch checks if a top-level field has the same value in both documents.
func fieldValuesMatch(a, b *StructuredData, fieldName string) bool {
	fieldA, hasA := a.Children[fieldName]
	fieldB, hasB := b.Children[fieldName]

	if !hasA && !hasB {
		return true // Neither has the field
	}
	if !hasA || !hasB {
		return false // Only one has the field
	}
	if fieldA.Type != fieldB.Type {
		return false
	}

	return fieldA.Value == fieldB.Value
}

// metadataFieldValuesMatch checks if a field under metadata has the same value in both documents.
func metadataFieldValuesMatch(a, b *StructuredData, fieldName string) bool {
	metaA, hasMetaA := a.Children["metadata"]
	metaB, hasMetaB := b.Children["metadata"]

	if !hasMetaA && !hasMetaB {
		return true // Neither has metadata
	}
	if !hasMetaA || !hasMetaB {
		return false // Only one has metadata
	}
	if metaA.Type != TypeObject || metaB.Type != TypeObject {
		return false
	}

	return fieldValuesMatch(metaA, metaB, fieldName)
}

// shouldDoLineDiff determines if we should do line-by-line diff for multiline strings.
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

// compareMultilineStrings compares multiline strings line by line.
func (e *DiffEngine) compareMultilineStrings(a, b *StructuredData, path []string) *DiffResult {
	aStr, ok := a.Value.(string)
	if !ok {
		return nil
	}
	bStr, ok := b.Value.(string)
	if !ok {
		return nil
	}

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
					lineNum := lastPath[1 : len(lastPath)-1]
					child.Path[len(child.Path)-1] = "line " + lineNum
				}
			}
			result.Children = append(result.Children, child)
		}
	}

	return result
}
