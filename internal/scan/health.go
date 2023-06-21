package scan

import (
	"context"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/doggystylez/penpal/internal/alert"
	"github.com/doggystylez/penpal/internal/config"
)

func healthCheck(cfg config.Health, alertChan chan<- alert.Alert, alerted *bool) {
	client := &http.Client{}
	for {
		for _, address := range cfg.Nodes {
			request, err := http.NewRequestWithContext(context.Background(), "GET", address+"/health", nil)
			if err != nil {
				log.Println("health check failed:", err)
				continue
			}
			request.Header.Set("agent", "penpal")
			resp, err := client.Do(request)
			if err != nil {
				*alerted = true
				log.Println("health check failed:", err)
				alertChan <- alert.Unhealthy(cfg.Interval, address)
			} else {
				if resp.StatusCode != http.StatusOK {
					*alerted = true
					log.Println("health check for", address, "failed: status code", resp.StatusCode)
					alertChan <- alert.Unhealthy(cfg.Interval, address)
				} else {
					log.Println("health check for", address, "succeeds")
					if *alerted {
						alertChan <- alert.Healthy(cfg.Interval, address)
					}
				}
				_ = resp.Body.Close()
			}
		}
		time.Sleep(time.Duration(cfg.Interval) * time.Hour)
	}
}

func healthServer(cfg config.Config) {
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("agent") == "penpal" {
			var alerted bool
			i := rand.Intn(len(cfg.Networks))
			a := checkNetwork(cfg.Networks[i], &alerted)
			if a.AlertType == 0 || a.AlertType == 1 || a.AlertType == 3 {
				w.WriteHeader(http.StatusOK)
				_, err := w.Write([]byte("OK"))
				if err != nil {
					log.Println("failed writing http response", err)
				}
			} else {
				w.WriteHeader(http.StatusServiceUnavailable)
				_, err := w.Write([]byte("NOTOK"))
				if err != nil {
					log.Println("failed writing http response", err)
				}
			}
		} else {
			w.WriteHeader(http.StatusUnauthorized)
			_, err := w.Write([]byte("NOT AUTHORIZED"))
			if err != nil {
				log.Println("failed writing http response", err)
			}
		}
	})
	server := &http.Server{
		Addr:              ":" + cfg.Health.Port,
		ReadHeaderTimeout: 3 * time.Second,
	}
	if err := server.ListenAndServe(); err != nil {
		log.Println("http server failed", err)
		return
	}
}
