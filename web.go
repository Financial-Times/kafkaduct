package kafkaduct

import (
	"github.com/Financial-Times/service-status-go/gtg"
	"github.com/Financial-Times/service-status-go/httphandlers"
	"github.com/gorilla/mux"
	"net/http"
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

	go func() {
		http.ListenAndServe(appConfig.Web.Port, nil)
	}()
}

func registerApi(router *mux.Router) {
	// nothing here yet
}

func goodToGo() gtg.Status {
	return gtg.Status{GoodToGo: true}
}

func error(code int, message string) func(http.ResponseWriter, *http.Request) {
	return func(resp http.ResponseWriter, req *http.Request) { http.Error(resp, message, code) }
}
