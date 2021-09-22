package rri

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
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
	// QueryFieldNameType denotes the query field name for type.
	QueryFieldNameType QueryFieldName = "type"
	// QueryFieldNameName denotes the query field name for name.
	QueryFieldNameName QueryFieldName = "name"
	// QueryFieldNameOrganisation denotes the query field name for organisation.
	QueryFieldNameOrganisation QueryFieldName = "organisation"
	// QueryFieldNameAddress denotes the query field name for address.
	QueryFieldNameAddress QueryFieldName = "address"
	// QueryFieldNamePostalCode denotes the query field name for postalcode.
	QueryFieldNamePostalCode QueryFieldName = "postalcode"
	// QueryFieldNameCity denotes the query field name for city.
	QueryFieldNameCity QueryFieldName = "city"
	// QueryFieldNameCountryCode denotes the query field name for countrycode.
	QueryFieldNameCountryCode QueryFieldName = "countrycode"
	// QueryFieldNameEMail denotes the query field name for email.
	QueryFieldNameEMail QueryFieldName = "email"
	// QueryFieldNameSSOEMail denotes the e-mail address to link after CSP registration for vChecked.
	QueryFieldNameSSOEMail QueryFieldName = "ssoemail"
	// QueryFieldNameBusinessNumber denotes the business number of a legal entity.
	QueryFieldNameBusinessNumber QueryFieldName = "businessnumber"
	// QueryFieldNameAuthSigFirstName denotes the first name of an authorized signatory.
	QueryFieldNameAuthSigFirstName QueryFieldName = "authorizedsignatoryfirstname"
	// QueryFieldNameAuthSigLastName denotes the last name of an authorized signatory.
	QueryFieldNameAuthSigLastName QueryFieldName = "authorizedsignatorylastname"
	// QueryFieldNameAuthSigEMail denotes the email address of an authorized signatory.
	QueryFieldNameAuthSigEMail QueryFieldName = "authorizedsignatoryemail"
	// QueryFieldNameAuthSigDateOfBirth denotes the date of birth of an authorized signatory.
	QueryFieldNameAuthSigDateOfBirth QueryFieldName = "authorizedsignatorydateofbirth"
	// QueryFieldNameAuthSigCountryCode denotes the country of an authorized signatory.
	QueryFieldNameAuthSigCountryCode QueryFieldName = "authorizedsignatorycountrycode"
	// QueryFieldNameAuthSigCity denotes the city of an authorized signatory.
	QueryFieldNameAuthSigCity QueryFieldName = "authorizedsignatorycity"
	// QueryFieldNameAuthSigPostalCode denotes the postal code of an authorized signatory.
	QueryFieldNameAuthSigPostalCode QueryFieldName = "authorizedsignatorypostalcode"
	// QueryFieldNameAuthSigStreet denotes the street of an authorized signatory.
	QueryFieldNameAuthSigStreet QueryFieldName = "authorizedsignatorystreet"
	// QueryFieldNameAuthSigPhone denotes the phone number of an authorized signatory.
	QueryFieldNameAuthSigPhone QueryFieldName = "authorizedsignatoryphone"
	// QueryFieldNamePersonFirstName denotes the first name of an authorized signatory.
	QueryFieldNamePersonFirstName QueryFieldName = "personfirstname"
	// QueryFieldNamePersonLastName denotes the last name of an authorized signatory.
	QueryFieldNamePersonLastName QueryFieldName = "personlastname"
	// QueryFieldNamePersonEMail denotes the email address of an authorized signatory.
	QueryFieldNamePersonEMail QueryFieldName = "personemail"
	// QueryFieldNamePersonDateOfBirth denotes the date of birth of an authorized signatory.
	QueryFieldNamePersonDateOfBirth QueryFieldName = "persondateofbirth"
	// QueryFieldNamePersonCountryCode denotes the country of an authorized signatory.
	QueryFieldNamePersonCountryCode QueryFieldName = "personcountrycode"
	// QueryFieldNamePersonCity denotes the city of an authorized signatory.
	QueryFieldNamePersonCity QueryFieldName = "personcity"
	// QueryFieldNamePersonPostalCode denotes the postal code of an authorized signatory.
	QueryFieldNamePersonPostalCode QueryFieldName = "personpostalcode"
	// QueryFieldNamePersonStreet denotes the street of an authorized signatory.
	QueryFieldNamePersonStreet QueryFieldName = "personstreet"
	// QueryFieldNamePersonPhone denotes the phone number of an authorized signatory.
	QueryFieldNamePersonPhone QueryFieldName = "personphone"
	// QueryFieldNameMsgID denotes the query field name for an message id.
	QueryFieldNameMsgID QueryFieldName = "msgid"

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
	// ActionChangeHolder denotes the action value for change holder.
	ActionChangeHolder QueryAction = "CHHOLDER"
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
	// ActionChangeProvider denotes the action value for change provider.
	ActionChangeProvider QueryAction = "CHPROV"
	// ActionVerifyLegalEntity denotes the action value to verify a legal entity.
	ActionVerifyLegalEntity QueryAction = "VERIFY-LEGAL-ENTITY"
	// ActionVerifyNaturalPerson denotes the action value to verify a natural person.
	ActionVerifyNaturalPerson QueryAction = "VERIFY-NATURAL-PERSON"
	// ActionVerifyQueueRead denotes the action value to read from the vChecked message queue.
	ActionVerifyQueueRead QueryAction = "VERIFY-QUEUE-READ"
	// ActionVerifyQueueDelete denotes the action value to delete from the vChecked message queue.
	ActionVerifyQueueDelete QueryAction = "VERIFY-QUEUE-DELETE"

	// ContactTypePerson denotes a person.
	ContactTypePerson ContactType = "PERSON"
	// ContactTypeOrganisation denotes an organisation.
	ContactTypeOrganisation ContactType = "ORG"
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

