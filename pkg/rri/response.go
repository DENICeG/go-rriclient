package rri

import (
	"fmt"
	"strings"
)

const (
	// ResultSuccess denotes a successful result.
	ResultSuccess Result = "success"
	// ResultFailure dontes a failed result.
	ResultFailure Result = "failure"

	// FieldNameResult denotes the response field name for result.
	FieldNameResult ResponseFieldName = "RESULT"
	// FieldNameSTID denotes the response field name for STID.
	FieldNameSTID ResponseFieldName = "STID"
	// FieldNameInfoMsg denotes the response field name for info message.
	FieldNameInfoMsg ResponseFieldName = "INFO"
	// FieldNameErrorMsg denotes the response field name for error message.
	FieldNameErrorMsg ResponseFieldName = "ERROR"
)

// Response represents an RRI response.
type Response struct {
	result   Result
	stid     string
	infoMsg  string
	errorMsg string
	fields   map[ResponseFieldName][]string
}

// Result represents the result of a query response.
type Result string

// ResponseFieldName represents the field name of a query response.
type ResponseFieldName string

// IsSuccessful returns whether the response is successfull.
func (r *Response) IsSuccessful() bool {
	return r.result == ResultSuccess
}

// Result returns the returned result.
func (r *Response) Result() Result {
	return r.result
}

// InfoMsg returns the info message.
func (r *Response) InfoMsg() string {
	return r.infoMsg
}

// ErrorMsg returns the error message.
func (r *Response) ErrorMsg() string {
	return r.errorMsg
}

// EncodeKV returns the Key-Value representation as used for RRI communication.
func (r *Response) EncodeKV() string {
	var sb strings.Builder
	sb.WriteString(string(FieldNameResult))
	sb.WriteString(": ")
	sb.WriteString(string(r.result))
	return sb.String()
}

// Fields returns all additional response fields.
func (r *Response) Fields() map[ResponseFieldName][]string {
	return r.fields
}

// Field returns all values defined for a field name.
func (r *Response) Field(fieldName ResponseFieldName) []string {
	fieldValues, ok := r.fields[fieldName]
	if !ok {
		return []string{}
	}
	return fieldValues
}

// FirstField returns the first field value or an empty string for a field name.
func (r *Response) FirstField(fieldName ResponseFieldName) string {
	fieldValues, ok := r.fields[fieldName]
	if !ok || len(fieldValues) == 0 {
		return ""
	}
	return fieldValues[0]
}

// ParseResponseKV parses a response object from the given key-value response string.
func ParseResponseKV(msg string) (*Response, error) {
	lines := strings.Split(msg, "\n")
	fields := make(map[ResponseFieldName][]string)
	for _, line := range lines {
		if len(line) > 0 {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) != 2 {
				// additional information follows here -> abort parsing
				break
				//fmt.Println(line)
				//return nil, fmt.Errorf("response line must be key-value separated by ':'")
			}

			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			fieldValues, ok := fields[ResponseFieldName(key)]
			if !ok {
				fieldValues = []string{value}
			} else {
				fieldValues = append(fieldValues, value)
			}

			fields[ResponseFieldName(key)] = fieldValues
		}
	}

	resultValues, ok := fields[FieldNameResult]
	if !ok || len(resultValues) == 0 {
		return nil, fmt.Errorf("%s key is missing", FieldNameResult)
	}
	if len(resultValues) > 1 {
		return nil, fmt.Errorf("multiple %s values", FieldNameResult)
	}
	delete(fields, FieldNameResult)

	stid := ""
	stidValues, ok := fields[FieldNameSTID]
	if ok {
		if len(stidValues) > 0 {
			stid = stidValues[0]
		}
		delete(fields, FieldNameSTID)
	}

	infoMsg := ""
	infoMsgValues, ok := fields[FieldNameInfoMsg]
	if ok {
		if len(infoMsgValues) > 0 {
			infoMsg = infoMsgValues[0]
		}
		delete(fields, FieldNameInfoMsg)
	}

	errorMsg := ""
	errorMsgValues, ok := fields[FieldNameErrorMsg]
	if ok {
		if len(errorMsgValues) > 0 {
			errorMsg = errorMsgValues[0]
		}
		delete(fields, FieldNameErrorMsg)
	}

	return &Response{Result(resultValues[0]), stid, infoMsg, errorMsg, fields}, nil
}
