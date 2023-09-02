package scan

import (
	"context"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/cordtus/penpal/internal/alert"
	"github.com/cordtus/penpal/internal/config"
	"github.com/cordtus/penpal/internal/rpc"
)

func Monitor(cfg config.Config) {
	alertChan := make(chan alert.Alert)
	exit := make(chan bool)
	client := &http.Client{
		Timeout: time.Second * 5,
	}
	latestBlock := fetchLatestBlock(cfg.Network.Rpcs[0], client) // Fetch the latest block data once

	for _, validator := range cfg.Validators {
		go scanValidator(validator, cfg.Network, alertChan, client, latestBlock)
	}
	if cfg.Health.Interval != 0 {
		go healthServer(cfg.Health.Port)
		go healthCheck(cfg.Health, alertChan, client, latestBlock)
	}

	<-exit
}

func fetchLatestBlock(url string, client *http.Client) rpc.Block {
	block, err := rpc.GetLatestBlock(url, client)
	if err != nil {
		log.Printf("Error fetching latest block: %v", err)
		return rpc.Block{}
	}
	return block
}

func scanValidator(validator config.Validator, network config.Network, alertChan chan<- alert.Alert, client *http.Client, latestBlock rpc.Block) {
	var (
		interval int
		alerted  bool
	)

	for {
		checkNetwork(validator, network, client, &alerted, alertChan, latestBlock)

		if alerted && network.Interval > 2 {
			interval = 2
		} else {
			interval = network.Interval
		}

		// Sleep for the specified interval
		time.Sleep(time.Duration(interval) * time.Minute)
	}
}

func checkNetwork(validator config.Validator, network config.Network, client *http.Client, alerted *bool, alertChan chan<- alert.Alert, latestBlock rpc.Block) {
	// Use the latestBlock data here for network checks
	chainID := latestBlock.Result.Block.Header.ChainID
	blockTime := latestBlock.Result.Block.Header.Time

	// The rest of your checkNetwork function remains the same
	// ...

	heightInt, _ := strconv.Atoi(latestBlock.Result.Block.Header.Height)
	alertChan <- backCheck(validator, network, heightInt, alerted, network.Rpcs[0], client)
	// ...
}

// ...