// ContactType represents the type of a contact handle.
type ContactType string

// Normalize returns the normalized representation of the given ContactType.
func (t ContactType) Normalize() ContactType {
	return ContactType(strings.ToUpper(string(t)))
}

// ParseContactType parses a contact type from string.
func ParseContactType(str string) (ContactType, error) {
	switch strings.ToUpper(str) {
	case "PERSON":
		return ContactTypePerson, nil
	case "ORG":
		return ContactTypeOrganisation, nil
	default:
		return "", fmt.Errorf("invalid contact type")
	}
}

// DenicHandle represents a handle like DENIC-1000006-SOME-CODE
type DenicHandle struct {
	RegAccID    int
	ContactCode string
}

func (h DenicHandle) String() string {
	if h.IsEmpty() {
		return ""
	}
	return fmt.Sprintf("DENIC-%d-%s", h.RegAccID, strings.ToUpper(h.ContactCode))
}

// IsEmpty returns true when the given denic handle is unset.
func (h DenicHandle) IsEmpty() bool {
	return h.RegAccID == 0 && len(h.ContactCode) == 0
}

// NewDenicHandle assembles a new denic handle.
func NewDenicHandle(regAccID int, contactCode string) DenicHandle {
	return DenicHandle{regAccID, strings.ToUpper(contactCode)}
}

// EmptyDenicHandle returns an empty denic handle.
func EmptyDenicHandle() DenicHandle {
	return DenicHandle{}
}

// ParseDenicHandle tries to parse a handle like DENIC-1000006-SOME-CODE. Returns an empty denic handle if str is empty.
func ParseDenicHandle(str string) (DenicHandle, error) {
	if len(str) == 0 {
		return EmptyDenicHandle(), nil
	}

	parts := strings.SplitN(str, "-", 3)
	if len(parts) != 3 {
		return DenicHandle{}, fmt.Errorf("invalid handle")
	}

	if strings.ToUpper(parts[0]) != "DENIC" {
		return DenicHandle{}, fmt.Errorf("invalid handle")
	}

	regAccID, err := strconv.Atoi(parts[1])
	if err != nil {
		return DenicHandle{}, fmt.Errorf("invalid handle")
	}

	return NewDenicHandle(regAccID, strings.ToUpper(parts[2])), nil
}

