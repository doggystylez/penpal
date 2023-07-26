package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/doggystylez/penpal/internal/config"
	"github.com/doggystylez/penpal/internal/scan"
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
	for _, network := range cfg.Networks {
		if network.StallTime == 1 {
			fmt.Println("warning! stall time for", network.Name, "is set to 1 minutes, this may cause more frequent false alerts")
		} else if network.StallTime == 0 {
			fmt.Println("warning! stall check for", network.Name, "is disabled")
		}
	}
	scan.Monitor(cfg)
}
