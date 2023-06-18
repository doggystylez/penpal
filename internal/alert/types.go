package alert

import "strconv"

type AlertType int

const (
	None AlertType = iota
	Clear
	RpcError
	Miss
	Jail
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

func Nil(signed int, check int, chain string) (a Alert) {
	a.AlertType, a.Message = 0, "found "+strconv.Itoa(signed)+" of "+strconv.Itoa(check)+" signed blocks on "+chain
	return
}

func Cleared(signed int, check int, chain string) (a Alert) {
	a.AlertType, a.Message = 1, "😌 alert resolved. found "+strconv.Itoa(signed)+" of "+strconv.Itoa(check)+" signed blocks on "+chain
	return
}

func NoRpc(chain string) (a Alert) {
	a.AlertType, a.Message = 2, "📡 no rpcs available for "+chain
	return
}

func RpcDown(url string) (a Alert) {
	a.AlertType, a.Message = 2, "📡 rpc "+url+" is down or malfunctioning"
	return
}

func Missed(missed int, check int, chain string) (a Alert) {
	a.AlertType, a.Message = 3, "❌ missed "+strconv.Itoa(missed)+" of last "+strconv.Itoa(check)+" blocks on "+chain
	return
}
