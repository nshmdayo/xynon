package proxy

import (
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	requestCount = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "xynon_requests_total",
			Help: "Total number of HTTP requests processed, partitioned by status code, method and normalized path.",
		},
		[]string{"status", "method", "path"},
	)
	requestLatency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "xynon_request_duration_seconds",
			Help:    "Histogram of request latencies.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"status", "method", "path"},
	)
)

// normalizePath replaces segments that look like IDs with placeholders to prevent cardinality explosion.
func normalizePath(path string) string {
	segments := strings.Split(path, "/")
	for i, segment := range segments {
		// Replace numeric segments with {id}
		if _, err := strconv.Atoi(segment); err == nil {
			segments[i] = "{id}"
			continue
		}
		// Replace UUID segments with {id}
		if len(segment) == 36 && strings.Count(segment, "-") == 4 {
			segments[i] = "{id}"
		}
	}
	return strings.Join(segments, "/")
}

// RecordMetrics records Prometheus metrics for a single request.
func RecordMetrics(method, path string, status int, duration time.Duration) {
	normPath := normalizePath(path)
	statusStr := strconv.Itoa(status)
	
	requestCount.WithLabelValues(statusStr, method, normPath).Inc()
	requestLatency.WithLabelValues(statusStr, method, normPath).Observe(duration.Seconds())
}
