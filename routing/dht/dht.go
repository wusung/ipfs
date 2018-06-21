package dht

import (
	"sync"
	"time"

	peer	"../../peer"
	swarm	"../../swarm"
	u		"../../util"
	identify "../../identify"

	ma "github.com/jbenet/go-multiaddr"

	ds "github.com/jbenet/datastore.go"

	"code.google.com/p/goprotobuf/proto"
	"github.com/miekg/dns"
)

// TODO. SEE https://github.com/jbenet/node-ipfs/blob/master/submodules/ipfs-dht/index.js

// IpfsDHT is an implementation of Kademlia with Coral and S/Kademlia modifications.
// It is used to implement the base IpfsRouting module.
type IpfsDHT struct {
	routes RoutingTable

	network *swarm.Swarm

	// Local peer (yourself)
	self *peer.Peer

	// Local data
	datastore ds.Datastore

	// map of channels waiting for reply messages
	listeners  map[uint64]chan *swarm.Message
	listenLock sync.RWMutex

	// Signal to shutdown dht
	shutdown chan struct{}
}

func NewDHT(p *peer.Peer) (*IpfsDHT, error) {
	dht := new(IpfsDHT)

	dht.network = swarm.NewSwarm(p)
	//TODO: should Listen return an error?
	dht.network.Listen()

	dht.datastore = ds.NewMapDatastore()
	dht.self = p
	dht.listeners = make(map[uint64]chan *swarm.Message)
	dht.shutdown = make(chan struct{})
	return dht, nil
}

func (dht *IpfsDHT) Start()  {
	go dht.handleMessages()
}

// Connect to a new peer at the given address
func (dht *IpfsDHT) Connect(addr *ma.Multiaddr) error {
	peer := new(peer.Peer)
	peer.AddAddress(addr)

	conn, err := swarm.Dial("tcp", peer)
	if err != nil {
		return err
	}

	err = identify.Handshake(dht.self, peer, conn.Incoming.MsgChan, conn.Outgoing.MsgChan)
	if err != nil {
		return err
	}

	dht.network.StartConn(conn)

	// TODO: Add this peer to our routing table
	return nil
}

// Read in all messages from swarm and handle them appropriately
// NOTE: this function is just a quick sketch
func (dht *IpfsDHT) handleMessages() {
	u.DOut("Being message handling routine")
	for {
		select {
		case mes := <-dht.network.Chan.Incoming:
			u.DOut("recieved message from swarm.")
			pmes := new(DHTMessage)
			err := proto.Unmarshal(mes.Data, pmes)
			if err != nil {
				u.PErr("Failed to decode protobuf message: %s", err)
				continue
			}

			// Note: not sure if this is the correct place for this
			if pmes.GetResponse() {
				dht.listenLock.RLock()
				ch, ok := dht.listeners[pmes.GetId()]
				dht.listenLock.RUnlock()
				if ok {
					ch <- mes
				}

				// this is expected behaviour during a timeout
				u.DOut("Received response with nobody listening...")
				continue
			}
			//

			switch pmes.GetType() {
			case DHTMessage_GET_VALUE:
				dht.handleGetValue(mes.Peer, pmes)
			case DHTMessage_PUT_VALUE:
				dht.handlePutValue(mes.Peer, pmes)
			case DHTMessage_FIND_NODE:
				dht.handleFindNode(mes.Peer, pmes)
			case DHTMessage_ADD_PROVIDER:
			case DHTMessage_GET_PROVIDERS:
			case DHTMessage_PING:
				dht.handleFindNode(mes.Peer, pmes)
			}
		case err := <-dht.network.Chan.Errors:
			panic(err)
		case <-dht.shutdown:
			return
		}
	}
}

func (dht *IpfsDHT) handleGetValue(p *peer.Peer, pmes *DHTMessage) {
	dskey := ds.NewKey(pmes.GetKey())
	i_val, err := dht.datastore.Get(dskey)
	if err == nil {
		isResponse := true
		resp := new(DHTMessage)
		resp.Response = &isResponse
		resp.Id = pmes.Id
		resp.Key = pmes.Key

		val := i_val.([]byte)
		resp.Value = val

		mes := new(swarm.Message)
		mes.Peer = p
		mes.Data = []byte(resp.String())
	} else if err == ds.ErrNotFound {
		// Find closest node(s) to desired key and reply with that info
		// TODO: this will need some other metadata in the protobuf message
		//			to signal to the querying node that the data its receiving
		//			is actually a list of other nodes
	}
}

// Store a value in this nodes local storage
func (dht *IpfsDHT) handlePutValue(p *peer.Peer, pmes *DHTMessage) {
	dskey := ds.NewKey(pmes.GetKey())
	err := dht.datastore.Put(dskey, pmes.GetValue())
	if err != nil {
		// For now, just panic, handle this better later maybe
		panic(err)
	}
}

func (dht *IpfsDHT) handlePing(p *peer.Peer, pmes *DHTMessage) {
	isResponse := true
	resp := new(DHTMessage)
	resp.Id = pmes.Id
	resp.Response = &isResponse
	resp.Type = pmes.Type

	dht.network.Chan.Outgoing <- swarm.NewMessage(p, []byte(resp.String()))
}

func (dht *IpfsDHT) handleFindNode(p * peer.Peer, pmes *DHTMessage)  {
	panic("Not implemented.")
}

func (dht *IpfsDHT) handleGetProviders(p * peer.Peer, pmes *DHTMessage)  {
	panic("Not implemented.")
}

func (dht *IpfsDHT) handleAddProvider(p * peer.Peer, pmes *DHTMessage)  {
	panic("Not implemented.")
}

// Register a handler for a specific message ID, used for getting replies
// to certain messages (i.e. response to a GET_VALUE message)
func (dht *IpfsDHT) ListenFor(mesid uint64) <-chan *swarm.Message {
	lchan := make(chan *swarm.Message)
	dht.listenLock.Lock()
	dht.listeners[mesid] = lchan
	dht.listenLock.Unlock()
	return lchan
}

func (dht *IpfsDHT) Unlisten(mesid uint64) {
	dht.listenLock.Lock()
	ch, ok := dht.listeners[mesid]
	if ok {
		delete(dht.listeners, mesid)
	}
	dht.listenLock.Unlock()
	close(ch)
}

// Stop all communications from this node and shut down
func (dht *IpfsDHT) Halt() {
	dht.shutdown <- struct{}{}
	dht.network.Close()
}

// Ping a node, log the time it took
func (dht *IpfsDHT) Ping(p *peer.Peer, timeout time.Duration)  {
	// Thoughts: maybe this should accept an ID and do a peer lookup?
	u.DOut("Enter Ping.")
	id := GenerateMessageID()
	mes_type := DHTMessage_PING
	pmes := new(DHTMessage)
	pmes.Id = &id
	pmes.Type = &mes_type

	mes := new(swarm.Message)
	mes.Peer = p
	mes.Data = []byte(pmes.String())

	before := time.Now()
	response_chan := dht.ListenFor(id)
	dht.network.Chan.Outgoing <- mes

	tout := time.After(timeout)
	select {
	case <-response_chan:
		roundtrip := time.Since(before)
		u.DOut("Ping took %s.", roundtrip.String())
	case <-tout:
		// Timed out, think about removing node from network
		u.DOut("Ping node timed out.")
	}
}
