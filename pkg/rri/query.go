package rri

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"golang.org/x/net/idna"
)

const (
	// LatestVersion denotes the latest RRI version supported by the client.
	LatestVersion Version = "3.0"

	// QueryFieldNameVersion denotes the query field name for version.
	QueryFieldNameVersion QueryFieldName = "version"
	// QueryFieldNameAction denotes the query field name for action.
	QueryFieldNameAction QueryFieldName = "action"
	// QueryFieldNameUser denotes the query field name for login user.
	QueryFieldNameUser QueryFieldName = "user"
	// QueryFieldNamePassword denotes the query field name for login password.
	QueryFieldNamePassword QueryFieldName = "password"
	// QueryFieldNameDomainIDN denotes the query field name for IDN domain name.
	QueryFieldNameDomainIDN QueryFieldName = "domain"
	// QueryFieldNameDomainACE denotes the query field name for ACE domain name.
	QueryFieldNameDomainACE QueryFieldName = "domain-ace"
	// QueryFieldNameHolder denotes the query field name for holder handle.
	QueryFieldNameHolder QueryFieldName = "holder"
	// QueryFieldNameGeneralRequest denotes the query field name for general request handle.
	QueryFieldNameGeneralRequest QueryFieldName = "generalrequest"
	// QueryFieldNameAbuseContact denotes the query field name for abuse contact handle.
	QueryFieldNameAbuseContact QueryFieldName = "abusecontact"
	// QueryFieldNameNameServer denotes the query field name for name servers.
	QueryFieldNameNameServer QueryFieldName = "nserver"
	// QueryFieldNameHandle denotes the query field name for denic handles.
	QueryFieldNameHandle QueryFieldName = "handle"
	// QueryFieldNameDisconnect denotes the query field name for disconnect.
	QueryFieldNameDisconnect QueryFieldName = "disconnect"
	// QueryFieldNameAuthInfoHash denotes the query field name for auth info hash.
	QueryFieldNameAuthInfoHash QueryFieldName = "authinfohash"
	// QueryFieldNameAuthInfoExpire denotes the query field name for auth info expire.
	QueryFieldNameAuthInfoExpire QueryFieldName = "authinfoexpire"
	// QueryFieldNameAuthInfo denotes the query field name for auth info hash.
	QueryFieldNameAuthInfo QueryFieldName = "authinfo"
	// QueryFieldNameAuthSigFirstName denotes the first name of an authorized signatory.
	QueryFieldNameAuthSigFirstName QueryFieldName = "authorizedsignatoryfirstname"
	// QueryFieldNameAuthSigLastName denotes the last name of an authorized signatory.
	QueryFieldNameAuthSigLastName QueryFieldName = "authorizedsignatorylastname"
	// QueryFieldNameAuthSigEMail denotes the email address of an authorized signatory.
	QueryFieldNameAuthSigEMail QueryFieldName = "authorizedsignatoryemail"
	// QueryFieldNameAuthSigDateOfBirth denotes the date of birth of an authorized signatory.
	QueryFieldNameAuthSigDateOfBirth QueryFieldName = "authorizedsignatorydateofbirth"
	// QueryFieldNameAuthSigPlaceOfBirth denotes the place of birth of an authorized signatory.
	QueryFieldNameAuthSigPlaceOfBirth QueryFieldName = "authorizedsignatoryplaceofbirth"
	// QueryFieldNameAuthSigPhone denotes the phone number of an authorized signatory.
	QueryFieldNameAuthSigPhone QueryFieldName = "authorizedsignatoryphone"

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
	// ActionUpdate denotes the action value for update.
	ActionUpdate QueryAction = "UPDATE"
	// ActionDelete deontes the action value for delete.
	ActionDelete QueryAction = "DELETE"
	// ActionRestore deontes the action value for restore.
	ActionRestore QueryAction = "RESTORE"
	// ActionTransit deontes the action value for transit.
	ActionTransit QueryAction = "TRANSIT"
	// ActionCreateAuthInfo1 denotes the action value for create AuthInfo1.
	ActionCreateAuthInfo1 QueryAction = "CREATE-AUTHINFO1"
	// ActionCreateAuthInfo2 denotes the action value for create AuthInfo2.
	ActionCreateAuthInfo2 QueryAction = "CREATE-AUTHINFO2"
	// ActionVerify denotes the action value for verify.
	ActionVerify QueryAction = "VERIFY"
	// ActionChangeProvider denotes the action value for change provider.
	ActionChangeProvider QueryAction = "CHPROV"
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

// DomainData holds domain information.
type DomainData struct {
	HolderHandles         []string
	GeneralRequestHandles []string
	AbuseContactHandles   []string
	NameServers           []string
}

func (domainData *DomainData) putToQueryFields(fields map[QueryFieldName][]string) {
	fields[QueryFieldNameHolder] = domainData.HolderHandles
	fields[QueryFieldNameGeneralRequest] = domainData.GeneralRequestHandles
	fields[QueryFieldNameAbuseContact] = domainData.AbuseContactHandles
	fields[QueryFieldNameNameServer] = domainData.NameServers
}

// Query represents a RRI request.
type Query struct {
	fields QueryFieldList
}

// Version returns the query version.
func (q *Query) Version() Version {
	return Version(q.FirstField(QueryFieldNameVersion)).Normalize()
}

// Action returns the query action.
func (q *Query) Action() QueryAction {
	return QueryAction(q.FirstField(QueryFieldNameAction)).Normalize()
}

// String returns a human readable representation of the query.
func (q *Query) String() string {
	//TODO shortened, single line representation
	var sb strings.Builder
	for _, f := range q.fields {
		if sb.Len() > 0 {
			sb.WriteString("; ")
		}
		sb.WriteString(string(f.Name))
		sb.WriteString("=")
		sb.WriteString(f.Value)
	}
	return fmt.Sprintf("%sv%s{%s}", q.Action(), q.Version(), sb.String())
}

// EncodeKV returns the Key-Value representation as used for RRI communication.
func (q *Query) EncodeKV() string {
	var sb strings.Builder
	for _, f := range q.fields {
		if sb.Len() > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(string(f.Name))
		sb.WriteString(": ")
		sb.WriteString(f.Value)
	}
	return sb.String()
}

// Fields returns all additional response fields.
func (q *Query) Fields() QueryFieldList {
	return q.fields
}

// Field returns all values defined for a field name.
func (q *Query) Field(fieldName QueryFieldName) []string {
	return q.fields.Values(fieldName)
}

// FirstField returns the first field value or an empty string for a field name.
func (q *Query) FirstField(fieldName QueryFieldName) string {
	return q.fields.FirstValue(fieldName)
}

// NewQuery returns a query with the given parameters.
func NewQuery(version Version, action QueryAction, fields map[QueryFieldName][]string) *Query {
	//TODO instantiate from QueryFieldList instead of map
	newFields := newQueryFieldList()
	newFields.Add(QueryFieldNameVersion, string(version.Normalize()))
	newFields.Add(QueryFieldNameAction, string(action.Normalize()))
	if fields != nil {
		for key, values := range fields {
			for _, value := range values {
				newFields.Add(key, value)
			}
		}
	}
	return &Query{newFields}
}

// NewLoginQuery returns a login query for the given credentials.
func NewLoginQuery(username, password string) *Query {
	fields := make(map[QueryFieldName][]string)
	fields[QueryFieldNameUser] = []string{username}
	fields[QueryFieldNamePassword] = []string{password}
	return NewQuery(LatestVersion, ActionLogin, fields)
}

// NewLogoutQuery returns a logout query.
func NewLogoutQuery() *Query {
	return NewQuery(LatestVersion, ActionLogout, nil)
}

// NewCheckHandleQuery returns a check query.
func NewCheckHandleQuery(handle string) *Query {
	fields := make(map[QueryFieldName][]string)
	fields[QueryFieldNameHandle] = []string{handle}
	return NewQuery(LatestVersion, ActionCheck, fields)
}

// NewInfoHandleQuery returns a check query.
func NewInfoHandleQuery(handle string) *Query {
	fields := make(map[QueryFieldName][]string)
	fields[QueryFieldNameHandle] = []string{handle}
	return NewQuery(LatestVersion, ActionCheck, fields)
}

func putDomainToQueryFields(fields map[QueryFieldName][]string, domain string) {
	if strings.HasPrefix(strings.ToLower(domain), "xn--") {
		fields[QueryFieldNameDomainACE] = []string{domain}
		if idn, err := idna.ToUnicode(domain); err == nil {
			fields[QueryFieldNameDomainIDN] = []string{idn}
		}

	} else {
		fields[QueryFieldNameDomainIDN] = []string{domain}
		if ace, err := idna.ToASCII(domain); err == nil {
			fields[QueryFieldNameDomainACE] = []string{ace}
		}
	}
}

// NewCreateDomainQuery returns a query to create a domain.
func NewCreateDomainQuery(domain string, domainData DomainData) *Query {
	fields := make(map[QueryFieldName][]string)
	putDomainToQueryFields(fields, domain)
	domainData.putToQueryFields(fields)
	return NewQuery(LatestVersion, ActionCreate, fields)
}

// NewCheckDomainQuery returns a check query.
func NewCheckDomainQuery(domain string) *Query {
	fields := make(map[QueryFieldName][]string)
	putDomainToQueryFields(fields, domain)
	return NewQuery(LatestVersion, ActionCheck, fields)
}

// NewInfoDomainQuery returns an info query.
func NewInfoDomainQuery(domain string) *Query {
	fields := make(map[QueryFieldName][]string)
	putDomainToQueryFields(fields, domain)
	return NewQuery(LatestVersion, ActionInfo, fields)
}

// NewUpdateDomainQuery returns a query to update a domain.
func NewUpdateDomainQuery(domain string, domainData DomainData) *Query {
	fields := make(map[QueryFieldName][]string)
	putDomainToQueryFields(fields, domain)
	domainData.putToQueryFields(fields)
	return NewQuery(LatestVersion, ActionUpdate, fields)
}

// NewDeleteDomainQuery returns a delete query.
func NewDeleteDomainQuery(domain string) *Query {
	fields := make(map[QueryFieldName][]string)
	putDomainToQueryFields(fields, domain)
	return NewQuery(LatestVersion, ActionDelete, fields)
}

// NewRestoreDomainQuery returns a restore query.
func NewRestoreDomainQuery(domain string) *Query {
	fields := make(map[QueryFieldName][]string)
	putDomainToQueryFields(fields, domain)
	return NewQuery(LatestVersion, ActionRestore, fields)
}

// NewTransitDomainQuery returns a restore query.
func NewTransitDomainQuery(domain string, disconnect bool) *Query {
	fields := make(map[QueryFieldName][]string)
	putDomainToQueryFields(fields, domain)
	if disconnect {
		fields[QueryFieldNameDisconnect] = []string{"true"}
	} else {
		fields[QueryFieldNameDisconnect] = []string{"false"}
	}
	return NewQuery(LatestVersion, ActionTransit, fields)
}

// NewCreateAuthInfo1Query returns a create AuthInfo1 query.
func NewCreateAuthInfo1Query(domain, authInfo string, expireDay time.Time) *Query {
	fields := make(map[QueryFieldName][]string)
	putDomainToQueryFields(fields, domain)
	fields[QueryFieldNameAuthInfoHash] = []string{computeHashSHA256(authInfo)}
	fields[QueryFieldNameAuthInfoExpire] = []string{expireDay.Format("20060102")}
	return NewQuery(LatestVersion, ActionCreateAuthInfo1, fields)
}

func computeHashSHA256(str string) string {
	hasher := sha256.New()
	hasher.Write([]byte(str))
	hash := hasher.Sum(nil)
	return hex.EncodeToString(hash)
}

// NewCreateAuthInfo2Query returns a create AuthInfo2 query.
func NewCreateAuthInfo2Query(domain string) *Query {
	fields := make(map[QueryFieldName][]string)
	putDomainToQueryFields(fields, domain)
	return NewQuery(LatestVersion, ActionCreateAuthInfo2, fields)
}

// AuthorizedSignatory represents the authorized signatory for a VERIFY query.
type AuthorizedSignatory struct {
	FirstName    string
	LastName     string
	EMail        string
	DateOfBirth  time.Time
	PlaceOfBirth string
	Phone        string
}

// NewVerifyDomainQuery returns a query to create a verify domain.
func NewVerifyDomainQuery(domain string, authSignatory AuthorizedSignatory) *Query {
	fields := make(map[QueryFieldName][]string)
	putDomainToQueryFields(fields, domain)
	fields[QueryFieldNameAuthSigFirstName] = []string{authSignatory.FirstName}
	fields[QueryFieldNameAuthSigLastName] = []string{authSignatory.LastName}
	fields[QueryFieldNameAuthSigEMail] = []string{authSignatory.EMail}
	fields[QueryFieldNameAuthSigDateOfBirth] = []string{authSignatory.DateOfBirth.Format("2006-01-02")}
	fields[QueryFieldNameAuthSigPlaceOfBirth] = []string{authSignatory.PlaceOfBirth}
	if len(authSignatory.Phone) > 0 {
		fields[QueryFieldNameAuthSigPhone] = []string{authSignatory.Phone}
	}
	return NewQuery(LatestVersion, ActionVerify, fields)
}

// NewChangeProviderQuery returns a query to create a domain.
func NewChangeProviderQuery(domain, authInfo string, domainData DomainData) *Query {
	fields := make(map[QueryFieldName][]string)
	putDomainToQueryFields(fields, domain)
	domainData.putToQueryFields(fields)
	fields[QueryFieldNameAuthInfo] = []string{authInfo}
	return NewQuery(LatestVersion, ActionChangeProvider, fields)
}

// ParseQueryKV parses a single key-value encoded query.
func ParseQueryKV(str string) (*Query, error) {
	lines := strings.Split(str, "\n")
	fields := newQueryFieldList()
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

		fields.Add(QueryFieldName(key), value)
	}

	versionValues := fields.Values(QueryFieldNameVersion)
	if len(versionValues) == 0 {
		return nil, fmt.Errorf("%s key is missing", QueryFieldNameVersion)
	}
	if len(versionValues) > 1 {
		return nil, fmt.Errorf("multiple %s values", QueryFieldNameVersion)
	}

	actionValues := fields.Values(QueryFieldNameAction)
	if len(actionValues) == 0 {
		return nil, fmt.Errorf("%s key is missing", QueryFieldNameAction)
	}
	if len(actionValues) > 1 {
		return nil, fmt.Errorf("multiple %s values", QueryFieldNameAction)
	}

	return &Query{fields}, nil
}

// ParseQuery tries to detect the query format (KV or XML) and returns the parsed query.
func ParseQuery(str string) (*Query, error) {
	//TODO detect type
	return ParseQueryKV(str)
}
