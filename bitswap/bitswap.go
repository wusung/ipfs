package bitswap

import (
	"time"
	mh "github.com/multiformats/go-multihash"
	"../blocks"
	u "../util"
)

// aliases

type Ledger struct {
	Owner     mh.Multihash
	Partner   mh.Multihash
	BytesSent uint64
	BytesRecv uint64
	Timestamp *time.Time
}

type BitSwap struct {
	Ledgers  map[u.Key]*Ledger       // key is peer.ID
	HaveList map[u.Key]*blocks.Block // key is multihash
	WantList []*mh.Multihash
	// todo
}
