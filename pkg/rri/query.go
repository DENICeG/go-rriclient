package rri

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"golang.org/x/net/idna"
)

const (
	// LatestVersion denotes the latest RRI version supported by the client.
	LatestVersion Version = "3.0"

	// FieldNameVersion denotes the query field name for version.
	FieldNameVersion QueryFieldName = "version"
	// FieldNameAction denotes the query field name for action.
	FieldNameAction QueryFieldName = "action"
	// FieldNameUser denotes the query field name for login user.
	FieldNameUser QueryFieldName = "user"
	// FieldNamePassword denotes the query field name for login password.
	FieldNamePassword QueryFieldName = "password"
	// FieldNameDomainIDN denotes the query field name for IDN domain name.
	FieldNameDomainIDN QueryFieldName = "domain"
	// FieldNameDomainACE denotes the query field name for ACE domain name.
	FieldNameDomainACE QueryFieldName = "domain-ace"
	// FieldNameHolder denotes the query field name for holder handle.
	FieldNameHolder QueryFieldName = "holder"
	// FieldNameAbuseContact denotes the query field name for abuse contact handle.
	FieldNameAbuseContact QueryFieldName = "abusecontact"
	// FieldNameGeneralRequest denotes the query field name for general request handle.
	FieldNameGeneralRequest QueryFieldName = "generalrequest"
	// FieldNameNameServer denotes the query field name for name servers.
	FieldNameNameServer QueryFieldName = "nserver"
	// FieldNameHandle denotes the query field name for denic handles.
	FieldNameHandle QueryFieldName = "handle"
	// FieldNameAuthInfoHash denotes the query field name for auth info hash.
	FieldNameAuthInfoHash QueryFieldName = "authinfohash"
	// FieldNameAuthInfo denotes the query field name for auth info hash.
	FieldNameAuthInfo QueryFieldName = "authinfo"

	// ActionLogin denotes the action value for login.
	ActionLogin QueryAction = "LOGIN"
	// ActionLogout denotes the action value for logout.
	ActionLogout QueryAction = "LOGOUT"
	// ActionCheck denotes the action value for check.
	ActionCheck QueryAction = "CHECK"
	// ActionInfo denotes the action value for info.
	ActionInfo QueryAction = "INFO"
	// ActionCreate denotes the action value for create.
	ActionCreate QueryAction = "CREATE"
	// ActionDelete deontes the action value for delete.
	ActionDelete QueryAction = "DELETE"
	// ActionRestore deontes the action value for restore.
	ActionRestore QueryAction = "RESTORE"
	// ActionCreateAuthInfo1 denotes the action value for create AuthInfo1.
	ActionCreateAuthInfo1 QueryAction = "CREATE-AUTHINFO1"
	// ActionUpdate denotes the action value for update.
	ActionUpdate QueryAction = "UPDATE"
	// ActionChangeProvider denotes the action value for change provider.
	ActionChangeProvider QueryAction = "CHPROV"
	// ActionCreateAuthInfo2 denotes the action value for create AuthInfo2.
	ActionCreateAuthInfo2 QueryAction = "CREATE-AUTHINFO2"
)

var (
	handleFields = []QueryFieldName{FieldNameHolder, FieldNameAbuseContact, FieldNameGeneralRequest, FieldNameHandle}
)

// Version represents the RRI protocol version.
type Version string

// Normalize returns the normalized representation of the given Version.
func (v Version) Normalize() Version {
	return v
}

// QueryAction represents the action of a RRI query.
type QueryAction string

// Normalize returns the normalized representation of the given QueryAction.
func (q QueryAction) Normalize() QueryAction {
	return QueryAction(strings.ToUpper(string(q)))
}

// QueryFieldName represents a single data field of a query.
type QueryFieldName string

// Normalize returns the normalized representation of the given QueryFieldName.
func (q QueryFieldName) Normalize() QueryFieldName {
	return QueryFieldName(strings.ToLower(string(q)))
}

// Query represents a RRI request.
type Query struct {
	version Version
	action  QueryAction
	fields  map[QueryFieldName][]string
}

// Version returns the query version.
func (q *Query) Version() Version {
	return q.version
}

// Action returns the query action.
func (q *Query) Action() QueryAction {
	return q.action
}

// String returns a shortened, human-readable representation of the query.
func (q *Query) String() string {
	var sb strings.Builder
	for key, values := range q.fields {
		for _, value := range values {
			if sb.Len() > 0 {
				sb.WriteString("; ")
			}
			sb.WriteString(string(key))
			sb.WriteString("=")
			sb.WriteString(value)
		}
	}
	return fmt.Sprintf("%sv%s{%s}", q.action, q.version, sb.String())
}

// Export returns the query representation as used by Parse.
func (q *Query) Export() string {
	return q.EncodeKV()
}

