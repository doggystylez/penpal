package scan

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/cordtus/penpal/internal/alert"
	"github.com/cordtus/penpal/internal/config"
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
				}
				request.Header.Set("agent", "penpal")
				resp, err := client.Do(request)
				if err != nil {
					if !alerted {
						log.Println("health check for", a, "failed:", err)
						alerted = true
						alertChan <- alert.Unhealthy(2*time.Minute, a)
					}
				} else {
					if resp.StatusCode != http.StatusOK {
						if !alerted {
							log.Println("health check for", a, "failed: status code", resp.StatusCode)
							alerted = true
							alertChan <- alert.Unhealthy(2*time.Minute, a)
						}
					} else {
						if alerted {
							log.Println("health check for", a, "succeeded.")
							alerted = false
							interval = time.Duration(cfg.Interval) * time.Hour
							alertChan <- alert.Healthy(interval, a)
						}
					}
					_ = resp.Body.Close()
				}
				time.Sleep(interval)
			}
		}(address)
	}
}

func healthServer(port string) {
	once := sync.Once{}

	once.Do(func() {
		server := &http.Server{
			Addr:              ":" + port,
			ReadHeaderTimeout: 5 * time.Second,
		}

		if err := server.ListenAndServe(); err != nil {
			log.Println("http server failed", err)
			return
		}
	})
}
