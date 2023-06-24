package alert

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/doggystylez/penpal/internal/config"
)

const retries = 3

func Watch(alertChan <-chan Alert, notifiers config.Notifiers, client *http.Client) {
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
						err := b.send(client)
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

func (n notification) send(client *http.Client) (err error) {
	json, err := json.Marshal(n.Content)
	if err != nil {
		return
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

func telegramNoti(key, chat, message string) notification {
	return notification{Type: "telegram", Auth: "https://api.telegram.org/bot" + key + "/sendMessage", Content: telegramMessage{Chat: chat, Text: message}}
}

func discordNoti(url, message string) notification {
	return notification{Type: "discord", Auth: url, Content: discordMessage{Username: "penpal", Content: message}}
}

func Nil(message string) Alert {
	return Alert{AlertType: 0, Message: message}
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

func Healthy(interval time.Duration, address string) Alert {
	return Alert{AlertType: 5, Message: "🤝 penpal at " + address + " healthy. next check at " + timeInterval(interval)}
}

func Unhealthy(interval time.Duration, address string) Alert {
	return Alert{AlertType: 5, Message: "🤢 penpal at " + address + " unhealthy. next check at " + timeInterval(interval)}
}

func timeInterval(d time.Duration) string {
	return time.Now().UTC().Add(d).Format(time.RFC3339)
}
