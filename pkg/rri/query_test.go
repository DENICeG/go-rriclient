package rri_test

import (
	"testing"
	"time"

	"github.com/DENICeG/go-rriclient/pkg/rri"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewQueryNil(t *testing.T) {
	var qry *rri.Query
	require.NotPanics(t, func() {
		qry = rri.NewQuery(rri.LatestVersion, rri.ActionLogout, nil)
	})
	assert.Equal(t, 2, qry.Fields().Size())
}

func TestQueryToString(t *testing.T) {
	assert.Equal(t, "LOGIN{\"DENIC-1000011-TEST\"}", rri.NewLoginQuery("DENIC-1000011-TEST", "secret").String())
	assert.Equal(t, "LOGOUT{}", rri.NewLogoutQuery().String())
	// TODO other actions
}

func TestQueryEncodeKV(t *testing.T) {
	query, err := rri.ParseQuery("Version: 5.0\nAction: update\naddress: foo\nDomain: denic.de\nAddress: bar")
	require.NoError(t, err)
	require.NotNil(t, query)
	assert.Equal(t, "version: 5.0\naction: update\naddress: foo\ndomain: denic.de\naddress: bar", query.EncodeKV())
}

func TestNewLoginQuery(t *testing.T) {
	query := rri.NewLoginQuery("DENIC-1000011-TEST", "secret")
	require.NotNil(t, query)
	assert.Equal(t, rri.LatestVersion, query.Version())
	assert.Equal(t, rri.ActionLogin, query.Action())
	require.Len(t, query.Fields(), 4)
	assert.Equal(t, []string{string(rri.LatestVersion)}, query.Field(rri.QueryFieldNameVersion))
	assert.Equal(t, []string{string(rri.ActionLogin)}, query.Field(rri.QueryFieldNameAction))
	assert.Equal(t, []string{"DENIC-1000011-TEST"}, query.Field(rri.QueryFieldNameUser))
	assert.Equal(t, []string{"secret"}, query.Field(rri.QueryFieldNamePassword))
}

func TestNewLogoutQuery(t *testing.T) {
	query := rri.NewLogoutQuery()
	require.NotNil(t, query)
	assert.Equal(t, rri.LatestVersion, query.Version())
	assert.Equal(t, rri.ActionLogout, query.Action())
	require.Len(t, query.Fields(), 2)
	assert.Equal(t, []string{string(rri.LatestVersion)}, query.Field(rri.QueryFieldNameVersion))
	assert.Equal(t, []string{string(rri.ActionLogout)}, query.Field(rri.QueryFieldNameAction))
}

func TestNewCheckHandleQuery(t *testing.T) {
	query := rri.NewCheckHandleQuery(rri.NewDenicHandle(1000011, "SOME-DUDE"))
	require.NotNil(t, query)
	assert.Equal(t, rri.LatestVersion, query.Version())
	assert.Equal(t, rri.ActionCheck, query.Action())
	require.Len(t, query.Fields(), 3)
	assert.Equal(t, []string{string(rri.LatestVersion)}, query.Field(rri.QueryFieldNameVersion))
	assert.Equal(t, []string{string(rri.ActionCheck)}, query.Field(rri.QueryFieldNameAction))
	assert.Equal(t, []string{"DENIC-1000011-SOME-DUDE"}, query.Field(rri.QueryFieldNameHandle))
}

func TestNewInfoHandleQuery(t *testing.T) {
	query := rri.NewInfoHandleQuery(rri.NewDenicHandle(1000011, "SOME-DUDE"))
	require.NotNil(t, query)
	assert.Equal(t, rri.LatestVersion, query.Version())
	assert.Equal(t, rri.ActionInfo, query.Action())
	require.Len(t, query.Fields(), 3)
	assert.Equal(t, []string{string(rri.LatestVersion)}, query.Field(rri.QueryFieldNameVersion))
	assert.Equal(t, []string{string(rri.ActionInfo)}, query.Field(rri.QueryFieldNameAction))
	assert.Equal(t, []string{"DENIC-1000011-SOME-DUDE"}, query.Field(rri.QueryFieldNameHandle))
}

func TestPutDomainToQueryFields(t *testing.T) {
	fieldsFromIDN := rri.NewQueryFieldList()
	rri.PutDomainToQueryFields(&fieldsFromIDN, "dönic.de")
	require.Len(t, fieldsFromIDN, 2)
	assert.Equal(t, []string{"dönic.de"}, fieldsFromIDN.Values(rri.QueryFieldNameDomainIDN))
	assert.Equal(t, []string{"xn--dnic-5qa.de"}, fieldsFromIDN.Values(rri.QueryFieldNameDomainACE))

	fieldsFromACE := rri.NewQueryFieldList()
	rri.PutDomainToQueryFields(&fieldsFromACE, "xn--dnic-5qa.de")
	require.Len(t, fieldsFromACE, 2)
	assert.Equal(t, []string{"dönic.de"}, fieldsFromACE.Values(rri.QueryFieldNameDomainIDN))
	assert.Equal(t, []string{"xn--dnic-5qa.de"}, fieldsFromACE.Values(rri.QueryFieldNameDomainACE))
}

func TestNewCreateDomainQuery(t *testing.T) {
	query := rri.NewCreateDomainQuery("denic.de", rri.DomainData{
		HolderHandles:         []rri.DenicHandle{rri.NewDenicHandle(1000011, "HOLDER-DUDE")},
		GeneralRequestHandles: []rri.DenicHandle{rri.NewDenicHandle(1000011, "REQUEST-DUDE")},
		AbuseContactHandles:   []rri.DenicHandle{rri.NewDenicHandle(1000011, "ABUSE-DUDE")},
		NameServers:           []string{"ns1.denic.de", "ns2.denic.de"},
	})
	require.NotNil(t, query)
	assert.Equal(t, rri.LatestVersion, query.Version())
	assert.Equal(t, rri.ActionCreate, query.Action())
	require.Len(t, query.Fields(), 9)
	assert.Equal(t, []string{string(rri.LatestVersion)}, query.Field(rri.QueryFieldNameVersion))
	assert.Equal(t, []string{string(rri.ActionCreate)}, query.Field(rri.QueryFieldNameAction))
	assert.Equal(t, []string{"denic.de"}, query.Field(rri.QueryFieldNameDomainIDN))
	assert.Equal(t, []string{"denic.de"}, query.Field(rri.QueryFieldNameDomainACE))
	assert.Equal(t, []string{"DENIC-1000011-HOLDER-DUDE"}, query.Field(rri.QueryFieldNameHolder))
	assert.Equal(t, []string{"DENIC-1000011-REQUEST-DUDE"}, query.Field(rri.QueryFieldNameGeneralRequest))
	assert.Equal(t, []string{"DENIC-1000011-ABUSE-DUDE"}, query.Field(rri.QueryFieldNameAbuseContact))
	assert.Equal(t, []string{"ns1.denic.de", "ns2.denic.de"}, query.Field(rri.QueryFieldNameNameServer))
}

func TestNewUpdateDomainQuery(t *testing.T) {
	query := rri.NewUpdateDomainQuery("denic.de", rri.DomainData{
		HolderHandles:         []rri.DenicHandle{rri.NewDenicHandle(1000011, "HOLDER-DUDE")},
		GeneralRequestHandles: []rri.DenicHandle{rri.NewDenicHandle(1000011, "REQUEST-DUDE")},
		AbuseContactHandles:   []rri.DenicHandle{rri.NewDenicHandle(1000011, "ABUSE-DUDE")},
		NameServers:           []string{"ns1.denic.de", "ns2.denic.de"},
	})
	require.NotNil(t, query)
	assert.Equal(t, rri.LatestVersion, query.Version())
	assert.Equal(t, rri.ActionUpdate, query.Action())
	require.Len(t, query.Fields(), 9)
	assert.Equal(t, []string{string(rri.LatestVersion)}, query.Field(rri.QueryFieldNameVersion))
	assert.Equal(t, []string{string(rri.ActionUpdate)}, query.Field(rri.QueryFieldNameAction))
	assert.Equal(t, []string{"denic.de"}, query.Field(rri.QueryFieldNameDomainIDN))
	assert.Equal(t, []string{"denic.de"}, query.Field(rri.QueryFieldNameDomainACE))
	assert.Equal(t, []string{"DENIC-1000011-HOLDER-DUDE"}, query.Field(rri.QueryFieldNameHolder))
	assert.Equal(t, []string{"DENIC-1000011-REQUEST-DUDE"}, query.Field(rri.QueryFieldNameGeneralRequest))
	assert.Equal(t, []string{"DENIC-1000011-ABUSE-DUDE"}, query.Field(rri.QueryFieldNameAbuseContact))
	assert.Equal(t, []string{"ns1.denic.de", "ns2.denic.de"}, query.Field(rri.QueryFieldNameNameServer))
}

func TestNewChangeHolderQuery(t *testing.T) {
	query := rri.NewChangeHolderQuery("denic.de", rri.DomainData{
		HolderHandles:         []rri.DenicHandle{rri.NewDenicHandle(1000011, "HOLDER-DUDE")},
		GeneralRequestHandles: []rri.DenicHandle{rri.NewDenicHandle(1000011, "REQUEST-DUDE")},
		AbuseContactHandles:   []rri.DenicHandle{rri.NewDenicHandle(1000011, "ABUSE-DUDE")},
		NameServers:           []string{"ns1.denic.de", "ns2.denic.de"},
	})
	require.NotNil(t, query)
	assert.Equal(t, rri.LatestVersion, query.Version())
	assert.Equal(t, rri.ActionChangeHolder, query.Action())
	require.Len(t, query.Fields(), 9)
	assert.Equal(t, []string{string(rri.LatestVersion)}, query.Field(rri.QueryFieldNameVersion))
	assert.Equal(t, []string{string(rri.ActionChangeHolder)}, query.Field(rri.QueryFieldNameAction))
	assert.Equal(t, []string{"denic.de"}, query.Field(rri.QueryFieldNameDomainIDN))
	assert.Equal(t, []string{"denic.de"}, query.Field(rri.QueryFieldNameDomainACE))
	assert.Equal(t, []string{"DENIC-1000011-HOLDER-DUDE"}, query.Field(rri.QueryFieldNameHolder))
	assert.Equal(t, []string{"DENIC-1000011-REQUEST-DUDE"}, query.Field(rri.QueryFieldNameGeneralRequest))
	assert.Equal(t, []string{"DENIC-1000011-ABUSE-DUDE"}, query.Field(rri.QueryFieldNameAbuseContact))
	assert.Equal(t, []string{"ns1.denic.de", "ns2.denic.de"}, query.Field(rri.QueryFieldNameNameServer))
}

func TestNewTransitDomainQuery(t *testing.T) {
	query := rri.NewTransitDomainQuery("dönic.de", false)
	require.NotNil(t, query)
	assert.Equal(t, rri.LatestVersion, query.Version())
	assert.Equal(t, rri.ActionTransit, query.Action())
	require.Len(t, query.Fields(), 5)
	assert.Equal(t, []string{string(rri.LatestVersion)}, query.Field(rri.QueryFieldNameVersion))
	assert.Equal(t, []string{string(rri.ActionTransit)}, query.Field(rri.QueryFieldNameAction))
	assert.Equal(t, []string{"dönic.de"}, query.Field(rri.QueryFieldNameDomainIDN))
	assert.Equal(t, []string{"xn--dnic-5qa.de"}, query.Field(rri.QueryFieldNameDomainACE))
	assert.Equal(t, []string{"false"}, query.Field(rri.QueryFieldNameDisconnect))
}

func TestNewTransitDomainWithDisconnectQuery(t *testing.T) {
	query := rri.NewTransitDomainQuery("dönic.de", true)
	require.NotNil(t, query)
	assert.Equal(t, rri.LatestVersion, query.Version())
	assert.Equal(t, rri.ActionTransit, query.Action())
	require.Len(t, query.Fields(), 5)
	assert.Equal(t, []string{string(rri.LatestVersion)}, query.Field(rri.QueryFieldNameVersion))
	assert.Equal(t, []string{string(rri.ActionTransit)}, query.Field(rri.QueryFieldNameAction))
	assert.Equal(t, []string{"dönic.de"}, query.Field(rri.QueryFieldNameDomainIDN))
	assert.Equal(t, []string{"xn--dnic-5qa.de"}, query.Field(rri.QueryFieldNameDomainACE))
	assert.Equal(t, []string{"true"}, query.Field(rri.QueryFieldNameDisconnect))
}

func TestNewCreateAuthInfo1Query(t *testing.T) {
	query := rri.NewCreateAuthInfo1Query("denic.de", "a-secret-auth-info", time.Date(2020, time.September, 25, 0, 0, 0, 0, time.Local))
	require.NotNil(t, query)
	assert.Equal(t, rri.LatestVersion, query.Version())
	assert.Equal(t, rri.ActionCreateAuthInfo1, query.Action())
	require.Len(t, query.Fields(), 6)
	assert.Equal(t, []string{string(rri.LatestVersion)}, query.Field(rri.QueryFieldNameVersion))
	assert.Equal(t, []string{string(rri.ActionCreateAuthInfo1)}, query.Field(rri.QueryFieldNameAction))
	assert.Equal(t, []string{"denic.de"}, query.Field(rri.QueryFieldNameDomainIDN))
	assert.Equal(t, []string{"denic.de"}, query.Field(rri.QueryFieldNameDomainACE))
	assert.Equal(t, []string{"78152947f3751ab6baf0fb54c3c508d9b959f707999cbab855caaac231628c7f"}, query.Field(rri.QueryFieldNameAuthInfoHash))
	assert.Equal(t, []string{"20200925"}, query.Field(rri.QueryFieldNameAuthInfoExpire))
}

func TestNewQueueReadQuery(t *testing.T) {
	query := rri.NewQueueReadQuery("")
	require.NotNil(t, query)
	assert.Equal(t, rri.LatestVersion, query.Version())
	assert.Equal(t, rri.ActionQueueRead, query.Action())
	require.Len(t, query.Fields(), 2)
	assert.Equal(t, []string{string(rri.LatestVersion)}, query.Field(rri.QueryFieldNameVersion))
	assert.Equal(t, []string{string(rri.ActionQueueRead)}, query.Field(rri.QueryFieldNameAction))
}

func TestNewQueueReadQueryWithType(t *testing.T) {
	query := rri.NewQueueReadQuery("authInfo2Delete")
	require.NotNil(t, query)
	assert.Equal(t, rri.LatestVersion, query.Version())
	assert.Equal(t, rri.ActionQueueRead, query.Action())
	require.Len(t, query.Fields(), 3)
	assert.Equal(t, []string{string(rri.LatestVersion)}, query.Field(rri.QueryFieldNameVersion))
	assert.Equal(t, []string{string(rri.ActionQueueRead)}, query.Field(rri.QueryFieldNameAction))
	assert.Equal(t, []string{string("authInfo2Delete")}, query.Field(rri.QueryFieldNameMsgType))
}

func TestNewQueueDeleteQuery(t *testing.T) {
	query := rri.NewQueueDeleteQuery("5c214b14-c919-11eb-a37b-0242ac130003", "")
	require.NotNil(t, query)
	assert.Equal(t, rri.LatestVersion, query.Version())
	assert.Equal(t, rri.ActionQueueDelete, query.Action())
	require.Len(t, query.Fields(), 3)
	assert.Equal(t, []string{string(rri.LatestVersion)}, query.Field(rri.QueryFieldNameVersion))
	assert.Equal(t, []string{string(rri.ActionQueueDelete)}, query.Field(rri.QueryFieldNameAction))
	assert.Equal(t, []string{"5c214b14-c919-11eb-a37b-0242ac130003"}, query.Field(rri.QueryFieldNameMsgID))
}

func TestNewQueueDeleteQueryWithType(t *testing.T) {
	query := rri.NewQueueDeleteQuery("5c214b14-c919-11eb-a37b-0242ac130003", "expireWarning")
	require.NotNil(t, query)
	assert.Equal(t, rri.LatestVersion, query.Version())
	assert.Equal(t, rri.ActionQueueDelete, query.Action())
	require.Len(t, query.Fields(), 4)
	assert.Equal(t, []string{string(rri.LatestVersion)}, query.Field(rri.QueryFieldNameVersion))
	assert.Equal(t, []string{string(rri.ActionQueueDelete)}, query.Field(rri.QueryFieldNameAction))
	assert.Equal(t, []string{"5c214b14-c919-11eb-a37b-0242ac130003"}, query.Field(rri.QueryFieldNameMsgID))
	assert.Equal(t, []string{string("expireWarning")}, query.Field(rri.QueryFieldNameMsgType))
}

func TestQueryNormalization(t *testing.T) {
	query, err := rri.ParseQuery("Version:   5.0   \nAction: iNfO\ndIsCoNnEcT: tRuE")
	require.NoError(t, err)
	require.NotNil(t, query)
	assert.Equal(t, rri.LatestVersion, query.Version())
	assert.Equal(t, rri.ActionInfo, query.Action())
	require.Len(t, query.Fields(), 3)
	assert.Equal(t, []string{"5.0"}, query.Field(rri.QueryFieldNameVersion))
	assert.Equal(t, []string{"iNfO"}, query.Field(rri.QueryFieldNameAction))
	assert.Equal(t, []string{"tRuE"}, query.Field(rri.QueryFieldNameDisconnect))
}

func TestParseQueryCasing(t *testing.T) {
	query, err := rri.ParseQuery("Version: 5.0\nAction: login\nUser: DENIC-1000042-TEST\nPassword: very-secure")
	require.NoError(t, err)
	require.NotNil(t, query)
	assert.Equal(t, rri.LatestVersion, query.Version())
	assert.Equal(t, rri.ActionLogin, query.Action())
	require.Len(t, query.Fields(), 4)
	assert.Equal(t, []string{string(rri.LatestVersion)}, query.Field(rri.QueryFieldNameVersion))
	assert.Equal(t, []string{"login"}, query.Field(rri.QueryFieldNameAction))
	assert.Equal(t, []string{"DENIC-1000042-TEST"}, query.Field("uSeR"))
	assert.Equal(t, []string{"very-secure"}, query.Field(rri.QueryFieldNamePassword))
}

func TestParseQueryWhitespaces(t *testing.T) {
	query, err := rri.ParseQuery("  version: \t5.0  \n\n\naction:    LOGIN\n   user: DENIC-1000042-TEST\npassword: very-secure    \n")
	require.NoError(t, err)
	require.NotNil(t, query)
	assert.Equal(t, rri.LatestVersion, query.Version())
	assert.Equal(t, rri.ActionLogin, query.Action())
	require.Len(t, query.Fields(), 4)
	assert.Equal(t, []string{string(rri.LatestVersion)}, query.Field(rri.QueryFieldNameVersion))
	assert.Equal(t, []string{string(rri.ActionLogin)}, query.Field(rri.QueryFieldNameAction))
	assert.Equal(t, []string{"DENIC-1000042-TEST"}, query.Field(rri.QueryFieldNameUser))
	assert.Equal(t, []string{"very-secure"}, query.Field(rri.QueryFieldNamePassword))
}

func TestParseQueryOrder(t *testing.T) {
	query, err := rri.ParseQuery("action: LOGIN\ncustom: 1\nversion: 5.0\nuser: DENIC-1000042-TEST\nstuff: foobar\npassword: very-secure\ncustom: 2")
	require.NoError(t, err)
	require.NotNil(t, query)
	assert.Equal(t, rri.LatestVersion, query.Version())
	assert.Equal(t, rri.ActionLogin, query.Action())
	require.Len(t, query.Fields(), 7)
	assert.Equal(t, rri.QueryField{rri.QueryFieldName("action"), string(rri.ActionLogin)}, query.Fields()[0])
	assert.Equal(t, rri.QueryField{rri.QueryFieldName("custom"), "1"}, query.Fields()[1])
	assert.Equal(t, rri.QueryField{rri.QueryFieldName("version"), string(rri.LatestVersion)}, query.Fields()[2])
	assert.Equal(t, rri.QueryField{rri.QueryFieldNameUser, "DENIC-1000042-TEST"}, query.Fields()[3])
	assert.Equal(t, rri.QueryField{rri.QueryFieldName("stuff"), "foobar"}, query.Fields()[4])
	assert.Equal(t, rri.QueryField{rri.QueryFieldNamePassword, "very-secure"}, query.Fields()[5])
	assert.Equal(t, rri.QueryField{rri.QueryFieldName("custom"), "2"}, query.Fields()[6])
}
