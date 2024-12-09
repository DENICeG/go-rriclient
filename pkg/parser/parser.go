package parser

import (
	"bytes"
	"strings"

	"github.com/DENICeG/go-rriclient/pkg/rri"
)

// ParseQueriesKV parses multiple queries separated by a =-= line from a string.
func ParseQueriesKV(queryStrings []string) ([]*rri.Query, error) {
	queries := make([]*rri.Query, len(queryStrings))

	for i, queryString := range queryStrings {
		query, err := rri.ParseQueryKV(strings.TrimSpace(queryString))
		if err != nil {
			return nil, err
		}

		queries[i] = query
	}

	return queries, nil
}

// SplitLines splits a byte slice into lines.
func SplitLines(input []byte) [][]byte {
	return bytes.Split(input, []byte("\n"))
}

// SplitQueries splits a slice of lines into individual queries.
// query divider is the character combination of: =-=
func SplitQueries(lines [][]byte) []string {
	// each string in queryStrings contains a single, unparsed query
	queryStrings := make([]string, 0)
	appendQueryString := func(str string) {
		str = strings.TrimSpace(str)
		if len(str) > 0 {
			queryStrings = append(queryStrings, str)
		}
	}

	// separate at lines beginning with =-=
	var sb strings.Builder
	for _, line := range lines {
		stringLine := string(line)
		if strings.HasPrefix(stringLine, "=-=") {
			appendQueryString(sb.String())
			sb.Reset()
		} else {
			sb.WriteString(stringLine)
			sb.WriteString("\n")
		}
	}
	if sb.Len() > 0 {
		appendQueryString(sb.String())
	}

	return queryStrings
}
