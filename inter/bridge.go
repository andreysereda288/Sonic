package inter

import "github.com/ethereum/go-ethereum/common"

type BridgeSignature [65]byte

type BridgeVote struct {
	Hash common.Hash
	Signature BridgeSignature
}
