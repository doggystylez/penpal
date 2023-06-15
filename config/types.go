package config

type (
	Config struct {
		Networks  []Network `json:"networks,omitempty"`
		Notifiers Notifiers `json:"notifiers,omitempty"`
	}
	Network struct {
		Name      string   `json:"name,omitempty"`
		ChainId   string   `json:"chain_id,omitempty"`
		Address   string   `json:"address,omitempty"`
		Rpcs      []string `json:"rpcs,omitempty"`
		BackCheck int      `json:"back_check,omitempty"`
		Interval  int      `json:"interval,omitempty"`
	}

	Notifiers struct {
		Telegram struct {
			Key  string `json:"key,omitempty"`
			Chat string `json:"chat_id,omitempty"`
		} `json:"telegram,omitempty"`
	}
)
