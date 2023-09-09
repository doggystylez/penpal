package settings

type (
	Config struct {
		Network    []Network    `json:"network,omitempty"`
		Notifiers  Notifiers    `json:"notifiers,omitempty"`
		Validators []Validators `json:"validators,omitempty"`
	}

	Network struct {
		ChainId        string   `json:"chain_id,omitempty"`
		Rpcs           []string `json:"rpcs,omitempty"`
		BackCheck      int      `json:"back_check,omitempty"`
		AlertThreshold int      `json:"alert_threshold,omitempty"`
		Interval       int      `json:"interval,omitempty"`
		StallTime      int      `json:"stall_time,omitempty"`
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

	Validators struct {
		Moniker string `json:"moniker,omitempty"`
		Address string `json:"address,omitempty"`
	}
)
