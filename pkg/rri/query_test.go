package rri

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	qryEncode       = "Version: 3.0\nAction: update\naddress: foo\nDomain: denic.de\nAddress: bar"
	qryNormalize    = "Version:   3.0   \nAction: iNfO\ndIsCoNnEcT: tRuE"
	qryIgnoreCasing = "Version: 3.0\nAction: login\nUser: DENIC-1000042-TEST\nPassword: very-secure"
	qryWhitespaces  = "  version: \t3.0  \n\n\naction:    LOGIN\n   user: DENIC-1000042-TEST\npassword: very-secure    \n"
	qryOrder        = "action: LOGIN\ncustom: 1\nversion: 3.0\nuser: DENIC-1000042-TEST\nstuff: foobar\npassword: very-secure\ncustom: 2"
)

func TestNewQueryNil(t *testing.T) {
	var qry *Query
	require.NotPanics(t, func() {
		qry = NewQuery(LatestVersion, ActionLogout, nil)
	})
	assert.Equal(t, 2, qry.Fields().Size())
}

func TestPutDomainToQueryFields(t *testing.T) {
	fieldsFromIDN := NewQueryFieldList()
	putDomainToQueryFields(&fieldsFromIDN, "dönic.de")
	require.Len(t, fieldsFromIDN, 2)
	assert.Equal(t, []string{"dönic.de"}, fieldsFromIDN.Values(QueryFieldNameDomainIDN))
	assert.Equal(t, []string{"xn--dnic-5qa.de"}, fieldsFromIDN.Values(QueryFieldNameDomainACE))

	fieldsFromACE := NewQueryFieldList()
	putDomainToQueryFields(&fieldsFromACE, "xn--dnic-5qa.de")
	require.Len(t, fieldsFromACE, 2)
	assert.Equal(t, []string{"dönic.de"}, fieldsFromACE.Values(QueryFieldNameDomainIDN))
	assert.Equal(t, []string{"xn--dnic-5qa.de"}, fieldsFromACE.Values(QueryFieldNameDomainACE))
}

func TestQueryEncodeKV(t *testing.T) {
	query, err := ParseQuery(qryEncode)
	require.NoError(t, err)
	assert.Equal(t, "version: 3.0\naction: update\naddress: foo\ndomain: denic.de\naddress: bar", query.EncodeKV())
}

func TestQueryNormalization(t *testing.T) {
	query, err := ParseQuery(qryNormalize)
	require.NoError(t, err)
	assert.Equal(t, LatestVersion, query.Version())
	assert.Equal(t, ActionInfo, query.Action())
	require.Len(t, query.Fields(), 3)
	assert.Equal(t, []string{"3.0"}, query.Field(QueryFieldNameVersion))
	assert.Equal(t, []string{"iNfO"}, query.Field(QueryFieldNameAction))
	assert.Equal(t, []string{"tRuE"}, query.Field(QueryFieldNameDisconnect))
}

func TestNewTransitDomainQuery(t *testing.T) {
	query := NewTransitDomainQuery("dönic.de", true)
	assert.Equal(t, LatestVersion, query.Version())
	assert.Equal(t, ActionTransit, query.Action())
	require.Len(t, query.Fields(), 5)
	assert.Equal(t, []string{string(LatestVersion)}, query.Field(QueryFieldNameVersion))
	assert.Equal(t, []string{string(ActionTransit)}, query.Field(QueryFieldNameAction))
	assert.Equal(t, []string{"dönic.de"}, query.Field(QueryFieldNameDomainIDN))
	assert.Equal(t, []string{"xn--dnic-5qa.de"}, query.Field(QueryFieldNameDomainACE))
	assert.Equal(t, []string{"true"}, query.Field(QueryFieldNameDisconnect))
}

func TestNewCreateAuthInfo1Query(t *testing.T) {
	query := NewCreateAuthInfo1Query("denic.de", "a-secret-auth-info", time.Date(2020, time.September, 25, 0, 0, 0, 0, time.Local))
	assert.Equal(t, LatestVersion, query.Version())
	assert.Equal(t, ActionCreateAuthInfo1, query.Action())
	require.Len(t, query.Fields(), 6)
	assert.Equal(t, []string{string(LatestVersion)}, query.Field(QueryFieldNameVersion))
	assert.Equal(t, []string{string(ActionCreateAuthInfo1)}, query.Field(QueryFieldNameAction))
	assert.Equal(t, []string{"denic.de"}, query.Field(QueryFieldNameDomainIDN))
	assert.Equal(t, []string{"denic.de"}, query.Field(QueryFieldNameDomainACE))
	assert.Equal(t, []string{"78152947f3751ab6baf0fb54c3c508d9b959f707999cbab855caaac231628c7f"}, query.Field(QueryFieldNameAuthInfoHash))
	assert.Equal(t, []string{"20200925"}, query.Field(QueryFieldNameAuthInfoExpire))
}

