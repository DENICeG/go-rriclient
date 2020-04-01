package rri

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	qryLoginLogout  = "version: 3.0\naction: LOGIN\nuser: DENIC-1000042-TEST\npassword: very-secure\n=-= now log out\nversion: 3.0\naction: LOGOUT\n"
	qryIgnoreCasing = "Version: 3.0\nAction: login\nUser: DENIC-1000042-TEST\nPassword: very-secure"
	qryWhitespaces  = "  version: \t3.0  \n\n\naction:    LOGIN\n   user: DENIC-1000042-TEST\npassword: very-secure    \n"
)

func TestParseQueries(t *testing.T) {
	queries, err := ParseQueries(qryLoginLogout)
	assert.NoError(t, err)
	if assert.Len(t, queries, 2) {
		assert.Equal(t, LatestVersion, queries[0].Version())
		assert.Equal(t, ActionLogin, queries[0].Action())
		assert.Len(t, queries[0].Fields(), 2)
		assert.Equal(t, queries[0].Field(FieldNameUser), []string{"DENIC-1000042-TEST"})
		assert.Equal(t, queries[0].Field(FieldNamePassword), []string{"very-secure"})

		assert.Equal(t, LatestVersion, queries[1].Version())
		assert.Equal(t, ActionLogout, queries[1].Action())
		assert.Len(t, queries[1].Fields(), 0)
	}
}

func TestParseQueriesCasing(t *testing.T) {
	query, err := ParseQuery(qryIgnoreCasing)
	if assert.NoError(t, err) {
		assert.Equal(t, LatestVersion, query.Version())
		assert.Equal(t, ActionLogin, query.Action())
		assert.Equal(t, query.Field("uSeR"), []string{"DENIC-1000042-TEST"})
		assert.Equal(t, query.Field(FieldNamePassword), []string{"very-secure"})
	}
}

func TestParseQueriesWhitespaces(t *testing.T) {
	query, err := ParseQuery(qryWhitespaces)
	if assert.NoError(t, err) {
		assert.Equal(t, LatestVersion, query.Version())
		assert.Equal(t, ActionLogin, query.Action())
		assert.Equal(t, query.Field(FieldNameUser), []string{"DENIC-1000042-TEST"})
		assert.Equal(t, query.Field(FieldNamePassword), []string{"very-secure"})
	}
}
