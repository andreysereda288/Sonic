package valkeystore

import (
	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/crypto"

	"github.com/Fantom-foundation/go-opera/inter/validatorpk"
	"github.com/Fantom-foundation/go-opera/valkeystore/encryption"
)

type SignerI interface {
	Sign(pubkey validatorpk.PubKey, digest []byte) ([]byte, error)
	SignBridge(pubkey validatorpk.PubKey, digest []byte) ([]byte, error)
}

type Signer struct {
	backend KeystoreI
}

func NewSigner(backend KeystoreI) *Signer {
	return &Signer{
		backend: backend,
	}
}

// Sign produce signature in [R || S] format.
func (s *Signer) Sign(pubkey validatorpk.PubKey, digest []byte) ([]byte, error) {
	if pubkey.Type != validatorpk.Types.Secp256k1 {
		return nil, encryption.ErrNotSupportedType
	}
	key, err := s.backend.GetUnlocked(pubkey)
	if err != nil {
		return nil, err
	}

	secp256k1Key := key.Decoded.(*ecdsa.PrivateKey)

	sigRSV, err := crypto.Sign(digest, secp256k1Key)
	if err != nil {
		return nil, err
	}
	sigRS := sigRSV[:64]
	return sigRS, err
}

// SignBridge produce signature in [R || S || V] format where V is 27 or 28.
func (s *Signer) SignBridge(pubkey validatorpk.PubKey, digest []byte) ([]byte, error) {
	if pubkey.Type != validatorpk.Types.Secp256k1 {
		return nil, encryption.ErrNotSupportedType
	}
	key, err := s.backend.GetUnlocked(pubkey)
	if err != nil {
		return nil, err
	}

	secp256k1Key := key.Decoded.(*ecdsa.PrivateKey)

	sigRSV, err := crypto.Sign(digest, secp256k1Key)
	if err != nil {
		return nil, err
	}
	sigRSV[64] += 27 // the V value must be 27 or 28 as an Ethereum convention
	return sigRSV, err
}
