package main

import (
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"http-handler-prometheus/handler"
	"log"
	"net/http"
)

// Simple app that demonstrates capabilities of the library
// Build and run test app
// run curl http://localhost:8081/test
// run curl http://localhost:8081/metrics | grep test
func main() {
	h, err := handler.NewHandler("/test", []string{"GET"}, actualHandler)
	if err != nil {
		log.Fatal(err)
	}
	hWithError, err := handler.NewHandler("/test/error", []string{"GET"}, actualHandlerWithError)
	router := mux.NewRouter()
	router.HandleFunc(h.Path, h.HandleRequest).Methods(h.Methods...)
	router.HandleFunc(hWithError.Path, hWithError.HandleRequest).Methods(hWithError.Methods...)
	http.Handle("/", router)
	router.Handle("/metrics", promhttp.Handler())

	server := http.Server{Addr: ":8081", Handler: nil}
	err = server.ListenAndServe()
	log.Fatal(err)
}

func actualHandlerWithError(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
}

func actualHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("Hello, world!"))
}
