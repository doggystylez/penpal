package scan

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"runtime"
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

	go func() {
		for {
			time.Sleep(time.Second * 10)
			printStackTraces()
		}
	}()

	<-exit
}

func printStackTraces() {
	stack := make([]byte, 1024)
	length := runtime.Stack(stack, false)

	fmt.Printf("Stack Trace:\n%s\n", stack[:length])
}

func scanNetwork(cfg config.Config, network config.Network, alertChan chan<- alert.Alert, client *http.Client) {
	var (
		interval  int
		alerted   bool
		validator config.Validators
	)
	for {
		checkNetwork(cfg, network, client, &alerted, alertChan, validator.Moniker, validator.Address)
		if alerted && network.Interval > 2 {
			interval = 2
		} else {
			interval = network.Interval
		}
		time.Sleep(time.Duration(interval) * time.Minute)
	}
}

func checkNetwork(cfg config.Config, network config.Network, client *http.Client, alerted *bool, alertChan chan<- alert.Alert, moniker string, address string) {
	var (
		chainId string
		height  string
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

	chainId, blocktime, err := rpc.GetLatestBlockTime(url, client)

	if err != nil || chainId != network.ChainId {
		log.Println("err - failed to check latest block time for", network.ChainId)
	} else if network.StallTime != 0 && time.Since(blocktime) > time.Minute*time.Duration(network.StallTime) {
		log.Println("last block time on", network.ChainId, "is", blocktime, "- sending alert")

		*alerted = true
		alertChan <- alert.Stalled(blocktime, network.ChainId)
	}

	heightInt, _ := strconv.Atoi(height)

	alertChan <- backCheck(cfg, network, heightInt, alerted, url, client, chainId, blocktime, moniker, address)
}

func backCheck(cfg config.Config, network config.Network, height int, alerted *bool, url string, client *http.Client, chainID string, LatestBlockTime time.Time, moniker string, address string) alert.Alert {
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
		if checkSig(address, block) {
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
	} else {
		return alert.Nil("found " + strconv.Itoa(signed) + " of " + strconv.Itoa(network.BackCheck) + " signed for " + moniker)
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
