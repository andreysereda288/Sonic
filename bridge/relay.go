package bridge

import (
	"context"
	"fmt"
	"github.com/Fantom-foundation/go-opera/evmcore"
	"github.com/Fantom-foundation/go-opera/gossip/emitter"
	"github.com/Fantom-foundation/go-opera/inter"
	"github.com/Fantom-foundation/go-opera/logger"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/event"
	"math/big"
	"time"
)

type Relay struct {
	ethClient *ethclient.Client
	ethChainId *big.Int
	localChainId *big.Int
	localNewBlockChan chan evmcore.ChainHeadNotify
	localNewBlockSub event.Subscription
	ethereumNewBlockChan chan *types.Header
	ethereumSub     ethereum.Subscription
	stopRunningLoop context.CancelFunc
	emitter *emitter.Emitter
	logger logger.Periodic
}

func MakeRelay(ethUrl string, ethChainId *big.Int, localChainId *big.Int) (*Relay, error) {
	ethClient, err := ethclient.Dial(ethUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Ethereum node: %w", err)
	}

	defer func() {
		if err != nil {
			ethClient.Close()
		}
	}()

	actualChainId, err := ethClient.ChainID(context.Background())
	if err != nil {
		return nil, err
	}
	if actualChainId.Cmp(ethChainId) != 0 {
		return nil, fmt.Errorf("unexpected chainId of the connected Ethereum node: %s != %s", actualChainId, ethChainId)
	}

	return &Relay{
		ethClient: ethClient,
		ethChainId: ethChainId,
		localChainId: localChainId,
		logger: logger.Periodic{Instance: logger.New()},
	}, nil
}

func (r *Relay) SetNewBlockChan(ch chan evmcore.ChainHeadNotify, sub event.Subscription) {
	r.localNewBlockChan = ch
	r.localNewBlockSub = sub
}

func (r *Relay) SetEmitter(em *emitter.Emitter) {
	r.emitter = em
}

// GetBridgeVotes to be called from emitter, when a new event is being emitted
func (r *Relay) GetBridgeVotes() ([]inter.BridgeVote, error) {
	fmt.Printf("Trying to get new bridge votes\n")
	return nil, nil
}

func (r *Relay) Start() error {
	fmt.Printf("Starting Relay\n")
	var ctx context.Context
	ctx, r.stopRunningLoop = context.WithCancel(context.Background())

	var err error
	r.ethereumNewBlockChan = make(chan *types.Header)
	r.ethereumSub, err = r.ethClient.SubscribeNewHead(ctx, r.ethereumNewBlockChan)
	if err != nil {
		return err
	}

	go r.run(ctx)
	return nil
}

func (r *Relay) Stop() {
	fmt.Printf("Closing bridge relay\n")
	r.stopRunningLoop()
	r.localNewBlockSub.Unsubscribe()
	r.ethereumSub.Unsubscribe()
}

func (r *Relay) run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case header := <-r.ethereumNewBlockChan:
			fmt.Printf("new ethereum header: %s\n", header.Number)
			hash, err := GetUpdateHash(header.Number, header.Root, r.ethChainId, nil)
			if err != nil {
				r.logger.Error(time.Second,"Failed to get update hash", "err", err)
				continue
			}
			r.emitter.SetBridgeVote(inter.BridgeVote{
				BlockNum:  header.Number.Uint64(),
				ChainID:   r.ethChainId.Uint64(),
				Hash:      hash,
			})
		case header := <-r.localNewBlockChan:
			fmt.Printf("new local header: %s\n", header.Block.Number)
			hash, err := GetUpdateHash(header.Block.Number, header.Block.Root, r.ethChainId, nil)
			if err != nil {
				r.logger.Error(time.Second,"Failed to get update hash", "err", err)
				continue
			}
			r.emitter.SetBridgeVote(inter.BridgeVote{
				BlockNum:  header.Block.Number.Uint64(),
				ChainID:   r.localChainId.Uint64(),
				Hash:      hash,
			})
		case err := <-r.ethereumSub.Err():
			fmt.Printf("ethereum subscription err: %s\n", err)
		case err := <-r.localNewBlockSub.Err():
			fmt.Printf("local subscription err: %s\n", err)
		}
	}
}

func (r *Relay) OnBridgeVotesReceive(bridgeVotes []inter.BridgeVote) {
	for _, bridgeVote := range bridgeVotes {
		fmt.Printf("received bridge vote: %v\n", bridgeVote)
	}
}
