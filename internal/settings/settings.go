package settings

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
