package alert

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/doggystylez/penpal/internal/config"
)

const retries = 3

func (a Alert) Handle(notifiers config.Notifiers) {
	a.telegramAlert(notifiers.Telegram.Key, notifiers.Telegram.Chat)
}

func (a Alert) telegramAlert(key, chat string) {
	if a.AlertType != None {
		for i := 0; i < retries; i++ {
			json, err := json.Marshal(TelegramMessage{Chat: chat, Text: a.Message})
			if err != nil {
				panic(err)
			}
			client := &http.Client{
				Timeout: time.Second * 2,
			}
			req, err := http.NewRequestWithContext(context.Background(), "POST", "https://api.telegram.org/bot"+key+"/sendMessage", bytes.NewBuffer(json))
			if err != nil {
				panic(err)
			}
			req.Header.Set("Content-Type", "application/json")
			resp, err := client.Do(req)
			if err != nil {
				log.Println(err)
			}
			err = req.Body.Close()
			if err != nil {
				panic(err)
			}
			if err == nil && resp.Status == "200 OK" {
				log.Println("sent alert to telegram", a.Message)
				return
			}
			time.Sleep(1 * time.Second)
		}
		log.Println("error sending message", a.Message, "to telegram after", retries, "tries")
	} else {
		log.Println(a.Message)
	}
}
