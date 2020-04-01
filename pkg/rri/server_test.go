package rri

import (
	"crypto/tls"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestServer(t *testing.T) {
	port := 31298
	tlsConfig, err := NewMockTLSConfig()
	if err != nil {
		panic(err)
	}
	server, err := NewServer(port, tlsConfig)
	if err != nil {
		panic(err)
	}

	queryCount := 0
	var lastQuery *Query

	server.Handler = func(s *Session, q *Query) (*Response, error) {
		queryCount++
		lastQuery = q
		return &Response{result: ResultSuccess}, nil
	}

	go func() {
		assert.NoError(t, server.Run())
	}()
	defer server.Close()

	client := &Client{address: fmt.Sprintf("localhost:%d", port)}
	client.tlsConfig = &tls.Config{
		MinVersion:         tls.VersionSSL30,
		CipherSuites:       []uint16{tls.TLS_RSA_WITH_AES_128_CBC_SHA},
		InsecureSkipVerify: true,
	}

	client.connection, err = tls.Dial("tcp", client.address, client.tlsConfig)
	if assert.NoError(t, err) {
		client.connection.Write(prepareMessage("version: 3.0\naction: LOGIN\nuser: user\npassword: secret"))

		// let some time pass for the query to be processed
		time.Sleep(50 * time.Millisecond)

		if assert.Equal(t, 1, queryCount, "expected to receive exactly one query") {
			assert.Equal(t, ActionLogin, lastQuery.Action())
			assert.Equal(t, "user", lastQuery.FirstField(FieldNameUser))
			assert.Equal(t, "secret", lastQuery.FirstField(FieldNamePassword))

			msg, err := readMessage(client.connection)
			if assert.NoError(t, err) {
				response, err := ParseResponseKV(msg)
				if assert.NoError(t, err) {
					assert.Equal(t, ResultSuccess, response.Result())
				}
			}
		}
	}
}

func TestServerSession(t *testing.T) {
	port := 31298
	tlsConfig, err := NewMockTLSConfig()
	if err != nil {
		panic(err)
	}
	server, err := NewServer(port, tlsConfig)
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
			s.Set("user", q.FirstField(FieldNameUser))

		} else {
			logoutQueryCount++
			user, ok := s.GetString("user")
			assert.True(t, ok)
			assert.Equal(t, expectedUser, user)
		}

		return &Response{result: ResultSuccess}, nil
	}

	go func() {
		assert.NoError(t, server.Run())
	}()
	defer server.Close()

	client, err := NewClient(fmt.Sprintf("localhost:%d", port))
	if assert.NoError(t, err) {
		if assert.NoError(t, client.Login(expectedUser, "secret")) {
			if assert.NoError(t, client.Logout()) {
				assert.Equal(t, 1, loginQueryCount)
				assert.Equal(t, 1, logoutQueryCount)
			}
		}
	}
}

func TestServerConcurrentConnections(t *testing.T) {
	port := 31298
	tlsConfig, err := NewMockTLSConfig()
	if err != nil {
		panic(err)
	}
	server, err := NewServer(port, tlsConfig)
	if err != nil {
		panic(err)
	}

	loggedIn := make(map[string]int)
	loggedOut := make(map[string]int)

	server.Handler = func(s *Session, q *Query) (*Response, error) {
		if q.Action() == ActionLogin {
			num, _ := loggedIn[q.FirstField(FieldNameUser)]
			loggedIn[q.FirstField(FieldNameUser)] = num + 1
			_, ok := s.GetString("user")
			assert.False(t, ok)
			s.Set("user", q.FirstField(FieldNameUser))

		} else {
			user, ok := s.GetString("user")
			assert.True(t, ok)
			num, _ := loggedOut[user]
			loggedOut[user] = num + 1
		}

		return &Response{result: ResultSuccess}, nil
	}

	go func() {
		assert.NoError(t, server.Run())
	}()
	defer server.Close()

	client1, err := NewClient(fmt.Sprintf("localhost:%d", port))
	if err != nil {
		panic(err)
	}

	client2, err := NewClient(fmt.Sprintf("localhost:%d", port))
	if err != nil {
		panic(err)
	}

	if assert.NoError(t, client1.Login("user1", "secret")) {
		defer client1.Close()
		if assert.NoError(t, client2.Login("user2", "secret")) {
			defer client2.Close()
			if assert.NoError(t, client1.Logout()) {
				if assert.NoError(t, client2.Logout()) {
					if assert.Len(t, loggedIn, 2) {
						if assert.Len(t, loggedOut, 2) {
							assert.Contains(t, loggedIn, "user1")
							assert.Contains(t, loggedIn, "user2")
							assert.Contains(t, loggedOut, "user1")
							assert.Contains(t, loggedOut, "user2")
							if !t.Failed() {
								assert.Equal(t, 1, loggedIn["user1"])
								assert.Equal(t, 1, loggedIn["user2"])
								assert.Equal(t, 1, loggedOut["user1"])
								assert.Equal(t, 1, loggedOut["user2"])
							}
						}
					}
				}
			}
		}
	}
}
