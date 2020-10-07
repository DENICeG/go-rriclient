# Go RRI Package

This package provides types and functionality to send and receive RRI queries. Use this package to easily implement custom applications that interact with RRI.

## Client

The RRI client can be used to connect to a RRI server, send queries and interpret the response. It also handles the connection and authentication state and automatically re-connects after a lost connection. See the following, minimal example application for a quick start guide:

```go
package main

import (
    "log"

    "github.com/DENICeG/go-rriclient/pkg/rri"
)

func main() {
    // instantiate new RRI client and connect to server
    rriClient, err := rri.NewClient("rri.denic.de:51131", nil)
    if err != nil {
        log.Fatalln("failed to connect:", err.Error())
    }
    // close connection after you are done
    defer rriClient.Close()
    // send LOGIN query. these credentials are automatically re-used
    // when restoring a lost connection
    if err := rriClient.Login("DENIC-1000001-RRI", "secret"); err != nil {
        log.Fatalln("failed to log in:", err.Error())
    }
    // now you can use the client for any other queries
    log.Println(client.SendQuery(rri.NewInfoDomainQuery("denic.de")))
}
```

Pass `&rri.ClientConfig{Insecure: true}` as second parameter to `rri.NewClient` if you want to test an RRI server with self-signed certificate.

## Server

You can also instantiate a RRI server to receive queries and pass them to a custom handler. The RRI server implementation in this package does **not** implement user authentication, business logic or response codes, it solely offers functionality to handle incoming connections and read queries from them. See the following, minimal example application:

```go
package main

import (
    "log"

    "github.com/DENICeG/go-rriclient/pkg/rri"
)

func main() {
    // generate random server certificate and key for a mocked server.
    // for a real-world application you should use a valid certificate here
    tlsConfig, _ := rri.NewMockTLSConfig()
    // now prepare the tls listener
    rriServer, err := rri.NewServer(":51131", tlsConfig)
    if err != nil {
        log.Fatalln("failed to create server:", err.Error())
    }
    defer rriServer.Close()
    // register the handler that is called for every received query
    rriServer.Handler = func(s *Session, q *Query) (*Response, error) {
        // this method is called for every valid query.
        // malformed queries are ignored by the server
        log.Println("received query:", q)
        // return the response object that is sent to the client.
        // returning an error here will instantly close the connection
        // without sending a response the client
        return rri.NewResponse(ResultSuccess, nil), nil
    }
    // now run the listener loop to handle incoming connections and queries
    if err := rriServer.Run(); err != nil {
        log.Fatalln("failed to serve:", err.Error())
    }
}
```

You can use the `Session` parameter in your `Handler` func to persist information across all queries in the same TLS connection. A common use-case would be to store the username for that connection after a successful `LOGIN` query has been handled.