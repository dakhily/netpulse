// Copyright (c) 2025 Dakhil Y.
package main

import (
	"context"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (

	//Failures
	FailureNone = "none"

	FailureContextCanceled = "context_canceled"
	FailureContextDeadline = "context_deadline"

	FailureDNSNotFound = "dns_not_found"
	FailureDNSTimeout  = "dns_timeout"
	FailureDNSError    = "dns_error"

	FailureTLSHostnameMismatch = "tls_hostname_mismatch"
	FailureTLSUntrustedCA      = "tls_untrusted_ca"
	FailureTLSCertInvalid      = "tls_cert_invalid"

	FailureTimeout           = "timeout"
	FailureConnectionRefused = "connection_refused"
	FailureNetworkError      = "network_error"

	FailureHTTP4xx = "http_4xx"
	FailureHTTP5xx = "http_5xx"

	FailureUnknown = "unknown"
)

var pingLatency = promauto.NewHistogramVec(
	prometheus.HistogramOpts{
		Namespace: "netpulse",
		Name:      "latency_seconds",
		Buckets: []float64{
			0.01, 0.05, 0.1, 0.2, 0.3, 0.4, 0.5, 1.0, 2.5, 5.0,
		},
	},
	[]string{"target", "status", "error_reason"},
)

var pingCount = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "netpulse_requests_total",
	Help: "Total number of pings sent",
}, []string{"target"})

var probeErrorsTotal = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "probe_errors_total",
		Help: "Total number of probe errors by error reason",
	},
	[]string{"error_reason"},
)

func classifyTransportError(err error) string {
	var dnsErr *net.DNSError
	var netErr net.Error
	var hostErr x509.HostnameError
	var authErr x509.UnknownAuthorityError
	var certErr x509.CertificateInvalidError
	var opErr *net.OpError

	switch {

	case errors.Is(err, context.Canceled):
		return FailureContextCanceled

	case errors.Is(err, context.DeadlineExceeded):
		return FailureContextDeadline

	case errors.As(err, &dnsErr):
		if dnsErr.IsNotFound {
			return FailureDNSNotFound
		}
		if dnsErr.IsTimeout {
			return FailureDNSTimeout
		}
		return FailureDNSError

	case errors.As(err, &hostErr):
		return FailureTLSHostnameMismatch

	case errors.As(err, &authErr):
		return FailureTLSUntrustedCA

	case errors.As(err, &certErr):
		return FailureTLSCertInvalid

	case errors.As(err, &netErr) && netErr.Timeout():
		return FailureTimeout

	case errors.As(err, &opErr):
		if opErr.Err != nil &&
			strings.Contains(opErr.Err.Error(), "connection refused") {
			return FailureConnectionRefused
		}
		return FailureNetworkError

	default:
		return FailureUnknown
	}
}

func classifyHTTPStatus(code int) string {
	switch {
	case code >= 400 && code < 500:
		return FailureHTTP4xx
	case code >= 500:
		return FailureHTTP5xx
	default:
		return FailureNone
	}
}

func probe(target string) {
	pingCount.WithLabelValues(target).Inc()

	start := time.Now()
	client := http.Client{Timeout: 5 * time.Second}

	resp, err := client.Get(target)
	duration := time.Since(start).Seconds()

	status := "success"
	errorReason := FailureNone

	if err != nil {
		status = "transport_error"
		errorReason = classifyTransportError(err)

		probeErrorsTotal.WithLabelValues(errorReason).Inc()
		pingLatency.WithLabelValues(target, status, errorReason).Observe(duration)

		fmt.Printf("Transport error probing %s: %v\n", target, err)
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		status = "http_error"
		errorReason = classifyHTTPStatus(resp.StatusCode)
		probeErrorsTotal.WithLabelValues(errorReason).Inc()
	}

	pingLatency.WithLabelValues(target, status, errorReason).Observe(duration)

	fmt.Printf("Target: %s | Status: %s | Code: %d | Latency: %.3fs\n",
		target, status, resp.StatusCode, duration)
	fmt.Printf("Target: %s | Latency: %v\n", target, duration)
}

func main() {
	targets := []string{
		"https://www.google.com",
		"https://www.facebook.com",
		"https://www.github.com",
		"https://www.giub.com/",
		"https://localhost:8080",
		"https://tools-httpstatus.pickup-services.com/404",
		"https://tools-httpstatus.pickup-services.com/503",
		"https://tools-httpstatus.pickup-services.com/200?sleep=5000",
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
