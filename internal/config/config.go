package config

import (
	"encoding/json"
	"errors"
	"net/url"
	"os"
	"path/filepath"
)

func Load(file string) (config Config, err error) {
	configFile := filepath.Clean(file)
	if _, err = os.Stat(configFile); os.IsNotExist(err) {
		err = errors.New("no config found - run with `-c` `/path/to/config`, or `-init` to generate one")
		return
	}
	data, err := os.ReadFile(configFile)
	if err != nil {
		return
	}
	err = json.Unmarshal(data, &config)
	warn := config.validate()
	if warn != "" {
		err = errors.New("check config - " + warn)
		return
	}
	return
}

func (c Config) validate() string {
	if len(c.Validators) == 0 {
		return "no validators found - check config"
	}

	network := c.Network[0]

	if network.ChainId == "" {
		return "chain-id value invalid - check config"
	}
	if network.BackCheck <= 0 {
		return "backcheck value invalid - check config"
	}
	if network.AlertThreshold <= 0 || network.AlertThreshold > network.BackCheck {
		return "alert threshold value invalid - check config"
	}
	if network.Interval <= 0 {
		return "check interval value invalid - check config"
	}
	if network.StallTime < 0 {
		return "stall time value invalid - check config"
	}

	for _, rpcURL := range network.Rpcs {
		parsedURL, err := url.Parse(rpcURL)
		if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" || (parsedURL.Scheme != "https" && parsedURL.Scheme != "http") {
			return "rpc \"" + rpcURL + "\" invalid for the network"
		}
	}

	if c.Notifiers.Telegram.Key != "" && c.Notifiers.Telegram.Chat == "" {
		return "telegram chat id missing - check config"
	}
	if c.Notifiers.Telegram.Key == "" && c.Notifiers.Discord.Webhook == "" {
		return "telegram or discord notifier missing - check config"
	}

	return ""
}

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
