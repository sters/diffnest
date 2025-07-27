package diffnest

// DataType represents the type of structured data.
type DataType int

const (
	TypeNull DataType = iota
	TypeBool
	TypeNumber
	TypeString
	TypeArray
	TypeObject
)

// StructuredData represents format-agnostic structured data.
type StructuredData struct {
	Type     DataType
	Value    any                        // Actual value for primitives
	Children map[string]*StructuredData // For objects
	Elements []*StructuredData          // For arrays
	Meta     *Metadata                  // Format-specific metadata
}

// Metadata contains format-specific information.
type Metadata struct {
	Format       string        // "json", "yaml", "toml"
	Location     *Location     // Position in source file
	Comments     []string      // Comments (YAML/TOML)
	StringStyle  StringStyle   // Style of string representation (for YAML)
}

// StringStyle represents YAML string representation style.
type StringStyle int

const (
	StringStyleUnknown StringStyle = iota
	StringStyleQuoted              // "string" or 'string'
	StringStyleLiteral             // |
	StringStyleFolded              // >
	StringStylePlain               // no quotes
)

// Location represents position in source file.
type Location struct {
	Line   int
	Column int
}

// DiffStatus represents the status of a diff.
type DiffStatus int

const (
	StatusSame DiffStatus = iota
	StatusModified
	StatusAdded
	StatusDeleted
)

// DiffResult represents the result of comparing two structures.
type DiffResult struct {
	Status   DiffStatus
	Path     []string // Path to this element
	From     *StructuredData
	To       *StructuredData
	Children []*DiffResult // For nested structures
	Meta     *DiffMeta
}

// DiffMeta contains additional diff information.
type DiffMeta struct {
	DiffCount int // Size of the difference
	Note      string
}