// DomainData holds domain information.
type DomainData struct {
	HolderHandles         []DenicHandle
	GeneralRequestHandles []DenicHandle
	AbuseContactHandles   []DenicHandle
	NameServers           []string
}

func (domainData *DomainData) putToQueryFields(fields *QueryFieldList) {
	putHandlesToQueryFields := func(fieldName QueryFieldName, handles []DenicHandle) {
		for _, h := range handles {
			if !h.IsEmpty() {
				fields.Add(fieldName, h.String())
			}
		}
	}

	putHandlesToQueryFields(QueryFieldNameHolder, domainData.HolderHandles)
	putHandlesToQueryFields(QueryFieldNameGeneralRequest, domainData.GeneralRequestHandles)
	putHandlesToQueryFields(QueryFieldNameAbuseContact, domainData.AbuseContactHandles)
	fields.Add(QueryFieldNameNameServer, domainData.NameServers...)
}

// ContactData holds information of a contact handle.
type ContactData struct {
	Type         ContactType
	Name         string
	Organisation string
	Address      string
	PostalCode   string
	City         string
	CountryCode  string
	EMail        []string
}

func (contactData *ContactData) putToQueryFields(fields *QueryFieldList) {
	fields.Add(QueryFieldNameType, string(contactData.Type.Normalize()))
	fields.Add(QueryFieldNameName, contactData.Name)
	fields.Add(QueryFieldNameOrganisation, splitLines(contactData.Organisation)...)
	fields.Add(QueryFieldNameAddress, splitLines(contactData.Address)...)
	fields.Add(QueryFieldNamePostalCode, contactData.PostalCode)
	fields.Add(QueryFieldNameCity, contactData.City)
	fields.Add(QueryFieldNameCountryCode, contactData.CountryCode)
	fields.Add(QueryFieldNameEMail, contactData.EMail...)
}

