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
	server, err := NewServer(port, NewMockTLSConfig())
	if assert.NoError(t, err) {
		queryCount := 0
		var lastQuery *Query

		server.Handler = func(q *Query) (*Response, error) {
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
}
