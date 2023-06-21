package alert

const (
	None AlertType = iota
	Clear
	RpcError
	Miss
	Jail
	Health
	Unknown
)

type (
	AlertType int

	Alert struct {
		AlertType AlertType
		Message   string
	}

	notification struct {
		Type    string
		Auth    string
		Content interface{}
	}

	telegramMessage struct {
		Chat string `json:"chat_id,omitempty"`
		Text string `json:"text,omitempty"`
	}

	discordMessage struct {
		Username string `json:"username"`
		Content  string `json:"content"`
	}
)
