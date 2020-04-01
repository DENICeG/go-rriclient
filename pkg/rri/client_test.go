package rri

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClient(t *testing.T) {
	mustWithMockServer(func(server *MockServer) {
		server.AddUser("test", "secret")

		var client *Client

		t.Run("NewClient", func(t *testing.T) {
			var err error
			client, err = NewClient(server.Address())
			if assert.NoError(t, err) {
				assert.Equal(t, server.Address(), client.RemoteAddress())
				assert.False(t, client.IsLoggedIn())
			}
		})

		if client == nil {
			// just to prevent nil errors
			return
		}

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

		/*t.Run("NormalizeQueryHandles", func(t *testing.T) {
			q := NewCheckHandleQuery("asdf")
			client.NormalizeQueryHandles(q)
			assert.Equal(t, "DENIC-asdf", q.FirstField("handle"))
		})*/
	})
}
