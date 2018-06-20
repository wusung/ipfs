package swarm

import (
	"fmt"
	"../peer"
	u "../util"
	"github.com/jbenet/go-msgio"
	ma "github.com/jbenet/go-multiaddr"
	"net"
)

const ChanBuffer = 10

type Conn struct {
	Peer *peer.Peer
	Addr *ma.Multiaddr
	Conn net.Conn

	Closed   chan bool
	Outgoing *msgio.Chan
	Incoming *msgio.Chan
}

type ConnMap map[u.Key]*Conn

// ConnMap maps Keys (Peer.IDs) to Connections.

func Dial(network string, peer *peer.Peer) (*Conn, error) {
	addr := peer.NetAddress(network)
	if addr == nil {
		return nil, fmt.Errorf("No address for network %s", network)
	}

	network, host, err := addr.DialArgs()
	if err != nil {
		return nil, err
	}

	nconn, err := net.Dial(network, host)
	if err != nil {
		return nil, err
	}

	conn := &Conn{
		Peer: peer,
		Addr: addr,
		Conn: nconn,
	}

	newConnChans(conn)
	return conn, nil
}

// Construct new channels for given Conn.
func newConnChans(c *Conn) error {
	if c.Outgoing != nil || c.Incoming != nil {
		return fmt.Errorf("Conn already initialized")
	}

	c.Outgoing = msgio.NewChan(10)
	c.Incoming = msgio.NewChan(10)
	c.Closed = make(chan bool, 1)

	go c.Outgoing.WriteTo(c.Conn)
	go c.Incoming.ReadFrom(c.Conn, 1<<12)

	return nil
}

func (s *Conn) Close() error {
	if s.Conn == nil {
		return fmt.Errorf("Already closed") // already closed
	}

	// closing net connection
	err := s.Conn.Close()
	s.Conn = nil
	// closing channels
	s.Incoming.Close()
	s.Outgoing.Close()
	s.Closed <- true
	return err
}
