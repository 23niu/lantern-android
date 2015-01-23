package client

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"strconv"

	"github.com/getlantern/balancer"
)

const (
	HttpConnectMethod = "CONNECT"
	XFlashlightQOS    = "X-Flashlight-QOS"
)

// Client is a HTTP proxy that accepts connections from local programs and
// proxies these via remote flashlight servers.
type Client struct {
	Addr           string
	frontedServers []*frontedServer
	ln             *Listener

	rpCh          chan *httputil.ReverseProxy
	rpInitialized bool

	balInitialized bool
	balCh          chan *balancer.Balancer
}

// AddFrontedServer adds a fronted server to the list.
func (client *Client) AddFrontedServer(fs *frontedServer) error {
	client.frontedServers = append(client.frontedServers, fs)
	// TODO: Find the best way to add this server to the balancer list.
	return nil
}

// NewClient creates a proxy client.
func NewClient(addr string) *Client {
	client := &Client{Addr: addr}

	client.frontedServers = make([]*frontedServer, 0, len(defaultFrontedServerList))

	log.Printf("Adding %d domain fronted servers.", len(defaultFrontedServerList))

	// Adding default fronted servers.
	for _, fs := range defaultFrontedServerList {
		log.Printf("Adding %s:%d.", fs.Host, fs.Port)
		client.AddFrontedServer(&fs)
	}

	// Starting up balancer.
	client.initBalancer()

	// Starting reverse proxy
	client.initReverseProxy()

	return client
}

// ServeHTTP implements the method from interface http.Handler using the latest
// handler available from getHandler() and latest ReverseProxy available from
// getReverseProxy().
func (client *Client) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	if req.Method == HttpConnectMethod {
		client.intercept(resp, req)
	} else {
		client.getReverseProxy().ServeHTTP(resp, req)
	}
}

// ListenAndServe spawns the HTTP proxy and makes it listen for incoming
// connections.
func (client *Client) ListenAndServe() (err error) {
	addr := client.Addr

	if addr == "" {
		addr = ":http"
	}

	if client.ln, err = NewListener(addr); err != nil {
		return err
	}

	httpServer := &http.Server{
		Addr:    client.Addr,
		Handler: client,
	}

	log.Printf("Starting proxy server at %s...", addr)

	return httpServer.Serve(client.ln)
}

func targetQOS(req *http.Request) int {
	requestedQOS := req.Header.Get(XFlashlightQOS)
	if requestedQOS != "" {
		rqos, err := strconv.Atoi(requestedQOS)
		if err == nil {
			return rqos
		}
	}
	return 0
}

// intercept intercepts an HTTP CONNECT request, hijacks the underlying client
// connetion and starts piping the data over a new net.Conn obtained from the
// given dial function.
func (client *Client) intercept(resp http.ResponseWriter, req *http.Request) {
	if req.Method != HttpConnectMethod {
		panic("Intercept used for non-CONNECT request!")
	}

	// Hijack underlying connection
	clientConn, _, err := resp.(http.Hijacker).Hijack()
	if err != nil {
		respondBadGateway(resp, fmt.Sprintf("Unable to hijack connection: %s", err))
		return
	}
	defer clientConn.Close()

	addr := hostIncludingPort(req, 443)

	// Establish outbound connection
	connOut, err := client.getBalancer().DialQOS("tcp", addr, targetQOS(req))
	if err != nil {
		respondBadGateway(clientConn, fmt.Sprintf("Unable to handle CONNECT request: %s", err))
		return
	}
	defer connOut.Close()

	// Pipe data
	pipeData(clientConn, connOut, req)
}

// Stop is currently not implemented but should make the listener stop
// accepting new connections and then kill all active connections.
func (client *Client) Stop() error {
	log.Printf("Stopping proxy server...")
	return client.ln.Stop()
}

func respondBadGateway(w io.Writer, msg string) error {
	log.Printf("Responding BadGateway: %v", msg)
	resp := &http.Response{
		StatusCode: http.StatusBadGateway,
		ProtoMajor: 1,
		ProtoMinor: 1,
	}
	err := resp.Write(w)
	if err == nil {
		_, err = w.Write([]byte(msg))
	}
	return err
}

// hostIncludingPort extracts the host:port from a request.  It fills in a
// a default port if none was found in the request.
func hostIncludingPort(req *http.Request, defaultPort int) string {
	_, port, err := net.SplitHostPort(req.Host)
	if port == "" || err != nil {
		return req.Host + ":" + strconv.Itoa(defaultPort)
	}
	return req.Host
}

// pipeData pipes data between the client and proxy connections.  It's also
// responsible for responding to the initial CONNECT request with a 200 OK.
func pipeData(clientConn net.Conn, connOut net.Conn, req *http.Request) {
	// Start piping to proxy
	go io.Copy(connOut, clientConn)

	// Respond OK
	err := respondOK(clientConn, req)
	if err != nil {
		log.Printf("Unable to respond OK: %s", err)
		return
	}

	// Then start coyping from out to client
	io.Copy(clientConn, connOut)
}

func respondOK(writer io.Writer, req *http.Request) error {
	defer req.Body.Close()
	resp := &http.Response{
		StatusCode: http.StatusOK,
		ProtoMajor: 1,
		ProtoMinor: 1,
	}
	return resp.Write(writer)
}
