package handler

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewHandler(t *testing.T) {
	h1, err := NewHandler("/path", []string{"GET"}, func(w http.ResponseWriter, r *http.Request) {})
	assert.NoError(t, err, "NewHandler did not produce any error")
	h2, err := NewHandler("/path", []string{"POST"}, func(w http.ResponseWriter, r *http.Request) {})
	assert.NoError(t, err, "Second invocation of NewHandler did not produce any error")
	h1.Cleanup()
	h2.Cleanup()
}

func TestMetricsHandler_HandleRequest(t *testing.T) {
	h, err := NewHandler("/path", []string{"GET"}, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	assert.NoError(t, err, "NewHandler did not produce any error")
	assert.HTTPSuccess(t, h.handler, "GET", "/path", nil, "Request should return status 'OK'")
	h.Cleanup()
	h, err = NewHandler("/path", []string{"GET"}, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("Hello, world!"))
	})
	assert.NoError(t, err, "NewHandler did not produce any error")
	rr := httptest.NewRecorder()
	request, err := http.NewRequest("GET", "/path", nil)
	assert.NoError(t, err, "Request has been created")
	h.HandleRequest(rr, request)
	assert.Equal(t, http.StatusOK, rr.Code, "Request should return status 'OK'")
	body, err := rr.Body.ReadString(byte('\n'))
	if err != nil {
		assert.IsType(t, io.EOF, err, "Body has been read")
	} else {
		assert.NoError(t, err, "Body has been read")
	}
	assert.Equal(t, "Hello, world!", body, "Handler return correct body")
	h.Cleanup()
}

func TestMetricsHandler_Cleanup(t *testing.T) {
	h, err := NewHandler("/path", []string{"GET", "POST"}, func(w http.ResponseWriter, r *http.Request) {})
	assert.NoError(t, err, "NewHandler did not produce any error")
	h.Cleanup()
	err = prometheus.DefaultRegisterer.Register(h.TotalErrors)
	assert.NoError(t, err, "Collector was successfully unregistered")
	err = prometheus.DefaultRegisterer.Register(h.TotalRequests)
	assert.NoError(t, err, "Collector was successfully unregistered")
	err = prometheus.DefaultRegisterer.Register(h.ResponseTime)
	assert.NoError(t, err, "Collector was successfully unregistered")
}

func TestMetricsHandlerResponseProxy_WriteHeader(t *testing.T) {
	rr := httptest.NewRecorder()
	proxy := newProxy(rr)
	proxy.WriteHeader(http.StatusBadRequest)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestMetricsHandlerResponseProxy_Write(t *testing.T) {
	rr := httptest.NewRecorder()
	proxy := newProxy(rr)
	_, err := proxy.Write([]byte("Hello, world!"))
	assert.NoError(t, err, "Proxy was able to write bytes")
	response, err := rr.Body.ReadString(byte('\n'))
	if err != nil {
		assert.IsType(t, io.EOF, err, "Should error happens, it's EOF")
	} else {
		assert.NoError(t, err)
	}
	assert.Equal(t, "Hello, world!", response)

	erProxy := ResponseWriterWithErrorOnWrite{}
	proxy = newProxy(erProxy)
	_, err = proxy.Write(nil)
	assert.Error(t, err, "Proxy caught an error")
	assert.NotEqual(t, http.StatusOK, proxy.statusCode, "Proxy status should be not HTTP OK")
}

func TestMetricsHandlerResponseProxy_Header(t *testing.T) {
	rr := httptest.NewRecorder()
	proxy := newProxy(rr)
	proxy.WriteHeader(http.StatusBadRequest)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.NotNil(t, proxy.Header())
}

type ResponseWriterWithErrorOnWrite struct{}

func (p ResponseWriterWithErrorOnWrite) Header() http.Header {
	return nil
}

func (p ResponseWriterWithErrorOnWrite) Write([]byte) (int, error) {
	return 0, io.EOF
}

func (p ResponseWriterWithErrorOnWrite) WriteHeader(statusCode int) {}
