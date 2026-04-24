package telemetry

import (
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "boilerplate_http_requests_total",
			Help: "Total number of handled HTTP requests.",
		},
		[]string{"service", "method", "route", "status"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "boilerplate_http_request_duration_seconds",
			Help:    "HTTP request latency in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"service", "method", "route"},
	)

	httpInFlight = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "boilerplate_http_in_flight_requests",
			Help: "Current number of in-flight HTTP requests.",
		},
		[]string{"service"},
	)

	dbQueriesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "boilerplate_db_queries_total",
			Help: "Total number of database queries.",
		},
		[]string{"query", "status"},
	)

	dbQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "boilerplate_db_query_duration_seconds",
			Help:    "Database query latency in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"query"},
	)

	appInfo = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "boilerplate_app_info",
			Help: "Boilerplate application info.",
		},
		[]string{"service", "version", "env"},
	)
)

func ObserveHTTPRequest(service, method, route string, status int, duration time.Duration) {
	httpRequestsTotal.WithLabelValues(service, method, route, strconv.Itoa(status)).Inc()
	httpRequestDuration.WithLabelValues(service, method, route).Observe(duration.Seconds())
}

func InFlightRequests(service string) func() {
	httpInFlight.WithLabelValues(service).Inc()
	return func() {
		httpInFlight.WithLabelValues(service).Dec()
	}
}

func ObserveDBQuery(query string, duration time.Duration, err error) {
	status := "success"
	if err != nil {
		status = "error"
	}

	dbQueriesTotal.WithLabelValues(query, status).Inc()
	dbQueryDuration.WithLabelValues(query).Observe(duration.Seconds())
}

func MarkAppInfo(service, version, env string) {
	appInfo.WithLabelValues(service, version, env).Set(1)
}
