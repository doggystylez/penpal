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
	LatestBlockTime := latestBlock.Result.Block.Header.Time

	heightInt, _ := strconv.Atoi(latestBlock.Result.Block.Header.Height)
	alertChan <- backCheck(validator, network, heightInt, alerted, network.Rpcs[0], client, chainID, LatestBlockTime)
}

func backCheck(validator config.Validator, network config.Network, height int, alerted *bool, url string, client *http.Client, chainID string, LatestBlockTime time.Time) alert.Alert {
	var (
		signed    int
		rpcErrors int
	)
	for checkHeight := height - network.BackCheck + 1; checkHeight <= height; checkHeight++ {
		block, err := rpc.GetBlockFromHeight(strconv.Itoa(checkHeight), url, client)
		if err != nil || block.Error != nil {
			rpcErrors++
			network.BackCheck--
			continue
		}
		if checkSig(validator.Address, block) {
			signed++
		}
	}
	if rpcErrors > network.BackCheck || network.BackCheck == 0 && network.RpcAlert {
		if !*alerted {
			*alerted = true
			return alert.RpcDown(url)
		} else {
			return alert.Nil("repeat alert suppressed - RpcDown on " + network.ChainId)
		}
	} else if !network.Reverse {
		if network.BackCheck-signed > network.AlertThreshold {
			*alerted = true
			return alert.Missed(validator.Name, (network.BackCheck - signed), network.BackCheck, validator.Name)
		} else if *alerted {
			*alerted = false
			return alert.Cleared(signed, network.BackCheck, validator.Name)
		} else {
			return alert.Nil("found " + strconv.Itoa(signed) + " of " + strconv.Itoa(network.BackCheck) + " signed on " + validator.Name)
		}
	} else {
		if signed > 1 {
			return alert.Signed(signed, network.BackCheck, validator.Name)
		} else {
			return alert.Nil("found " + strconv.Itoa(signed) + " of " + strconv.Itoa(network.BackCheck) + " signed on " + validator.Name)
		}
	}
}

func checkSig(address string, block rpc.Block) bool {
	for _, sig := range block.Result.Block.LastCommit.Signatures {
		if sig.ValidatorAddress == address {
			return true
		}
	}
	return false
}
