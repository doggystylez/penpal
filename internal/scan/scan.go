package scan

import (
	"net/http"
	"strconv"
	"time"

	"github.com/cordtus/penpal/internal/alert"
	"github.com/cordtus/penpal/internal/config"
	"github.com/cordtus/penpal/internal/rpc"
)

func Monitor(cfg config.Config, latestBlock rpc.Block) {
	alertChan := make(chan alert.Alert)
	exit := make(chan bool)
	client := &http.Client{
		Timeout: time.Second * 5,
	}

	for _, validator := range cfg.Validators {
		go scanValidator(validator, cfg.Network, alertChan, client, latestBlock)
	}
	if cfg.Health.Interval != 0 {
		go healthServer(cfg.Health.Port)
		go healthCheck(cfg.Health, alertChan, client)
	}

	<-exit
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

		time.Sleep(time.Duration(interval) * time.Minute)
	}
}

func checkNetwork(validator config.Validator, network config.Network, client *http.Client, alerted *bool, alertChan chan<- alert.Alert, latestBlock rpc.Block) {
	chainID := latestBlock.Result.Block.Header.ChainID
	blockTime := latestBlock.Result.Block.Header.Time

	heightInt, _ := strconv.Atoi(latestBlock.Result.Block.Header.Height)
	alertChan <- backCheck(validator, network, heightInt, alerted, network.Rpcs[0], client)
}
