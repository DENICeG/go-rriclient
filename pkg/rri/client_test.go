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
		server.AddUser("test", "secret")

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
			assert.NoError(t, client.Login("test", "secret"))
			assert.True(t, client.IsLoggedIn())
			assert.Equal(t, "test", client.CurrentUser())
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
	defer client.Close()

	require.NoError(t, err)
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
	defer client.Close()

	require.NoError(t, err)
	assert.Equal(t, 1, dialCount)
}

func TestClientAutoRetry(t *testing.T) {
	//TODO this test requires deterministic response creation
	/*dialCount := 0
	conn := newMockReadWriteCloser(t, []readResponse{
		{[]byte{0, 0, 0, 4}, nil},
		{[]byte("RESULT: success"), nil},
	}, []writeResponse{
		{b64("AAAAQ3ZlcnNpb246IDMuMAphY3Rpb246IExPR0lOCnVzZXI6IERFTklDLTEwMDAwMTEtUlJJCnBhc3N3b3JkOiBzZWNyZXQ="), nil},
	})

	client, err := NewClient("localhost", &ClientConfig{
		Insecure:      true,
		MinTLSVersion: tls.VersionTLS12,
		TLSDialHandler: func(network, addr string, config *tls.Config) (TLSConnection, error) {
			dialCount++
			return conn, nil
		},
	})
	defer client.Close()
	require.NoError(t, err)

	require.NoError(t, client.Login("DENIC-1000011-RRI", "secret"))

	assert.Equal(t, 2, dialCount)*/
}

type mockReadWriteCloser struct {
	ReadResponses  []readResponse
	ReadIndex      int
	WriteResponses []writeResponse
	WriteIndex     int
	Closed         bool
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

func (m *mockReadWriteCloser) Read(p []byte) (int, error) {
	m.m.Lock()
	defer m.m.Unlock()

	if m.ReadIndex >= len(m.ReadResponses) {
		require.Fail(m.t, fmt.Sprintf("more than %d read operations detected", len(m.ReadResponses)))
	}
	if len(m.ReadResponses[m.ReadIndex].Data) > len(p) {
		panic("read response does not fit destination array")
	}
	m.ReadIndex++
	if m.ReadResponses[m.ReadIndex-1].Error != nil {
		return 0, m.ReadResponses[m.ReadIndex-1].Error
	}
	return len(m.ReadResponses[m.ReadIndex-1].Data), nil
}

func (m *mockReadWriteCloser) Write(p []byte) (int, error) {
	m.m.Lock()
	defer m.m.Unlock()

	if m.WriteIndex >= len(m.WriteResponses) {
		require.Fail(m.t, fmt.Sprintf("more than %d write operations detected", len(m.WriteResponses)))
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
