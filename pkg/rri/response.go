package rri

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	// ResultSuccess denotes a successful result.
	ResultSuccess Result = "success"
	// ResultFailure denotes a failed result.
	ResultFailure Result = "failure"

	// ResponseFieldNameResult denotes the response field name for result.
	ResponseFieldNameResult ResponseFieldName = "RESULT"
	// ResponseFieldNameSTID denotes the response field name for STID.
	ResponseFieldNameSTID ResponseFieldName = "STID"
	// ResponseFieldNameInfo denotes the response field name for info message.
	ResponseFieldNameInfo ResponseFieldName = "INFO"
	// ResponseFieldNameError denotes the response field name for error message.
	ResponseFieldNameError ResponseFieldName = "ERROR"
	// ResponseFieldNameWarning denotes the response field name for warning message.
	ResponseFieldNameWarning ResponseFieldName = "WARNING"

	// ResponseEntityNameHolder denotes the entity name of a holder.
	ResponseEntityNameHolder ResponseEntityName = "holder"
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

// BusinessMessage represents a response message with id as returned in INFO and ERROR.
type BusinessMessage struct {
	message string
	id      int64
}

// NewBusinessMessage creates a new BusinessMessage with id and message.
func NewBusinessMessage(id int64, msg string) BusinessMessage {
	return BusinessMessage{id: id, message: msg}
}

// ID returns the message id.
func (bm BusinessMessage) ID() int64 {
	return bm.id
}

// Message returns the message string.
func (bm BusinessMessage) Message() string {
	return bm.message
}

// String returns both ID and Message.
func (bm BusinessMessage) String() string {
	return fmt.Sprintf("%v %s", bm.id, bm.message)
}

// ResponseEntityName represents a response entity name.
type ResponseEntityName string

// Normalize returns the normalized representation of the given ResponseEntityName.
func (r ResponseEntityName) Normalize() ResponseEntityName {
	return ResponseEntityName(strings.ToLower(string(r)))
}

// ResponseEntity is used to encapsulate further information.
type ResponseEntity struct {
	name   ResponseEntityName
	fields ResponseFieldList
}

// Response represents an RRI response.
type Response struct {
	fields   ResponseFieldList
	entities []ResponseEntity
}

// IsSuccessful returns whether the response is successfull.
func (r *Response) IsSuccessful() bool {
	return r.Result() == ResultSuccess
}

// Result returns the returned result.
func (r *Response) Result() Result {
	return Result(r.FirstField(ResponseFieldNameResult)).Normalize()
}

// InfoMessages returns all info messages.
func (r *Response) InfoMessages() []BusinessMessage {
	// ignore parse errors here, should be accounted for during response parsing
	messages := make([]BusinessMessage, 0)
	for _, msg := range r.Field(ResponseFieldNameInfo) {
		bm, _ := ParseBusinessMessageKV(msg)
		messages = append(messages, bm)
	}
	return messages
}

// ErrorMessages returns all error messages.
func (r *Response) ErrorMessages() []BusinessMessage {
	// ignore parse errors here, should be accounted for during response parsing
	messages := make([]BusinessMessage, 0)
	for _, msg := range r.Field(ResponseFieldNameError) {
		bm, _ := ParseBusinessMessageKV(msg)
		messages = append(messages, bm)
	}
	return messages
}

// WarningMessages returns all warning messages.
func (r *Response) WarningMessages() []BusinessMessage {
	// ignore parse errors here, should be accounted for during response parsing
	messages := make([]BusinessMessage, 0)
	for _, msg := range r.Field(ResponseFieldNameWarning) {
		bm, _ := ParseBusinessMessageKV(msg)
		messages = append(messages, bm)
	}
	return messages
}

// STID return the server transaction id.
func (r *Response) STID() string {
	return r.FirstField(ResponseFieldNameSTID)
}

// String returns a human readable representation of the response.
func (r *Response) String() string {
	//TODO shortened, single line representation
	return r.EncodeKV()
}

