package rri

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	qryLoginLogout  = "version: 3.0\naction: LOGIN\nuser: DENIC-1000042-TEST\npassword: very-secure\n=-= now log out\nversion: 3.0\naction: LOGOUT\n"
	qryIgnoreCasing = "Version: 3.0\nAction: login\nUser: DENIC-1000042-TEST\nPassword: very-secure"
	qryWhitespaces  = "  version: \t3.0  \n\n\naction:    LOGIN\n   user: DENIC-1000042-TEST\npassword: very-secure    \n"
	qryOrder        = "action: LOGIN\ncustom: 1\nversion: 3.0\nuser: DENIC-1000042-TEST\nstuff: foobar\npassword: very-secure\ncustom: 2"
)

func TestNewTransitDomainQuery(t *testing.T) {
	query := NewTransitDomainQuery("dönic.de", true)
	assert.Equal(t, LatestVersion, query.Version())
	assert.Equal(t, ActionTransit, query.Action())
	assert.Len(t, query.Fields(), 3)
	assert.Equal(t, []string{"dönic.de"}, query.Field(FieldNameDomainIDN))
	assert.Equal(t, []string{"xn--dnic-5qa.de"}, query.Field(FieldNameDomainACE))
	assert.Equal(t, []string{"true"}, query.Field(FieldNameDisconnect))
}

func TestNewCreateAuthInfo1Query(t *testing.T) {
	query := NewCreateAuthInfo1Query("denic.de", "a-secret-auth-info", time.Date(2020, time.September, 25, 0, 0, 0, 0, time.Local))
	assert.Equal(t, LatestVersion, query.Version())
	assert.Equal(t, ActionCreateAuthInfo1, query.Action())
	assert.Len(t, query.Fields(), 4)
	assert.Equal(t, []string{"denic.de"}, query.Field(FieldNameDomainIDN))
	assert.Equal(t, []string{"denic.de"}, query.Field(FieldNameDomainACE))
	assert.Equal(t, []string{"78152947f3751ab6baf0fb54c3c508d9b959f707999cbab855caaac231628c7f"}, query.Field(FieldNameAuthInfoHash))
	assert.Equal(t, []string{"20200925"}, query.Field(FieldNameAuthInfoExpire))
}

func TestParseQueryCasing(t *testing.T) {
	query, err := ParseQuery(qryIgnoreCasing)
	if assert.NoError(t, err) {
		assert.Equal(t, LatestVersion, query.Version())
		assert.Equal(t, ActionLogin, query.Action())
		assert.Len(t, query.Fields(), 2)
		assert.Equal(t, []string{"DENIC-1000042-TEST"}, query.Field("uSeR"))
		assert.Equal(t, []string{"very-secure"}, query.Field(FieldNamePassword))
	}
}

func TestParseQueryWhitespaces(t *testing.T) {
	query, err := ParseQuery(qryWhitespaces)
	if assert.NoError(t, err) {
		assert.Equal(t, LatestVersion, query.Version())
		assert.Equal(t, ActionLogin, query.Action())
		assert.Len(t, query.Fields(), 2)
		assert.Equal(t, []string{"DENIC-1000042-TEST"}, query.Field(FieldNameUser))
		assert.Equal(t, []string{"very-secure"}, query.Field(FieldNamePassword))
	}
}

func TestParseQueryOrder(t *testing.T) {
	query, err := ParseQuery(qryOrder)
	if assert.NoError(t, err) {
		assert.Equal(t, LatestVersion, query.Version())
		assert.Equal(t, ActionLogin, query.Action())
		if assert.Len(t, query.Fields(), 5) {
			assert.Equal(t, QueryField{QueryFieldName("custom"), "1"}, query.Fields()[0])
			assert.Equal(t, QueryField{FieldNameUser, "DENIC-1000042-TEST"}, query.Fields()[1])
			assert.Equal(t, QueryField{QueryFieldName("stuff"), "foobar"}, query.Fields()[2])
			assert.Equal(t, QueryField{FieldNamePassword, "very-secure"}, query.Fields()[3])
			assert.Equal(t, QueryField{QueryFieldName("custom"), "2"}, query.Fields()[4])
		}
	}
}
