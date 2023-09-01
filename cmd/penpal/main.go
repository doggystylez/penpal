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

	for _, validator := range cfg.Validators {
		validatorConfig := createValidatorConfig(validator, cfg.Network)
		go scan.Monitor(validatorConfig)
	}
	select {}
}

func createValidatorConfig(validator config.Validator, network config.Network) config.Config {
	return config.Config{
		Validators: []config.Validator{
			validator,
		},
		Network:   network, // Use the common network config for all validators
		Notifiers: cfg.Notifiers,
		Health:    cfg.Health,
	}
}
