package metrics

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
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

func PrometheusHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		c.Next()

		statusCode := strconv.Itoa(c.Response().StatusCode())
		duration := float64(time.Since(start)) / float64(time.Second)

		counter.WithLabelValues(statusCode).Inc()
		histogram.WithLabelValues(c.Method(), statusCode).Observe(duration)

		return nil
	}
}
