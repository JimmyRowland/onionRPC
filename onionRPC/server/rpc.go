package server

import (
	"fmt"
	"io"
	"net/http"
	"net/rpc"
	"net/rpc/jsonrpc"
)

//https://gist.github.com/maiiz/e8c44472a8a09b3284db9f80470afacb
const (
	maxHTTPRequestContentLength = 1024 * 128
)

type httpReadWriteNopCloser struct {
	io.Reader
	io.Writer
}

func (c *httpReadWriteNopCloser) Close() error { return nil }

// RPCServer wrappers net/rpc RPCServer.
type RPCServer struct {
	*rpc.Server
}

// NewServer returns a JSON-RPC RPCServer instance.
func NewRPCServer() *RPCServer {
	return &RPCServer{rpc.NewServer()}
}

// ServeHTTP implements an http.Handler that answers every RPC requests.
// Overwrite this method, we can wrapper a simple JSON-RPC framework.
func (server *RPCServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.ContentLength > maxHTTPRequestContentLength {
		http.Error(w,
			fmt.Sprintf("content length too large (%d>%d)", req.ContentLength, maxHTTPRequestContentLength),
			http.StatusRequestEntityTooLarge)
		return
	}
	w.Header().Set("content-type", "application/json")
	server.ServeRequest(jsonrpc.NewServerCodec(&httpReadWriteNopCloser{req.Body, w}))
}

// NewHTTPServer creates a new HTTP RPC server around an API provider.
func NewHTTPServer(srv *RPCServer) *http.Server {
	return &http.Server{Handler: srv}
}