// EncodeKV returns the Key-Value representation as used for RRI communication.
func (r *Response) EncodeKV() string {
	var sb strings.Builder
	for _, f := range r.fields {
		if sb.Len() > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(string(f.Name))
		sb.WriteString(": ")
		sb.WriteString(f.Value)
	}
	//TODO encode entities
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

// Entities returns a list of entities contained in this response.
func (r *Response) Entities() []ResponseEntity {
	return r.entities
}

// Name return the name of this response entity.
func (e *ResponseEntity) Name() ResponseEntityName {
	return e.name
}

// Fields returns all additional response entity fields.
func (e *ResponseEntity) Fields() ResponseFieldList {
	return e.fields
}

// Field returns all values defined for a field name.
func (e *ResponseEntity) Field(fieldName ResponseFieldName) []string {
	return e.fields.Values(fieldName)
}

// FirstField returns the first field value or an empty string for a field name.
func (e *ResponseEntity) FirstField(fieldName ResponseFieldName) string {
	return e.fields.FirstValue(fieldName)
}

// NewResponse returns a new Response with the given result code.
func NewResponse(result Result, fields ResponseFieldList) *Response {
	newFields := NewResponseFieldList()
	newFields.Add(ResponseFieldNameResult, string(result.Normalize()))
	if fields != nil {
		fields.CopyTo(&newFields)
	}
	return &Response{newFields, nil}
}

// NewResponseWithInfo returns a new Response with the given result code and attached info messages.
func NewResponseWithInfo(result Result, fields ResponseFieldList, infos ...BusinessMessage) *Response {
	newFields := NewResponseFieldList()
	newFields.Add(ResponseFieldNameResult, string(result.Normalize()))
	if fields != nil {
		fields.CopyTo(&newFields)
	}
	for _, msg := range infos {
		newFields.Add(ResponseFieldNameInfo, msg.String())
	}
	return &Response{newFields, nil}
}

// NewResponseWithError returns a new Response with the given result code and attached error messages.
func NewResponseWithError(result Result, fields ResponseFieldList, errors ...BusinessMessage) *Response {
	newFields := NewResponseFieldList()
	newFields.Add(ResponseFieldNameResult, string(result.Normalize()))
	if fields != nil {
		fields.CopyTo(&newFields)
	}
	for _, msg := range errors {
		newFields.Add(ResponseFieldNameError, msg.String())
	}
	return &Response{newFields, nil}
}

// ParseResponseKV parses a response object from the given key-value response string.
func ParseResponseKV(msg string) (*Response, error) {
	lines := strings.Split(msg, "\n")
	fields := NewResponseFieldList()
	entities := make([]ResponseEntity, 0)
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) > 0 {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 1 && strings.HasPrefix(parts[0], "[") && strings.HasSuffix(parts[0], "]") {
				// begin of new entity
				entities = append(entities, ResponseEntity{ResponseEntityName(parts[0][1 : len(parts[0])-1]).Normalize(), NewResponseFieldList()})
				continue
			}
			if len(parts) != 2 {
				return nil, fmt.Errorf("malformed key-value pair in line %d", i)
			}

			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			if len(entities) > 0 {
				entities[len(entities)-1].fields.Add(ResponseFieldName(key), value)
			} else {
				fields.Add(ResponseFieldName(key), value)
			}
		}
	}

	resultValues := fields.Values(ResponseFieldNameResult)
	if len(resultValues) == 0 {
		return nil, fmt.Errorf("%s key is missing", ResponseFieldNameResult)
	}
	if len(resultValues) > 1 {
		return nil, fmt.Errorf("multiple %s values", ResponseFieldNameResult)
	}

	stidValues := fields.Values(ResponseFieldNameSTID)
	if len(stidValues) > 1 {
		return nil, fmt.Errorf("multiple %s values", ResponseFieldNameSTID)
	}

	for _, msg := range fields.Values(ResponseFieldNameInfo) {
		if _, err := ParseBusinessMessageKV(msg); err != nil {
			return nil, fmt.Errorf("invalid info message: %s", err.Error())
		}
	}

	for _, msg := range fields.Values(ResponseFieldNameError) {
		if _, err := ParseBusinessMessageKV(msg); err != nil {
			return nil, fmt.Errorf("invalid error message: %s", err.Error())
		}
	}

	return &Response{fields, entities}, nil
}

