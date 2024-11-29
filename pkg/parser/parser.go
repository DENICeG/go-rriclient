package parser

import (
	"strings"

	"github.com/DENICeG/go-rriclient/pkg/rri"
)

// ParseQueriesKV parses multiple queries separated by a =-= line from a string.
func ParseQueriesKV(str string) ([]*rri.Query, error) {
	lines := strings.Split(str, "\n")

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
		if strings.HasPrefix(line, "=-=") {
			appendQueryString(sb.String())
			sb.Reset()
		} else {
			sb.WriteString(line)
			sb.WriteString("\n")
		}
	}
	if sb.Len() > 0 {
		appendQueryString(sb.String())
	}

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
