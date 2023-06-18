package scan

import (
	"log"
	"math/rand"
	"strconv"
	"time"

	"github.com/doggystylez/penpal/internal/alert"
	"github.com/doggystylez/penpal/internal/config"
	"github.com/doggystylez/penpal/internal/rpc"
)

const signThreshold = 0.95

func Monitor(cfg config.Config) {
	exit := make(chan bool)
	for _, network := range cfg.Networks {
		alertChan := make(chan alert.Alert)
		go func(n config.Network, aChan chan<- alert.Alert) {
			var alerted bool
			for {
				checkNetwork(n, aChan, &alerted)
				time.Sleep(time.Duration(n.Interval) * time.Minute)
			}

		}(network, alertChan)
		go func(aChan <-chan alert.Alert) {
			for {
				a := <-aChan
				a.Handle(cfg.Notifiers)
			}
		}(alertChan)
	}
	<-exit
}

func checkNetwork(network config.Network, alertChan chan<- alert.Alert, alerted *bool) {
	var (
		chainId string
		height  string
		err     error
	)
	client := rpc.New()
	rpcs := network.Rpcs
	if len(rpcs) > 1 {
		for {
			var i int
			var nRpcs []string
			if len(rpcs) == 0 {
				*alerted = true
				alertChan <- alert.NoRpc(network.ChainId)
				return
			} else {
				i = rand.Intn(len(rpcs))
				for _, r := range rpcs {
					if r != rpcs[i] {
						nRpcs = append(nRpcs, r)
					}
				}
				client.Url = rpcs[i]
				rpcs = nRpcs
				chainId, height, err = rpc.GetLastestHeight(client)
				if err != nil {
					log.Println(err)
				}
				if chainId == network.ChainId {
					break
				}
			}
		}
	} else if len(rpcs) == 1 {
		client.Url = network.Rpcs[0]
		chainId, height, err = rpc.GetLastestHeight(client)
		if err != nil {
			log.Println(err)
			*alerted = true
			alertChan <- alert.NoRpc(network.ChainId)
			return
		}
		if chainId != network.ChainId {
			*alerted = true
			alertChan <- alert.NoRpc(network.ChainId)
			return
		}
	}
	heightInt, err := strconv.Atoi(height)
	if err != nil {
		return
	}
	alertChan <- backCheck(client, network, heightInt, alerted)

}

func backCheck(client rpc.Client, cfg config.Network, height int, alerted *bool) alert.Alert {
	var (
		signed    int
		rpcErrors int
	)
	for checkHeight := height - cfg.BackCheck + 1; checkHeight <= height; checkHeight++ {
		block, err := rpc.GetBlockFromHeight(client, strconv.Itoa(checkHeight))
		if err != nil || block.Error != nil {
			log.Println(err, block.Error)
			rpcErrors++
			cfg.BackCheck--
			continue
		}
		if checkSig(cfg.Address, block) {
			signed++
		}
	}
	if rpcErrors > cfg.BackCheck {
		*alerted = true
		return alert.RpcDown(client.Url)
	} else if float64(signed)/float64(cfg.BackCheck) < signThreshold {
		*alerted = true
		return alert.Missed((cfg.BackCheck - signed), cfg.BackCheck, cfg.ChainId)
	} else if *alerted {
		*alerted = false
		return alert.Cleared(signed, cfg.BackCheck, cfg.ChainId)
	} else {
		return alert.Nil(signed, cfg.BackCheck, cfg.ChainId)
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