func TestNewVerifyDomainQuery(t *testing.T) {
	query := NewVerifyDomainQuery("denic.de", AuthorizedSignatory{
		FirstName:    "Donald",
		LastName:     "Duck",
		EMail:        "donald@duck.de",
		DateOfBirth:  time.Date(1934, time.June, 9, 0, 0, 0, 0, time.UTC),
		PlaceOfBirth: "Entenhausen",
		Phone:        "+0123 456789",
	})
	assert.Equal(t, LatestVersion, query.Version())
	assert.Equal(t, ActionVerify, query.Action())
	require.Len(t, query.Fields(), 10)
	assert.Equal(t, []string{string(LatestVersion)}, query.Field(QueryFieldNameVersion))
	assert.Equal(t, []string{string(ActionVerify)}, query.Field(QueryFieldNameAction))
	assert.Equal(t, []string{"denic.de"}, query.Field(QueryFieldNameDomainIDN))
	assert.Equal(t, []string{"denic.de"}, query.Field(QueryFieldNameDomainACE))
	assert.Equal(t, []string{"Donald"}, query.Field(QueryFieldNameAuthSigFirstName))
	assert.Equal(t, []string{"Duck"}, query.Field(QueryFieldNameAuthSigLastName))
	assert.Equal(t, []string{"donald@duck.de"}, query.Field(QueryFieldNameAuthSigEMail))
	assert.Equal(t, []string{"1934-06-09"}, query.Field(QueryFieldNameAuthSigDateOfBirth))
	assert.Equal(t, []string{"Entenhausen"}, query.Field(QueryFieldNameAuthSigPlaceOfBirth))
	assert.Equal(t, []string{"+0123 456789"}, query.Field(QueryFieldNameAuthSigPhone))
}

func TestParseQueryCasing(t *testing.T) {
	query, err := ParseQuery(qryIgnoreCasing)
	require.NoError(t, err)
	assert.Equal(t, LatestVersion, query.Version())
	assert.Equal(t, ActionLogin, query.Action())
	require.Len(t, query.Fields(), 4)
	assert.Equal(t, []string{string(LatestVersion)}, query.Field(QueryFieldNameVersion))
	assert.Equal(t, []string{"login"}, query.Field(QueryFieldNameAction))
	assert.Equal(t, []string{"DENIC-1000042-TEST"}, query.Field("uSeR"))
	assert.Equal(t, []string{"very-secure"}, query.Field(QueryFieldNamePassword))
}

func TestParseQueryWhitespaces(t *testing.T) {
	query, err := ParseQuery(qryWhitespaces)
	require.NoError(t, err)
	assert.Equal(t, LatestVersion, query.Version())
	assert.Equal(t, ActionLogin, query.Action())
	require.Len(t, query.Fields(), 4)
	assert.Equal(t, []string{string(LatestVersion)}, query.Field(QueryFieldNameVersion))
	assert.Equal(t, []string{string(ActionLogin)}, query.Field(QueryFieldNameAction))
	assert.Equal(t, []string{"DENIC-1000042-TEST"}, query.Field(QueryFieldNameUser))
	assert.Equal(t, []string{"very-secure"}, query.Field(QueryFieldNamePassword))
}

func TestParseQueryOrder(t *testing.T) {
	query, err := ParseQuery(qryOrder)
	require.NoError(t, err)
	assert.Equal(t, LatestVersion, query.Version())
	assert.Equal(t, ActionLogin, query.Action())
	require.Len(t, query.Fields(), 7)
	assert.Equal(t, QueryField{QueryFieldName("action"), string(ActionLogin)}, query.Fields()[0])
	assert.Equal(t, QueryField{QueryFieldName("custom"), "1"}, query.Fields()[1])
	assert.Equal(t, QueryField{QueryFieldName("version"), string(LatestVersion)}, query.Fields()[2])
	assert.Equal(t, QueryField{QueryFieldNameUser, "DENIC-1000042-TEST"}, query.Fields()[3])
	assert.Equal(t, QueryField{QueryFieldName("stuff"), "foobar"}, query.Fields()[4])
	assert.Equal(t, QueryField{QueryFieldNamePassword, "very-secure"}, query.Fields()[5])
	assert.Equal(t, QueryField{QueryFieldName("custom"), "2"}, query.Fields()[6])
}
