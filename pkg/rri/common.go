package rri

import (
	"encoding/binary"
	"fmt"
	"io"
)

func prepareMessage(msg string) []byte {
	// prepare data packet: 4 byte message length + actual message
	data := []byte(msg)
	buffer := make([]byte, 4+len(data))
	binary.BigEndian.PutUint32(buffer[0:4], uint32(len(data)))
	copy(buffer[4:], data)
	return buffer
}

func readMessage(r io.Reader) (string, error) {
	lenBuffer, err := readBytes(r, 4)
	if err != nil {
		return "", err
	}
	len := binary.BigEndian.Uint32(lenBuffer)
	if len == 0 {
		return "", fmt.Errorf("message is empty")
	}
	if len > 65536 {
		return "", fmt.Errorf("message too large")
	}

	buffer, err := readBytes(r, int(len))
	if err != nil {
		return "", err
	}

	return string(buffer), nil
}

func readBytes(r io.Reader, count int) ([]byte, error) {
	buffer := make([]byte, count)
	received := 0

	for received < count {
		len, err := r.Read(buffer[received:])
		if err != nil {
			return nil, err
		}
		if len == 0 {
			return nil, fmt.Errorf("failed to read %d bytes from connection", count)
		}

		received += len
	}

	return buffer, nil
}
