package gateway

import (
	"github.com/mediocregopher/radix.v2/pool"
	"github.com/mediocregopher/radix.v2/redis"
	"net"
	"testing"
)

func TestHandlerError(t *testing.T) {

	// setup gateway
	l, err := net.Listen("tcp", "127.0.0.1:25010")
	if err != nil {
		t.Error(err)
		t.FailNow()
		return
	}

	g := NewGateway(l)
	go g.Listen()
	defer g.Close()

	// connect to gateway
	p, err := pool.New("tcp", "127.0.0.1:25010", 10)
	if err != nil {
		t.Error(err)
		t.FailNow()
		return
	}

	conn, err := p.Get()
	if err != nil {
		t.Error(err)
		t.FailNow()
		return
	}
	defer p.Put(conn)

	r := p.Cmd("GET", "somekey")
	if r.Err == nil || r.Err.Error() != ErrHandlerUnset.Error() {
		t.Errorf("handler was set or wrong error: %s", r.Err)
		t.FailNow()
	}

}

func TestBlacklistResolver(t *testing.T) {
	// setup gateway
	l, err := net.Listen("tcp", "127.0.0.1:25011")
	if err != nil {
		t.Error(err)
		t.FailNow()
		return
	}

	g := NewGateway(l)
	go g.Listen()
	defer g.Close()

	// connect to gateway
	p, err := pool.New("tcp", "127.0.0.1:25011", 10)
	if err != nil {
		t.Error(err)
		t.FailNow()
		return
	}

	conn, err := p.Get()
	if err != nil {
		t.Error(err)
		t.FailNow()
		return
	}
	defer p.Put(conn)

	r := p.Cmd("SCRIPT")
	if r.Err == nil || r.Err.Error() != ErrBlacklisted.Error() {
		t.Errorf("blacklist didn't resolve for SCRIPT: %s", r.Err)
		t.FailNow()
	}
}

func TestHandlers(t *testing.T) {
	// setup gateway
	l, err := net.Listen("tcp", "127.0.0.1:25012")
	if err != nil {
		t.Error(err)
		t.FailNow()
		return
	}

	g := NewGateway(l)

	g.RegisterHandler(Read, func(tc net.Conn, _ string) {
		redis.NewResp("OK").WriteTo(tc)
	})

	g.RegisterHandler(Write, func(tc net.Conn, r string) {
		redis.NewResp(r).WriteTo(tc)
	})

	go g.Listen()
	defer g.Close()

	// connect to gateway
	p, err := pool.New("tcp", "127.0.0.1:25012", 10)
	if err != nil {
		t.Error(err)
		t.FailNow()
		return
	}

	conn, err := p.Get()
	if err != nil {
		t.Error(err)
		t.FailNow()
		return
	}
	defer p.Put(conn)

	r, err := p.Cmd("GET", "hellowlrd").Str()
	if err != nil {
		t.Error(err)
		t.FailNow()
		return
	}

	if r != "OK" {
		t.Errorf("expected `OK` got `%s`", r)
		t.FailNow()
		return
	}

	r2 := p.Cmd("SET", "keyname", "keyval")
	if r2.Err != nil {
		t.Error(r2.Err)
		t.FailNow()
		return
	}

	r2s, _ := r2.Str()

	if r2s != "SET keyname keyval" {
		t.Errorf("expected `SET keyname keyval` got `%s`", r2s)
		t.FailNow()
		return
	}

}
