package rri

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	qryLoginLogout  = "version: 3.0\naction: LOGIN\nuser: DENIC-1000042-TEST\npassword: very-secure\n=-= now log out\nversion: 3.0\naction: LOGOUT\n"
	qryIgnoreCasing = "Version: 3.0\nAction: login\nUser: DENIC-1000042-TEST\nPassword: very-secure"
	qryWhitespaces  = "  version: \t3.0  \n\n\naction:    LOGIN\n   user: DENIC-1000042-TEST\npassword: very-secure    \n"
	qryOrder        = "action: LOGIN\ncustom: 1\nversion: 3.0\nuser: DENIC-1000042-TEST\nstuff: foobar\npassword: very-secure\ncustom: 2"
)

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
