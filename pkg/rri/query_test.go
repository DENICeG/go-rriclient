package rri

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewQueryNil(t *testing.T) {
	var qry *Query
	require.NotPanics(t, func() {
		qry = NewQuery(LatestVersion, ActionLogout, nil)
	})
	assert.Equal(t, 2, qry.Fields().Size())
}

func TestQueryToString(t *testing.T) {
	assert.Equal(t, "LOGIN{\"DENIC-1000011-TEST\"}", NewLoginQuery("DENIC-1000011-TEST", "secret").String())
	assert.Equal(t, "LOGOUT{}", NewLogoutQuery().String())
	//TODO other actions
}

func TestQueryEncodeKV(t *testing.T) {
	query, err := ParseQuery("Version: 5.0\nAction: update\naddress: foo\nDomain: denic.de\nAddress: bar")
	require.NoError(t, err)
	require.NotNil(t, query)
	assert.Equal(t, "version: 5.0\naction: update\naddress: foo\ndomain: denic.de\naddress: bar", query.EncodeKV())
}

func TestNewLoginQuery(t *testing.T) {
	query := NewLoginQuery("DENIC-1000011-TEST", "secret")
	require.NotNil(t, query)
	assert.Equal(t, LatestVersion, query.Version())
	assert.Equal(t, ActionLogin, query.Action())
	require.Len(t, query.Fields(), 4)
	assert.Equal(t, []string{string(LatestVersion)}, query.Field(QueryFieldNameVersion))
	assert.Equal(t, []string{string(ActionLogin)}, query.Field(QueryFieldNameAction))
	assert.Equal(t, []string{"DENIC-1000011-TEST"}, query.Field(QueryFieldNameUser))
	assert.Equal(t, []string{"secret"}, query.Field(QueryFieldNamePassword))
}

func TestNewLogoutQuery(t *testing.T) {
	query := NewLogoutQuery()
	require.NotNil(t, query)
	assert.Equal(t, LatestVersion, query.Version())
	assert.Equal(t, ActionLogout, query.Action())
	require.Len(t, query.Fields(), 2)
	assert.Equal(t, []string{string(LatestVersion)}, query.Field(QueryFieldNameVersion))
	assert.Equal(t, []string{string(ActionLogout)}, query.Field(QueryFieldNameAction))
}

func TestNewCheckHandleQuery(t *testing.T) {
	query := NewCheckHandleQuery(NewDenicHandle(1000011, "SOME-DUDE"))
	require.NotNil(t, query)
	assert.Equal(t, LatestVersion, query.Version())
	assert.Equal(t, ActionCheck, query.Action())
	require.Len(t, query.Fields(), 3)
	assert.Equal(t, []string{string(LatestVersion)}, query.Field(QueryFieldNameVersion))
	assert.Equal(t, []string{string(ActionCheck)}, query.Field(QueryFieldNameAction))
	assert.Equal(t, []string{"DENIC-1000011-SOME-DUDE"}, query.Field(QueryFieldNameHandle))
}

func TestNewInfoHandleQuery(t *testing.T) {
	query := NewInfoHandleQuery(NewDenicHandle(1000011, "SOME-DUDE"))
	require.NotNil(t, query)
	assert.Equal(t, LatestVersion, query.Version())
	assert.Equal(t, ActionInfo, query.Action())
	require.Len(t, query.Fields(), 3)
	assert.Equal(t, []string{string(LatestVersion)}, query.Field(QueryFieldNameVersion))
	assert.Equal(t, []string{string(ActionInfo)}, query.Field(QueryFieldNameAction))
	assert.Equal(t, []string{"DENIC-1000011-SOME-DUDE"}, query.Field(QueryFieldNameHandle))
}

