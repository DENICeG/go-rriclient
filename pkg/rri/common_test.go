package rri_test

import (
	"bytes"
	"encoding/base64"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/DENICeG/go-rriclient/pkg/rri"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func b64(str string) []byte {
	data, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		panic(err)
	}
	return data
}

func b64enc(data []byte) string {
	if data == nil {
		return ""
	}
	return base64.StdEncoding.EncodeToString(data)
}

func TestPrepareMessage(t *testing.T) {
	// hardcoded RRI packet with size prefix and query "version: 5.0\naction: LOGIN\nuser: user\npassword: secret"
	expected := b64("AAAANnZlcnNpb246IDUuMAphY3Rpb246IExPR0lOCnVzZXI6IHVzZXIKcGFzc3dvcmQ6IHNlY3JldA==")
	assert.Equal(t, expected, rri.PrepareMessage("version: 5.0\naction: LOGIN\nuser: user\npassword: secret"))
}

func TestReadMessage(t *testing.T) {
	expectedMsg := "version: 5.0\naction: LOGIN\nuser: user\npassword: secret"
	r := bytes.NewReader(rri.PrepareMessage(expectedMsg))
	msg, err := rri.ReadMessage(r)
	require.NoError(t, err)
	assert.Equal(t, expectedMsg, msg)
}

func TestReadMessageEmpty(t *testing.T) {
	_, err := rri.ReadMessage(bytes.NewReader(rri.PrepareMessage("")))
	assert.Error(t, err)
}

func TestReadMessageTooLong(t *testing.T) {
	_, err := rri.ReadMessage(bytes.NewReader(rri.PrepareMessage(strings.Repeat("a", 70000))))
	assert.Error(t, err)
}

func TestReadMessageNoData(t *testing.T) {
	_, err := rri.ReadMessage(bytes.NewReader([]byte{}))
	assert.Error(t, err)
}

func TestReadMessageIncompleteSize(t *testing.T) {
	_, err := rri.ReadMessage(bytes.NewReader([]byte{0}))
	assert.Error(t, err)
}

func TestReadMessageIncompleteMessage(t *testing.T) {
	msg := b64("AAAANnZlcnNp")
	_, err := rri.ReadMessage(bytes.NewReader(msg))
	assert.Error(t, err)
}

func TestReadBytes(t *testing.T) {
	r, w := io.Pipe()
	go func() {
		for i := 0; i < 10; i++ {
			// do not send all data in a single packet to test fragmented reading
			w.Write([]byte{byte(i)})
			time.Sleep(100 * time.Microsecond)
		}
	}()

	data, err := rri.ReadBytes(r, 10)
	require.NoError(t, err)
	assert.Equal(t, []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}, data)
}

func TestCensorRawMessage(t *testing.T) {
	tt := []struct {
		description string
		input       string
		expected    string
	}{
		{description: "no sensitive data", input: "version: 5.0\naction: info\ndomain: denic.de", expected: "version: 5.0\naction: info\ndomain: denic.de"},
		{description: "no sensitive data - keyword", input: "version: 5.0\naction: info\nno-password: foobar\ndomain: denic.de", expected: "version: 5.0\naction: info\nno-password: foobar\ndomain: denic.de"},
		{description: "empty password", input: "version: 5.0\naction: info\npassword:\ndomain: denic.de", expected: "version: 5.0\naction: info\npassword:\ndomain: denic.de"},
		{description: "actual password", input: "password: secret-password\nversion: 5.0\naction: LOGIN\nuser: DENIC-1000011-RRI", expected: "password: ******\nversion: 5.0\naction: LOGIN\nuser: DENIC-1000011-RRI"},
		{description: "actual password 2", input: "version: 5.0\naction: LOGIN\npassword: secret-password\nuser: DENIC-1000011-RRI", expected: "version: 5.0\naction: LOGIN\npassword: ******\nuser: DENIC-1000011-RRI"},
		{description: "actual password 3", input: "version: 5.0\naction: LOGIN\nuser: DENIC-1000011-RRI\npassword: secret-password", expected: "version: 5.0\naction: LOGIN\nuser: DENIC-1000011-RRI\npassword: ******"},
		{description: "actual password 4", input: "password: secret-password\nversion: 5.0\npassword: secret-password\naction: LOGIN\nuser: DENIC-1000011-RRI\npassword: secret-password", expected: "password: ******\nversion: 5.0\npassword: ******\naction: LOGIN\nuser: DENIC-1000011-RRI\npassword: ******"},
	}

	for _, tc := range tt {
		t.Run(tc.description, func(t *testing.T) {
			assert.Equal(t, tc.expected, rri.CensorRawMessage(tc.input))
		})
	}
}
