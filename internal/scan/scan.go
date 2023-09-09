package scan

import (
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

	network := cfg.Network[0]
	go scanNetwork(cfg, network, alertChan, client)
	go alert.Watch(alertChan, cfg.Notifiers, client)

	if cfg.Health.Interval != 0 {
		go healthServer(cfg.Health.Port)
		go healthCheck(cfg.Health, alertChan, client)
	}

	<-exit
}

func scanNetwork(cfg config.Config, network config.Network, alertChan chan<- alert.Alert, client *http.Client) {
	var (
		interval int
		alerted  bool
	)
	url := network.Rpcs[0]
	heightStr, err := rpc.GetLatestHeight(url, client)
	if err != nil {
		log.Println("Error getting latest height:", err)
		return
	}
	height, err := strconv.Atoi(heightStr)
	if err != nil {
		log.Println("Error converting height to int:", err)
		return
	}
	startHeight := height - network.BackCheck + 1

	signedBlocks := backCheck(cfg, network, startHeight, client)

	for _, validator := range cfg.Validators {
		alertChan <- alert.Signed(signedBlocks, network.BackCheck, validator.Moniker)
	}

	if alerted && network.Interval > 2 {
		interval = 2
	} else {
		interval = network.Interval
	}
	time.Sleep(time.Duration(interval) * time.Minute)
}

func checkNetwork(cfg config.Config, network config.Network, client *http.Client, alerted *bool, alertChan chan<- alert.Alert, moniker string, address string) {
	var (
		chainId string
		height  int
		url     string
		err     error
	)
	rpcs := network.Rpcs

	if len(rpcs) == 0 && !*alerted && network.RpcAlert {
		*alerted = true
		alertChan <- alert.NoRpc(network.ChainId)
		return
	}

	if len(rpcs) > 0 {
		i := rand.Intn(len(rpcs))
		url = rpcs[i]
	} else {
		url = network.Rpcs[0]
	}

	_, err = rpc.GetLatestHeight(url, client)

	if err != nil && !*alerted && network.RpcAlert {
		log.Println("err - failed to check latest height for", network.ChainId, "err - ", err)
		*alerted = true
		alertChan <- alert.NoRpc(network.ChainId)
		return
	}

	if chainId != network.ChainId && !*alerted && network.RpcAlert {
		log.Println("err - chain id validation failed for rpc", url, "on", network.ChainId)
		*alerted = true
		alertChan <- alert.NoRpc(network.ChainId)
		return
	}

	chainId, blocktime, err := rpc.GetLatestBlockTime(url, client)

	if err != nil || chainId != network.ChainId {
		log.Println("err - failed to check latest block time for", network.ChainId)
	} else if network.StallTime != 0 && time.Since(blocktime) > time.Minute*time.Duration(network.StallTime) {
		log.Println("last block time on", network.ChainId, "is", blocktime, "- sending alert")

		*alerted = true
		alertChan <- alert.Stalled(blocktime, network.ChainId)
	}

	alert := alert.Signed(backCheck(cfg, network, height, client), network.BackCheck, moniker)
	alertChan <- alert
}

func backCheck(cfg config.Config, network config.Network, height int, client *http.Client) int {
	signedBlocks := 0

	for checkHeight := height; checkHeight <= height+network.BackCheck-1; checkHeight++ {
		block, err := rpc.GetBlockFromHeight(strconv.Itoa(checkHeight), network.Rpcs[0], client)
		if err != nil || block.Error != nil {
			log.Println("Error fetching block at height", checkHeight, ":", err)
			continue
		}

		for _, validator := range cfg.Validators {
			if checkSig(validator.Address, block) {
				signedBlocks++
			}
		}
	}

	return signedBlocks
}

func checkSig(address string, block rpc.Block) bool {
	for _, sig := range block.Result.Block.LastCommit.Signatures {
		if sig.ValidatorAddress == address {
			return true
		}
	}
	return false
}