// EncodeKV returns the Key-Value representation as used for RRI communication.
func (q *Query) EncodeKV() string {
	var sb strings.Builder
	sb.WriteString(string(FieldNameVersion))
	sb.WriteString(": ")
	sb.WriteString(string(q.version))
	sb.WriteString("\n")
	sb.WriteString(string(FieldNameAction))
	sb.WriteString(": ")
	sb.WriteString(string(q.action))
	for key, values := range q.fields {
		for _, value := range values {
			sb.WriteString("\n")
			sb.WriteString(string(key))
			sb.WriteString(": ")
			sb.WriteString(value)
		}
	}
	return sb.String()
}

// Fields returns all additional response fields.
func (q *Query) Fields() map[QueryFieldName][]string {
	return q.fields
}

// Field returns all values defined for a field name.
func (q *Query) Field(fieldName QueryFieldName) []string {
	fieldValues, ok := q.fields[fieldName.Normalize()]
	if !ok {
		return []string{}
	}
	return fieldValues
}

// FirstField returns the first field value or an empty string for a field name.
func (q *Query) FirstField(fieldName QueryFieldName) string {
	fieldValues, ok := q.fields[fieldName.Normalize()]
	if !ok || len(fieldValues) == 0 {
		return ""
	}
	return fieldValues[0]
}

// NewQuery returns a query with the given parameters.
func NewQuery(version Version, action QueryAction, fields map[QueryFieldName][]string) *Query {
	newFields := make(map[QueryFieldName][]string)
	if fields != nil {
		for key, value := range fields {
			newFields[key.Normalize()] = value
		}
	}
	return &Query{version, action, newFields}
}

// NewLoginQuery returns a login query for the given credentials.
func NewLoginQuery(username, password string) *Query {
	fields := make(map[QueryFieldName][]string)
	fields[FieldNameUser] = []string{username}
	fields[FieldNamePassword] = []string{password}
	return NewQuery(LatestVersion, ActionLogin, fields)
}

// NewLogoutQuery returns a logout query.
func NewLogoutQuery() *Query {
	return NewQuery(LatestVersion, ActionLogout, nil)
}

// NewCheckHandleQuery returns a check query.
func NewCheckHandleQuery(handle string) *Query {
	fields := make(map[QueryFieldName][]string)
	fields[FieldNameHandle] = []string{handle}
	return NewQuery(LatestVersion, ActionCheck, fields)
}

// NewCreateDomainQuery returns a query to create a domain.
func NewCreateDomainQuery(idnDomain string, holderHandles, abuseContactHandles, generalRequestHandles []string, nameServers ...string) *Query {
	fields := make(map[QueryFieldName][]string)
	fields[FieldNameDomainIDN] = []string{idnDomain}
	if ace, err := idna.ToASCII(idnDomain); err == nil {
		fields[FieldNameDomainACE] = []string{ace}
	}
	fields[FieldNameHolder] = holderHandles
	fields[FieldNameAbuseContact] = abuseContactHandles
	fields[FieldNameGeneralRequest] = generalRequestHandles
	fields[FieldNameNameServer] = nameServers
	return NewQuery(LatestVersion, ActionCreate, fields)
}

// NewCheckDomainQuery returns a check query.
func NewCheckDomainQuery(idnDomain string) *Query {
	fields := make(map[QueryFieldName][]string)
	fields[FieldNameDomainIDN] = []string{idnDomain}
	if ace, err := idna.ToASCII(idnDomain); err == nil {
		fields[FieldNameDomainACE] = []string{ace}
	}
	return NewQuery(LatestVersion, ActionCheck, fields)
}

// NewInfoDomainQuery returns an info query.
func NewInfoDomainQuery(idnDomain string) *Query {
	fields := make(map[QueryFieldName][]string)
	fields[FieldNameDomainIDN] = []string{idnDomain}
	if ace, err := idna.ToASCII(idnDomain); err == nil {
		fields[FieldNameDomainACE] = []string{ace}
	}
	return NewQuery(LatestVersion, ActionInfo, fields)
}

// NewUpdateDomainQuery returns a query to update a domain.
func NewUpdateDomainQuery(idnDomain string, holderHandles, abuseContactHandles, generalRequestHandles []string, nameServers ...string) *Query {
	fields := make(map[QueryFieldName][]string)
	fields[FieldNameDomainIDN] = []string{idnDomain}
	if ace, err := idna.ToASCII(idnDomain); err == nil {
		fields[FieldNameDomainACE] = []string{ace}
	}
	fields[FieldNameHolder] = holderHandles
	fields[FieldNameAbuseContact] = abuseContactHandles
	fields[FieldNameGeneralRequest] = generalRequestHandles
	fields[FieldNameNameServer] = nameServers
	return NewQuery(LatestVersion, ActionUpdate, fields)
}

