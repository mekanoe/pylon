// Gateway test mechanism (also a redis proxy)
package main

import (
	"github.com/kayteh/pylon/cmd/pylond/run/gateway"
	"github.com/mediocregopher/radix.v2/pool"
	rp "github.com/mediocregopher/radix.v2/redis"
	"log"
	"net"
	"os"
	"strings"
)

var r *pool.Pool

func main() {
	l, err := net.Listen("tcp", "127.0.0.1:16379")
	if err != nil {
		log.Fatal(err)
	}

	r, err = pool.New("tcp", os.Args[1], 10)
	if err != nil {
		log.Fatal(err)
	}

	g := gateway.NewGateway(l)

	g.RegisterHandler(gateway.Read, sendRedis)
	g.RegisterHandler(gateway.Write, sendRedis)

	g.Listen()
}

func echo(conn net.Conn, cmd string) {
	rp.NewResp(cmd).WriteTo(conn)
}

func sendRedis(conn net.Conn, cmd string) {
	rc, err := r.Get()
	if err != nil {
		rp.NewResp(err).WriteTo(conn)
	}
	defer r.Put(rc)

	ss := strings.Split(cmd, " ")

	if len(ss) == 1 {
		rc.Cmd(cmd).WriteTo(conn)
	}

	si := make([]interface{}, 0, len(ss[1:]))
	for _, s := range ss[1:] {
		si = append(si, s)
	}

	rc.Cmd(ss[0], si...).WriteTo(conn)
}
