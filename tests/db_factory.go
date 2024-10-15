package tests

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/state/snapshot"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/triedb"
	"github.com/ethereum/go-ethereum/triedb/hashdb"
	"github.com/ethereum/go-ethereum/triedb/pathdb"
	"github.com/holiman/uint256"
)

// TestStateDB allows for switching database implementation for running tests.
// It is an extension ov vm.StateDB with  additional methods that were originally
// available only at an implementation level of the geth database.
type TestStateDB interface {
	vm.StateDB

	// Logs returns the logs of the current transaction.
	Logs() []*types.Log

	// SetBalance sets the balance of the given account.
	SetBalance(addr common.Address, amount *uint256.Int, reason tracing.BalanceChangeReason)

	// IntermediateRoot returns current state root hash.
	IntermediateRoot(deleteEmptyObjects bool) common.Hash

	// Commit commits the state to the underlying trie database and returns state root hash.
	Commit(block uint64, deleteEmptyObjects bool) (common.Hash, error)
}

// TestComponentsFactory is an interface for creating test components
// such as database.
type TestComponentsFactory interface {

	// NewTestStateDB creates a new TestStateDB and applies the given pre-state.
	NewTestStateDB(makePreState func(db TestStateDB)) StateTestState
}

// GethFactory is a factory for creating geth database.
type GethFactory struct {
	snapshotter bool
	scheme      string
}

// NewGethFactory creates a new GethFactory.
func NewGethFactory(snapshotter bool, scheme string) GethFactory {
	return GethFactory{snapshotter, scheme}

}

func (f GethFactory) NewTestStateDB(makePreState func(db TestStateDB)) StateTestState {
	tconf := &triedb.Config{Preimages: true}
	if f.scheme == rawdb.HashScheme {
		tconf.HashDB = hashdb.Defaults
	} else {
		tconf.PathDB = pathdb.Defaults
	}

	db := rawdb.NewMemoryDatabase()
	triedb := triedb.NewDatabase(db, tconf)
	sdb := state.NewDatabaseWithNodeDB(db, triedb)
	statedb, _ := state.New(types.EmptyRootHash, sdb, nil)

	makePreState(statedb)

	// Commit and re-open to start with a clean state.
	root, _ := statedb.Commit(0, false)

	// If snapshot is requested, initialize the snapshotter and use it in state.
	var snaps *snapshot.Tree
	if f.snapshotter {
		snapconfig := snapshot.Config{
			CacheSize:  1,
			Recovery:   false,
			NoBuild:    false,
			AsyncBuild: false,
		}
		snaps, _ = snapshot.New(snapconfig, db, triedb, root)
	}
	statedb, _ = state.New(root, sdb, snaps)

	return StateTestState{statedb, triedb, snaps}
}
