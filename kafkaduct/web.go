package main

import (
	"encoding/json"
	"net/http"
	"io/ioutil"
	"strings"
	"compress/gzip"

	"github.com/Financial-Times/service-status-go/gtg"
	"github.com/Financial-Times/service-status-go/httphandlers"
	"github.com/Shopify/sarama"
	"github.com/gorilla/mux"

	"github.com/davecgh/go-spew/spew"

)

func StartServer(appConfig *AppConfig) {
	root := mux.NewRouter()

	root.HandleFunc(httphandlers.GTGPath, httphandlers.NewGoodToGoHandler(goodToGo))
	root.HandleFunc(httphandlers.BuildInfoPath, httphandlers.BuildInfoHandler)
	root.HandleFunc(httphandlers.BuildInfoPathDW, httphandlers.BuildInfoHandler)

	protected := root.NewRoute().Headers("X-Api-Key", appConfig.Web.APIKey).Subrouter()

	root.NewRoute().Headers("X-Api-Key", "").HandlerFunc(errorFunc(403, "Missing or Invalid Api Key")) // invalid api-key

	registerAPI(protected, appConfig)

	protected.NewRoute().HandlerFunc(http.NotFound) // valid api-key, but invalid URL

	http.Handle("/", root)


	logger.Printf("starting server on %s", appConfig.Web.Port)

	logger.Fatalf("%s", http.ListenAndServe(":"+appConfig.Web.Port, nil))
}

func registerAPI(router *mux.Router, appConfig *AppConfig) {

	var service service
	service.kafkaClient = newKafkaClient(appConfig)

	router.HandleFunc("/write", service.handle)
}

type service struct {
	kafkaClient sarama.SyncProducer
}

func unmarshal(r *http.Request) (res messageRequest, err error) {

	defer r.Body.Close()

	var jsonBody []byte

	if(strings.Contains(r.Header.Get("Content-Encoding"),"gzip")) {
		var gz *gzip.Reader
		gz, err = gzip.NewReader(r.Body)

		if err == nil {
			defer gz.Close()

			jsonBody, err = ioutil.ReadAll(gz)
		}

	} else {
		jsonBody, err = ioutil.ReadAll(r.Body)
	}

	if err == nil {
		err = json.Unmarshal(jsonBody, &res)
	}

	return res, err
}

func (s *service) handle(w http.ResponseWriter, r *http.Request) {

	requestId := r.Header.Get("X-Request-Id")

	logger.Printf("%s Handling write %s %s", requestId, r.Host, r.UserAgent())

	res, err := unmarshal(r)

	// Spew(res)
	if err != nil  {
		logger.Printf("%s ERROR unmarshalling: %s", requestId, err)
		http.Error(w, err.Error(), http.StatusBadRequest )
		return
	}

	logger.Printf("%s Messages topic=%s count=%d", requestId, res.Topic, len(res.Messages))

	for _, element := range res.Messages {

		logger.Printf("%s Processing %s %s", requestId, element.MessageType, element.MessageID)

		data, err := json.Marshal(element)

		if err != nil {
			logger.Printf("%s ERROR Failed to Marshal: %s", requestId, err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		partition, offset, err := s.kafkaClient.SendMessage(&sarama.ProducerMessage{
			Topic: res.Topic,
			Value: sarama.StringEncoder(data),
		})

		if err != nil {
			logger.Printf("%s ERROR Failed to send message error=%s", requestId, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return

		} else {
			logger.Printf("%s Sent %s %s with partition=%d offset=%d", requestId, element.MessageType, element.MessageID, partition, offset)
		}
	}
}

type message struct {
	Body               string `json:"body"`
	ContentType        string `json:"contentType"`
	MessageID          string `json:"messageId"`
	MessageTimestamp   string `json:"messageTimestamp"`
	MessageType        string `json:"messageType"`
	OriginHost         string `json:"originHost"`
	OriginHostLocation string `json:"originHostLocation"`
	OriginSystemID     string `json:"originSystemId"`
}

type messageRequest struct {
	Messages []message `json:"messages"`
	Topic    string    `json:"topic"`
}

func goodToGo() gtg.Status {
	return gtg.Status{GoodToGo: true}
}

func errorFunc(code int, message string) func(http.ResponseWriter, *http.Request) {
	return func(resp http.ResponseWriter, req *http.Request) { http.Error(resp, message, code) }
}

func Spew(a ...interface{}) {
	spew.Dump(a)
}
