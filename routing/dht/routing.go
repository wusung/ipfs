package dht

import (
	"math/rand"
	"time"
	"bytes"
	"encoding/json"

	proto "github.com/golang/protobuf/proto"

	ma "github.com/jbenet/go-multiaddr"

	peer "../../peer"
	swarm "../../swarm"
	u "../../util"
	kb "../kbucket"
	"golang.org/x/text/collate"
	"fmt"
	"errors"
)

// Pool size is the number of nodes used for group find/set RPC calls
var PoolSize = 6

// TODO: determine a way of creating and managing message IDs
func GenerateMessageID() uint64 {
	//return (uint64(rand.Uint32()) << 32) & uint64(rand.Uint32())
	return uint64(rand.Uint32())
}

// This file implements the Routing interface for the IpfsDHT struct.

// Basic Put/Get

// PutValue adds value corresponding to given Key.
// This is the top level "Store" operation of the DHT
func (s *IpfsDHT) PutValue(key u.Key, value []byte) {
	complete := make(chan struct{})
	for i, route := range s.routes {
		p := route.NearestPeer(kb.ConvertKey(key))
		if p == nil {
			s.network.Chan.Errors <- fmt.Errorf("No peer found on level %d", i)
			continue
			go func() {
				complete <- struct{}{}
			}()
		}
		go func() {
			err := s.putValueToNetwork(p, string(key), value)
			if err != nil {
				s.network.Chan.Errors <- err
			}
			complete <- struct{}{}
		}()
	}
	for _, _ = range s.routes {
		<-complete
	}
}

// GetValue searches for the value corresponding to given Key.
// If the search does not succeed, a multiaddr string of a closer peer is
// returned along with util.ErrSearchIncomplete
func (s *IpfsDHT) GetValue(key u.Key, timeout time.Duration) ([]byte, error) {
	for _, route := range s.routes {
		var p *peer.Peer
		p = route.NearestPeer(kb.ConvertKey(key))
		if p == nil {
			return nil, errors.New("Table returned nil peer!")
		}

		b, err := s.getValueSingle(p, key, timeout)
		if err == nil {
			return b, nil
		}
		if err != u.ErrSearchIncomplete {
			return nil, err
		}
	}
	return nil, u.ErrNotFound
}

// Value provider layer of indirection.
// This is what DSHTs (Coral and MainlineDHT) do to store large values in a DHT.

// Announce that this node can provide value for given key
func (s *IpfsDHT) Provide(key u.Key) error {
	peers := s.routes[0].NearestPeers(kb.ConvertKey(key), PoolSize)
	if len(peers) == 0 {
		//return an error
	}

	pmes := DHTMessage{
		Type: PBDHTMessage_ADD_PROVIDER,
		Key:  string(key),
	}
	pbmes := pmes.ToProtobuf()

	for _, p := range peers {
		mes := swarm.NewMessage(p, pbmes)
		s.network.Chan.Outgoing <- mes
	}
	return nil
}

// FindProviders searches for peers who can provide the value for given key.
func (s *IpfsDHT) FindProviders(key u.Key, timeout time.Duration) ([]*peer.Peer, error) {
	p := s.routes[0].NearestPeer(kb.ConvertKey(key))

	pmes := DHTMessage{
		Type: PBDHTMessage_GET_PROVIDERS,
		Key:  string(key),
		Id:   GenerateMessageID(),
	}

	mes := swarm.NewMessage(p, pmes.ToProtobuf())

	listenChan := s.ListenFor(pmes.Id, 1, time.Minute)
	u.DOut("Find providers for: '%s'", key)
	s.network.Chan.Outgoing <- mes
	after := time.After(timeout)
	select {
	case <-after:
		s.Unlisten(pmes.Id)
		return nil, u.ErrTimeout
	case resp := <-listenChan:
		u.DOut("FindProviders: got response.")
		pmes_out := new(PBDHTMessage)
		err := proto.Unmarshal(resp.Data, pmes_out)
		if err != nil {
			return nil, err
		}
		var addrs map[u.Key]string
		err = json.Unmarshal(pmes_out.GetValue(), &addrs)
		if err != nil {
			return nil, err
		}

		var prov_arr []*peer.Peer
		for pid, addr := range addrs {
			p := s.network.Find(pid)
			if p == nil {
				maddr, err := ma.NewMultiaddr(addr)
				if err != nil {
					u.PErr("error connecting to new peer: %s", err)
					continue
				}
				p, err = s.Connect(maddr)
				if err != nil {
					u.PErr("error connecting to new peer: %s", err)
					continue
				}
			}
			s.addProviderEntry(key, p)
			prov_arr = append(prov_arr, p)
		}

		return prov_arr, nil
	}
}