// NewDeleteDomainQuery returns a delete query.
func NewDeleteDomainQuery(idnDomain string) *Query {
	fields := make(map[QueryFieldName][]string)
	fields[FieldNameDomainIDN] = []string{idnDomain}
	if ace, err := idna.ToASCII(idnDomain); err == nil {
		fields[FieldNameDomainACE] = []string{ace}
	}
	return NewQuery(LatestVersion, ActionDelete, fields)
}

// NewRestoreDomainQuery returns a restore query.
func NewRestoreDomainQuery(idnDomain string) *Query {
	fields := make(map[QueryFieldName][]string)
	fields[FieldNameDomainIDN] = []string{idnDomain}
	if ace, err := idna.ToASCII(idnDomain); err == nil {
		fields[FieldNameDomainACE] = []string{ace}
	}
	return NewQuery(LatestVersion, ActionRestore, fields)
}

// NewCreateAuthInfo1Query returns a create AuthInfo1 query.
func NewCreateAuthInfo1Query(idnDomain, authInfo string) *Query {
	fields := make(map[QueryFieldName][]string)
	fields[FieldNameDomainIDN] = []string{idnDomain}
	if ace, err := idna.ToASCII(idnDomain); err == nil {
		fields[FieldNameDomainACE] = []string{ace}
	}
	fields[FieldNameAuthInfoHash] = []string{computeHashSHA256(authInfo)}
	return NewQuery(LatestVersion, ActionCreateAuthInfo1, fields)
}

func computeHashSHA256(str string) string {
	hasher := sha256.New()
	hasher.Write([]byte(str))
	hash := hasher.Sum(nil)
	return hex.EncodeToString(hash)
}

// NewCreateAuthInfo2Query returns a create AuthInfo2 query.
func NewCreateAuthInfo2Query(idnDomain string) *Query {
	fields := make(map[QueryFieldName][]string)
	fields[FieldNameDomainIDN] = []string{idnDomain}
	if ace, err := idna.ToASCII(idnDomain); err == nil {
		fields[FieldNameDomainACE] = []string{ace}
	}
	return NewQuery(LatestVersion, ActionCreateAuthInfo2, fields)
}

// NewChangeProviderQuery returns a query to create a domain.
func NewChangeProviderQuery(idnDomain, authInfo string, holderHandles, abuseContactHandles, generalRequestHandles []string, nameServers ...string) *Query {
	fields := make(map[QueryFieldName][]string)
	fields[FieldNameDomainIDN] = []string{idnDomain}
	if ace, err := idna.ToASCII(idnDomain); err == nil {
		fields[FieldNameDomainACE] = []string{ace}
	}
	fields[FieldNameHolder] = holderHandles
	fields[FieldNameAbuseContact] = abuseContactHandles
	fields[FieldNameGeneralRequest] = generalRequestHandles
	fields[FieldNameNameServer] = nameServers
	fields[FieldNameAuthInfo] = []string{authInfo}
	return NewQuery(LatestVersion, ActionChangeProvider, fields)
}

// ParseQueryKV parses a single key-value encoded query.
func ParseQueryKV(str string) (*Query, error) {
	lines := strings.Split(str, "\n")
	fields := make(map[QueryFieldName][]string)
	for _, line := range lines {
		// trim spaces and ignore empty lines
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("query line must be key-value separated by ':'")
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		fieldValues, ok := fields[QueryFieldName(key).Normalize()]
		if !ok {
			fieldValues = []string{value}
		} else {
			fieldValues = append(fieldValues, value)
		}

		fields[QueryFieldName(key).Normalize()] = fieldValues
	}

	versionValues, ok := fields[FieldNameVersion]
	if !ok || len(versionValues) == 0 {
		return nil, fmt.Errorf("%s key is missing", FieldNameVersion)
	}
	if len(versionValues) > 1 {
		return nil, fmt.Errorf("multiple %s values", FieldNameVersion)
	}
	delete(fields, FieldNameVersion)

	actionValues, ok := fields[FieldNameAction]
	if !ok || len(actionValues) == 0 {
		return nil, fmt.Errorf("%s key is missing", FieldNameAction)
	}
	if len(actionValues) > 1 {
		return nil, fmt.Errorf("multiple %s values", FieldNameAction)
	}
	delete(fields, FieldNameAction)

	return &Query{Version(versionValues[0]).Normalize(), QueryAction(actionValues[0]).Normalize(), fields}, nil
}

// ParseQuery parses a single query from a string.
func ParseQuery(str string) (*Query, error) {
	return ParseQueryKV(str)
}

// ParseQueries parses multiple queries separated by a =-= line from a string.
func ParseQueries(str string) ([]*Query, error) {
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

	queries := make([]*Query, len(queryStrings))
	for i, queryString := range queryStrings {
		query, err := ParseQuery(strings.TrimSpace(queryString))
		if err != nil {
			return nil, err
		}
		queries[i] = query
	}
	return queries, nil
}