// ParseBusinessMessageKV parses a BusinessMessage from a single KV entry.
func ParseBusinessMessageKV(str string) (BusinessMessage, error) {
	parts := strings.SplitN(str, " ", 2)
	if len(parts) != 2 {
		return BusinessMessage{}, fmt.Errorf("missing id or message part")
	}
	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return BusinessMessage{}, err
	}
	return BusinessMessage{id: id, message: parts[1]}, nil
}

// ParseResponse tries to detect the response format (KV or XML) and returns the parsed response.
func ParseResponse(str string) (*Response, error) {
	//TODO detect type
	return ParseResponseKV(str)
}

// ExtractVerificationInformation extracts the VerificationInformation from the Response.
func (r *Response) ExtractVerificationInformation() ([]*VerificationInformation, error) {
	var verificationInformation []*VerificationInformation
	for _, eachEntity := range r.Entities() {
		if string(eachEntity.Name().Normalize()) == string(QueryEntityVerificationInformation.Normalize()) {
			extractedVerificationInformation, extractErr := eachEntity.ExtractVerificationInformation()
			if extractErr != nil {
				return nil, extractErr
			}
			verificationInformation = append(verificationInformation, extractedVerificationInformation)
		}
	}
	return verificationInformation, nil
}

// ExtractVerificationInformation extracts the VerificationInformation from a ResponseEntity.
func (e *ResponseEntity) ExtractVerificationInformation() (*VerificationInformation, error) {
	responseFieldClaims := e.Field(ResponseFieldName(QueryFieldNameVerifiedClaim))
	verificationClaims := make([]VerificationClaim, len(responseFieldClaims))
	for i := range responseFieldClaims {
		stringClaim, claimErr := ParseVerificationClaim(responseFieldClaims[i])
		if claimErr != nil {
			return nil, claimErr
		}
		verificationClaims[i] = stringClaim
	}

	responseFieldTimestamp := e.FirstField(ResponseFieldName(QueryFieldNameVerificationTimestamp))
	parsedTime, parsedTimeErr := time.Parse(VerificationInformationTimestampFormat, responseFieldTimestamp)
	if parsedTimeErr != nil {
		return nil, fmt.Errorf("error while parsing verification timestamp %v: %v", responseFieldTimestamp, parsedTimeErr)
	}

	verificationResult, verificationResultErr := ParseVerificationResult(e.FirstField(ResponseFieldName(QueryFieldNameVerificationResult)))
	if verificationResultErr != nil {
		return nil, verificationResultErr
	}

	verificationEvidence, verificationEvidenceErr := ParseVerificationEvidence(e.FirstField(ResponseFieldName(QueryFieldNameVerificationEvidence)))
	if verificationEvidenceErr != nil {
		return nil, verificationEvidenceErr
	}

	verificationMethod, verificationMethodErr := ParseVerificationMethod(e.FirstField(ResponseFieldName(QueryFieldNameVerificationMethod)))
	if verificationMethodErr != nil {
		return nil, verificationMethodErr
	}

	trustFramework, trustFrameworkErr := ParseTrustFramework(e.FirstField(ResponseFieldName(QueryFieldNameTrustFramework)))
	if trustFrameworkErr != nil {
		return nil, trustFrameworkErr
	}

	return &VerificationInformation{
		VerifiedClaim:         verificationClaims,
		VerificationResult:    verificationResult,
		VerificationReference: e.FirstField(ResponseFieldName(QueryFieldNameVerificationReference)),
		VerificationTimestamp: parsedTime,
		VerificationEvidence:  verificationEvidence,
		VerificationMethod:    verificationMethod,
		TrustFramework:        trustFramework,
	}, nil
}
