package blocks

import (
	"github.com/jbenet/go-ipfs/bitswap"
	"github.com/jbenet/go-ipfs/storage"
)

type BlockService struct {
	Local *storage.Storage
	Remote *bitswap.BitSwap
}
