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
	for _, validator := range cfg.Validators {
		go scanValidator(validator, cfg.Network, alertChan, client)
		go alert.Watch(alertChan, cfg.Notifiers, client)
	}
	if cfg.Health.Interval != 0 {
		go healthServer(cfg.Health.Port)
		go healthCheck(cfg.Health, alertChan, client)
	}
	<-exit
}

func scanValidator(validator config.Validator, network config.Network, alertChan chan<- alert.Alert, client *http.Client) {
	var (
		interval int
		alerted  bool
	)
	for {
		checkNetwork(validator, network, client, &alerted, alertChan)
		if alerted && network.Interval > 2 {
			interval = 2
		} else {
			interval = network.Interval
		}
		time.Sleep(time.Duration(interval) * time.Minute)
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

func checkNetwork(validator config.Validator, network config.Network, client *http.Client, alerted *bool, alertChan chan<- alert.Alert) {
	var (
		chainId string
		height  string
		url     string
		err     error
	)
	rpcs := network.Rpcs
	if len(rpcs) > 1 {
		for {
			var i int
			var nRpcs []string
			if len(rpcs) == 0 && !*alerted && network.RpcAlert {
				*alerted = true
				alertChan <- alert.NoRpc(network.ChainId)
				return
			} else {
				i = rand.Intn(len(rpcs)) //nolint
				for _, r := range rpcs {
					if r != rpcs[i] {
						nRpcs = append(nRpcs, r)
					}
				}
				url = rpcs[i]
				rpcs = nRpcs
				chainId, height, err = rpc.GetLatestHeight(url, client)
				if err != nil {
					continue
				}
				if chainId == network.ChainId {
					break
				}
			}
		}
	} else if len(rpcs) == 1 {
		url = network.Rpcs[0]
		chainId, height, err = rpc.GetLatestHeight(url, client)
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
	}

	// Extract the "time" field from the JSON response
	var blockTime time.Time
	chainId, blockTime, err = rpc.GetLatestBlockTime(url, client)
	if err != nil || chainId != network.ChainId {
		log.Println("err - failed to check latest block time for", network.ChainId)
	} else if network.StallTime != 0 && time.Since(blockTime) > time.Minute*time.Duration(network.StallTime) {
		log.Println("last block time on", network.ChainId, "is", blockTime, "- sending alert")
		*alerted = true
		alertChan <- alert.Stalled(blockTime, network.ChainId)
	}
	heightInt, _ := strconv.Atoi(height)
	alertChan <- backCheck(validator, network, heightInt, alerted, url, client)
}

func backCheck(validator config.Validator, network config.Network, height int, alerted *bool, url string, client *http.Client) alert.Alert {
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
			return alert.Cleared(signed, network.BackCheck, network.ChainId, validator.Name)
		} else {
			return alert.Nil("found " + strconv.Itoa(signed) + " of " + strconv.Itoa(network.BackCheck) + " signed on " + network.ChainId + " for validator " + validator.Name)
		}
	} else {
		if signed > 1 {
			return alert.Signed(signed, network.BackCheck, network.ChainId, validator.Name)
		} else {
			return alert.Nil("found " + strconv.Itoa(signed) + " of " + strconv.Itoa(network.BackCheck) + " signed on " + network.ChainId + " for validator " + validator.Name)
		}
	}
}
