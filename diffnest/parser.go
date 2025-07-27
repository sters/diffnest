package diffnest

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/goccy/go-yaml"
)

// Format constants.
const (
	FormatJSON = "json"
	FormatYAML = "yaml"
	FormatTOML = "toml"
)

// Errors.
var (
	ErrUnsupportedFormat = errors.New("unsupported format")
)

// Parser interface for different formats.
type Parser interface {
	Parse(reader io.Reader) ([]*StructuredData, error)
	Format() string
}

// DetectFormatFromFilename detects the format from filename extension.
func DetectFormatFromFilename(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".json":
		return FormatJSON
	case ".yaml", ".yml":
		return FormatYAML
	case ".toml":
		return FormatTOML
	default:
		return FormatYAML
	}
}

// ParseWithFormat parses content from reader with specified format.
func ParseWithFormat(reader io.Reader, format string) ([]*StructuredData, error) {
	var parser Parser
	switch format {
	case FormatJSON:
		parser = &JSONParser{}
	case FormatYAML:
		parser = &YAMLParser{}
	case FormatTOML:
		return nil, fmt.Errorf("%w: TOML parser not implemented yet", ErrUnsupportedFormat)
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedFormat, format)
	}

	result, err := parser.Parse(reader)
	if err != nil {
		return nil, fmt.Errorf("parse error: %w", err)
	}

	return result, nil
}

// JSONParser implements Parser for JSON.
type JSONParser struct{}

func (p *JSONParser) Format() string {
	return FormatJSON
}

func (p *JSONParser) Parse(reader io.Reader) ([]*StructuredData, error) {
	decoder := json.NewDecoder(reader)
	results := make([]*StructuredData, 0, 1)

	for {
		var raw any
		err := decoder.Decode(&raw)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to decode JSON: %w", err)
		}

		structured := convertToStructured(raw, "json")
		results = append(results, structured)
	}

	return results, nil
}

// YAMLParser implements Parser for YAML.
type YAMLParser struct{}

func (p *YAMLParser) Format() string {
	return FormatYAML
}

func (p *YAMLParser) Parse(reader io.Reader) ([]*StructuredData, error) {
	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read content: %w", err)
	}

	docs := strings.Split(string(content), "\n---\n")
	results := make([]*StructuredData, 0, len(docs))

	for _, doc := range docs {
		doc = strings.TrimSpace(doc)
		if doc == "" {
			continue
		}

		var raw any
		err := yaml.Unmarshal([]byte(doc), &raw)
		if err != nil {
			return nil, fmt.Errorf("failed to decode YAML: %w", err)
		}

		structured := convertToStructured(raw, "yaml")
		results = append(results, structured)
	}

	return results, nil
}

// convertToStructured converts raw data to StructuredData.
func convertToStructured(raw any, format string) *StructuredData {
	if raw == nil {
		return &StructuredData{
			Type: TypeNull,
			Meta: &Metadata{Format: format},
		}
	}

	switch v := raw.(type) {
	case bool:
		return &StructuredData{
			Type:  TypeBool,
			Value: v,
			Meta:  &Metadata{Format: format},
		}

	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64:
		return &StructuredData{
			Type:  TypeNumber,
			Value: v,
			Meta:  &Metadata{Format: format},
		}

	case string:
		return &StructuredData{
			Type:  TypeString,
			Value: v,
			Meta:  &Metadata{Format: format},
		}

	case []any:
		elements := make([]*StructuredData, len(v))
		for i, elem := range v {
			elements[i] = convertToStructured(elem, format)
		}

		return &StructuredData{
			Type:     TypeArray,
			Elements: elements,
			Meta:     &Metadata{Format: format},
		}

	case map[string]any:
		children := make(map[string]*StructuredData)
		for key, val := range v {
			children[key] = convertToStructured(val, format)
		}

		return &StructuredData{
			Type:     TypeObject,
			Children: children,
			Meta:     &Metadata{Format: format},
		}

	default:
		if ms, ok := v.(yaml.MapSlice); ok {
			children := make(map[string]*StructuredData)
			for _, item := range ms {
				if key, ok := item.Key.(string); ok {
					children[key] = convertToStructured(item.Value, format)
				}
			}

			return &StructuredData{
				Type:     TypeObject,
				Children: children,
				Meta:     &Metadata{Format: format},
			}
		}

		return &StructuredData{
			Type:  TypeString,
			Value: fmt.Sprint(v),
			Meta:  &Metadata{Format: format},
		}
	}
}
