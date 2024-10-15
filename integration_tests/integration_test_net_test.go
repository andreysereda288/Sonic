package integration_tests

import (
	"context"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/holiman/uint256"
)

func TestIntegrationTestNet_CanStartAndStopIntegrationTestNet(t *testing.T) {
	dataDir := t.TempDir()
	net, err := StartIntegrationTestNet(dataDir)
	if err != nil {
		t.Fatalf("Failed to start the fake network: %v", err)
	}
	net.stop()
}

func TestIntegrationTestNet_CanStartAndStopMultipleConsecutiveIntegrationTestNetInstances(t *testing.T) {
	for i := 0; i < 2; i++ {
		dataDir := t.TempDir()
		net, err := StartIntegrationTestNet(dataDir)
		if err != nil {
			t.Fatalf("Failed to start the fake network: %v", err)
		}
		net.stop()
	}
}

func TestIntegrationTestNet_CanFetchInformationFromTheNetwork(t *testing.T) {
	dataDir := t.TempDir()
	net, err := StartIntegrationTestNet(dataDir)
	if err != nil {
		t.Fatalf("Failed to start the fake network: %v", err)
	}
	defer net.stop()

	client, err := net.getClient()
	if err != nil {
		t.Fatalf("Failed to connect to the integration test network: %v", err)
	}

	block, err := client.BlockNumber(context.Background())
	if err != nil {
		t.Fatalf("Failed to get block number: %v", err)
	}

	if block == 0 || block > 1000 {
		t.Errorf("Unexpected block number: %v", block)
	}
}

func TestIntegrationTestNet_CanEndowAccountsWithTokens(t *testing.T) {
	dataDir := t.TempDir()
	net, err := StartIntegrationTestNet(dataDir)
	if err != nil {
		t.Fatalf("Failed to start the fake network: %v", err)
	}
	defer net.stop()

	client, err := net.getClient()
	if err != nil {
		t.Fatalf("Failed to connect to the integration test network: %v", err)
	}

	address := common.Address{0x01}
	balance, err := client.BalanceAt(context.Background(), address, nil)
	if err != nil {
		t.Fatalf("Failed to get balance for account: %v", err)
	}

	for i := 0; i < 10; i++ {
		increment := uint256.NewInt(1000)
		if err := net.endowAccount(address, increment); err != nil {
			t.Fatalf("Failed to endow account 1: %v", err)
		}
		want := balance.Add(balance, increment.ToBig())
		balance, err = client.BalanceAt(context.Background(), address, nil)
		if err != nil {
			t.Fatalf("Failed to get balance for account: %v", err)
		}
		if want, got := want, balance; want.Cmp(got) != 0 {
			t.Fatalf("Unexpected balance for account, got %v, wanted %v", got, want)
		}
		balance = want
	}
}
