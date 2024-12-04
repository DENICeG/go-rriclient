package highlight

import (
	"bytes"

	"github.com/alecthomas/chroma/v2/quick"
)

const (
	// YAML Format needed for the lexer to enable highlighting
	YAML Format = "yaml"
	// XML Format needed for the lexer to enable highlighting
	XML Format = "xml"
)

// Format Type for the lexer to enable highlighting
type Format string

// Transform function to highlight an input string. Can return error
func Transform(input string, format Format) (string, error) {
	var buf []byte
	buffer := bytes.NewBuffer(buf)

	err := quick.Highlight(buffer, input, string(format), "terminal16m", "monokai")
	if err != nil {
		return "", err
	}

	result := buffer.String()
	return result, nil
}
