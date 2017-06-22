package metrics

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// HTTPLogger logs HTTP requests and collects request related metrics
type HTTPLogger struct {
	Handler         http.Handler
	RequestDuration prometheus.Histogram
}

func (l HTTPLogger) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	l.Handler.ServeHTTP(w, r)

	duration := time.Since(start)
	l.RequestDuration.Observe(float64(duration) / float64(time.Millisecond))
}
