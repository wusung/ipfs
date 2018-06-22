package core

import (
	"fmt"
	ds "github.com/ipfs/go-datastore"
	"../blocks"
	"../config"
	"../merkledag"
	path "../path"
	"../peer"
)

// IPFS Core module. It represents an IPFS instance.

type IpfsNode struct {
	// the node's configuration
	Config *config.Config

	// the local node's identity
	Identity *peer.Peer

	// the map of other nodes (Peer instances)
	PeerMap *peer.Map

	// the local datastore
	Datastore ds.Datastore

	// the network message stream
	// Network *netmux.Netux

	// the routing system. recommend ipfs-dht
	// Routing *routing.Routing

	// the block exchange + strategy (bitswap)
	// BitSwap *bitswap.BitSwap

	// the block service, get/add blocks.
	Blocks *blocks.BlockService

	// the merkle dag service, get/add objects.
	DAG *merkledag.DAGService

	// the path resolution system
	Resolver *path.Resolver

	// the name system, resolves paths to hashes
	// Namesys *namesys.Namesys
}

func NewIpfsNode(cfg *config.Config) (*IpfsNode, error) {
	if cfg == nil {
		return nil, fmt.Errorf("configuration required.")
	}

	d, err := makeDatastore(cfg.Datastore)
	if err != nil {
		return nil, err
	}

	bs, err := blocks.NewBlockService(d)
	if err != nil {
		return nil, err
	}

	dag := &merkledag.DAGService{Blocks: bs}

	n := &IpfsNode{
		Config:    cfg,
		PeerMap:   &peer.Map{},
		Datastore: d,
		Blocks:    bs,
		DAG:       dag,
		Resolver:  &path.Resolver{DAG: dag},
	}

	return n, nil
}
