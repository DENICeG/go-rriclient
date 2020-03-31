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
	"net"
	"time"
)

// NewMockTLSConfig returns a new, random TLS key and certificate pair for mock services.
//
// DO NOT USE IN PRODUCTION!
func NewMockTLSConfig() *tls.Config {
	privKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		panic(fmt.Sprintf("failed to generate RSA key: %s", err))
	}

	keyData, err := x509.MarshalPKCS8PrivateKey(privKey)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal private key: %s", err))
	}
	pemKeyData := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyData})

	serialNumber, err := rand.Int(rand.Reader, big.NewInt(128))
	if err != nil {
		panic(fmt.Sprintf("failed to generate serial number: %s", err))
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
		panic(fmt.Sprintf("failed to create certificate: %s", err))
	}
	pemCertData := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certData})

	tlsCert, err := tls.X509KeyPair(pemCertData, pemKeyData)
	if err != nil {
		panic(fmt.Sprintf("failed to load key pair: %s", err))
	}

	return &tls.Config{Certificates: []tls.Certificate{tlsCert}}
}

// QueryHandler is called for incoming RRI queryies by the server and expects a result as return value.
// If an error is returned instead, it is written to log and the connection is closed immmediately.
type QueryHandler func(*Query) (*Response, error)

// Server represents a basic RRI client to receive RRI queries and send responses.
type Server struct {
	listener net.Listener
	isClosed bool
	Handler  QueryHandler
}

// NewServer returns a new RRI server for the given TLS config listening on the given port.
func NewServer(port int, tlsConfig *tls.Config) (*Server, error) {
	listener, err := tls.Listen("tcp", fmt.Sprintf(":%d", port), tlsConfig)
	if err != nil {
		return nil, err
	}

	return &Server{listener, false, nil}, nil
}

// Close gracefully shuts down the server.
func (srv *Server) Close() error {
	srv.isClosed = true
	return srv.listener.Close()
}

// Run starts accepting client connections to pass to the configured query handler and blocks until the server is stopped.
func (srv *Server) Run() error {
	for {
		conn, err := srv.listener.Accept()
		if err != nil {
			if srv.isClosed {
				return nil
			}
			return err
		}

		if err := func() error {
			for {
				msg, err := readMessage(conn)
				if err != nil {
					return err
				}

				if srv.Handler != nil {
					query, err := ParseQuery(string(msg))
					if err != nil {
						return err
					}

					response, err := srv.Handler(query)
					if err != nil {
						return err
					}

					responseMsg := prepareMessage(response.EncodeKV())
					if _, err := conn.Write([]byte(responseMsg)); err != nil {
						return err
					}
				} else {
					return fmt.Errorf("no RRI query handler defined")
				}
			}
		}(); err != nil {
			//TODO handle error
		}

		conn.Close()
	}
}
