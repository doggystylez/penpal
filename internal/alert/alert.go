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

	"github.com/cordtus/penpal/internal/config"
)

const (
	maxRepeatAlerts = 5
	maxRetries      = 5
	initialBackoff  = 1 * time.Second
)

func Watch(alertChan <-chan Alert, cfg config.Config, client *http.Client) {
	backoffAttempts := make(map[string]int)

	var lastAlertTime = time.Now().Add(-time.Hour)

	alertDelay := 250 * time.Millisecond

	for {
		a := <-alertChan
		if a.AlertType == None {
			log.Println(a.Message)
			continue
		}

		if backoffAttempts[a.Message] >= maxRepeatAlerts {
			log.Printf("Maximum repeat attempts reached for alert: %s. Skipping further notifications.", a.Message)
			continue
		}

		timeElapsed := time.Since(lastAlertTime)

		interval := time.Duration(cfg.Network[0].Interval) * time.Second

		if timeElapsed >= interval {
			var notifications []notification
			if cfg.Notifiers.Telegram.Key != "" {
				notifications = append(notifications, telegramNoti(cfg.Notifiers.Telegram.Key, cfg.Notifiers.Telegram.Chat, a.Message))
			}
			if cfg.Notifiers.Discord.Webhook != "" {
				notifications = append(notifications, discordNoti(cfg.Notifiers.Discord.Webhook, a.Message))
			}

			for _, n := range notifications {
				go func(b notification, alertMsg string) {
					backoffDuration := initialBackoff
					for i := 0; i < maxRetries; i++ {
						err := b.send(client)
						if err == nil {
							log.Println("Sent alert to", b.Type, alertMsg)
							delete(backoffAttempts, alertMsg)
							lastAlertTime = time.Now()
							return
						}
						time.Sleep(backoffDuration)
						backoffDuration *= 2
						log.Printf("Error sending message %s to %s. Retrying in %v", alertMsg, b.Type, backoffDuration)
					}

					backoffAttempts[alertMsg]++
					log.Printf("Error sending message %s to %s after maximum retries. Skipping further notifications.", alertMsg, b.Type)
				}(n, a.Message)
				time.Sleep(alertDelay)
			}
		} else {
			log.Printf("Last alert was too recent: %s. Waiting for cool-down.", a.Message)
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
	if err != nil {
		log.Println(err)
		return
	}
	defer resp.Body.Close()
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
	return Alert{AlertType: None, Message: message}
}

func Missed(missed int, check int, validatorMoniker string) Alert {
	return Alert{AlertType: Miss, Message: validatorMoniker + "❌ missed " + strconv.Itoa(missed) + " of last " + strconv.Itoa(check) + " blocks"}
}

func Cleared(signed int, check int, validatorMoniker string) Alert {
	return Alert{AlertType: Clear, Message: ":face_exhaling:  alert resolved. found " + strconv.Itoa(signed) + " of " + strconv.Itoa(check) + " signed blocks for " + validatorMoniker}
}

func Signed(signed int, check int, validatorMoniker string) Alert {
	return Alert{AlertType: Clear, Message: ":white_check_mark:  blocks! found " + strconv.Itoa(signed) + " of " + strconv.Itoa(check) + " signed blocks for " + validatorMoniker}
}

func NoRpc(ChainId string) Alert {
	return Alert{AlertType: RpcError, Message: "📡 no rpcs available for " + ChainId}
}

func RpcDown(url string) Alert {
	return Alert{AlertType: RpcError, Message: "📡 rpc " + url + " is down or malfunctioning "}
}

func Stalled(blocktime time.Time, ChainId string) Alert {
	return Alert{AlertType: Stall, Message: "⏰ warning - last block found for " + ChainId + " was " + blocktime.Format(time.RFC1123)}
}

func Healthy(interval time.Duration, address string) Alert {
	return Alert{AlertType: Health, Message: "🤝 penpal at " + address + " healthy. next check at " + timeInterval(interval)}
}

func Unhealthy(interval time.Duration, address string) Alert {
	return Alert{AlertType: Health, Message: "🤢 penpal at " + address + " unhealthy. next check at " + timeInterval(interval)}
}

func timeInterval(d time.Duration) string {
	return time.Now().UTC().Add(d).Format(time.RFC3339)
}
