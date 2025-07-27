package diffnest

import (
	"fmt"
	"io"
)

// Controller handles the core diff logic.
type Controller struct {
	reader1   io.Reader
	reader2   io.Reader
	format1   string
	format2   string
	diffOpts  DiffOptions
	formatter Formatter
	writer    io.Writer
}

// NewController creates a new Controller.
func NewController(reader1, reader2 io.Reader, format1, format2 string, diffOpts DiffOptions, formatter Formatter, writer io.Writer) *Controller {
	return &Controller{
		reader1:   reader1,
		reader2:   reader2,
		format1:   format1,
		format2:   format2,
		diffOpts:  diffOpts,
		formatter: formatter,
		writer:    writer,
	}
}

// Run executes the diff process and returns whether differences were found.
func (c *Controller) Run() (bool, error) {
	docs1, err := ParseWithFormat(c.reader1, c.format1)
	if err != nil {
		return false, fmt.Errorf("error parsing first file: %w", err)
	}

	docs2, err := ParseWithFormat(c.reader2, c.format2)
	if err != nil {
		return false, fmt.Errorf("error parsing second file: %w", err)
	}

	results := Compare(docs1, docs2, c.diffOpts)

	if err := c.formatter.Format(c.writer, results); err != nil {
		return false, fmt.Errorf("error formatting output: %w", err)
	}

	return HasDifferences(results), nil
}

// HasDifferences checks if there are any differences in the results.
func HasDifferences(results []*DiffResult) bool {
	for _, result := range results {
		if result.Status != StatusSame {
			return true
		}
	}

	return false
}
