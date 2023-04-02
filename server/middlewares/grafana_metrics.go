// Package middlewares for middleware between api and backend
package middlewares

import "github.com/prometheus/client_golang/prometheus"

// Requests metrics
var Requests = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_request_count", // metric name
		Help: "Count of status returned by user.",
	},
	[]string{"method", "uri"}, // labels
)
