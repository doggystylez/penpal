package scan

import (
	"log"
	"math/rand"
	"strconv"
	"time"

	"github.com/doggystylez/penpal/alert"
	"github.com/doggystylez/penpal/config"
	"github.com/doggystylez/penpal/rpc"
)

const signThreshold = 0.95

func Monitor(cfg config.Config) {
	exit := make(chan bool)
	for _, network := range cfg.Networks {
		alertChan := make(chan alert.Alert)
		go func(n config.Network, aChan chan<- alert.Alert) {
			for {
				checkNetwork(n, aChan)
				time.Sleep(time.Duration(n.Interval) * time.Minute)
			}

		}(network, alertChan)
		go func(aChan <-chan alert.Alert) {
			var alerted bool
			for {
				a := <-aChan
				a.Handle(cfg.Notifiers, &alerted)
			}
		}(alertChan)
	}
	<-exit
}

func checkNetwork(network config.Network, alertChan chan<- alert.Alert) {
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
				alertChan <- alert.Alert{AlertType: alert.RpcError, Message: "no rpcs available for " + network.ChainId}
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
			alertChan <- alert.Alert{AlertType: alert.RpcError, Message: "no rpcs available for " + network.ChainId}
			return
		}
		if chainId != network.ChainId {
			alertChan <- alert.Alert{AlertType: alert.RpcError, Message: "no rpcs available for " + network.ChainId}
			return
		}
	}
	heightInt, err := strconv.Atoi(height)
	if err != nil {
		return
	}
	alertChan <- backCheck(client, network, heightInt)
}

func backCheck(client rpc.Client, cfg config.Network, height int) (a alert.Alert) {
	var (
		signed    int
		rpcErrors int
	)
	for checkHeight := height - cfg.BackCheck + 1; checkHeight <= height; checkHeight++ {
		block, err := rpc.GetBlockFromHeight(client, strconv.Itoa(checkHeight))
		if err != nil || block.Error != nil {
			log.Println(err, block.Error)
			rpcErrors++
			continue
		}
		if rpcErrors > cfg.BackCheck/2 {
			a = alert.Alert{AlertType: alert.RpcError, Message: "rpc " + client.Url + " is down"}
			return
		}
		if checkSig(cfg.Address, block) {
			signed++
		}
	}
	if float64(signed)/float64(cfg.BackCheck) < signThreshold {
		a.AlertType = alert.Missed
		a.Message = "missed " + strconv.Itoa(cfg.BackCheck-signed) + " blocks on " + cfg.ChainId
	} else {
		a.AlertType = alert.None
		a.Message = "found " + strconv.Itoa(signed) + " of " + strconv.Itoa(cfg.BackCheck) + " signed blocks on " + cfg.ChainId
	}
	return
}

func checkSig(address string, block rpc.Block) bool {
	for _, sig := range block.Result.Block.LastCommit.Signatures {
		if sig.ValidatorAddress == address {
			return true
		}
	}
	return false
}