func TestPutDomainToQueryFields(t *testing.T) {
	fieldsFromIDN := NewQueryFieldList()
	PutDomainToQueryFields(&fieldsFromIDN, "dönic.de")
	require.Len(t, fieldsFromIDN, 2)
	assert.Equal(t, []string{"dönic.de"}, fieldsFromIDN.Values(QueryFieldNameDomainIDN))
	assert.Equal(t, []string{"xn--dnic-5qa.de"}, fieldsFromIDN.Values(QueryFieldNameDomainACE))

	fieldsFromACE := NewQueryFieldList()
	PutDomainToQueryFields(&fieldsFromACE, "xn--dnic-5qa.de")
	require.Len(t, fieldsFromACE, 2)
	assert.Equal(t, []string{"dönic.de"}, fieldsFromACE.Values(QueryFieldNameDomainIDN))
	assert.Equal(t, []string{"xn--dnic-5qa.de"}, fieldsFromACE.Values(QueryFieldNameDomainACE))
}

func TestNewCreateDomainQuery(t *testing.T) {
	query := NewCreateDomainQuery("denic.de", DomainData{
		HolderHandles:         []DenicHandle{NewDenicHandle(1000011, "HOLDER-DUDE")},
		GeneralRequestHandles: []DenicHandle{NewDenicHandle(1000011, "REQUEST-DUDE")},
		AbuseContactHandles:   []DenicHandle{NewDenicHandle(1000011, "ABUSE-DUDE")},
		NameServers:           []string{"ns1.denic.de", "ns2.denic.de"},
	})
	require.NotNil(t, query)
	assert.Equal(t, LatestVersion, query.Version())
	assert.Equal(t, ActionCreate, query.Action())
	require.Len(t, query.Fields(), 9)
	assert.Equal(t, []string{string(LatestVersion)}, query.Field(QueryFieldNameVersion))
	assert.Equal(t, []string{string(ActionCreate)}, query.Field(QueryFieldNameAction))
	assert.Equal(t, []string{"denic.de"}, query.Field(QueryFieldNameDomainIDN))
	assert.Equal(t, []string{"denic.de"}, query.Field(QueryFieldNameDomainACE))
	assert.Equal(t, []string{"DENIC-1000011-HOLDER-DUDE"}, query.Field(QueryFieldNameHolder))
	assert.Equal(t, []string{"DENIC-1000011-REQUEST-DUDE"}, query.Field(QueryFieldNameGeneralRequest))
	assert.Equal(t, []string{"DENIC-1000011-ABUSE-DUDE"}, query.Field(QueryFieldNameAbuseContact))
	assert.Equal(t, []string{"ns1.denic.de", "ns2.denic.de"}, query.Field(QueryFieldNameNameServer))
}

func TestNewUpdateDomainQuery(t *testing.T) {
	query := NewUpdateDomainQuery("denic.de", DomainData{
		HolderHandles:         []DenicHandle{NewDenicHandle(1000011, "HOLDER-DUDE")},
		GeneralRequestHandles: []DenicHandle{NewDenicHandle(1000011, "REQUEST-DUDE")},
		AbuseContactHandles:   []DenicHandle{NewDenicHandle(1000011, "ABUSE-DUDE")},
		NameServers:           []string{"ns1.denic.de", "ns2.denic.de"},
	})
	require.NotNil(t, query)
	assert.Equal(t, LatestVersion, query.Version())
	assert.Equal(t, ActionUpdate, query.Action())
	require.Len(t, query.Fields(), 9)
	assert.Equal(t, []string{string(LatestVersion)}, query.Field(QueryFieldNameVersion))
	assert.Equal(t, []string{string(ActionUpdate)}, query.Field(QueryFieldNameAction))
	assert.Equal(t, []string{"denic.de"}, query.Field(QueryFieldNameDomainIDN))
	assert.Equal(t, []string{"denic.de"}, query.Field(QueryFieldNameDomainACE))
	assert.Equal(t, []string{"DENIC-1000011-HOLDER-DUDE"}, query.Field(QueryFieldNameHolder))
	assert.Equal(t, []string{"DENIC-1000011-REQUEST-DUDE"}, query.Field(QueryFieldNameGeneralRequest))
	assert.Equal(t, []string{"DENIC-1000011-ABUSE-DUDE"}, query.Field(QueryFieldNameAbuseContact))
	assert.Equal(t, []string{"ns1.denic.de", "ns2.denic.de"}, query.Field(QueryFieldNameNameServer))
}

