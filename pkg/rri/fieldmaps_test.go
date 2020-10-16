package rri

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQueryFieldList(t *testing.T) {
	l := newQueryFieldList()
	require.Equal(t, 0, l.Size())

	assert.Equal(t, []string{}, l.Values(QueryFieldNameDomainIDN))
	assert.Equal(t, "", l.FirstValue(QueryFieldNameDomainIDN))

	l.Add(QueryFieldNameDomainIDN, "denic.de")
	require.Equal(t, 1, l.Size())
	assert.Equal(t, []string{"denic.de"}, l.Values(QueryFieldNameDomainIDN))
	assert.Equal(t, "denic.de", l.FirstValue(QueryFieldNameDomainIDN))

	l.Add(QueryFieldNameNameServer, "ns1.denic.de", "ns2.denic.de")
	require.Equal(t, 3, l.Size())
	assert.Equal(t, []string{"ns1.denic.de", "ns2.denic.de"}, l.Values(QueryFieldNameNameServer))
	assert.Equal(t, "ns1.denic.de", l.FirstValue(QueryFieldNameNameServer))

	l.Add("dOmAiN", "some-other", "stuff")
	require.Equal(t, 5, l.Size())
	assert.Equal(t, []string{"denic.de", "some-other", "stuff"}, l.Values("DoMaIn"))
	assert.Equal(t, "denic.de", l.FirstValue("DoMaIn"))

	l.RemoveAll(QueryFieldNameNameServer)
	require.Equal(t, 3, l.Size())
	assert.Equal(t, []string{}, l.Values(QueryFieldNameNameServer))
	assert.Equal(t, "", l.FirstValue(QueryFieldNameNameServer))
	assert.Equal(t, []string{"denic.de", "some-other", "stuff"}, l.Values(QueryFieldNameDomainIDN))
	assert.Equal(t, "denic.de", l.FirstValue(QueryFieldNameDomainIDN))
}

func TestResponseFieldList(t *testing.T) {
	l := newResponseFieldList()
	require.Equal(t, 0, l.Size())

	assert.Equal(t, []string{}, l.Values(ResponseFieldNameResult))
	assert.Equal(t, "", l.FirstValue(ResponseFieldNameResult))

	l.Add(ResponseFieldNameError, "foobar")
	require.Equal(t, 1, l.Size())
	assert.Equal(t, []string{"foobar"}, l.Values(ResponseFieldNameError))
	assert.Equal(t, "foobar", l.FirstValue(ResponseFieldNameError))

	l.Add(ResponseFieldNameInfo, "foo", "bar")
	require.Equal(t, 3, l.Size())
	assert.Equal(t, []string{"foo", "bar"}, l.Values(ResponseFieldNameInfo))
	assert.Equal(t, "foo", l.FirstValue(ResponseFieldNameInfo))

	l.Add("eRrOr", "some-other", "stuff")
	require.Equal(t, 5, l.Size())
	assert.Equal(t, []string{"foobar", "some-other", "stuff"}, l.Values("ErRoR"))
	assert.Equal(t, "foobar", l.FirstValue("ErRoR"))

	l.RemoveAll(ResponseFieldNameInfo)
	require.Equal(t, 3, l.Size())
	assert.Equal(t, []string{}, l.Values(ResponseFieldNameInfo))
	assert.Equal(t, "", l.FirstValue(ResponseFieldNameInfo))
	assert.Equal(t, []string{"foobar", "some-other", "stuff"}, l.Values(ResponseFieldNameError))
	assert.Equal(t, "foobar", l.FirstValue(ResponseFieldNameError))
}
