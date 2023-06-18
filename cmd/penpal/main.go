package main

import (
	"flag"
	"fmt"

	"github.com/doggystylez/penpal/internal/config"
	"github.com/doggystylez/penpal/internal/scan"
)

func main() {
	var filepath string
	flag.StringVar(&filepath, "config", "", "path to config (shorthand)")
	flag.StringVar(&filepath, "c", "", "path to config (shorthand)")
	flag.Parse()
	if filepath == "" {
		fmt.Println("specify a config file using the -config/-c flag")
		return
	}
	cfg, err := config.Load(filepath)
	if err != nil {
		panic(err)
	}
	scan.Monitor(cfg)
}
