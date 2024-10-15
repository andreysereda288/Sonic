package integration_tests

import (
	"context"
	"testing"
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
