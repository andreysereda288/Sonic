package integration_tests

import (
	"context"
	"fmt"
	"os"
	"syscall"
	"time"

	sonicd "github.com/Fantom-foundation/go-opera/cmd/sonicd/cmd"
	sonictool "github.com/Fantom-foundation/go-opera/cmd/sonictool/cmd"
	"github.com/ethereum/go-ethereum/ethclient"
)

type IntegrationTestNet struct {
	done <-chan struct{}
}

func (n *IntegrationTestNet) stop() {
	syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	<-n.done
}

func (n *IntegrationTestNet) getClient() (*ethclient.Client, error) {
	return ethclient.Dial("http://localhost:18545")
}

// StartIntegrationTestNet starts a single-node test network for integration tests.
// The node serving the network is started in the same process as the caller. This
// is intended to facilitate debugging of client code in the context of a running
// node.
func StartIntegrationTestNet(directory string) (*IntegrationTestNet, error) {
	done := make(chan struct{})
	go func() {
		defer close(done)

		originalArgs := os.Args
		defer func() { os.Args = originalArgs }()

		// initialize the data directory for the single node on the test network
		// equivalent to running `sonictool --datadir <dataDir> genesis fake 1`
		os.Args = []string{"sonictool", "--datadir", directory, "genesis", "fake", "1"}
		sonictool.RunSonicTool()

		// start the fakenet sonic node
		// equivalent to running `sonicd ...` but in this local process
		os.Args = []string{
			"sonicd",
			"--datadir", directory,
			"--fakenet", "1/1",
			"--http", "--http.addr", "0.0.0.0", "--http.port", "18545",
			"--http.api", "admin,eth,web3,net,txpool,ftm,trace,debug",
			"--ws", "--ws.addr", "0.0.0.0", "--ws.port", "18546", "--ws.api", "admin,eth,ftm",
			"--pprof", "--pprof.addr", "0.0.0.0",
		}
		sonicd.RunSonicd()
	}()

	result := &IntegrationTestNet{done}

	// connect to blockchain network
	client, err := result.getClient()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to the Ethereum client: %w", err)
	}
	defer client.Close()

	// wait for the node to be ready to serve requests
	for i := 0; i < 30; i++ {
		id, err := client.ChainID(context.Background())
		if err != nil {
			time.Sleep(1 * time.Second)
			continue
		}
		fmt.Printf("Managed to get the chain ID: %d\n", id)
		return result, nil
	}

	return nil, fmt.Errorf("failed to connect to successfully start up a test network")
}
