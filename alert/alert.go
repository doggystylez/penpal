package alert

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/doggystylez/penpal/config"
)

func (a Alert) Handle(notifiers config.Notifiers, alerted *bool) (err error) {
	err = a.telegramAlert(notifiers.Telegram.Key, notifiers.Telegram.Chat, alerted)
	return
}

func (a Alert) telegramAlert(key, chat string, alerted *bool) (err error) {
	if a.AlertType != None {
		*alerted = true
	} else if a.AlertType == None && *alerted {
		*alerted = false
	} else {
		log.Println(a.Message)
		return
	}
	json, err := json.Marshal(TelegramMessage{Chat: chat, Text: a.Message})
	if err != nil {
		return
	}
	client := &http.Client{
		Timeout: time.Second * 2,
	}
	req, err := http.NewRequestWithContext(context.Background(), "POST", "https://api.telegram.org/bot"+key+"/sendMessage", bytes.NewBuffer(json))
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return
	}
	err = req.Body.Close()
	if err != nil {
		panic(err)
	}
	if resp.Status == "200 OK" {
		log.Println("sent alert to telegram", a.Message)
	} else {
		log.Println("failed sending alert to telegram")
	}
	return
}
