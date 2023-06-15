package alert

type AlertType int

const (
	None AlertType = iota
	RpcError
	Missed
	Jailed
	Unknown
)

type Alert struct {
	AlertType AlertType
	Message   string
}

type TelegramMessage struct {
	Chat string `json:"chat_id,omitempty"`
	Text string `json:"text,omitempty"`
}
