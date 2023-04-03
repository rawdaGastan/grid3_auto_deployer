// Package middlewares for middleware between api and backend
package middlewares

import "github.com/prometheus/client_golang/prometheus"

// Requests metrics
var Requests = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_request_count", // metric name
		Help: "Count of requests done by user.",
	},
	[]string{"method", "uri", "status"}, // labels
)

// UserCreations metrics
var UserCreations = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_request_create_user", // metric name
		Help: "Count of users registered.",
	},
	[]string{"user", "email", "college", "team"}, // labels
)

// VoucherActivated metrics
var VoucherActivated = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_request_activate_voucher", // metric name
		Help: "Count of activated voucher.",
	},
	[]string{"user", "voucher", "vms", "public_ips"}, // labels
)

// VoucherApplied metrics
var VoucherApplied = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_request_apply_voucher", // metric name
		Help: "Count of applied voucher.",
	},
	[]string{"user", "voucher", "vms", "public_ips"}, // labels
)

// Deployments metrics
var Deployments = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_request_deploy", // metric name
		Help: "Count of deployments.",
	},
	[]string{"user", "resources", "type"}, // labels
)

// Deletions metrics
var Deletions = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_request_delete", // metric name
		Help: "Count of deletions.",
	},
	[]string{"user", "type"}, // labels
)
