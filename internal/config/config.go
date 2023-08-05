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
		err = errors.New("config file does not exist. use `-init` to generate a new one")
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
	if len(c.Networks) == 0 {
		return "no networks configured"
	}
	for _, network := range c.Networks {
		if network.Name == "" {
			return "name missing"
		} else if network.Address == "" {
			return "address missing"
		} else if network.ChainId == "" {
			return "chain-id missing for " + network.Name
		} else if network.BackCheck <= 0 {
			return "backcheck missing or invalid for " + network.Name
		} else if network.AlertThreshold <= 0 || network.AlertThreshold > network.BackCheck {
			return "alert threshold missing or invalid for " + network.Name
		} else if network.Interval <= 0 {
			return "check interval missing or invalid for " + network.Name
		} else if network.StallTime < 0 {
			return "stall time invalid for " + network.Name
		} else {
			for _, rpc := range network.Rpcs {
				parsed, err := url.Parse(rpc)
				if err != nil || parsed.Scheme == "" || parsed.Host == "" || (parsed.Scheme != "https" && parsed.Scheme != "http") {
					return "rpc \"" + rpc + "\" invalid for" + network.Name
				}
			}
		}
	}
	if c.Notifiers.Telegram.Key != "" && c.Notifiers.Telegram.Chat == "" {
		return "telegram chat id missing"
	}
	if c.Notifiers.Telegram.Key == "" && c.Notifiers.Discord.Webhook == "" {
		return "telegram or discord notifier must be configured"
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
		Networks: []Network{{
			Name:           "Network1",
			ChainId:        "network-1",
			Address:        "AAAABBBBCCCCDDDD",
			Rpcs:           []string{"rpc1", "rpc2"},
			RpcAlert:       true,
			BackCheck:      10,
			AlertThreshold: 5,
			Interval:       15,
			StallTime:      30,
		}, {
			Name:           "Network2",
			ChainId:        "network-2",
			Address:        "AAAABBBBCCCCDDDD",
			Rpcs:           []string{"rpc1", "rpc2"},
			RpcAlert:       true,
			BackCheck:      5,
			AlertThreshold: 5,
			Interval:       15,
			StallTime:      30,
		}},
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
			}},
		Health: Health{
			Interval: 1,
			Port:     "8080",
			Nodes:    []string{"http://192.168.1.1:8080"},
		},
	})
	if err == nil {
		err = errors.New("generated a new config at " + file)
	}
	return
}
