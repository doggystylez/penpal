package scan

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/doggystylez/penpal/internal/alert"
	"github.com/doggystylez/penpal/internal/config"
)

func healthCheck(interval int, addresses []string, alertChan chan<- alert.Alert) {
	client := &http.Client{}
	for {
		for _, address := range addresses {
			request, err := http.NewRequestWithContext(context.Background(), "GET", address+"/health", nil)
			if err != nil {
				panic(err)
			}
			resp, err := client.Do(request)
			if err != nil {
				log.Println("health check failed:", err)
				alertChan <- alert.Unhealthy(address)
			} else {
				if resp.StatusCode != http.StatusOK {
					log.Println("health check for", address, "failed: status code", resp.StatusCode)
					alertChan <- alert.Unhealthy(address)
				} else {
					alertChan <- alert.Healthy(address)
				}
				_ = resp.Body.Close()
			}
		}
		time.Sleep(time.Duration(interval) * time.Hour)
	}
}

func healthServer(cfg config.Config) {
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		var alerted bool
		i := rand.Intn(len(cfg.Networks))
		a := checkNetwork(cfg.Networks[i], &alerted)
		if a.AlertType == 0 || a.AlertType == 1 || a.AlertType == 3 {
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, "OK")
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprint(w, "NOTOK")
		}
	})
	server := &http.Server{
		Addr:              ":" + cfg.Health.Port,
		ReadHeaderTimeout: 3 * time.Second,
	}
	if err := server.ListenAndServe(); err != nil {
		panic(err)
	}
}
