// Gateway Layer speaks Redis to the app, and does code-level read/write splitting leveraging radix's RESP parser.
// Hat tip to https://github.com/mediocregopher/breadis for lots of base code.
package gateway

import (
	"errors"
	rp "github.com/mediocregopher/radix.v2/redis"
	"log"
	"net"
	"strings"
	"time"
)

const (
	Unknown = 0
	Read    = iota
	Write
	Blacklist
	Silenced
	Internal
)

var (
	ErrBadCmd        = errors.New("ERR pylon/gateway: bad command")
	ErrBlacklisted   = errors.New("ERR pylon/gateway: blacklisted command")
	ErrHandlerUnset  = errors.New("ERR pylon/gateway: handler is not set")
	ErrResolverUnset = errors.New("ERR pylon/gateway: resolver is not set")
)

// The Gateway type is a regular TCP socket actively listening for and speaking RESP (REdis Serialization Protocol.)
// It requires two handlers to work, both should be individually registered with g.RegisterHandler(), a Read and a Write handler.
// The gateway is aware of all Redis commands as of 3.2.0, see this repo's README for more information.
//
// Ensure that you close a gateway when you're done listening to it to free resources.
type Gateway struct {
	l      net.Listener
	h      map[int]GatewayHandler
	r      HandlerResolver
	closed bool
}

// A GatewayHandler is a contract for the Read/Write handlers. It accepts a TCP connection and the input command as a string.
type GatewayHandler func(net.Conn, string)

// A HandlerResolver takes in the first section of the command, and routes it accordingly.
type HandlerResolver func(string) (int, error)

// Returns a new gateway with the attached net.TCPListener.
func NewGateway(l net.Listener) *Gateway {
	return &Gateway{
		l: l,
		h: make(map[int]GatewayHandler),
		r: PylonRWResolver,
	}
}

// Registers a handler for either reads or writes. If you cannot import gateway for some reason,
// the first argument is either 1 for reads or 2 for writes. This can also be completely arbitrary if you register your own resolver.
func (g *Gateway) RegisterHandler(t int, fn GatewayHandler) {
	g.h[t] = fn
}

// Registers the resolver to be called to decide on how
func (g *Gateway) RegisterResolver(fn HandlerResolver) {
	g.r = fn
}

// Calls the specified handler or returns an error to the socket.
func (g *Gateway) callHandler(t int, cmd string, conn net.Conn) {

	h, ok := g.h[t]

	if !ok {
		rp.NewResp(ErrHandlerUnset).WriteTo(conn)
		return
	}

	go h(conn, cmd)
}

func (g *Gateway) callResolver(cmd string) (int, error) {

	if g.r == nil {
		return -1, ErrResolverUnset
	}

	i, e := g.r(cmd)
	return i, e

}

// Listens to TCP connections. This is the main event loop for the Gateway layer.
func (g *Gateway) Listen() {
	l := g.l
	g.closed = false
	defer l.Close()

	for {
		if g.closed {
			break
		}
		c, err := l.Accept()
		if err != nil {
			log.Print(err)
			continue
		}

		go g.handleConnection(c)
	}
}

// Closes the gateway on next loop.
func (g *Gateway) Close() {
	g.closed = true
}

// Handles a connection. Essentially reads then parses the RESP line into individual parts,
// figures out how to route it (e.g. reads, writes, blacklisted, etc,) then recompiles it into a string form,
// and finally calls it's handler.
func (g *Gateway) handleConnection(conn net.Conn) {
	defer conn.Close()

	rr := rp.NewRespReader(conn)
	for {
		err := conn.SetReadDeadline(time.Now().Add(1 * time.Second))
		if err != nil {
			log.Print(err)
			return
		}

		r := rr.Read()
		if rp.IsTimeout(r) {
			continue
		} else if r.IsType(rp.IOErr) {
			return
		}

		ms, err := r.Array()
		if err != nil {
			rp.NewResp(ErrBadCmd).WriteTo(conn)
		}

		cmd, err := ms[0].Str()
		if err != nil {
			rp.NewResp(ErrBadCmd).WriteTo(conn)
		}
		cmd = strings.ToUpper(cmd)

		t, err := g.callResolver(cmd)
		if err != nil {
			rp.NewResp(err).WriteTo(conn)
			return
		}

		args := make([]string, 0, len(ms[1:]))
		for _, argm := range ms[1:] {
			arg, err := argm.Str()
			if err != nil {
				rp.NewResp(ErrBadCmd).WriteTo(conn)
			}
			args = append(args, arg)
		}

		fullCmd := cmd + " " + strings.Join(args, " ")

		g.callHandler(t, fullCmd, conn)

	}
}
