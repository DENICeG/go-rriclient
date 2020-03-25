package rri

import (
	"crypto/tls"
	"encoding/binary"
	"fmt"
	"io"
)

// QueryProcessor is used to process a query directly before sending. The returned query is sent to RRI server. Return nil to abort processing.
type QueryProcessor func(*Query) *Query

// RawQueryPrinter is called to print a raw outgoing or incoming query string.
type RawQueryPrinter func(msg string, isOutgoing bool)

// Client represents a stateful connection to a specific RRI Server.
type Client struct {
	address            string
	connection         *tls.Conn
	tlsConfig          *tls.Config
	currentUser        string
	lastUser, lastPass string
	Processor          QueryProcessor
	RawQueryPrinter    RawQueryPrinter
	XMLMode            bool
}

// NewClient returns a new Client object for the given RRI Server.
func NewClient(address string) (*Client, error) {
	client := &Client{
		address: address,
	}

	// taken from https://github.com/sebidude/go-rri
	client.tlsConfig = &tls.Config{
		MinVersion:         tls.VersionSSL30,
		CipherSuites:       []uint16{tls.TLS_RSA_WITH_AES_128_CBC_SHA},
		InsecureSkipVerify: true,
	}

	if err := client.setupConnection(); err != nil {
		return nil, err
	}

	return client, nil
}

func (client *Client) setupConnection() error {
	if client.connection == nil {
		var err error
		client.connection, err = tls.Dial("tcp", client.address, client.tlsConfig)
		client.currentUser = ""
		return err
	}
	return nil
}

// RemoteAddress returns the RRI server address and port.
func (client *Client) RemoteAddress() string {
	return client.address
}

// IsLoggedIn returns whether the client is currently logged in.
func (client *Client) IsLoggedIn() bool {
	return len(client.currentUser) > 0
}

// CurrentUser returns the currently logged in user.
func (client *Client) CurrentUser() string {
	return client.currentUser
}

// Close closes the underlying connection.
func (client *Client) Close() error {
	if client.connection != nil {
		return client.connection.Close()
	}
	return nil
}

// Login sends a login request to the server and checks for a success result.
func (client *Client) Login(username, password string) error {
	r, err := client.SendQuery(NewLoginQuery(username, password))
	if err != nil {
		return err
	}

	if r != nil && !r.IsSuccessful() {
		return fmt.Errorf("login failed")
	}

	return err
}

// Logout sends a logout request to the server.
func (client *Client) Logout() error {
	_, err := client.SendQuery(NewLogoutQuery())
	return err
}

// SendQuery sends a query to the server and returns the response.
//
// Only technical errors are returned. You need to check Response.Result to check for RRI error responses.
func (client *Client) SendQuery(query *Query) (*Response, error) {
	if client.XMLMode {
		return nil, fmt.Errorf("XML mode not yet supported")
	}

	if client.Processor != nil {
		query = client.Processor(query)
		if query == nil {
			return nil, nil
		}
	}

	if !client.IsLoggedIn() && query.action != ActionLogin {
		return nil, fmt.Errorf("need to log in before sending action %s", query.action)
	}
	if client.IsLoggedIn() && query.action == ActionLogin {
		return nil, fmt.Errorf("already logged in")
	}

	if query.action == ActionLogout {
		defer func() {
			// after action logout the connection and session are closed
			client.connection = nil
			client.currentUser = ""
			client.lastUser = ""
			client.lastPass = ""
		}()
	}

	rawResponse, err := client.SendRaw(query.EncodeKV())
	if err != nil {
		if err == io.EOF && query.action == ActionLogout {
			// the server will immediately close the connection once LOGOUT is received
			return nil, nil
		}
		return nil, err
	}

	response, err := ParseResponseKV(rawResponse)
	if err != nil {
		return nil, fmt.Errorf("received malformed response: %s", err.Error())
	}

	if query.action == ActionLogin && response.IsSuccessful() {
		client.currentUser = query.FirstField(FieldNameUser)
		// save credentials to restore session after lost connections
		client.lastUser = client.currentUser
		pwField := query.Field(FieldNamePassword)
		if pwField != nil && len(pwField) > 0 {
			client.lastPass = query.Field(FieldNamePassword)[0]
		} else {
			client.lastPass = ""
		}
	}

	return response, nil
}

// SendRaw sends a raw message to RRI and reads the returns the raw response.
//
// This method should be used with caution as it does not update the client login state.
func (client *Client) SendRaw(msg string) (string, error) {
	if client.RawQueryPrinter != nil {
		client.RawQueryPrinter(msg, true)
	}

	// ensure connection is established
	if err := client.setupConnection(); err != nil {
		return "", err
	}

	// prepare data packet: 4 byte message length + actual message
	data := []byte(msg)
	buffer := make([]byte, 4+len(data))
	binary.BigEndian.PutUint32(buffer[0:4], uint32(len(data)))
	copy(buffer[4:], data)

	response, err := client.sendAndReceive(buffer)
	if err != nil {
		// try re-establishing lost connection once
		if client.connection != nil {
			// ignore close errors (connection will be discarded anyway)
			client.connection.Close()
			client.connection = nil
		}
		if err := client.setupConnection(); err != nil {
			return "", fmt.Errorf("failed to restore lost connection: %s", err.Error())
		}
		// restore authenticated session if it existed before
		if len(client.lastUser) > 0 && len(client.lastPass) > 0 {
			if err := client.Login(client.lastUser, client.lastPass); err != nil {
				return "", fmt.Errorf("failed to restore session: %s", err.Error())
			}
		}
		// retry sending request once
		response, err = client.sendAndReceive(buffer)
		if err != nil {
			return "", err
		}
	}

	if client.RawQueryPrinter != nil {
		client.RawQueryPrinter(response, false)
	}
	return response, nil
}

func (client *Client) sendAndReceive(msg []byte) (string, error) {
	n, err := client.connection.Write(msg)
	if err != nil {
		return "", err
	}
	if n != len(msg) {
		return "", fmt.Errorf("failed to send %d bytes", len(msg))
	}
	return client.readResponse()
}

func (client *Client) readResponse() (string, error) {
	lenBuffer, err := client.readBytes(4)
	if err != nil {
		return "", err
	}
	len := binary.BigEndian.Uint32(lenBuffer)
	if len == 0 {
		return "", fmt.Errorf("server response is empty")
	}
	if len > 65536 {
		return "", fmt.Errorf("server response too large")
	}

	buffer, err := client.readBytes(int(len))
	if err != nil {
		return "", err
	}

	return string(buffer), nil
}

func (client *Client) readBytes(count int) ([]byte, error) {
	buffer := make([]byte, count)
	received := 0

	for received < count {
		len, err := client.connection.Read(buffer[received:])
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
