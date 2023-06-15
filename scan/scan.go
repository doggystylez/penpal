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
	alertChan := make(chan alert.Alert)
	for _, network := range cfg.Networks {
		go func(n config.Network, a chan<- alert.Alert) {
			for {
				checkNetwork(n, a)
				time.Sleep(time.Duration(n.Interval) * time.Minute)
			}

		}(network, alertChan)
	}
	var alerted bool
	for {
		a := <-alertChan
		err := a.Handle(cfg.Notifiers, &alerted)
		if err != nil {
			return
		}
		log.Println(alerted)
	}
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
	signed, err := backCheck(client, network.Address, network.BackCheck, heightInt)
	if err != nil {
		log.Println(err)
		return
	}
	alertChan <- checkSigning(signed, network.BackCheck, network.ChainId)
}

func checkSigning(signed int, check int, chainId string) (a alert.Alert) {
	if float64(signed)/float64(check) < signThreshold {
		a.AlertType = alert.Missed
		a.Message = "missed " + strconv.Itoa(check-signed) + " blocks on " + chainId
	} else {
		a.AlertType = alert.None
		a.Message = "found " + strconv.Itoa(signed) + " of " + strconv.Itoa(check) + " signed blocks on " + chainId
	}
	return
}

func backCheck(client rpc.Client, address string, checkBack int, height int) (signed int, err error) {
	for checkHeight := height - checkBack + 1; checkHeight <= height; checkHeight++ {
		var block rpc.Block
		block, err = rpc.GetBlockFromHeight(client, strconv.Itoa(checkHeight))
		if err != nil {
			return
		}
		if checkSig(address, block) {
			signed++
		}
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
