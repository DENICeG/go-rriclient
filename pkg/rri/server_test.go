package rri

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServer(t *testing.T) {
	port := 31298
	server, err := NewServer(port, NewMockTLSConfig())
	if assert.NoError(t, err) {
		server.Handler = func(q *Query) (*Response, error) {
			return &Response{result: ResultSuccess}, nil
		}

		go server.Run()
		defer server.Close()

		client, err := NewClient(fmt.Sprintf("localhost:%d", port))
		if assert.NoError(t, err) {
			assert.NoError(t, client.Login("user", "secret"))
		}
	}
}
