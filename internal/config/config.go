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
			return "chain-id missing"
		} else if network.BackCheck <= 0 {
			return "backcheck missing or invalid"
		} else if network.Interval <= 0 {
			return "check interval missing or invalid"
		} else {
			for _, rpc := range network.Rpcs {
				parsed, err := url.Parse(rpc)
				if err != nil || parsed.Scheme == "" || parsed.Host == "" || (parsed.Scheme != "https" && parsed.Scheme != "http") {
					return "rpc \"" + rpc + "\" invalid"
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
	err = encoder.Encode(struct {
		Networks []struct {
			Name      string   `json:"name"`
			ChainId   string   `json:"chain_id"`
			Address   string   `json:"address"`
			Rpcs      []string `json:"rpcs"`
			BackCheck int      `json:"back_check"`
			Interval  int      `json:"interval"`
		} `json:"networks"`
		Notifiers struct {
			Telegram struct {
				Key  string `json:"key"`
				Chat string `json:"chat_id,omitempty"`
			} `json:"telegram"`
			Discord struct {
				Webhook string `json:"webhook"`
			} `json:"discord"`
		} `json:"notifiers"`
		Health struct {
			Interval int      `json:"interval,omitempty"`
			Port     string   `json:"port,omitempty"`
			Nodes    []string `json:"nodes,omitempty"`
		} `json:"health,omitempty"`
	}{
		Networks: []struct {
			Name      string   `json:"name"`
			ChainId   string   `json:"chain_id"`
			Address   string   `json:"address"`
			Rpcs      []string `json:"rpcs"`
			BackCheck int      `json:"back_check"`
			Interval  int      `json:"interval"`
		}{
			{
				Name:      "Network1",
				ChainId:   "network-1",
				Address:   "AAAABBBBCCCCDDDD",
				Rpcs:      []string{"rpc1", "rpc2"},
				BackCheck: 5,
				Interval:  15,
			},
			{
				Name:      "Network2",
				ChainId:   "network-2",
				Address:   "AAAABBBBCCCCDDDD",
				Rpcs:      []string{"rpc1", "rpc2"},
				BackCheck: 5,
				Interval:  15,
			},
		},
		Notifiers: struct {
			Telegram struct {
				Key  string `json:"key"`
				Chat string `json:"chat_id,omitempty"`
			} `json:"telegram"`
			Discord struct {
				Webhook string `json:"webhook"`
			} `json:"discord"`
		}{
			Telegram: struct {
				Key  string `json:"key"`
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
		Health: struct {
			Interval int      `json:"interval,omitempty"`
			Port     string   `json:"port,omitempty"`
			Nodes    []string `json:"nodes,omitempty"`
		}{
			Interval: 1,
			Port:     "8080",
			Nodes:    []string{"http://192.168.1.1:8080"},
		}})
	if err == nil {
		err = errors.New("generated a new config at " + file)
	}
	return
}