// Find specific Peer

// FindPeer searches for a peer with given ID.
func (s *IpfsDHT) FindPeer(id peer.ID, timeout time.Duration) (*peer.Peer, error) {
	p := s.routes[0].NearestPeer(kb.ConvertPeerID(id))

	pmes := DHTMessage{
		Type: PBDHTMessage_FIND_NODE,
		Key:  string(id),
		Id:   GenerateMessageID(),
	}

	mes := swarm.NewMessage(p, pmes.ToProtobuf())

	listenChan := s.ListenFor(pmes.Id, 1, time.Minute)
	s.network.Chan.Outgoing <- mes
	after := time.After(timeout)
	select {
	case <-after:
		s.Unlisten(pmes.Id)
		return nil, u.ErrTimeout
	case resp := <-listenChan:
		pmes_out := new(PBDHTMessage)
		err := proto.Unmarshal(resp.Data, pmes_out)
		if err != nil {
			return nil, err
		}
		addr := string(pmes_out.GetValue())
		maddr, err := ma.NewMultiaddr(addr)
		if err != nil {
			return nil, err
		}

		found_peer, err := s.Connect(maddr)
		if err != nil {
			u.POut("Found peer but couldnt connect.")
			return nil, err
		}

		if !found_peer.ID.Equal(id) {
			u.POut("FindPeer: searching for '%s' but found '%s'", id.Pretty(), found_peer.ID.Pretty())
			return found_peer, u.ErrSearchIncomplete
		}

		return found_peer, nil
	}
}

// Ping a peer, log the time it took
func (dht *IpfsDHT) Ping(p *peer.Peer, timeout time.Duration) error {
	// Thoughts: maybe this should accept an ID and do a peer lookup?
	u.DOut("Enter Ping.")

	pmes := DHTMessage{Id: GenerateMessageID(), Type: PBDHTMessage_PING}
	mes := swarm.NewMessage(p, pmes.ToProtobuf())

	before := time.Now()
	response_chan := dht.ListenFor(pmes.Id, 1, time.Minute)
	dht.network.Chan.Outgoing <- mes

	tout := time.After(timeout)
	select {
	case <-response_chan:
		roundtrip := time.Since(before)
		p.SetLatency(roundtrip)
		u.POut("Ping took %s.", roundtrip.String())
		return nil
	case <-tout:
		// Timed out, think about removing peer from network
		u.DOut("Ping peer timed out.")
		dht.Unlisten(pmes.Id)
		return u.ErrTimeout
	}
}

func (dht *IpfsDHT) GetDiagnostic(timeout time.Duration) ([]*diagInfo, error) {
	u.DOut("Begin Diagnostic")
	//Send to N closest peers
	targets := dht.routes[0].NearestPeers(kb.ConvertPeerID(dht.self.ID), 10)

	// TODO: Add timeout to this struct so nodes know when to return
	pmes := DHTMessage{
		Type: PBDHTMessage_DIAGNOSTIC,
		Id:   GenerateMessageID(),
	}

	listenChan := dht.ListenFor(pmes.Id, len(targets), time.Minute*2)

	pbmes := pmes.ToProtobuf()
	for _, p := range targets {
		mes := swarm.NewMessage(p, pbmes)
		dht.network.Chan.Outgoing <- mes
	}

	var out []*diagInfo
	after := time.After(timeout)
	for count := len(targets); count > 0; {
		select {
		case <-after:
			u.DOut("Diagnostic request timed out.")
			return out, u.ErrTimeout
		case resp := <-listenChan:
			pmes_out := new(PBDHTMessage)
			err := proto.Unmarshal(resp.Data, pmes_out)
			if err != nil {
				// NOTE: here and elsewhere, need to audit error handling,
				//		some errors should be continued on from
				return out, err
			}

			dec := json.NewDecoder(bytes.NewBuffer(pmes_out.GetValue()))
			for {
				di := new(diagInfo)
				err := dec.Decode(di)
				if err != nil {
					break
				}

				out = append(out, di)
			}
		}
	}

	return nil, nil
}
