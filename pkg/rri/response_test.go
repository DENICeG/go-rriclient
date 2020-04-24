package rri

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	respEntity = "RESULT: success\nSTID: 10459b07-861a-11ea-b33a-d9ddb946cb7c\n\nDomain: de-registrylock.de\nDomain-Ace: de-registrylock.de\nNserver: ns1.denic.de.\nNserver: ns2.denic.de.\nNserver: ns3.denic.de.\nStatus: connect\nRegistryLock: true\nRegAccId: DENIC-1000006\nRegAccName: DENIC eG\nChanged: 2020-04-23T09:58:11+02:00\n\n[Holder]\nHandle: DENIC-1000006-DENIC\nType: ORG\nName: DENIC eG\nAddress: Kaiserstrasse 75-77\nCity: Frankfurt am Main\nPostalCode: 60329\nCountryCode: DE\nEmail: info@denic.de\nChanged: 2019-04-05T10:26:06+02:00\n"
)

func TestResponseEntity(t *testing.T) {
	response, err := ParseResponseKV(respEntity)
	if assert.NoError(t, err) {
		//TODO assert fields
		assert.Equal(t, ResultSuccess, response.Result())
		if assert.Equal(t, []ResponseEntityName{EntityNameHolder}, response.EntityNames()) {
			entity := response.Entity(EntityNameHolder)
			if assert.NotNil(t, entity) {
				//TODO assert fields
			}
		}
	}
}
