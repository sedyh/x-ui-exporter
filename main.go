package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
	"x-ui-exporter/api"
	"x-ui-exporter/config"
	"x-ui-exporter/metrics"

	"github.com/go-co-op/gocron"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	version = "unknown"
	commit  = "unknown"
)

func init() { //
	prometheus.MustRegister(
		// User-related metrics
		metrics.OnlineUsersCount,
		// Client-related metrics
		metrics.InboundUp,
		metrics.InboundDown,
		metrics.ClientUp,
		metrics.ClientDown,
		// System-related metrics
		metrics.XrayVersion,
		metrics.PanelThreads,
		metrics.PanelMemory,
		metrics.PanelUptime,
	)
}

func BasicAuthMiddleware(username, password string, protectedMetrics bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if protectedMetrics {
				user, pass, ok := r.BasicAuth()
				if !ok || user != username || pass != password {
					w.Header().Set("WWW-Authenticate", `Basic realm="metrics"`)
					http.Error(w, "Unauthorized.", http.StatusUnauthorized)
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

func main() {
	cliConfig, err := config.Parse(version, commit)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("3X-UI Exporter (https://github.com/hteppl/3x-ui-exporter/)", version)

	s := gocron.NewScheduler(time.Local)
	defer s.Stop()

	client := api.NewAPIClient(api.APIConfig{
		BaseURL:            cliConfig.BaseURL,
		ApiUsername:        cliConfig.ApiUsername,
		ApiPassword:        cliConfig.ApiPassword,
		InsecureSkipVerify: cliConfig.InsecureSkipVerify,
	})

	s.Every(cliConfig.UpdateInterval).Seconds().Do(func() {
		token, err := client.GetAuthToken()
		if err != nil {
			log.Println("Error getting auth token:", err)
			return
		}

		log.Print("Starting to collect metrics")

		log.Print("Collecting UsersStats metrics")
		client.FetchOnlineUsersCount(token)
		log.Print("Finished collecting UsersStats metrics")

		log.Print("Collecting Server and Panel metrics")
		client.FetchServerStatus(token)
		log.Print("Finished collecting Server and Panel metrics")

		log.Print("Collecting Inbounds metrics")
		client.FetchInboundsList(token)
		log.Print("Finished collecting Inbounds metrics")

		log.Print("Finished all metric collection\n\n")
	})

	go s.StartAsync()

	http.Handle("/metrics", BasicAuthMiddleware(
		cliConfig.MetricsUsername,
		cliConfig.MetricsPassword,
		cliConfig.ProtectedMetrics,
	)(promhttp.Handler()))
	log.Printf("Starting server on %s:%s", cliConfig.Ip, cliConfig.Port)
	log.Fatal(http.ListenAndServe(cliConfig.Ip+":"+cliConfig.Port, nil))
}
