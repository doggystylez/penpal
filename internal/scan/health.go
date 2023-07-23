package scan

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/doggystylez/penpal/internal/alert"
	"github.com/doggystylez/penpal/internal/config"
)

func healthCheck(cfg config.Health, alertChan chan<- alert.Alert, client *http.Client) {
	for _, address := range cfg.Nodes {
		go func(a string) {
			var (
				interval time.Duration
				alerted  bool
			)
			for {
				request, err := http.NewRequestWithContext(context.Background(), "GET", a+"/health", nil)
				if err != nil {
					log.Println("health check failed:", err)
					panic(err)
				}
				request.Header.Set("agent", "penpal")
				resp, err := client.Do(request)
				if err != nil {
					alerted = true
					log.Println("health check for", a, "failed:", err, "next check in two minutes")
				} else {
					if resp.StatusCode != http.StatusOK {
						alerted = true
						log.Println("health check for", a, "failed: status code", resp.StatusCode, "next check in two minutes")
					} else {
						log.Println("health check for", a, "succeeded. next check at", time.Now().UTC().Add(interval).Format(time.RFC3339))
						interval = time.Duration(cfg.Interval) * time.Hour
						if alerted {
							alerted = false
							alertChan <- alert.Healthy(interval, a)
						}
					}
					_ = resp.Body.Close()
				}
				if alerted {
					interval = 2 * time.Minute
					alertChan <- alert.Unhealthy(interval, a)
				}
				time.Sleep(interval)
			}
		}(address)
	}
}

func healthServer(port string) {
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("agent") == "penpal" {
			_, err := w.Write([]byte("OK"))
			if err != nil {
				log.Println("failed writing http response", err)
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
		Addr:              ":" + port,
		ReadHeaderTimeout: 5 * time.Second,
	}
	if err := server.ListenAndServe(); err != nil {
		log.Println("http server failed", err)
		return
	}
}
