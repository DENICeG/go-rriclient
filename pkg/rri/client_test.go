package rri

import (
	"crypto/tls"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient(t *testing.T) {
	mustWithMockServer(func(server *MockServer) {
		server.AddUser("DENIC-1000011-TEST", "secret")

		var client *Client

		t.Run("NewClient", func(t *testing.T) {
			var err error
			client, err = NewClient(server.Address(), &ClientConfig{Insecure: true})
			require.NoError(t, err)
			assert.Equal(t, server.Address(), client.RemoteAddress())
			assert.False(t, client.IsLoggedIn())
		})

		require.NotNil(t, client)

		t.Run("QueryBeforeLogin", func(t *testing.T) {
			_, err := client.SendQuery(NewInfoDomainQuery("denic.de"))
			assert.Error(t, err)
		})

		t.Run("Login", func(t *testing.T) {
			assert.Error(t, client.Login("asdf", "foobar"))
			assert.False(t, client.IsLoggedIn())
			assert.NoError(t, client.Login("DENIC-1000011-TEST", "secret"))
			assert.True(t, client.IsLoggedIn())
			assert.Equal(t, "DENIC-1000011-TEST", client.CurrentUser())
			regAccID, err := client.CurrentRegAccID()
			require.NoError(t, err)
			assert.Equal(t, 1000011, regAccID)
		})
	})
}

func TestClientConfDefaults(t *testing.T) {
	dialCount := 0
	client, err := NewClient("localhost", &ClientConfig{
		TLSDialHandler: func(network, addr string, config *tls.Config) (TLSConnection, error) {
			assert.Equal(t, "tcp", network)
			assert.Equal(t, "localhost:51131", addr)
			assert.Equal(t, uint16(tls.VersionTLS13), config.MinVersion)
			assert.False(t, config.InsecureSkipVerify)
			dialCount++
			return nil, nil
		},
	})
	require.NoError(t, err)
	defer client.Close()

	assert.Equal(t, 1, dialCount)
}

func TestClientConf(t *testing.T) {
	dialCount := 0
	client, err := NewClient("localhost:12345", &ClientConfig{
		Insecure:      true,
		MinTLSVersion: tls.VersionTLS12,
		TLSDialHandler: func(network, addr string, config *tls.Config) (TLSConnection, error) {
			assert.Equal(t, "tcp", network)
			assert.Equal(t, "localhost:12345", addr)
			assert.Equal(t, uint16(tls.VersionTLS12), config.MinVersion)
			assert.True(t, config.InsecureSkipVerify)
			dialCount++
			return nil, nil
		},
	})
	require.NoError(t, err)
	defer client.Close()

	assert.Equal(t, 1, dialCount)
}

func TestClientNoAutoRetry(t *testing.T) {
	dialCount := 0
	conn := newMockReadWriteCloser(t, []readResponse{
		{[]byte{0, 0, 0, 15}, nil},
		{[]byte("RESULT: success"), nil},
	}, []writeResponse{
		{b64("AAAAQ3ZlcnNpb246IDQuMAphY3Rpb246IExPR0lOCnVzZXI6IERFTklDLTEwMDAwMTEtUlJJCnBhc3N3b3JkOiBzZWNyZXQ="), nil},
		{b64("AAAAP3ZlcnNpb246IDQuMAphY3Rpb246IElORk8KZG9tYWluOiBkZW5pYy5kZQpkb21haW4tYWNlOiBkZW5pYy5kZQ=="), fmt.Errorf("broken pipe")},
	})

	client, err := NewClient("localhost", &ClientConfig{
		TLSDialHandler: func(network, addr string, config *tls.Config) (TLSConnection, error) {
			dialCount++
			return conn, nil
		},
	})
	require.NoError(t, err)
	defer client.Close()

	require.NoError(t, client.Login("DENIC-1000011-RRI", "secret"))
	client.NoAutoRetry = true
	_, err = client.SendQuery(NewInfoDomainQuery("denic.de"))
	require.Error(t, err)

	assert.Equal(t, 1, dialCount, "unexpected number of tls connections")
	conn.AssertComplete()
}

func TestClientAutoRetry(t *testing.T) {
	dialCount := 0
	conn := newMockReadWriteCloser(t, []readResponse{
		{[]byte{0, 0, 0, 15}, nil},
		{[]byte("RESULT: success"), nil},
		{[]byte{0, 0, 0, 15}, nil},
		{[]byte("RESULT: success"), nil},
		{[]byte{0, 0, 0, 39}, nil},
		{[]byte("RESULT: success\nINFO: 12345 only a test"), nil},
	}, []writeResponse{
		{b64("AAAAQ3ZlcnNpb246IDQuMAphY3Rpb246IExPR0lOCnVzZXI6IERFTklDLTEwMDAwMTEtUlJJCnBhc3N3b3JkOiBzZWNyZXQ="), nil},
		{b64("AAAAP3ZlcnNpb246IDQuMAphY3Rpb246IElORk8KZG9tYWluOiBkZW5pYy5kZQpkb21haW4tYWNlOiBkZW5pYy5kZQ=="), fmt.Errorf("broken pipe")},
		{b64("AAAAQ3ZlcnNpb246IDQuMAphY3Rpb246IExPR0lOCnVzZXI6IERFTklDLTEwMDAwMTEtUlJJCnBhc3N3b3JkOiBzZWNyZXQ="), nil},
		{b64("AAAAP3ZlcnNpb246IDQuMAphY3Rpb246IElORk8KZG9tYWluOiBkZW5pYy5kZQpkb21haW4tYWNlOiBkZW5pYy5kZQ=="), nil},
	})

	client, err := NewClient("localhost", &ClientConfig{
		TLSDialHandler: func(network, addr string, config *tls.Config) (TLSConnection, error) {
			dialCount++
			return conn, nil
		},
	})
	require.NoError(t, err)
	defer client.Close()

	require.NoError(t, client.Login("DENIC-1000011-RRI", "secret"))
	resp, err := client.SendQuery(NewInfoDomainQuery("denic.de"))
	require.NoError(t, err)
	require.Equal(t, 2, resp.Fields().Size())
	assert.Equal(t, []string{"success"}, resp.Field(ResponseFieldNameResult))
	assert.Equal(t, []string{"12345 only a test"}, resp.Field(ResponseFieldNameInfo))

	assert.Equal(t, 2, dialCount, "unexpected number of tls connections")
	conn.AssertComplete()
}

type mockReadWriteCloser struct {
	ReadResponses  []readResponse
	ReadIndex      int
	WriteResponses []writeResponse
	WriteIndex     int
	m              sync.Mutex
	t              *testing.T
}

func newMockReadWriteCloser(t *testing.T, readResponses []readResponse, writeResponses []writeResponse) *mockReadWriteCloser {
	return &mockReadWriteCloser{ReadResponses: readResponses, WriteResponses: writeResponses, t: t}
}

type readResponse struct {
	Data  []byte
	Error error
}

type writeResponse struct {
	ExpectedData []byte
	Error        error
}

func (m *mockReadWriteCloser) AssertComplete() bool {
	eq1 := assert.Equal(m.t, len(m.ReadResponses), m.ReadIndex)
	eq2 := assert.Equal(m.t, len(m.WriteResponses), m.WriteIndex)
	return eq1 && eq2
}

func (m *mockReadWriteCloser) Read(p []byte) (int, error) {
	m.m.Lock()
	defer m.m.Unlock()

	if m.ReadIndex >= len(m.ReadResponses) {
		require.Fail(m.t, fmt.Sprintf("unexpected read operation %d of %d", m.ReadIndex+1, len(m.ReadResponses)))
	}
	if len(m.ReadResponses[m.ReadIndex].Data) != len(p) {
		require.Fail(m.t, fmt.Sprintf("read response destination array of size %d does not match expected output of size %d", len(p), len(m.ReadResponses[m.ReadIndex].Data)))
	}
	m.ReadIndex++
	if m.ReadResponses[m.ReadIndex-1].Error != nil {
		return 0, m.ReadResponses[m.ReadIndex-1].Error
	}
	copy(p, m.ReadResponses[m.ReadIndex-1].Data)
	return len(m.ReadResponses[m.ReadIndex-1].Data), nil
}

func (m *mockReadWriteCloser) Write(p []byte) (int, error) {
	m.m.Lock()
	defer m.m.Unlock()

	if m.WriteIndex >= len(m.WriteResponses) {
		require.Fail(m.t, fmt.Sprintf("unexpected write operation %d of %d: %s", m.WriteIndex+1, len(m.WriteResponses), b64enc(p)))
	}
	require.Equal(m.t, b64enc(m.WriteResponses[m.WriteIndex].ExpectedData), b64enc(p))
	m.WriteIndex++
	if m.WriteResponses[m.WriteIndex-1].Error != nil {
		return 0, m.WriteResponses[m.WriteIndex-1].Error
	}
	return len(p), nil
}

func (m *mockReadWriteCloser) Close() error {
	return nil
}
