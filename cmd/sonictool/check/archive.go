package check

import (
	"fmt"
	"path/filepath"

	"github.com/Fantom-foundation/Carmen/go/database/mpt"
	"github.com/Fantom-foundation/Carmen/go/database/mpt/io"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/utils/cachescale"
	"github.com/ethereum/go-ethereum/log"
)

func CheckArchiveStateDb(dataDir string, cacheRatio cachescale.Func) error {
	// compare with blocks in the gdb
	if err := checkArchiveBlockRoots(dataDir, cacheRatio); err != nil {
		return err
	}
	log.Info("The archive states hashes matches with blocks in the gdb")

	archiveDir := filepath.Join(dataDir, "carmen", "archive")
	info, err := io.CheckMptDirectoryAndGetInfo(archiveDir)
	if err != nil {
		return fmt.Errorf("failed to check archive state dir: %w", err)
	}
	if err := mpt.VerifyArchive(archiveDir, info.Config, verificationObserver{}); err != nil {
		return fmt.Errorf("archive state verification failed: %w", err)
	}
	log.Info("Verification of the archive state succeed")
	return nil
}

func checkArchiveBlockRoots(dataDir string, cacheRatio cachescale.Func) error {
	gdb, dbs, err := createGdb(dataDir, cacheRatio, true, false)
	if err != nil {
		return err
	}
	defer gdb.Close()
	defer dbs.Close()

	invalidBlocks := 0
	lastBlockIdx := gdb.GetLatestBlockIndex()
	for i := idx.Block(1); i <= lastBlockIdx; i++ {
		block := gdb.GetBlock(i)
		if block == nil {
			return fmt.Errorf("verification failed - unable to get block %d from gdb", i)
		}
		err = gdb.EvmStore().CheckArchiveStateHash(i, block.Root)
		if err != nil {
			log.Error("Block root verification failed", "block", i, "err", err)
			invalidBlocks++
		}
		if i%1000 == 0 {
			log.Info("Block root verification OK", "block", i)
		}
	}
	if invalidBlocks != 0 {
		return fmt.Errorf("block root verification failed for %d blocks (from %d total blocks)", invalidBlocks, lastBlockIdx)
	}
	log.Info("Block root verification OK for all blocks", "blocks", lastBlockIdx)
	return nil
}
