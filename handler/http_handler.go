package handler

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"net/http"
	"sort"
	"strings"
	"time"
)

// MetricsHandler is a wrapper around http.handler which automatically creates, registers and updates prometheuse metrics
// for http requests, such as total requests, total errors and requests summary.
type MetricsHandler struct {
	Path          string
	Methods       []string
	handler       func(w http.ResponseWriter, r *http.Request)
	TotalRequests prometheus.Counter
	TotalErrors   prometheus.Counter
	ResponseTime  prometheus.Summary
}

// NewHandler creates new handler with specified request path, methods and callback function to you actual handler
func NewHandler(path string, methods []string, f func(w http.ResponseWriter, r *http.Request)) (handler *MetricsHandler, err error) {
	sort.Strings(methods)
	methodsFlattened := strings.ToLower(strings.Join(methods, "_"))
	methodsCommaFlattened := strings.Join(methods, ", ")
	pathFlattened := strings.ToLower(strings.Join(strings.FieldsFunc(path, func(r rune) bool {
		return r == rune('/')
	}), "_"))
	handler = &MetricsHandler{
		Path:    path,
		Methods: methods,
		TotalRequests: promauto.NewCounter(prometheus.CounterOpts{
			Name: fmt.Sprintf("%s_%s_%s", pathFlattened, methodsFlattened, "total_requests"),
			Help: fmt.Sprintf("%s %s %s", path, methodsCommaFlattened, "total number of requests"),
		}),
		TotalErrors: promauto.NewCounter(prometheus.CounterOpts{
			Name: fmt.Sprintf("%s_%s_%s", pathFlattened, methodsFlattened, "total_errors"),
			Help: fmt.Sprintf("%s %s %s", path, methodsCommaFlattened, "total number of errors"),
		}),
		ResponseTime: promauto.NewSummary(prometheus.SummaryOpts{
			Name: fmt.Sprintf("%s_%s_%s", pathFlattened, methodsFlattened, "response_time_ms"),
			Help: fmt.Sprintf("%s %s %s", path, methodsCommaFlattened, "response time in ms"),
		}),
		handler: f}
	return handler, nil
}

// HandleRequest handles request and updates metrics. Original request function will be called to serve actual request
// If actual request does not returns status code 2xx then "total errors" counter will be updated
// "total requests" counter counts all requests
func (h MetricsHandler) HandleRequest(writer http.ResponseWriter, request *http.Request) {
	startTime := time.Now()
	proxy := newProxy(writer)
	defer func() {
		duration := time.Now().Sub(startTime)
		h.ResponseTime.Observe(float64(duration.Nanoseconds()) / 1000000.0)
	}()
	h.handler(proxy, request)
	if proxy.err != nil || proxy.statusCode >= 300 || proxy.statusCode < 200 {
		h.TotalErrors.Inc()
	}
	h.TotalRequests.Inc()
}

// Cleanup is used to unregister prometheus counters
func (h MetricsHandler) Cleanup() {
	prometheus.Unregister(h.TotalRequests)
	prometheus.Unregister(h.TotalErrors)
	prometheus.Unregister(h.ResponseTime)
}

func newProxy(w http.ResponseWriter) *metricsHandlerResponseProxy {
	return &metricsHandlerResponseProxy{
		actual:      w,
		err:         nil,
		statusCode:  -1,
		wroteHeader: false,
	}
}

type metricsHandlerResponseProxy struct {
	actual      http.ResponseWriter
	err         error
	statusCode  int
	wroteHeader bool
}

func (rr *metricsHandlerResponseProxy) WriteHeader(statusCode int) {
	rr.statusCode, rr.wroteHeader = statusCode, true
	rr.actual.WriteHeader(statusCode)
}

func (rr *metricsHandlerResponseProxy) Header() http.Header {
	return rr.actual.Header()
}

func (rr *metricsHandlerResponseProxy) Write(bytes []byte) (int, error) {
	n, err := rr.actual.Write(bytes)
	if rr.err != nil {
		rr.err = err
	}
	if !rr.wroteHeader && err == nil {
		rr.WriteHeader(http.StatusOK)
	}
	return n, err
}
