package rri

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"time"
)

// NewMockTLSConfig returns a new, random TLS key and certificate pair for mock services.
//
// DO NOT USE IN PRODUCTION!
func NewMockTLSConfig() (*tls.Config, error) {
	privKey, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		return nil, fmt.Errorf("failed to generate RSA key: %s", err)
	}

	keyData, err := x509.MarshalPKCS8PrivateKey(privKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal private key: %s", err)
	}
	pemKeyData := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyData})

	serialNumber, err := rand.Int(rand.Reader, big.NewInt(128))
	if err != nil {
		return nil, fmt.Errorf("failed to generate serial number: %s", err)
	}

	template := x509.Certificate{
		SerialNumber:          serialNumber,
		Subject:               pkix.Name{Organization: []string{"DENIC eG"}},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(100 * 365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	certData, err := x509.CreateCertificate(rand.Reader, &template, &template, &privKey.PublicKey, privKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create certificate: %s", err)
	}
	pemCertData := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certData})

	tlsCert, err := tls.X509KeyPair(pemCertData, pemKeyData)
	if err != nil {
		return nil, fmt.Errorf("failed to load key pair: %s", err)
	}

	return &tls.Config{Certificates: []tls.Certificate{tlsCert}}, nil
}

// MockQueryHandler is called for every query, except LOGIN and LOGOUT.
type MockQueryHandler func(user string, session *Session, query *Query) (*Response, error)

// MockServer represents a mock RRI server with mocked user authentication.
type MockServer struct {
	server  *Server
	address string
	users   map[string]string
	Handler MockQueryHandler
}

// Run starts the underlying RRI server.
func (server *MockServer) Run() error {
	server.server.Handler = func(session *Session, query *Query) (*Response, error) {
		switch query.Action() {
		case ActionLogin:
			user := query.FirstField(QueryFieldNameUser)
			pass := query.FirstField(QueryFieldNamePassword)
			if userPass, ok := server.users[user]; ok && pass == userPass {
				session.Set("user", user)
				return NewResponse(ResultSuccess, nil), nil
			}
			return NewResponseWithError(ResultFailure, nil, NewBusinessMessage(83000000010, "Please login first")), nil

		case ActionLogout:
			return nil, ErrCloseConnection

		default:
			if server.Handler == nil {
				return NewResponse(ResultSuccess, nil), nil
			}
			user, ok := session.GetString("user")
			if !ok {
				return NewResponseWithError(ResultFailure, nil, NewBusinessMessage(83000000010, "Please login first")), nil
			}
			return server.Handler(user, session, query)
		}
	}

	return server.server.Run()
}

// Close closes the underlying RRI server.
func (server *MockServer) Close() error {
	return server.server.Close()
}

// AddUser adds a new user with given password or overwrites an existing one.
func (server *MockServer) AddUser(user, pass string) {
	server.users[user] = pass
}

// RemoveUser removes a user from the authentication list.
func (server *MockServer) RemoveUser(user string) {
	delete(server.users, user)
}

// Address returns the local address to use for an RRI client.
func (server *MockServer) Address() string {
	return server.address
}

// NewMockServer returns a mock server with user authentication for testing.
//
// DO NOT USE IN PRODUCTION!
func NewMockServer(port int) (*MockServer, error) {
	tlsConfig, err := NewMockTLSConfig()
	if err != nil {
		return nil, err
	}

	server, err := NewServer(fmt.Sprintf(":%d", port), tlsConfig)
	if err != nil {
		return nil, err
	}

	return &MockServer{server, fmt.Sprintf("localhost:%d", port), make(map[string]string), nil}, nil
}

// WithMockServer initializes and starts a mock server for the execution of f.
//
// DO NOT USE IN PRODUCTION!
func WithMockServer(port int, f func(server *MockServer) error) error {
	server, err := NewMockServer(port)
	if err != nil {
		return err
	}

	var runError error
	go func() {
		runError = server.Run()
	}()
	result := f(server)
	server.Close()

	if result != nil {
		return result
	}
	return runError
}

func mustWithMockServer(f func(server *MockServer)) {
	if err := WithMockServer(31298, func(server *MockServer) error {
		f(server)
		return nil
	}); err != nil {
		panic(err)
	}
}
