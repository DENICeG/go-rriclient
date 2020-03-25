package rri

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	qryLoginLogout = "version: 3.0\naction: LOGIN\nuser: DENIC-1000042-TEST\npassword: very-secure\n=-= now log out\nversion: 3.0\naction: LOGOUT\n"
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
