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
	flag.BoolVar(&init, "init", false, "initialize new config file")
	flag.StringVar(&file, "config", "./config.json", "path to config")
	flag.StringVar(&file, "c", "./config.json", "path to config [shorthand]")
	flag.Parse()
	args := os.Args
	if len(args) > 1 && !strings.HasPrefix(args[1], "-") {
		fmt.Println("invalid arg", os.Args[1])
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

	for _, validator := range cfg.Validators {
		go func(validator config.Validator) {
			validatorConfig := createValidatorConfig(validator, cfg.CommonRPCs)
			scan.Monitor(validatorConfig)
		}(validator)
	}
	select {}
}

func createValidatorConfig(validator config.Validator, commonRPCs []string) config.Config {
	return config.Config{
		Networks: []config.Network{
			{
				Name:           validator.Moniker,
				ChainId:        "common-chain-id", // Replace with the actual chain ID
				Address:        validator.Address,
				Rpcs:           commonRPCs,
				RpcAlert:       true,
				BackCheck:      10,
				AlertThreshold: 5,
				Interval:       15,
				StallTime:      30,
			},
		},
		Notifiers: cfg.Notifiers, // You can reuse the notifiers from the common config
		Health:    cfg.Health,    // You can reuse the health config from the common config
	}
}
