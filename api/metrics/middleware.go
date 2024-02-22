package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"net/http"
	"strconv"
	"time"
)

var (
	counter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "counter",
			Help: "Counter Metrics",
		},
		[]string{"name"},
	)

	histogram = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "timer",
		Help: "Histogram Metrics",
	}, []string{"name", "status"})
)

// PrometheusHandler is a middleware function that instruments HTTP request handling for Prometheus metrics.
func PrometheusHandler(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Call the next handler function in the chain with the custom ResponseWriter
		customResponseWriter := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next(customResponseWriter, r)

		statusCode := strconv.Itoa(customResponseWriter.statusCode)
		duration := float64(time.Since(start)) / float64(time.Second)

		counter.WithLabelValues(statusCode).Inc()
		histogram.WithLabelValues(r.Method, statusCode).Observe(duration)
	}
}

// Custom ResponseWriter to capture the status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}
