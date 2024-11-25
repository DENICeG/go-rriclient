package rri

import (
	"crypto/tls"
	"fmt"
	"net"
)

var (
	// ErrCloseConnection can be returned by QueryHandler to gracefully close the connection to the client.
	ErrCloseConnection = fmt.Errorf("gracefully shutdown connection to client")
)

// QueryHandler is called for incoming RRI queries by the server and expects a result as return value.
// If an error is returned instead, it is written to log and the connection is closed immediately.
type QueryHandler func(*Session, *Query) (*Response, error)

// Session is used to keep the state of an RRI connection.
type Session struct {
	values map[string]any
}

// Set sets a value for the current session across multiple queries.
func (s *Session) Set(key string, value any) {
	s.values[key] = value
}

// Get returns a value previously set for the current session.
func (s *Session) Get(key string) (any, bool) {
	value, ok := s.values[key]
	return value, ok
}

// GetString returns a string value previously set for the current session.
func (s *Session) GetString(key string) (string, bool) {
	if value, ok := s.values[key]; ok {
		if strVal, ok := value.(string); ok {
			return strVal, true
		}
	}
	return "", false
}

// GetInt returns an integer value previously set for the current session.
func (s *Session) GetInt(key string) (int, bool) {
	if value, ok := s.values[key]; ok {
		if intVal, ok := value.(int); ok {
			return intVal, true
		}
	}
	return 0, false
}

// GetBool returns a boolean value previously set for the current session.
func (s *Session) GetBool(key string) (bool, bool) {
	if value, ok := s.values[key]; ok {
		if boolVal, ok := value.(bool); ok {
			return boolVal, true
		}
	}
	return false, false
}

// Server represents a basic RRI client to receive RRI queries and send responses.
type Server struct {
	listener net.Listener
	Handler  QueryHandler
	isClosed bool
}

// NewServer returns a new RRI server for the given TLS config listening on the given port.
func NewServer(listenAddress string, tlsConfig *tls.Config) (*Server, error) {
	listener, err := tls.Listen("tcp", listenAddress, tlsConfig)
	if err != nil {
		return nil, err
	}

	return &Server{listener: listener, isClosed: false, Handler: nil}, nil
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

		go func() {
			session := &Session{make(map[string]any)}

			if err := func() error {
				for {
					msg, err := readMessage(conn)
					if err != nil {
						return err
					}

					if srv.Handler != nil {
						query, err := ParseQuery(msg)
						if err != nil {
							return err
						}

						response, err := srv.Handler(session, query)
						if err != nil {
							return err
						}

						// TODO answer in same type as the query (KV or XML)
						responseMsg := prepareMessage(response.EncodeKV())
						if _, err := conn.Write(responseMsg); err != nil {
							return err
						}
					} else {
						return fmt.Errorf("no RRI query handler defined")
					}
				}
			}(); err != nil {
				// TODO handle error
			}

			conn.Close()
		}()
	}
}
