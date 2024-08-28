package inter

import "github.com/ethereum/go-ethereum/common"

type BridgeSignature [65]byte

type BridgeVote struct {
	BlockNum uint64
	ChainID  uint64
	Hash common.Hash
	Signature BridgeSignature
}
