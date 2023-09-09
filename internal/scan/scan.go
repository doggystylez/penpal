package scan

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/cordtus/penpal/internal/alert"
	"github.com/cordtus/penpal/internal/rpc"
	"github.com/cordtus/penpal/internal/settings"
)

func Monitor(cfg settings.Config) {
	alertChan := make(chan alert.Alert)
	exit := make(chan bool)
	client := &http.Client{
		Timeout: time.Second * 5,
	}

	network := cfg.Network[0]
	for _, validator := range cfg.Validators {
		go scanValidator(network, client, validator, alertChan)
	}

	alert.Watch(alertChan, settings.Config{Notifiers: cfg.Notifiers}, client)
	<-exit
}

func scanValidator(network settings.Network, client *http.Client, validator settings.Validators, alertChan chan<- alert.Alert) {
	alerted := new(bool) // Initialize the alerted variable
	for {
		block, err := rpc.GetLatestBlock(network.Rpcs[0], client)

		if err != nil {
			log.Println("Failed to fetch the latest block data:", err)
			// Continue the loop without returning to keep retrying.
		}

		checkValidator(network, block, validator, alertChan, alerted) // Pass the alerted variable
		if network.Interval > 2 {
			time.Sleep(time.Minute * 2)
		} else {
			time.Sleep(time.Minute * time.Duration(network.Interval))
		}
	}
}

func checkValidator(network settings.Network, block rpc.Block, validator settings.Validators, alertChan chan<- alert.Alert, alerted *bool) {
	var (
		chainId   string
		height    string
		blocktime time.Time
	)

	height = block.Result.Block.Header.Height
	chainId = block.Result.Block.Header.ChainID
	blocktime = block.Result.Block.Header.Time

	if chainId != network.ChainId && *alerted {
		log.Println("err - chain id validation failed for rpc", network.Rpcs[0], "on", network.ChainId)
		alerted = new(bool)
		*alerted = true
		alertChan <- alert.NoRpc(network.ChainId)
		return
	}

	if network.StallTime != 0 && time.Since(blocktime) > time.Minute*time.Duration(network.StallTime) {
		log.Println("last block time on", network.ChainId, "is", blocktime, "- sending alert")
		alerted = new(bool)
		*alerted = true
		alertChan <- alert.Stalled(blocktime, network.ChainId)
	}

	alert := backCheck(network, height, validator, block, alerted) // Pass the alerted parameter here
	alertChan <- alert
}

func backCheck(network settings.Network, height string, validator settings.Validators, block rpc.Block, alerted *bool) alert.Alert {
	signedBlocks := 0
	missedBlocks := 0
	heightInt, _ := strconv.Atoi(height)

	for checkHeight := heightInt - network.BackCheck + 1; checkHeight <= heightInt; checkHeight++ {
		if checkSig(validator.Address, block) {
			signedBlocks++
		} else {
			missedBlocks++
		}
	}

	if missedBlocks > network.AlertThreshold {
		if !*alerted {
			*alerted = true
			return alert.Missed(missedBlocks, network.BackCheck, validator.Moniker)
		} else {
			return alert.Nil("repeat alert suppressed - Missed blocks for " + validator.Moniker)
		}
	} else if signedBlocks == network.BackCheck {
		if *alerted {
			*alerted = false
			return alert.Cleared(signedBlocks, network.BackCheck, validator.Moniker)
		} else {
			return alert.Signed(signedBlocks, network.BackCheck, validator.Moniker)
		}
	} else {
		return alert.Nil("found " + strconv.Itoa(signedBlocks) + " of " + strconv.Itoa(network.BackCheck) + " signed for " + validator.Moniker)
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
