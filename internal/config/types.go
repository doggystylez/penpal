package config

type (
	Config struct {
		Validators []Validators `json:"validators,omitempty"`
		Notifiers  Notifiers    `json:"notifiers,omitempty"`
		Health     Health       `json:"health,omitempty"`
		Network    Network      `json:"network,omitempty"`
	}

	Validators struct {
		Moniker string `json:"moniker,omitempty"`
		Address string `json:"address,omitempty"`
	}

	Network struct {
		ChainId        string   `json:"chain_id,omitempty"`
		Rpcs           []string `json:"rpcs,omitempty"`
		BackCheck      int      `json:"back_check,omitempty"`
		AlertThreshold int      `json:"alert_threshold,omitempty"`
		Interval       int      `json:"interval,omitempty"`
		StallTime      int      `json:"stall_time,omitempty"`
		Reverse        bool     `json:"reverse,omitempty"`
		RpcAlert       bool     `json:"rpc_alert,omitempty"`
	}

	Notifiers struct {
		Telegram struct {
			Key  string `json:"key,omitempty"`
			Chat string `json:"chat_id,omitempty"`
		} `json:"telegram,omitempty"`
		Discord struct {
			Webhook string `json:"webhook"`
		} `json:"discord"`
	}

	Health struct {
		Interval int      `json:"interval,omitempty"`
		Port     string   `json:"port,omitempty"`
		Nodes    []string `json:"nodes,omitempty"`
	}
)
