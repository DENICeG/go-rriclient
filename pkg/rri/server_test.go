package rri_test

import (
	"crypto/tls"
	"fmt"
	"testing"
	"time"

	"github.com/DENICeG/go-rriclient/pkg/rri"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServer(t *testing.T) {
	port := 31298
	tlsConfig, err := rri.NewMockTLSConfig()
	if err != nil {
		panic(err)
	}
	server, err := rri.NewServer(fmt.Sprintf(":%d", port), tlsConfig)
	if err != nil {
		panic(err)
	}

	queryCount := 0
	var lastQuery *rri.Query

	server.Handler = func(s *rri.Session, q *rri.Query) (*rri.Response, error) {
		queryCount++
		lastQuery = q
		return rri.NewResponse(rri.ResultSuccess, nil), nil
	}

	go func() {
		assert.NoError(t, server.Run())
	}()
	defer server.Close()

	address := fmt.Sprintf("localhost:%d", port)
	client, err := rri.NewClient(address, &rri.ClientConfig{Insecure: true, MinTLSVersion: tls.VersionTLS13})
	require.NoError(t, err)

	client.Connection().Write(rri.PrepareMessage("version: 5.0\naction: LOGIN\nuser: user\npassword: secret"))

	// let some time pass for the query to be processed
	time.Sleep(50 * time.Millisecond)

	require.Equal(t, 1, queryCount, "expected to receive exactly one query")
	assert.Equal(t, rri.ActionLogin, lastQuery.Action())
	assert.Equal(t, "user", lastQuery.FirstField(rri.QueryFieldNameUser))
	assert.Equal(t, "secret", lastQuery.FirstField(rri.QueryFieldNamePassword))

	msg, err := rri.ReadMessage(client.Connection())
	require.NoError(t, err)
	response, err := rri.ParseResponse(msg)
	require.NoError(t, err)
	assert.Equal(t, rri.ResultSuccess, response.Result())
}

func TestServerSession(t *testing.T) {
	port := 31298
	tlsConfig, err := rri.NewMockTLSConfig()
	if err != nil {
		panic(err)
	}
	server, err := rri.NewServer(fmt.Sprintf(":%d", port), tlsConfig)
	if err != nil {
		panic(err)
	}

	expectedUser := "captain-kirk"
	loginQueryCount := 0
	logoutQueryCount := 0

	server.Handler = func(s *rri.Session, q *rri.Query) (*rri.Response, error) {
		if q.Action() == rri.ActionLogin {
			loginQueryCount++
			_, ok := s.GetString("user")
			assert.False(t, ok)
			s.Set("user", q.FirstField(rri.QueryFieldNameUser))

		} else {
			logoutQueryCount++
			user, ok := s.GetString("user")
			assert.True(t, ok)
			assert.Equal(t, expectedUser, user)
		}

		return rri.NewResponse(rri.ResultSuccess, nil), nil
	}

	go func() {
		assert.NoError(t, server.Run())
	}()
	defer server.Close()

	client, err := rri.NewClient(fmt.Sprintf("localhost:%d", port), &rri.ClientConfig{Insecure: true})
	require.NoError(t, err)
	require.NoError(t, client.Login(expectedUser, "secret"))
	require.NoError(t, client.Logout())
	assert.Equal(t, 1, loginQueryCount)
	assert.Equal(t, 1, logoutQueryCount)
}

func TestServerConcurrentConnections(t *testing.T) {
	port := 31298
	tlsConfig, err := rri.NewMockTLSConfig()
	if err != nil {
		panic(err)
	}
	server, err := rri.NewServer(fmt.Sprintf(":%d", port), tlsConfig)
	if err != nil {
		panic(err)
	}

	loggedIn := make(map[string]int)
	loggedOut := make(map[string]int)

	server.Handler = func(s *rri.Session, q *rri.Query) (*rri.Response, error) {
		if q.Action() == rri.ActionLogin {
			num := loggedIn[q.FirstField(rri.QueryFieldNameUser)]
			loggedIn[q.FirstField(rri.QueryFieldNameUser)] = num + 1
			_, ok := s.GetString("user")
			assert.False(t, ok)
			s.Set("user", q.FirstField(rri.QueryFieldNameUser))

		} else {
			user, ok := s.GetString("user")
			assert.True(t, ok)
			num := loggedOut[user]
			loggedOut[user] = num + 1
		}

		return rri.NewResponse(rri.ResultSuccess, nil), nil
	}

	go func() {
		assert.NoError(t, server.Run())
	}()
	defer server.Close()

	client1, err := rri.NewClient(fmt.Sprintf("localhost:%d", port), &rri.ClientConfig{Insecure: true})
	if err != nil {
		panic(err)
	}

	client2, err := rri.NewClient(fmt.Sprintf("localhost:%d", port), &rri.ClientConfig{Insecure: true})
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
