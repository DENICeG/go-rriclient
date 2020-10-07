package rri

import (
	"crypto/tls"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServer(t *testing.T) {
	port := 31298
	tlsConfig, err := NewMockTLSConfig()
	if err != nil {
		panic(err)
	}
	server, err := NewServer(fmt.Sprintf(":%d", port), tlsConfig)
	if err != nil {
		panic(err)
	}

	queryCount := 0
	var lastQuery *Query

	server.Handler = func(s *Session, q *Query) (*Response, error) {
		queryCount++
		lastQuery = q
		return NewResponse(ResultSuccess, nil), nil
	}

	go func() {
		assert.NoError(t, server.Run())
	}()
	defer server.Close()

	client := &Client{address: fmt.Sprintf("localhost:%d", port)}
	client.tlsConfig = &tls.Config{
		MinVersion:         tls.VersionTLS13,
		InsecureSkipVerify: true,
	}

	client.connection, err = tls.Dial("tcp", client.address, client.tlsConfig)
	require.NoError(t, err)
	client.connection.Write(prepareMessage("version: 3.0\naction: LOGIN\nuser: user\npassword: secret"))

	// let some time pass for the query to be processed
	time.Sleep(50 * time.Millisecond)

	require.Equal(t, 1, queryCount, "expected to receive exactly one query")
	assert.Equal(t, ActionLogin, lastQuery.Action())
	assert.Equal(t, "user", lastQuery.FirstField(QueryFieldNameUser))
	assert.Equal(t, "secret", lastQuery.FirstField(QueryFieldNamePassword))

	msg, err := readMessage(client.connection)
	require.NoError(t, err)
	response, err := ParseResponse(msg)
	require.NoError(t, err)
	assert.Equal(t, ResultSuccess, response.Result())
}

func TestServerSession(t *testing.T) {
	port := 31298
	tlsConfig, err := NewMockTLSConfig()
	if err != nil {
		panic(err)
	}
	server, err := NewServer(fmt.Sprintf(":%d", port), tlsConfig)
	if err != nil {
		panic(err)
	}

	expectedUser := "captain-kirk"
	loginQueryCount := 0
	logoutQueryCount := 0

	server.Handler = func(s *Session, q *Query) (*Response, error) {
		if q.Action() == ActionLogin {
			loginQueryCount++
			_, ok := s.GetString("user")
			assert.False(t, ok)
			s.Set("user", q.FirstField(QueryFieldNameUser))

		} else {
			logoutQueryCount++
			user, ok := s.GetString("user")
			assert.True(t, ok)
			assert.Equal(t, expectedUser, user)
		}

		return NewResponse(ResultSuccess, nil), nil
	}

	go func() {
		assert.NoError(t, server.Run())
	}()
	defer server.Close()

	client, err := NewClient(fmt.Sprintf("localhost:%d", port), &ClientConfig{Insecure: true})
	require.NoError(t, err)
	require.NoError(t, client.Login(expectedUser, "secret"))
	require.NoError(t, client.Logout())
	assert.Equal(t, 1, loginQueryCount)
	assert.Equal(t, 1, logoutQueryCount)
}

func TestServerConcurrentConnections(t *testing.T) {
	port := 31298
	tlsConfig, err := NewMockTLSConfig()
	if err != nil {
		panic(err)
	}
	server, err := NewServer(fmt.Sprintf(":%d", port), tlsConfig)
	if err != nil {
		panic(err)
	}

	loggedIn := make(map[string]int)
	loggedOut := make(map[string]int)

	server.Handler = func(s *Session, q *Query) (*Response, error) {
		if q.Action() == ActionLogin {
			num, _ := loggedIn[q.FirstField(QueryFieldNameUser)]
			loggedIn[q.FirstField(QueryFieldNameUser)] = num + 1
			_, ok := s.GetString("user")
			assert.False(t, ok)
			s.Set("user", q.FirstField(QueryFieldNameUser))

		} else {
			user, ok := s.GetString("user")
			assert.True(t, ok)
			num, _ := loggedOut[user]
			loggedOut[user] = num + 1
		}

		return NewResponse(ResultSuccess, nil), nil
	}

	go func() {
		assert.NoError(t, server.Run())
	}()
	defer server.Close()

	client1, err := NewClient(fmt.Sprintf("localhost:%d", port), &ClientConfig{Insecure: true})
	if err != nil {
		panic(err)
	}

	client2, err := NewClient(fmt.Sprintf("localhost:%d", port), &ClientConfig{Insecure: true})
	if err != nil {
		panic(err)
	}

	require.NoError(t, client1.Login("user1", "secret"))
	defer client1.Close()
	require.NoError(t, client2.Login("user2", "secret"))
	defer client2.Close()
	require.NoError(t, client1.Logout())
	require.NoError(t, client2.Logout())
	require.Len(t, loggedIn, 2)
	require.Len(t, loggedOut, 2)
	require.Contains(t, loggedIn, "user1")
	require.Contains(t, loggedIn, "user2")
	require.Contains(t, loggedOut, "user1")
	require.Contains(t, loggedOut, "user2")
	assert.Equal(t, 1, loggedIn["user1"])
	assert.Equal(t, 1, loggedIn["user2"])
	assert.Equal(t, 1, loggedOut["user1"])
	assert.Equal(t, 1, loggedOut["user2"])
}
