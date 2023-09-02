package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/cordtus/penpal/internal/config"
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

	latestBlock := fetchLatestBlock(cfg.Network.Rpcs[0]) // Fetch the latest block data once

	for _, validator := range cfg.Validators {
		validatorConfig := createValidatorConfig(validator, cfg.Network, cfg.Notifiers, cfg.Health, latestBlock)
		go scan.Monitor(validatorConfig)
	}

	select {}
}

func createValidatorConfig(validator config.Validator, network config.Network, notifiers config.Notifiers, health config.Health, latestBlock rpc.Block) config.Config {
	return config.Config{
		Validators: []config.Validator{validator}, // Each validator has its own configuration
		Network:    network,                       // Use the common network config for all validators
		Notifiers:  notifiers,
		Health:     health,
		LatestBlock: latestBlock, // Pass the latest block data
	}
}

func fetchLatestBlock(url string) rpc.Block {
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
