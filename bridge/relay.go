package bridge

import (
	"context"
	"fmt"
	"github.com/Fantom-foundation/go-opera/evmcore"
	"github.com/Fantom-foundation/go-opera/inter"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/event"
	"math/big"
)

type Relay struct {
	ethClient *ethclient.Client
	localNewBlockChan chan evmcore.ChainHeadNotify
	localNewBlockSub event.Subscription
	ethereumNewBlockChan chan *types.Header
	ethereumSub     ethereum.Subscription
	stopRunningLoop context.CancelFunc
}

func MakeRelay(ethUrl string, ethChainId *big.Int) (*Relay, error) {
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
	}, nil
}

func (r *Relay) SetNewBlockChan(ch chan evmcore.ChainHeadNotify, sub event.Subscription) {
	r.localNewBlockChan = ch
	r.localNewBlockSub = sub
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
		case header := <-r.localNewBlockChan:
			fmt.Printf("new local header: %s\n", header.Block.Number)
		case err := <-r.ethereumSub.Err():
			fmt.Printf("ethereum subscription err: %s\n", err)
		case err := <-r.localNewBlockSub.Err():
			fmt.Printf("local subscription err: %s\n", err)
		}
	}
}
