package rri

import (
	"crypto/tls"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// TLSDialer is the callback function to open a new TLS connection. Maps tls.Dial by default.
type TLSDialer func(network, addr string, config *tls.Config) (TLSConnection, error)

// TLSConnection wraps a TLS connection as denoted by *tls.Conn.
type TLSConnection interface {
	io.ReadWriteCloser
}

// QueryProcessor is used to process a query directly before sending. The returned query is sent to RRI server. Return nil to abort processing.
type QueryProcessor func(*Query) *Query

// RawQueryPrinter is called to print a raw outgoing or incoming query string.
type RawQueryPrinter func(msg string, isOutgoing bool)

// ErrorPrinter is called to print uncritical errors.
type ErrorPrinter func(err error)

// Client represents a stateful connection to a specific RRI Server.
type Client struct {
	connection        TLSConnection
	dialer            TLSDialer
	tlsConfig         *tls.Config
	RawQueryPrinter   RawQueryPrinter
	InnerErrorPrinter ErrorPrinter
	address           string
	currentUser       string
	lastUser          string
	lastPass          string
	XMLMode           bool
	NoAutoRetry       bool
}

// ClientConfig can be used to further configure the RRI client.
type ClientConfig struct {
	// TLSDialHandler denotes the TLS dialer to use for the instanced RRI Client. Maps tls.Dial by default.
	TLSDialHandler TLSDialer
	// Insecure allows to accept self-signed SSL certificates.
	Insecure bool
	// MinTLSVersion denotes the minimum accepted TLS version.
	MinTLSVersion uint16
}

// NewClient returns a new Client object for the given RRI Server.
func NewClient(address string, conf *ClientConfig) (*Client, error) {
	var actualConf ClientConfig
	if conf != nil {
		// create copy of config to operate on
		actualConf = *conf
	}
	if !strings.ContainsRune(address, ':') {
		const defaultPort = ":51131"
		address += defaultPort
	}
	if actualConf.TLSDialHandler == nil {
		// use tls.Dial by default to establish a tls connection
		actualConf.TLSDialHandler = func(network, addr string, config *tls.Config) (TLSConnection, error) {
			return tls.Dial(network, addr, config)
		}
	}
	if actualConf.MinTLSVersion <= 0 {
		actualConf.MinTLSVersion = tls.VersionTLS13
	}

	client := &Client{
		address: address,
		dialer:  actualConf.TLSDialHandler,
		tlsConfig: &tls.Config{
			MinVersion:         actualConf.MinTLSVersion,
			InsecureSkipVerify: actualConf.Insecure,
		},
	}

	if err := client.setupConnection(); err != nil {
		return nil, err
	}

	return client, nil
}

func (client *Client) Connection() TLSConnection {
	return client.connection
}

func (client *Client) setupConnection() error {
	if client.connection == nil {
		var err error
		client.connection, err = client.dialer("tcp", client.address, client.tlsConfig)
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

// CurrentRegAccID tries to parse the RegAccID from CurrentUser.
func (client *Client) CurrentRegAccID() (int, error) {
	parts := strings.Split(client.currentUser, "-")
	if len(parts) < 2 {
		return 0, fmt.Errorf("malformed login name")
	}
	regAccID, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, fmt.Errorf("malformed login name")
	}
	return regAccID, nil
}

// Close closes the underlying connection.
func (client *Client) Close() error {
	if client.connection != nil {
		// TODO send LOGOUT while connected?
		return client.closeConnection()
	}
	return nil
}

func (client *Client) closeConnection() error {
	if client.connection == nil {
		// no connection? nothing to do
		return nil
	}

	// wrap tls connection close to also catch nil-pointer panics (regression in tls package)
	var err error
	func() {
		defer func() {
			if r := recover(); r != nil {
				// pass recovered panic as error
				err = fmt.Errorf("panic: %v", r)
			}
		}()

		err = client.connection.Close()
	}()
	client.connection = nil
	return err
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

	if !client.IsLoggedIn() && query.Action() != ActionLogin {
		return nil, fmt.Errorf("need to log in before sending action %s", query.Action())
	}
	if client.IsLoggedIn() && query.Action() == ActionLogin {
		return nil, fmt.Errorf("already logged in")
	}

	if query.Action() == ActionLogout {
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
		if err == io.EOF && query.Action() == ActionLogout {
			// the server will immediately close the connection once LOGOUT is received
			return nil, nil
		}
		return nil, err
	}

	response, err := ParseResponse(rawResponse)
	if err != nil {
		return nil, fmt.Errorf("received malformed response: %s", err.Error())
	}

	if query.Action() == ActionLogin && response.IsSuccessful() {
		client.currentUser = query.FirstField(QueryFieldNameUser)
		// save credentials to restore session after lost connections
		client.lastUser = client.currentUser
		pwField := query.Field(QueryFieldNamePassword)
		if len(pwField) > 0 {
			client.lastPass = query.Field(QueryFieldNamePassword)[0]
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
	// ensure connection is established
	if err := client.setupConnection(); err != nil {
		return "", err
	}

	buffer := PrepareMessage(msg)

	if client.RawQueryPrinter != nil {
		client.RawQueryPrinter(msg, true)
	}
	response, err := client.sendAndReceive(buffer)
	if err != nil {
		if client.NoAutoRetry {
			return "", err
		}

		if client.InnerErrorPrinter != nil {
			client.InnerErrorPrinter(fmt.Errorf("query failed: %s", err))
		}

		// try re-establishing lost connection once
		if client.connection != nil {
			// ignore close errors (connection will be discarded anyway)
			client.closeConnection()
		}
		if err = client.setupConnection(); err != nil {
			return "", fmt.Errorf("failed to restore lost connection: %s", err.Error())
		}
		// restore authenticated session if it existed before
		if len(client.lastUser) > 0 && len(client.lastPass) > 0 {
			if err = client.Login(client.lastUser, client.lastPass); err != nil {
				return "", fmt.Errorf("failed to restore session: %s", err.Error())
			}
		}
		// retry sending request once
		if client.RawQueryPrinter != nil {
			client.RawQueryPrinter(msg, true)
		}
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
	return ReadMessage(client.connection)
}
