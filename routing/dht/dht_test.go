package dht

import (
	"testing"
	peer "../../peer"
	ma "github.com/multiformats/go-multiaddr"
	u "../../util"

	"fmt"

	"time"
)

var _ = fmt.Println

func TestPing(t *testing.T)  {
	u.Debug = false
	addr_a, err := ma.NewMultiaddr("/ip4/127.0.0.1/tcp/1234")
	if err != nil {
		t.Fatal(err)
	}
	addr_b,err := ma.NewMultiaddr("/ip4/127.0.0.1/tcp/5678")
	if err != nil {
		t.Fatal(err)
	}

	peer_a := new(peer.Peer)
	peer_a.AddAddress(addr_a)
	peer_a.ID = peer.ID([]byte("peer_a"))

	peer_b := new(peer.Peer)
	peer_b.AddAddress(addr_b)
	peer_b.ID = peer.ID([]byte("peer_b"))

	dht_a, err := NewDHT(peer_a)
	if err != nil {
		t.Fatal(err)
	}

	dht_b, err := NewDHT(peer_b)
	if err != nil {
		t.Fatal(err)
	}

	dht_a.Start()
	dht_b.Start()

	_, err = dht_a.Connect(addr_b)
	if err != nil {
		t.Fatal(err)
	}

	//Test that we can ping the node
	err = dht_a.Ping(peer_b, time.Second * 2)
	if err != nil {
		t.Fatal(err)
	}

	dht_a.Halt()
	dht_b.Halt()
}

func TestValueGetSet(t *testing.T) {
	u.Debug = true
	addr_a, err := ma.NewMultiaddr("/ip4/127.0.0.1/tcp/1235")
	if err != nil {
		t.Fatal(err)
	}
	addr_b, err := ma.NewMultiaddr("/ip4/127.0.0.1/tcp/5678")
	if err != nil {
		t.Fatal(err)
	}

	peer_a := new(peer.Peer)
	peer_a.AddAddress(addr_a)
	peer_a.ID = peer.ID([]byte("peer_a"))

	peer_b := new(peer.Peer)
	peer_b.AddAddress(addr_b)
	peer_b.ID = peer.ID([]byte("peer_b"))

	dht_a, err := NewDHT(peer_a)
	if err != nil {
		t.Fatal(err)
	}

	dht_b, err := NewDHT(peer_b)
	if err != nil {
		t.Fatal(err)
	}

	dht_a.Start()
	dht_b.Start()

	_, err = dht_a.Connect(addr_b)
	if err != nil {
		t.Fatal(err)
	}

	err = dht_a.PutValue("hello", []byte("world"))
	if err != nil {
		t.Fatal(err)
	}

	val, err := dht_a.GetValue("hello", time.Second * 2)
	if err != nil {
		t.Fatal(err)
	}

	if string(val) != "world" {
		t.Fatal("Expected 'world' get %s", string(val))
	}
}