func TestNewChangeHolderQuery(t *testing.T) {
	query := NewChangeHolderQuery("denic.de", DomainData{
		HolderHandles:         []DenicHandle{NewDenicHandle(1000011, "HOLDER-DUDE")},
		GeneralRequestHandles: []DenicHandle{NewDenicHandle(1000011, "REQUEST-DUDE")},
		AbuseContactHandles:   []DenicHandle{NewDenicHandle(1000011, "ABUSE-DUDE")},
		NameServers:           []string{"ns1.denic.de", "ns2.denic.de"},
	})
	require.NotNil(t, query)
	assert.Equal(t, LatestVersion, query.Version())
	assert.Equal(t, ActionChangeHolder, query.Action())
	require.Len(t, query.Fields(), 9)
	assert.Equal(t, []string{string(LatestVersion)}, query.Field(QueryFieldNameVersion))
	assert.Equal(t, []string{string(ActionChangeHolder)}, query.Field(QueryFieldNameAction))
	assert.Equal(t, []string{"denic.de"}, query.Field(QueryFieldNameDomainIDN))
	assert.Equal(t, []string{"denic.de"}, query.Field(QueryFieldNameDomainACE))
	assert.Equal(t, []string{"DENIC-1000011-HOLDER-DUDE"}, query.Field(QueryFieldNameHolder))
	assert.Equal(t, []string{"DENIC-1000011-REQUEST-DUDE"}, query.Field(QueryFieldNameGeneralRequest))
	assert.Equal(t, []string{"DENIC-1000011-ABUSE-DUDE"}, query.Field(QueryFieldNameAbuseContact))
	assert.Equal(t, []string{"ns1.denic.de", "ns2.denic.de"}, query.Field(QueryFieldNameNameServer))
}

func TestNewTransitDomainQuery(t *testing.T) {
	query := NewTransitDomainQuery("dönic.de", false)
	require.NotNil(t, query)
	assert.Equal(t, LatestVersion, query.Version())
	assert.Equal(t, ActionTransit, query.Action())
	require.Len(t, query.Fields(), 5)
	assert.Equal(t, []string{string(LatestVersion)}, query.Field(QueryFieldNameVersion))
	assert.Equal(t, []string{string(ActionTransit)}, query.Field(QueryFieldNameAction))
	assert.Equal(t, []string{"dönic.de"}, query.Field(QueryFieldNameDomainIDN))
	assert.Equal(t, []string{"xn--dnic-5qa.de"}, query.Field(QueryFieldNameDomainACE))
	assert.Equal(t, []string{"false"}, query.Field(QueryFieldNameDisconnect))
}

