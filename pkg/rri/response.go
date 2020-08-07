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

	// EntityNameHolder denotes the entity name of a holder.
	EntityNameHolder ResponseEntityName = "holder"
)

// Result represents the result of a query response.
type Result string

// Normalize returns the normalized representation of the given Result.
func (r Result) Normalize() Result {
	return Result(strings.ToLower(string(r)))
}

// ResponseFieldName represents the field name of a query response.
type ResponseFieldName string

// Normalize returns the normalized representation of the given ResponseFieldName.
func (r ResponseFieldName) Normalize() ResponseFieldName {
	return ResponseFieldName(strings.ToUpper(string(r)))
}

// ResponseEntityName represents a response entity name.
type ResponseEntityName string

// Normalize returns the normalized representation of the given ResponseEntityName.
func (r ResponseEntityName) Normalize() ResponseEntityName {
	return ResponseEntityName(strings.ToLower(string(r)))
}

type responseEntity struct {
	Name   ResponseEntityName
	Fields ResponseFieldList
}

// Response represents an RRI response.
type Response struct {
	result   Result
	stid     string
	infoMsg  string
	errorMsg string
	fields   ResponseFieldList
	entities []responseEntity
}

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

// String returns a human readable representation of the response.
func (r *Response) String() string {
	//TODO shortened, single line representation
	return r.EncodeKV()
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
func (r *Response) Fields() ResponseFieldList {
	return r.fields
}

// Field returns all values defined for a field name.
func (r *Response) Field(fieldName ResponseFieldName) []string {
	return r.fields.Values(fieldName)
}

// FirstField returns the first field value or an empty string for a field name.
func (r *Response) FirstField(fieldName ResponseFieldName) string {
	return r.fields.FirstValue(fieldName)
}

// EntityNames returns a list of entity names contained in this response.
func (r *Response) EntityNames() []ResponseEntityName {
	names := make([]ResponseEntityName, len(r.entities))
	for i, e := range r.entities {
		names[i] = e.Name
	}
	return names
}

// Entity returns the named entity of this response or nil.
func (r *Response) Entity(entityName ResponseEntityName) ResponseFieldList {
	entityName = entityName.Normalize()
	for _, e := range r.entities {
		if e.Name == entityName {
			return e.Fields
		}
	}
	return nil
}

// ParseResponseKV parses a response object from the given key-value response string.
func ParseResponseKV(msg string) (*Response, error) {
	lines := strings.Split(msg, "\n")
	fields := newResponseFieldList()
	entities := make([]responseEntity, 0)
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) > 0 {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 1 && strings.HasPrefix(parts[0], "[") && strings.HasSuffix(parts[0], "]") {
				// begin of new entity
				entities = append(entities, responseEntity{ResponseEntityName(parts[0][1 : len(parts[0])-1]).Normalize(), newResponseFieldList()})
				continue
			}
			if len(parts) != 2 {
				return nil, fmt.Errorf("malformed query in line %d", i)
			}

			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			if len(entities) > 0 {
				entities[len(entities)-1].Fields.Add(ResponseFieldName(key), value)
			} else {
				fields.Add(ResponseFieldName(key), value)
			}
		}
	}

	resultValues := fields.Values(FieldNameResult)
	if len(resultValues) == 0 {
		return nil, fmt.Errorf("%s key is missing", FieldNameResult)
	}
	if len(resultValues) > 1 {
		return nil, fmt.Errorf("multiple %s values", FieldNameResult)
	}
	fields.RemoveAll(FieldNameResult)

	stid := ""
	stidValues := fields.Values(FieldNameSTID)
	if len(stidValues) > 0 {
		stid = stidValues[0]
	}
	fields.RemoveAll(FieldNameSTID)

	infoMsg := ""
	infoMsgValues := fields.Values(FieldNameInfoMsg)
	if len(infoMsgValues) > 0 {
		infoMsg = infoMsgValues[0]
	}
	fields.RemoveAll(FieldNameInfoMsg)

	errorMsg := ""
	errorMsgValues := fields.Values(FieldNameErrorMsg)
	if len(errorMsgValues) > 0 {
		errorMsg = errorMsgValues[0]
	}
	fields.RemoveAll(FieldNameErrorMsg)

	return &Response{Result(resultValues[0]).Normalize(), stid, infoMsg, errorMsg, fields, entities}, nil
}

// ParseResponse tries to detect the response format (KV or XML) and returns the parsed response.
func ParseResponse(str string) (*Response, error) {
	//TODO detect type
	return ParseResponseKV(str)
}
