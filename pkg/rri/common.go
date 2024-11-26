package rri

import (
	"encoding/binary"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/beevik/etree"
)

func PrepareMessage(msg string) []byte {
	// prepare data packet: 4 byte message length + actual message
	data := []byte(msg)
	buffer := make([]byte, 4+len(data))
	binary.BigEndian.PutUint32(buffer[0:4], uint32(len(data)))
	copy(buffer[4:], data)
	return buffer
}

func ReadMessage(r io.Reader) (string, error) {
	lenBuffer, err := ReadBytes(r, 4)
	if err != nil {
		return "", err
	}

	bytesRead := binary.BigEndian.Uint32(lenBuffer)
	if bytesRead == 0 {
		return "", fmt.Errorf("message is empty")
	}

	if bytesRead > 65536 || int(bytesRead) < 0 {
		return "", fmt.Errorf("message too large")
	}

	buffer, err := ReadBytes(r, int(bytesRead))
	if err != nil {
		return "", err
	}

	return string(buffer), nil
}

func ReadBytes(r io.Reader, count int) ([]byte, error) {
	buffer := make([]byte, count)
	received := 0

	for received < count {
		bytesRead, err := r.Read(buffer[received:])
		if err != nil {
			return nil, err
		}

		if bytesRead == 0 {
			return nil, fmt.Errorf("failed to read %d bytes from connection", count)
		}

		received += bytesRead
	}

	return buffer, nil
}

// CensorRawMessage replaces passwords in a raw query with '******'.
func CensorRawMessage(msg string) string {
	if strings.Contains(msg, "<xml") {
		doc := etree.NewDocument()
		err := doc.ReadFromString(msg)
		if err == nil {
			return DocToString(doc)
		}
	}

	pattern := regexp.MustCompile("([\r\n]|^)(password:[ \t]+)([^\r\n]*)([\r\n]|$)")
	return pattern.ReplaceAllStringFunc(msg, func(matchStr string) string {
		m := pattern.FindStringSubmatch(matchStr)
		return m[1] + m[2] + "******" + m[4]
	})
}

// DocToString converts an XML document to a string, replacing passwords with 'XXX'.
func DocToString(doc *etree.Document) string {
	results := doc.FindElements("//password")
	for _, result := range results {
		result.SetText("XXX")
	}

	result, _ := doc.WriteToString()

	return result
}
