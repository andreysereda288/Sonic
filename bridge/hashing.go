package bridge

import (
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"math/big"
)

var (
	uint256Type, _    = abi.NewType("uint256", "", nil)
	bytes32Type, _    = abi.NewType("bytes32", "", nil)
	validatorsType, _ = abi.NewType("tuple[]", "", []abi.ArgumentMarshaling{
		{Name: "addr", Type: "address"},
		{Name: "weight", Type: "uint256"},
	})
	updateHashAbi = abi.Arguments{
		{"blockNum", uint256Type, false},
		{"stateRoot", bytes32Type, false},
		{"chainId", uint256Type, false},
		{"newValidators", validatorsType, false},
	}
)

type RegistryValidator struct {
	Addr   common.Address
	Weight *big.Int
}

func GetUpdateHash(blockNumber *big.Int, stateRoot common.Hash, chainId *big.Int, newValidators []RegistryValidator) (common.Hash, error) {
	// abi encode
	abiEncoded, err := updateHashAbi.Pack(blockNumber, stateRoot, chainId, newValidators)
	if err != nil {
		return common.Hash{}, err
	}
	// hash
	return crypto.Keccak256Hash(abiEncoded), nil
}
