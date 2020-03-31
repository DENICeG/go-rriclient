package rri

import (
	"bytes"
	"encoding/base64"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPrepareMessage(t *testing.T) {
	// hardcoded RRI packet with size prefix and query "version: 3.0\naction: LOGIN\nuser: user\npassword: secret"
	expected, _ := base64.StdEncoding.DecodeString("AAAANnZlcnNpb246IDMuMAphY3Rpb246IExPR0lOCnVzZXI6IHVzZXIKcGFzc3dvcmQ6IHNlY3JldA==")
	assert.Equal(t, expected, prepareMessage("version: 3.0\naction: LOGIN\nuser: user\npassword: secret"))
}

func TestReadMessage(t *testing.T) {
	expectedMsg := "version: 3.0\naction: LOGIN\nuser: user\npassword: secret"
	r := bytes.NewReader(prepareMessage(expectedMsg))
	msg, err := readMessage(r)
	if assert.NoError(t, err) {
		assert.Equal(t, expectedMsg, msg)
	}
}

func TestReadMessageEmpty(t *testing.T) {
	_, err := readMessage(bytes.NewReader(prepareMessage("")))
	assert.Error(t, err)
}

func TestReadMessageTooLong(t *testing.T) {
	_, err := readMessage(bytes.NewReader(prepareMessage(strings.Repeat("a", 70000))))
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

	data, err := readBytes(r, 10)
	if assert.NoError(t, err) {
		assert.Equal(t, []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}, data)
	}
}