func splitLines(str string) []string {
	return strings.Split(strings.ReplaceAll(strings.ReplaceAll(str, "\r\n", "\n"), "\r", "\n"), "\n")
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
	var sb strings.Builder
	switch q.Action() {
	case ActionLogin:
		sb.WriteString(fmt.Sprintf("%q", q.FirstField(QueryFieldNameUser)))
	}
	return fmt.Sprintf("%s{%s}", q.Action(), sb.String())
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
func NewQuery(version Version, action QueryAction, fields QueryFieldList) *Query {
	newFields := NewQueryFieldList()
	newFields.Add(QueryFieldNameVersion, string(version.Normalize()))
	newFields.Add(QueryFieldNameAction, string(action.Normalize()))
	if fields != nil {
		fields.CopyTo(&newFields)
	}
	return &Query{newFields}
}

// NewLoginQuery returns a login query for the given credentials.
func NewLoginQuery(username, password string) *Query {
	fields := NewQueryFieldList()
	fields.Add(QueryFieldNameUser, username)
	fields.Add(QueryFieldNamePassword, password)
	return NewQuery(LatestVersion, ActionLogin, fields)
}

// NewLogoutQuery returns a logout query.
func NewLogoutQuery() *Query {
	return NewQuery(LatestVersion, ActionLogout, nil)
}

// NewCreateContactQuery returns a check query.
func NewCreateContactQuery(handle DenicHandle, contactData ContactData) *Query {
	fields := NewQueryFieldList()
	fields.Add(QueryFieldNameHandle, handle.String())
	contactData.putToQueryFields(&fields)
	return NewQuery(LatestVersion, ActionCreate, fields)
}

// NewCheckHandleQuery returns a check query for a contact or request contact handle.
func NewCheckHandleQuery(handle DenicHandle) *Query {
	fields := NewQueryFieldList()
	fields.Add(QueryFieldNameHandle, handle.String())
	return NewQuery(LatestVersion, ActionCheck, fields)
}

// NewInfoHandleQuery returns an info query for a contact or request contact handle.
func NewInfoHandleQuery(handle DenicHandle) *Query {
	fields := NewQueryFieldList()
	fields.Add(QueryFieldNameHandle, handle.String())
	return NewQuery(LatestVersion, ActionInfo, fields)
}

func putDomainToQueryFields(fields *QueryFieldList, domain string) {
	if strings.HasPrefix(strings.ToLower(domain), "xn--") {
		fields.Add(QueryFieldNameDomainACE, domain)
		if idn, err := idna.ToUnicode(domain); err == nil {
			fields.Add(QueryFieldNameDomainIDN, idn)
		}

	} else {
		fields.Add(QueryFieldNameDomainIDN, domain)
		if ace, err := idna.ToASCII(domain); err == nil {
			//TODO only add ace string if it differs from idn
			fields.Add(QueryFieldNameDomainACE, ace)
		}
	}
}

// NewCreateDomainQuery returns a query to create a domain.
func NewCreateDomainQuery(domain string, domainData DomainData) *Query {
	fields := NewQueryFieldList()
	putDomainToQueryFields(&fields, domain)
	domainData.putToQueryFields(&fields)
	return NewQuery(LatestVersion, ActionCreate, fields)
}

// NewCheckDomainQuery returns a check query.
func NewCheckDomainQuery(domain string) *Query {
	fields := NewQueryFieldList()
	putDomainToQueryFields(&fields, domain)
	return NewQuery(LatestVersion, ActionCheck, fields)
}

// NewInfoDomainQuery returns an info query.
func NewInfoDomainQuery(domain string) *Query {
	fields := NewQueryFieldList()
	putDomainToQueryFields(&fields, domain)
	return NewQuery(LatestVersion, ActionInfo, fields)
}

// NewUpdateDomainQuery returns a query to update a domain.
func NewUpdateDomainQuery(domain string, domainData DomainData) *Query {
	fields := NewQueryFieldList()
	putDomainToQueryFields(&fields, domain)
	domainData.putToQueryFields(&fields)
	return NewQuery(LatestVersion, ActionUpdate, fields)
}

// NewChangeHolderQuery returns a query to update a domain.
func NewChangeHolderQuery(domain string, domainData DomainData) *Query {
	fields := NewQueryFieldList()
	putDomainToQueryFields(&fields, domain)
	domainData.putToQueryFields(&fields)
	return NewQuery(LatestVersion, ActionChangeHolder, fields)
}

// NewDeleteDomainQuery returns a delete query.
func NewDeleteDomainQuery(domain string) *Query {
	fields := NewQueryFieldList()
	putDomainToQueryFields(&fields, domain)
	return NewQuery(LatestVersion, ActionDelete, fields)
}

// NewRestoreDomainQuery returns a restore query.
func NewRestoreDomainQuery(domain string) *Query {
	fields := NewQueryFieldList()
	putDomainToQueryFields(&fields, domain)
	return NewQuery(LatestVersion, ActionRestore, fields)
}

// NewTransitDomainQuery returns a restore query.
func NewTransitDomainQuery(domain string, disconnect bool) *Query {
	fields := NewQueryFieldList()
	putDomainToQueryFields(&fields, domain)
	if disconnect {
		fields.Add(QueryFieldNameDisconnect, "true")
	} else {
		fields.Add(QueryFieldNameDisconnect, "false")
	}
	return NewQuery(LatestVersion, ActionTransit, fields)
}

// NewCreateAuthInfo1Query returns a create AuthInfo1 query.
func NewCreateAuthInfo1Query(domain, authInfo string, expireDay time.Time) *Query {
	fields := NewQueryFieldList()
	putDomainToQueryFields(&fields, domain)
	fields.Add(QueryFieldNameAuthInfoHash, computeHashSHA256(authInfo))
	fields.Add(QueryFieldNameAuthInfoExpire, expireDay.Format("20060102"))
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
	fields := NewQueryFieldList()
	putDomainToQueryFields(&fields, domain)
	return NewQuery(LatestVersion, ActionCreateAuthInfo2, fields)
}

// NewChangeProviderQuery returns a query to create a domain.
func NewChangeProviderQuery(domain, authInfo string, domainData DomainData) *Query {
	fields := NewQueryFieldList()
	putDomainToQueryFields(&fields, domain)
	domainData.putToQueryFields(&fields)
	fields.Add(QueryFieldNameAuthInfo, authInfo)
	return NewQuery(LatestVersion, ActionChangeProvider, fields)
}

// PersonToVerify represents the natural person that will be verified for vChecked.
type PersonToVerify struct {
	FirstName   string
	LastName    string
	EMail       string
	DateOfBirth time.Time
	CountryCode string
	City        string
	PostalCode  string
	Street      string
	Phone       string
}

// NewVerifyLegalEntityQuery returns a query to verify a domain that is registered for a legal entity.
func NewVerifyLegalEntityQuery(domain string, ssoEMail string, authSignatory PersonToVerify, businessNumber string) *Query {
	fields := NewQueryFieldList()
	putDomainToQueryFields(&fields, domain)
	fields.Add(QueryFieldNameAuthSigFirstName, authSignatory.FirstName)
	fields.Add(QueryFieldNameAuthSigLastName, authSignatory.LastName)
	fields.Add(QueryFieldNameAuthSigEMail, authSignatory.EMail)
	fields.Add(QueryFieldNameAuthSigDateOfBirth, authSignatory.DateOfBirth.Format("2006-01-02"))
	fields.Add(QueryFieldNameAuthSigCountryCode, authSignatory.CountryCode)
	fields.Add(QueryFieldNameAuthSigCity, authSignatory.City)
	fields.Add(QueryFieldNameAuthSigPostalCode, authSignatory.PostalCode)
	fields.Add(QueryFieldNameAuthSigStreet, authSignatory.Street)
	if len(authSignatory.Phone) > 0 {
		fields.Add(QueryFieldNameAuthSigPhone, authSignatory.Phone)
	}
	if len(ssoEMail) > 0 {
		fields.Add(QueryFieldNameSSOEMail, ssoEMail)
	}
	if len(businessNumber) > 0 {
		fields.Add(QueryFieldNameBusinessNumber, businessNumber)
	}
	return NewQuery(LatestVersion, ActionVerifyLegalEntity, fields)
}

// VerifyNaturalPersonQuery returns a query to verify a domain that is registered for a natural person.
func NewVerifyNaturalPersonQuery(domain string, ssoEMail string, person PersonToVerify) *Query {
	fields := NewQueryFieldList()
	putDomainToQueryFields(&fields, domain)
	fields.Add(QueryFieldNamePersonFirstName, person.FirstName)
	fields.Add(QueryFieldNamePersonLastName, person.LastName)
	fields.Add(QueryFieldNamePersonEMail, person.EMail)
	fields.Add(QueryFieldNamePersonDateOfBirth, person.DateOfBirth.Format("2006-01-02"))
	fields.Add(QueryFieldNamePersonCountryCode, person.CountryCode)
	fields.Add(QueryFieldNamePersonCity, person.City)
	fields.Add(QueryFieldNamePersonPostalCode, person.PostalCode)
	fields.Add(QueryFieldNamePersonStreet, person.Street)
	if len(person.Phone) > 0 {
		fields.Add(QueryFieldNamePersonPhone, person.Phone)
	}
	if len(ssoEMail) > 0 {
		fields.Add(QueryFieldNameSSOEMail, ssoEMail)
	}
	return NewQuery(LatestVersion, ActionVerifyLegalEntity, fields)
}

// NewVerifyQueueReadQuery returns a query to read from the vChecked message queue.
func NewVerifyQueueReadQuery() *Query {
	return NewQuery(LatestVersion, ActionVerifyQueueRead, nil)
}

// NewVerifyQueueReadQuery returns a query to read from the vChecked message queue.
func NewVerifyQueueDeleteQuery(msgID string) *Query {
	fields := NewQueryFieldList()
	fields.Add(QueryFieldNameMsgID, msgID)
	return NewQuery(LatestVersion, ActionVerifyQueueDelete, fields)
}

// ParseQueryKV parses a single key-value encoded query.
func ParseQueryKV(str string) (*Query, error) {
	lines := strings.Split(str, "\n")
	fields := NewQueryFieldList()
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
