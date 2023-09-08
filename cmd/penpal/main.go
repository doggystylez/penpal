package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/cordtus/penpal/internal/config"
	"github.com/cordtus/penpal/internal/rpc"
	"github.com/cordtus/penpal/internal/scan"
)

func main() {
	var (
		file string
		init bool
	)
	flag.BoolVar(&init, "init", false, "initialize a new config file")
	flag.StringVar(&file, "config", "./config.json", "path to the config file")
	flag.StringVar(&file, "c", "./config.json", "path to the config file [shorthand]")
	flag.Parse()
	args := os.Args
	if len(args) > 1 && !strings.HasPrefix(args[1], "-") {
		fmt.Println("invalid argument:", os.Args[1])
		flag.Usage()
		return
	}
	if init {
		if err := config.New(file); err != nil {
			fmt.Println(err)
		}
		return
	}
	cfg, err := config.Load(file)
	if err != nil {
		fmt.Println(err)
		return
	}

	latestBlock := FetchLatestBlock(cfg.Network.Rpcs[0])

	for _, validator := range cfg.Validators {
		validatorConfig := createValidatorConfig(validator, cfg.Network, cfg.Notifiers, cfg.Health, latestBlock)
		go scan.Monitor(validatorConfig, latestBlock)
	}

	select {}

}

func FetchLatestBlock(url string) rpc.Block {
	client := &http.Client{
		Timeout: time.Second * 5,
	}
	block, err := rpc.GetLatestBlock(url, client)
	if err != nil {
		fmt.Println("Error fetching latest block:", err)
		return rpc.Block{}
	}
	return block
}

func createValidatorConfig(validator config.Validators, network config.Network, notifiers config.Notifiers, health config.Health, latestBlock rpc.Block) config.Config {
	return config.Config{
		Validators: []config.Validators{validator},
		Network:    network,
		Notifiers:  notifiers,
		Health:     health,
	}
}
