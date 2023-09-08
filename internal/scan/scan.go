package scan

import (
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
	latestBlock, err := rpc.GetLatestBlock(cfg.Network[0].Rpcs[0], client)
	if err != nil {
		return
	}

	for _, validator := range cfg.Validators {
		go scanValidator(validator, cfg.Network[0], alertChan, client, latestBlock)
	}
	if cfg.Health.Interval != 0 {
		go healthServer(cfg.Health.Port)
		go healthCheck(cfg.Health, alertChan, client)
	}

	<-exit
}

func scanValidator(validator config.Validators, network config.Network, alertChan chan<- alert.Alert, client *http.Client, latestBlock rpc.Block) {
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

func checkNetwork(validator config.Validators, network config.Network, client *http.Client, alerted *bool, alertChan chan<- alert.Alert, latestBlock rpc.Block) {
	chainID := latestBlock.Result.Block.Header.ChainID
	LatestBlockTime := latestBlock.Result.Block.Header.Time

	heightInt, _ := strconv.Atoi(latestBlock.Result.Block.Header.Height)
	alertChan <- backCheck(validator, network, heightInt, alerted, network.Rpcs[0], client, chainID, LatestBlockTime)
}

func backCheck(validator config.Validators, network config.Network, height int, alerted *bool, url string, client *http.Client, chainID string, LatestBlockTime time.Time, args ...interface{}) alert.Alert {
	var (
		signed    int
		rpcErrors int
	)

	var clearedSignedArgs []interface{}
	if len(args) >= 2 {
		clearedSignedArgs = args[:2]
	}

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
		if len(clearedSignedArgs) >= 2 {
			*alerted = true
			return alert.Missed((network.BackCheck - signed), network.BackCheck, validator.Moniker)
		} else if *alerted {
			*alerted = false
			return alert.Cleared(signed, network.BackCheck, validator.Moniker)
		} else {
			return alert.Nil("found " + strconv.Itoa(signed) + " of " + strconv.Itoa(network.BackCheck) + " signed on " + validator.Moniker)
		}
	} else {
		if len(clearedSignedArgs) >= 2 {
			return alert.Signed(signed, network.BackCheck, validator.Moniker)
		} else {
			return alert.Nil("found " + strconv.Itoa(signed) + " of " + strconv.Itoa(network.BackCheck) + " signed on " + validator.Moniker)
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
