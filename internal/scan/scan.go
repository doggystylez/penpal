package scan

import (
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/doggystylez/penpal/internal/alert"
	"github.com/doggystylez/penpal/internal/config"
	"github.com/doggystylez/penpal/internal/rpc"
)

func Monitor(cfg config.Config) {
	alertChan := make(chan alert.Alert)
	exit := make(chan bool)
	client := &http.Client{
		Timeout: time.Second * 5,
	}
	for _, network := range cfg.Networks {
		go scanNetwork(network, alertChan, client)
		go alert.Watch(alertChan, cfg.Notifiers, client)
	}
	if cfg.Health.Interval != 0 {
		go healthServer(cfg, client)
		go healthCheck(cfg.Health, alertChan, client)
	}
	<-exit
}

func scanNetwork(network config.Network, alertChan chan<- alert.Alert, client *http.Client) {
	var (
		interval int
		alerted  bool
	)
	for {
		alertChan <- checkNetwork(network, &alerted, client)
		if alerted && network.Interval > 2 {
			interval = 2
		} else {
			interval = network.Interval
		}
		time.Sleep(time.Duration(interval) * time.Minute)
	}
}

func checkNetwork(network config.Network, alerted *bool, client *http.Client) alert.Alert {
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
			if len(rpcs) == 0 && !*alerted {
				*alerted = true
				return alert.NoRpc(network.ChainId)
			} else {
				i = rand.Intn(len(rpcs))
				for _, r := range rpcs {
					if r != rpcs[i] {
						nRpcs = append(nRpcs, r)
					}
				}
				url = rpcs[i]
				rpcs = nRpcs
				chainId, height, err = rpc.GetLastestHeight(url, client)
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
		chainId, height, err = rpc.GetLastestHeight(url, client)
		if err != nil && !*alerted {
			*alerted = true
			return alert.NoRpc(network.ChainId)
		}
		if chainId != network.ChainId && !*alerted {
			*alerted = true
			return alert.NoRpc(network.ChainId)
		}
	}
	heightInt, _ := strconv.Atoi(height)
	return backCheck(network, heightInt, alerted, url, client)

}

func backCheck(cfg config.Network, height int, alerted *bool, url string, client *http.Client) alert.Alert {
	var (
		signed    int
		rpcErrors int
	)
	for checkHeight := height - cfg.BackCheck + 1; checkHeight <= height; checkHeight++ {
		block, err := rpc.GetBlockFromHeight(strconv.Itoa(checkHeight), url, client)
		if err != nil || block.Error != nil {
			rpcErrors++
			cfg.BackCheck--
			continue
		}
		if checkSig(cfg.Address, block) {
			signed++
		}
	}
	if rpcErrors > cfg.BackCheck || cfg.BackCheck == 0 {
		if !*alerted {
			*alerted = true
			return alert.RpcDown(url)
		} else {
			return alert.Nil("repeat alert suppressed - RpcDown on " + cfg.ChainId)
		}
	} else if cfg.BackCheck-signed > cfg.AlertThreshold {
		*alerted = true
		return alert.Missed((cfg.BackCheck - signed), cfg.BackCheck, cfg.ChainId)
	} else if *alerted {
		*alerted = false
		return alert.Cleared(signed, cfg.BackCheck, cfg.ChainId)
	} else {
		return alert.Nil("found " + strconv.Itoa(signed) + " of " + strconv.Itoa(cfg.BackCheck) + " signed on " + cfg.ChainId)
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
