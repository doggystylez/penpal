package alert

import (
	"strconv"
)

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

func telegramNoti(key, chat, message string) notification {
	return notification{Type: "telegram", Auth: "https://api.telegram.org/bot" + key + "/sendMessage", Content: telegramMessage{Chat: chat, Text: message}}
}

func discordNoti(url, message string) notification {
	return notification{Type: "discord", Auth: url, Content: discordMessage{Username: "penpal", Content: message}}
}

func Nil(signed int, check int, chain string) Alert {
	return Alert{AlertType: 0, Message: "found " + strconv.Itoa(signed) + " of " + strconv.Itoa(check) + " signed blocks on " + chain}
}

func Cleared(signed int, check int, chain string) Alert {
	return Alert{AlertType: 1, Message: "😌 alert resolved. found " + strconv.Itoa(signed) + " of " + strconv.Itoa(check) + " signed blocks on " + chain}
}

func NoRpc(chain string) Alert {
	return Alert{AlertType: 2, Message: "📡 no rpcs available for " + chain}
}

func RpcDown(url string) Alert {
	return Alert{AlertType: 2, Message: "📡 rpc " + url + " is down or malfunctioning"}
}

func Missed(missed int, check int, chain string) Alert {
	return Alert{AlertType: 3, Message: "❌ missed " + strconv.Itoa(missed) + " of last " + strconv.Itoa(check) + " blocks on " + chain}
}

func Healthy(address string) Alert {
	return Alert{AlertType: 5, Message: "🤝 penpal at " + address + " healthy"}
}

func Unhealthy(address string) Alert {
	return Alert{AlertType: 5, Message: "🤢 penpal at " + address + " unhealthy"}
}
