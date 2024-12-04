package rri_test

import (
	"testing"

	"github.com/DENICeG/go-rriclient/pkg/rri"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResponseEncodeKV(t *testing.T) {
	response, err := rri.ParseResponse("RESULT: success\nINFO: 13000000011 foo\nSTID: 554c2cd7-0885-11eb-a619-610f86f60bcb\nINFO: 13000000011 bar")
	require.NoError(t, err)
	assert.Equal(t, "RESULT: success\nINFO: 13000000011 foo\nSTID: 554c2cd7-0885-11eb-a619-610f86f60bcb\nINFO: 13000000011 bar", response.EncodeKV())
}

func TestResponseInfoMessages(t *testing.T) {
	response, err := rri.ParseResponse("RESULT: success\nINFO: 13000000011 Request was processed in test environment - not valid in real world [testing platform]\nSTID: 554c2cd7-0885-11eb-a619-610f86f60bcb")
	require.NoError(t, err)
	require.Len(t, response.InfoMessages(), 1)
	assert.Equal(t, []rri.BusinessMessage{rri.NewBusinessMessage(13000000011, "Request was processed in test environment - not valid in real world [testing platform]")}, response.InfoMessages())
	require.Len(t, response.ErrorMessages(), 0)
	assert.Equal(t, []rri.BusinessMessage{}, response.ErrorMessages())
}

func TestResponseErrorMessages(t *testing.T) {
	response, err := rri.ParseResponse("RESULT: failed\nSTID: d97b7af9-0886-11eb-a619-610f86f60bcb\nERROR: 63300062009 Domain doesn't exist [foobartestgibtsnet.de]\nINFO: 13000000011 Request was processed in test environment - not valid in real world [testing platform]")
	require.NoError(t, err)
	require.Len(t, response.InfoMessages(), 1)
	assert.Equal(t, []rri.BusinessMessage{rri.NewBusinessMessage(13000000011, "Request was processed in test environment - not valid in real world [testing platform]")}, response.InfoMessages())
	require.Len(t, response.ErrorMessages(), 1)
	assert.Equal(t, []rri.BusinessMessage{rri.NewBusinessMessage(63300062009, "Domain doesn't exist [foobartestgibtsnet.de]")}, response.ErrorMessages())
}

func TestResponseEntity(t *testing.T) {
	response, err := rri.ParseResponse("RESULT: success\nSTID: 10459b07-861a-11ea-b33a-d9ddb946cb7c\n\nDomain: de-registrylock.de\nDomain-Ace: de-registrylock.de\nNserver: ns1.denic.de.\nNserver: ns2.denic.de.\nNserver: ns3.denic.de.\nStatus: connect\nRegistryLock: true\nRegAccId: DENIC-1000006\nRegAccName: DENIC eG\nChanged: 2020-04-23T09:58:11+02:00\n\n[Holder]\nHandle: DENIC-1000006-DENIC\nType: ORG\nName: DENIC eG\nAddress: Kaiserstrasse 75-77\nCity: Frankfurt am Main\nPostalCode: 60329\nCountryCode: DE\nEmail: info@denic.de\nChanged: 2019-04-05T10:26:06+02:00\n")
	require.NoError(t, err)
	assert.Equal(t, rri.ResultSuccess, response.Result())
	assert.Equal(t, "10459b07-861a-11ea-b33a-d9ddb946cb7c", response.STID())
	require.Len(t, response.Fields(), 12)
	assert.Equal(t, []string{string(rri.ResultSuccess)}, response.Field(rri.ResponseFieldNameResult))
	assert.Equal(t, []string{"10459b07-861a-11ea-b33a-d9ddb946cb7c"}, response.Field(rri.ResponseFieldNameSTID))
	assert.Equal(t, []string{"de-registrylock.de"}, response.Field("Domain"))
	assert.Equal(t, []string{"de-registrylock.de"}, response.Field("Domain-Ace"))
	assert.Equal(t, []string{"ns1.denic.de.", "ns2.denic.de.", "ns3.denic.de."}, response.Field("Nserver"))
	assert.Equal(t, []string{"connect"}, response.Field("Status"))
	assert.Equal(t, []string{"true"}, response.Field("RegistryLock"))
	assert.Equal(t, []string{"DENIC-1000006"}, response.Field("RegAccId"))
	assert.Equal(t, []string{"DENIC eG"}, response.Field("RegAccName"))
	assert.Equal(t, []string{"2020-04-23T09:58:11+02:00"}, response.Field("Changed"))
	entities := response.Entities()
	require.Len(t, entities, 1)
	require.Len(t, entities[0].Fields(), 9)
	assert.Equal(t, rri.ResponseEntityNameHolder, entities[0].Name())
	assert.Equal(t, []string{"DENIC-1000006-DENIC"}, entities[0].Field("Handle"))
	assert.Equal(t, []string{"ORG"}, entities[0].Field("Type"))
	assert.Equal(t, []string{"DENIC eG"}, entities[0].Field("Name"))
	assert.Equal(t, []string{"Kaiserstrasse 75-77"}, entities[0].Field("Address"))
	assert.Equal(t, []string{"Frankfurt am Main"}, entities[0].Field("City"))
	assert.Equal(t, []string{"60329"}, entities[0].Field("PostalCode"))
	assert.Equal(t, []string{"DE"}, entities[0].Field("CountryCode"))
	assert.Equal(t, []string{"info@denic.de"}, entities[0].Field("Email"))
	assert.Equal(t, []string{"2019-04-05T10:26:06+02:00"}, entities[0].Field("Changed"))
}

func TestResponseEntityMultiHolder(t *testing.T) {
	response, err := rri.ParseResponse("RESULT: success\nINFO: 13000000011 Request was processed in test environment - not valid in real world [testing platform]\nSTID: 8792891a-c366-11eb-bca6-bbfdc472082a\n\nDomain: denic-opstt-29791.de\nDomain-Ace: denic-opstt-29791.de\nNserver: dns1.opsblau.de.\nNserver: dns3.opsblau.de.\nStatus: connect\nAuthInfo2: 2030-01-01T00:00:00+01:00\nRegAccId: DENIC-1000021\nRegAccName: DENIC eG - Operations Test 1000021\nChanged: 2011-08-22T14:39:34+02:00\n\n[Holder]\nHandle: DENIC-1000021-TEST-DAGOBERT\nType: PERSON\nName: Dagobert Duck\nOrganisation: Duck Industries\nAddress: Im Geldspeicher\nCity: Entenhausen\nPostalCode: 64542\nCountryCode: DE\nEmail: dagobert.duck@duck-industries.de\nChanged: 2020-12-23T07:13:04+01:00\n\n[Holder]\nHandle: DENIC-1000021-TEST-DAISY\nType: PERSON\nName: Daisy Duck\nOrganisation: Frauenverein\nAddress: Fliederweg 8\nCity: Entenhausen\nPostalCode: 64548\nCountryCode: DE\nEmail: daisy.duck@enten-netz.de\nChanged: 2020-12-23T07:09:19+01:00\n")
	require.NoError(t, err)
	assert.Equal(t, rri.ResultSuccess, response.Result())
	assert.Equal(t, "8792891a-c366-11eb-bca6-bbfdc472082a", response.STID())
	require.Len(t, response.Fields(), 12)
	assert.Equal(t, []string{string(rri.ResultSuccess)}, response.Field(rri.ResponseFieldNameResult))
	assert.Equal(t, []string{"8792891a-c366-11eb-bca6-bbfdc472082a"}, response.Field(rri.ResponseFieldNameSTID))
	assert.Equal(t, []string{"denic-opstt-29791.de"}, response.Field("Domain"))
	assert.Equal(t, []string{"denic-opstt-29791.de"}, response.Field("Domain-Ace"))
	assert.Equal(t, []string{"dns1.opsblau.de.", "dns3.opsblau.de."}, response.Field("Nserver"))
	assert.Equal(t, []string{"connect"}, response.Field("Status"))
	assert.Equal(t, []string{"2030-01-01T00:00:00+01:00"}, response.Field("AuthInfo2"))
	assert.Equal(t, []string{"DENIC-1000021"}, response.Field("RegAccId"))
	assert.Equal(t, []string{"DENIC eG - Operations Test 1000021"}, response.Field("RegAccName"))
	assert.Equal(t, []string{"2011-08-22T14:39:34+02:00"}, response.Field("Changed"))
	entities := response.Entities()
	require.Len(t, entities, 2)
	require.Len(t, entities[0].Fields(), 10)
	assert.Equal(t, rri.ResponseEntityNameHolder, entities[0].Name())
	assert.Equal(t, []string{"DENIC-1000021-TEST-DAGOBERT"}, entities[0].Field("Handle"))
	assert.Equal(t, []string{"PERSON"}, entities[0].Field("Type"))
	assert.Equal(t, []string{"Dagobert Duck"}, entities[0].Field("Name"))
	assert.Equal(t, []string{"Duck Industries"}, entities[0].Field("Organisation"))
	assert.Equal(t, []string{"Im Geldspeicher"}, entities[0].Field("Address"))
	assert.Equal(t, []string{"Entenhausen"}, entities[0].Field("City"))
	assert.Equal(t, []string{"64542"}, entities[0].Field("PostalCode"))
	assert.Equal(t, []string{"DE"}, entities[0].Field("CountryCode"))
	assert.Equal(t, []string{"dagobert.duck@duck-industries.de"}, entities[0].Field("Email"))
	assert.Equal(t, []string{"2020-12-23T07:13:04+01:00"}, entities[0].Field("Changed"))
	require.Len(t, entities[1].Fields(), 10)
	assert.Equal(t, rri.ResponseEntityNameHolder, entities[1].Name())
	assert.Equal(t, []string{"DENIC-1000021-TEST-DAISY"}, entities[1].Field("Handle"))
	assert.Equal(t, []string{"PERSON"}, entities[1].Field("Type"))
	assert.Equal(t, []string{"Daisy Duck"}, entities[1].Field("Name"))
	assert.Equal(t, []string{"Frauenverein"}, entities[1].Field("Organisation"))
	assert.Equal(t, []string{"Fliederweg 8"}, entities[1].Field("Address"))
	assert.Equal(t, []string{"Entenhausen"}, entities[1].Field("City"))
	assert.Equal(t, []string{"64548"}, entities[1].Field("PostalCode"))
	assert.Equal(t, []string{"DE"}, entities[1].Field("CountryCode"))
	assert.Equal(t, []string{"daisy.duck@enten-netz.de"}, entities[1].Field("Email"))
	assert.Equal(t, []string{"2020-12-23T07:09:19+01:00"}, entities[1].Field("Changed"))
}

func TestParseBusinessMessageKV(t *testing.T) {
	var bm rri.BusinessMessage
	var err error

	bm, err = rri.ParseBusinessMessageKV("13000000011 Request was processed in test environment - not valid in real world [testing platform]")
	assert.NoError(t, err)
	assert.Equal(t, rri.NewBusinessMessage(13000000011, "Request was processed in test environment - not valid in real world [testing platform]"), bm)

	_, err = rri.ParseBusinessMessageKV("13000000011")
	assert.Error(t, err)

	_, err = rri.ParseBusinessMessageKV("Request was processed in test environment - not valid in real world [testing platform]")
	assert.Error(t, err)

	_, err = rri.ParseBusinessMessageKV("")
	assert.Error(t, err)
}
