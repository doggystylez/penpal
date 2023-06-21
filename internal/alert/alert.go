package alert

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/doggystylez/penpal/internal/config"
)

const retries = 3

func Watch(alertChan <-chan Alert, notifiers config.Notifiers) {
	for {
		a := <-alertChan
		var notifications []notification
		if notifiers.Telegram.Key != "" {
			notifications = append(notifications, telegramNoti(notifiers.Telegram.Key, notifiers.Telegram.Chat, a.Message))
		}
		if notifiers.Discord.Webhook != "" {
			notifications = append(notifications, discordNoti(notifiers.Discord.Webhook, a.Message))
		}
		if a.AlertType != None {
			for _, n := range notifications {
				go func(b notification) {
					for i := 0; i < retries; i++ {
						err := b.send()
						if err == nil {
							log.Println("sent alert to", b.Type, a.Message)
							return
						}
						time.Sleep(1 * time.Second)
					}
					log.Println("error sending message", a.Message, "to", b.Type, "after", retries, "tries")
				}(n)
			}
		} else {
			log.Println(a.Message)
		}
	}
}

func (n notification) send() (err error) {
	json, err := json.Marshal(n.Content)
	if err != nil {
		return
	}
	client := &http.Client{
		Timeout: time.Second * 2,
	}
	req, err := http.NewRequestWithContext(context.Background(), "POST", n.Auth, bytes.NewBuffer(json))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	defer req.Body.Close()
	if err != nil {
		log.Println(err)
		return
	}
	if resp.StatusCode > 299 || resp.StatusCode < 200 {
		err = errors.New("code not 200")
	}
	return

}
