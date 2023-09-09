package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

func New(file string) (err error) {
	_, err = os.Stat(file)
	if err == nil {
		err = errors.New("config file already exists at " + file)
		return
	} else if !os.IsNotExist(err) {
		return
	}
	configFile, err := os.Create(filepath.Clean(file))
	if err != nil {
		return
	}
	encoder := json.NewEncoder(configFile)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(Config{
		Network: []Network{
			{
				ChainId:        "network-1",
				Rpcs:           []string{"rpc1", "rpc2"},
				BackCheck:      10,
				AlertThreshold: 5,
				Interval:       15,
				StallTime:      30,
			},
		},
		Notifiers: Notifiers{
			Telegram: struct {
				Key  string `json:"key,omitempty"`
				Chat string `json:"chat_id,omitempty"`
			}{
				Key:  "api_key",
				Chat: "chat_id",
			},
			Discord: struct {
				Webhook string `json:"webhook"`
			}{
				Webhook: "webhook_url",
			},
		},
		Health: Health{
			Interval: 1,
			Port:     "8080",
			Nodes:    []string{"http://192.168.1.1:8080"},
		},
		Validators: []Validators{
			{
				Moniker: "Validator1",
				Address: "AAAABBBBCCCCDDDD1",
			},
			{
				Moniker: "Validator2",
				Address: "AAAABBBBCCCCDDDD2",
			},
		},
	})
	if err == nil {
		err = errors.New("generated a new config at " + file)
	}
	return
}
