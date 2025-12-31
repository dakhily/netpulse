// Copyright (c) 2025 Dakhil Y.
package main

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var pingLatency = promauto.NewHistogramVec(
	prometheus.HistogramOpts{
		Namespace: "netpulse",
		Name:      "latency_seconds",
		Buckets: []float64{
			0.01, 0.05, 0.1, 0.2, 0.3, 0.4, 0.5, 1.0, 2.5, 5.0,
		},
	},
	[]string{"target", "status"},
)

var pingCount = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "netpulse_requests_total",
	Help: "Total number of pings sent",
}, []string{"target"})

func probe(target string) {
	pingCount.WithLabelValues(target).Inc()

	start := time.Now()
	client := http.Client{Timeout: 5 * time.Second}

	resp, err := client.Get(target)
	duration := time.Since(start).Seconds()

	if err != nil {
		fmt.Printf("Error probing %s: %v\n", target, err)
		pingLatency.WithLabelValues(target, "error").Observe(duration)
		return
	}
	defer resp.Body.Close()

	status := strconv.Itoa(resp.StatusCode)
	pingLatency.WithLabelValues(target, status).Observe(duration)

	fmt.Printf("Target: %s | Latency: %v\n", target, duration)
}

func main() {
	targets := []string{
		"https://www.google.com",
		"https://www.facebook.com",
		"https://www.github.com",
	}

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		fmt.Println("Metric server starting on :8080")
		http.ListenAndServe(":8080", nil)
	}()

	for {
		for _, t := range targets {
			probe(t)
			time.Sleep(1 * time.Second)
		}
	}
}
