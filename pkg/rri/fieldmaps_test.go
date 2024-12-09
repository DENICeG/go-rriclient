package rri_test

import (
	"testing"

	"github.com/DENICeG/go-rriclient/pkg/rri"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQueryFieldList(t *testing.T) {
	l := rri.NewQueryFieldList()
	require.Equal(t, 0, l.Size())

	assert.Equal(t, []string{}, l.Values(rri.QueryFieldNameDomainIDN))
	assert.Equal(t, "", l.FirstValue(rri.QueryFieldNameDomainIDN))

	l.Add(rri.QueryFieldNameDomainIDN, "denic.de")
	require.Equal(t, 1, l.Size())
	assert.Equal(t, []string{"denic.de"}, l.Values(rri.QueryFieldNameDomainIDN))
	assert.Equal(t, "denic.de", l.FirstValue(rri.QueryFieldNameDomainIDN))

	require.NotPanics(t, func() {
		l.Add(rri.QueryFieldNameDomainIDN, nil...)
	})
	require.Equal(t, 1, l.Size())
	assert.Equal(t, []string{"denic.de"}, l.Values(rri.QueryFieldNameDomainIDN))
	assert.Equal(t, "denic.de", l.FirstValue(rri.QueryFieldNameDomainIDN))

	l.Add(rri.QueryFieldNameNameServer, "ns1.denic.de", "ns2.denic.de")
	require.Equal(t, 3, l.Size())
	assert.Equal(t, []string{"ns1.denic.de", "ns2.denic.de"}, l.Values(rri.QueryFieldNameNameServer))
	assert.Equal(t, "ns1.denic.de", l.FirstValue(rri.QueryFieldNameNameServer))

	l.Add("dOmAiN", "some-other", "stuff")
	require.Equal(t, 5, l.Size())
	assert.Equal(t, []string{"denic.de", "some-other", "stuff"}, l.Values("DoMaIn"))
	assert.Equal(t, "denic.de", l.FirstValue("DoMaIn"))

	l.RemoveAll(rri.QueryFieldNameNameServer)
	require.Equal(t, 3, l.Size())
	assert.Equal(t, []string{}, l.Values(rri.QueryFieldNameNameServer))
	assert.Equal(t, "", l.FirstValue(rri.QueryFieldNameNameServer))
	assert.Equal(t, []string{"denic.de", "some-other", "stuff"}, l.Values(rri.QueryFieldNameDomainIDN))
	assert.Equal(t, "denic.de", l.FirstValue(rri.QueryFieldNameDomainIDN))
}

func TestQueryFieldsCopyTo(t *testing.T) {
	src := rri.NewQueryFieldList()
	src.Add(rri.QueryFieldNameAction, string(rri.ActionLogin))
	src.Add(rri.QueryFieldNameUser, "test")
	dst := rri.NewQueryFieldList()
	src.CopyTo(&dst)
	require.Equal(t, src.Size(), dst.Size())
	assert.Equal(t, src.Values(rri.QueryFieldNameAction), dst.Values(rri.QueryFieldNameAction))
	assert.Equal(t, src.Values(rri.QueryFieldNameUser), dst.Values(rri.QueryFieldNameUser))
}

func TestResponseFieldList(t *testing.T) {
	l := rri.NewResponseFieldList()
	require.Equal(t, 0, l.Size())

	assert.Equal(t, []string{}, l.Values(rri.ResponseFieldNameResult))
	assert.Equal(t, "", l.FirstValue(rri.ResponseFieldNameResult))

	l.Add(rri.ResponseFieldNameError, "foobar")
	require.Equal(t, 1, l.Size())
	assert.Equal(t, []string{"foobar"}, l.Values(rri.ResponseFieldNameError))
	assert.Equal(t, "foobar", l.FirstValue(rri.ResponseFieldNameError))

	require.NotPanics(t, func() {
		l.Add(rri.ResponseFieldNameSTID, nil...)
	})
	require.Equal(t, 1, l.Size())
	assert.Equal(t, []string{"foobar"}, l.Values(rri.ResponseFieldNameError))
	assert.Equal(t, "foobar", l.FirstValue(rri.ResponseFieldNameError))

	l.Add(rri.ResponseFieldNameInfo, "foo", "bar")
	require.Equal(t, 3, l.Size())
	assert.Equal(t, []string{"foo", "bar"}, l.Values(rri.ResponseFieldNameInfo))
	assert.Equal(t, "foo", l.FirstValue(rri.ResponseFieldNameInfo))

	l.Add("eRrOr", "some-other", "stuff")
	require.Equal(t, 5, l.Size())
	assert.Equal(t, []string{"foobar", "some-other", "stuff"}, l.Values("ErRoR"))
	assert.Equal(t, "foobar", l.FirstValue("ErRoR"))

	l.RemoveAll(rri.ResponseFieldNameInfo)
	require.Equal(t, 3, l.Size())
	assert.Equal(t, []string{}, l.Values(rri.ResponseFieldNameInfo))
	assert.Equal(t, "", l.FirstValue(rri.ResponseFieldNameInfo))
	assert.Equal(t, []string{"foobar", "some-other", "stuff"}, l.Values(rri.ResponseFieldNameError))
	assert.Equal(t, "foobar", l.FirstValue(rri.ResponseFieldNameError))
}

func TestResponseFieldsCopyTo(t *testing.T) {
	src := rri.NewResponseFieldList()
	src.Add(rri.ResponseFieldNameResult, string(rri.ResultFailure))
	src.Add(rri.ResponseFieldNameError, "12345 foobar")
	dst := rri.NewResponseFieldList()
	src.CopyTo(&dst)
	require.Equal(t, src.Size(), dst.Size())
	assert.Equal(t, src.Values(rri.ResponseFieldNameResult), dst.Values(rri.ResponseFieldNameResult))
	assert.Equal(t, src.Values(rri.ResponseFieldNameError), dst.Values(rri.ResponseFieldNameError))
}
