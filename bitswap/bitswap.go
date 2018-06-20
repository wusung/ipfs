package bitswap

import (
	peer "../peer"
	"github.com/jbenet/go-multihash"
)

// aliases

type Ledger struct {
	// todo
}

type BitSwap struct {
	Ledgers map[peer.ID]*Ledger
	HaveList map[multihash.Multihash]*block.Block
	WantList []*multihash.Multihash
}
