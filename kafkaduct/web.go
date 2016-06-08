package main

import (
	"bytes"
	"net/http"

	"github.com/Financial-Times/service-status-go/gtg"
	"github.com/Financial-Times/service-status-go/httphandlers"
	"github.com/gorilla/mux"
)

func StartServer(appConfig *AppConfig) {
	root := mux.NewRouter()

	root.HandleFunc(httphandlers.GTGPath, httphandlers.NewGoodToGoHandler(goodToGo))
	root.HandleFunc(httphandlers.BuildInfoPath, httphandlers.BuildInfoHandler)
	root.HandleFunc(httphandlers.BuildInfoPathDW, httphandlers.BuildInfoHandler)

	protected := root.NewRoute().Headers("X-Api-Key", appConfig.Web.ApiKey).Subrouter()

	root.NewRoute().Headers("X-Api-Key", "").HandlerFunc(error(403, "Missing or Invalid Api Key")) // invalid api-key

	registerApi(protected)

	protected.NewRoute().HandlerFunc(http.NotFound) // valid api-key, but invalid URL

	http.Handle("/", root)

	log.Log("FATAL", http.ListenAndServe(appConfig.Web.Port, protected))
}

func registerApi(router *mux.Router) {
	router.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		buf := new(bytes.Buffer)
		buf.ReadFrom(r.Body)
		s := buf.String()
		log.Log("EVENT", s)
	})
}

func goodToGo() gtg.Status {
	return gtg.Status{GoodToGo: true}
}

func error(code int, message string) func(http.ResponseWriter, *http.Request) {
	return func(resp http.ResponseWriter, req *http.Request) { http.Error(resp, message, code) }
}
