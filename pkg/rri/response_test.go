package rri

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	respEncode   = "RESULT: success\nINFO: 13000000011 foo\nSTID: 554c2cd7-0885-11eb-a619-610f86f60bcb\nINFO: 13000000011 bar"
	respInfoMsg  = "RESULT: success\nINFO: 13000000011 Request was processed in test environment - not valid in real world [testing platform]\nSTID: 554c2cd7-0885-11eb-a619-610f86f60bcb"
	respErrorMsg = "RESULT: failed\nSTID: d97b7af9-0886-11eb-a619-610f86f60bcb\nERROR: 63300062009 Domain doesn't exist [foobartestgibtsnet.de]\nINFO: 13000000011 Request was processed in test environment - not valid in real world [testing platform]"
	respEntity   = "RESULT: success\nSTID: 10459b07-861a-11ea-b33a-d9ddb946cb7c\n\nDomain: de-registrylock.de\nDomain-Ace: de-registrylock.de\nNserver: ns1.denic.de.\nNserver: ns2.denic.de.\nNserver: ns3.denic.de.\nStatus: connect\nRegistryLock: true\nRegAccId: DENIC-1000006\nRegAccName: DENIC eG\nChanged: 2020-04-23T09:58:11+02:00\n\n[Holder]\nHandle: DENIC-1000006-DENIC\nType: ORG\nName: DENIC eG\nAddress: Kaiserstrasse 75-77\nCity: Frankfurt am Main\nPostalCode: 60329\nCountryCode: DE\nEmail: info@denic.de\nChanged: 2019-04-05T10:26:06+02:00\n"
)

func TestResponseEncodeKV(t *testing.T) {
	response, err := ParseResponse(respEncode)
	require.NoError(t, err)
	assert.Equal(t, "RESULT: success\nINFO: 13000000011 foo\nSTID: 554c2cd7-0885-11eb-a619-610f86f60bcb\nINFO: 13000000011 bar", response.EncodeKV())
}

func TestResponseInfoMessages(t *testing.T) {
	response, err := ParseResponse(respInfoMsg)
	require.NoError(t, err)
	require.Len(t, response.InfoMessages(), 1)
	assert.Equal(t, []BusinessMessage{NewBusinessMessage(13000000011, "Request was processed in test environment - not valid in real world [testing platform]")}, response.InfoMessages())
	require.Len(t, response.ErrorMessages(), 0)
	assert.Equal(t, []BusinessMessage{}, response.ErrorMessages())
}

func TestResponseErrorMessages(t *testing.T) {
	response, err := ParseResponse(respErrorMsg)
	require.NoError(t, err)
	require.Len(t, response.InfoMessages(), 1)
	assert.Equal(t, []BusinessMessage{NewBusinessMessage(13000000011, "Request was processed in test environment - not valid in real world [testing platform]")}, response.InfoMessages())
	require.Len(t, response.ErrorMessages(), 1)
	assert.Equal(t, []BusinessMessage{NewBusinessMessage(63300062009, "Domain doesn't exist [foobartestgibtsnet.de]")}, response.ErrorMessages())
}

func TestResponseEntity(t *testing.T) {
	response, err := ParseResponse(respEntity)
	require.NoError(t, err)
	assert.Equal(t, ResultSuccess, response.Result())
	assert.Equal(t, "10459b07-861a-11ea-b33a-d9ddb946cb7c", response.STID())
	require.Len(t, response.Fields(), 12)
	assert.Equal(t, []string{string(ResultSuccess)}, response.Field(ResponseFieldNameResult))
	assert.Equal(t, []string{"10459b07-861a-11ea-b33a-d9ddb946cb7c"}, response.Field(ResponseFieldNameSTID))
	assert.Equal(t, []string{"de-registrylock.de"}, response.Field("Domain"))
	assert.Equal(t, []string{"de-registrylock.de"}, response.Field("Domain-Ace"))
	assert.Equal(t, []string{"ns1.denic.de.", "ns2.denic.de.", "ns3.denic.de."}, response.Field("Nserver"))
	assert.Equal(t, []string{"connect"}, response.Field("Status"))
	assert.Equal(t, []string{"true"}, response.Field("RegistryLock"))
	assert.Equal(t, []string{"DENIC-1000006"}, response.Field("RegAccId"))
	assert.Equal(t, []string{"DENIC eG"}, response.Field("RegAccName"))
	assert.Equal(t, []string{"2020-04-23T09:58:11+02:00"}, response.Field("Changed"))
	require.Equal(t, []ResponseEntityName{ResponseEntityNameHolder}, response.EntityNames())
	entity := response.Entity(ResponseEntityNameHolder)
	require.NotNil(t, entity)
	require.Len(t, entity, 9)
	assert.Equal(t, []string{"DENIC-1000006-DENIC"}, entity.Values("Handle"))
	assert.Equal(t, []string{"ORG"}, entity.Values("Type"))
	assert.Equal(t, []string{"DENIC eG"}, entity.Values("Name"))
	assert.Equal(t, []string{"Kaiserstrasse 75-77"}, entity.Values("Address"))
	assert.Equal(t, []string{"Frankfurt am Main"}, entity.Values("City"))
	assert.Equal(t, []string{"60329"}, entity.Values("PostalCode"))
	assert.Equal(t, []string{"DE"}, entity.Values("CountryCode"))
	assert.Equal(t, []string{"info@denic.de"}, entity.Values("Email"))
	assert.Equal(t, []string{"2019-04-05T10:26:06+02:00"}, entity.Values("Changed"))
}

//TODO test response and entity order

func TestParseBusinessMessageKV(t *testing.T) {
	var bm BusinessMessage
	var err error

	bm, err = ParseBusinessMessageKV("13000000011 Request was processed in test environment - not valid in real world [testing platform]")
	assert.NoError(t, err)
	assert.Equal(t, BusinessMessage{13000000011, "Request was processed in test environment - not valid in real world [testing platform]"}, bm)

	bm, err = ParseBusinessMessageKV("13000000011")
	assert.Error(t, err)

	bm, err = ParseBusinessMessageKV("Request was processed in test environment - not valid in real world [testing platform]")
	assert.Error(t, err)

	bm, err = ParseBusinessMessageKV("")
	assert.Error(t, err)
}