func TestNewTransitDomainWithDisconnectQuery(t *testing.T) {
	query := NewTransitDomainQuery("dönic.de", true)
	require.NotNil(t, query)
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
	require.NotNil(t, query)
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

func TestNewQueueReadQuery(t *testing.T) {
	query := NewQueueReadQuery("")
	require.NotNil(t, query)
	assert.Equal(t, LatestVersion, query.Version())
	assert.Equal(t, ActionQueueRead, query.Action())
	require.Len(t, query.Fields(), 2)
	assert.Equal(t, []string{string(LatestVersion)}, query.Field(QueryFieldNameVersion))
	assert.Equal(t, []string{string(ActionQueueRead)}, query.Field(QueryFieldNameAction))
}

func TestNewQueueReadQueryWithType(t *testing.T) {
	query := NewQueueReadQuery("authInfo2Delete")
	require.NotNil(t, query)
	assert.Equal(t, LatestVersion, query.Version())
	assert.Equal(t, ActionQueueRead, query.Action())
	require.Len(t, query.Fields(), 3)
	assert.Equal(t, []string{string(LatestVersion)}, query.Field(QueryFieldNameVersion))
	assert.Equal(t, []string{string(ActionQueueRead)}, query.Field(QueryFieldNameAction))
	assert.Equal(t, []string{string("authInfo2Delete")}, query.Field(QueryFieldNameMsgType))
}

func TestNewQueueDeleteQuery(t *testing.T) {
	query := NewQueueDeleteQuery("5c214b14-c919-11eb-a37b-0242ac130003", "")
	require.NotNil(t, query)
	assert.Equal(t, LatestVersion, query.Version())
	assert.Equal(t, ActionQueueDelete, query.Action())
	require.Len(t, query.Fields(), 3)
	assert.Equal(t, []string{string(LatestVersion)}, query.Field(QueryFieldNameVersion))
	assert.Equal(t, []string{string(ActionQueueDelete)}, query.Field(QueryFieldNameAction))
	assert.Equal(t, []string{"5c214b14-c919-11eb-a37b-0242ac130003"}, query.Field(QueryFieldNameMsgID))
}

func TestNewQueueDeleteQueryWithType(t *testing.T) {
	query := NewQueueDeleteQuery("5c214b14-c919-11eb-a37b-0242ac130003", "expireWarning")
	require.NotNil(t, query)
	assert.Equal(t, LatestVersion, query.Version())
	assert.Equal(t, ActionQueueDelete, query.Action())
	require.Len(t, query.Fields(), 4)
	assert.Equal(t, []string{string(LatestVersion)}, query.Field(QueryFieldNameVersion))
	assert.Equal(t, []string{string(ActionQueueDelete)}, query.Field(QueryFieldNameAction))
	assert.Equal(t, []string{"5c214b14-c919-11eb-a37b-0242ac130003"}, query.Field(QueryFieldNameMsgID))
	assert.Equal(t, []string{string("expireWarning")}, query.Field(QueryFieldNameMsgType))
}

func TestQueryNormalization(t *testing.T) {
	query, err := ParseQuery("Version:   5.0   \nAction: iNfO\ndIsCoNnEcT: tRuE")
	require.NoError(t, err)
	require.NotNil(t, query)
	assert.Equal(t, LatestVersion, query.Version())
	assert.Equal(t, ActionInfo, query.Action())
	require.Len(t, query.Fields(), 3)
	assert.Equal(t, []string{"5.0"}, query.Field(QueryFieldNameVersion))
	assert.Equal(t, []string{"iNfO"}, query.Field(QueryFieldNameAction))
	assert.Equal(t, []string{"tRuE"}, query.Field(QueryFieldNameDisconnect))
}

func TestParseQueryCasing(t *testing.T) {
	query, err := ParseQuery("Version: 5.0\nAction: login\nUser: DENIC-1000042-TEST\nPassword: very-secure")
	require.NoError(t, err)
	require.NotNil(t, query)
	assert.Equal(t, LatestVersion, query.Version())
	assert.Equal(t, ActionLogin, query.Action())
	require.Len(t, query.Fields(), 4)
	assert.Equal(t, []string{string(LatestVersion)}, query.Field(QueryFieldNameVersion))
	assert.Equal(t, []string{"login"}, query.Field(QueryFieldNameAction))
	assert.Equal(t, []string{"DENIC-1000042-TEST"}, query.Field("uSeR"))
	assert.Equal(t, []string{"very-secure"}, query.Field(QueryFieldNamePassword))
}

func TestParseQueryWhitespaces(t *testing.T) {
	query, err := ParseQuery("  version: \t5.0  \n\n\naction:    LOGIN\n   user: DENIC-1000042-TEST\npassword: very-secure    \n")
	require.NoError(t, err)
	require.NotNil(t, query)
	assert.Equal(t, LatestVersion, query.Version())
	assert.Equal(t, ActionLogin, query.Action())
	require.Len(t, query.Fields(), 4)
	assert.Equal(t, []string{string(LatestVersion)}, query.Field(QueryFieldNameVersion))
	assert.Equal(t, []string{string(ActionLogin)}, query.Field(QueryFieldNameAction))
	assert.Equal(t, []string{"DENIC-1000042-TEST"}, query.Field(QueryFieldNameUser))
	assert.Equal(t, []string{"very-secure"}, query.Field(QueryFieldNamePassword))
}

func TestParseQueryOrder(t *testing.T) {
	query, err := ParseQuery("action: LOGIN\ncustom: 1\nversion: 5.0\nuser: DENIC-1000042-TEST\nstuff: foobar\npassword: very-secure\ncustom: 2")
	require.NoError(t, err)
	require.NotNil(t, query)
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
