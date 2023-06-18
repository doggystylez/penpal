package config

import (
	"encoding/json"
	"errors"
	"net/url"
	"os"
	"path/filepath"
)

func Load(path string) (config Config, err error) {
	configFile := filepath.Join(filepath.Clean(path), "config.json")
	if _, err = os.Stat(configFile); os.IsNotExist(err) {
		err = new(configFile)
		if err != nil {
			panic(err)
		}
		err = errors.New("config file does not exist - generating a new one")
		return
	}
	data, err := os.ReadFile(configFile)
	if err != nil {
		panic(err)
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
	return ""
}

func new(file string) (err error) {
	configFile, err := os.Create(file)
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
				Webhook: "webhook",
			},
		},
	})
	return
}
