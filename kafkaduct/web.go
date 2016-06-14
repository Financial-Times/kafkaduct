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

func unmarshal(r *http.Request) (messageRequest, error) {
	res := messageRequest{}

	defer r.Body.Close()

	var jsonBody []byte
	var err error

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

	if err != nil {
		panic(err)
	}

	// Spew(jsonBody)

	err = json.Unmarshal(jsonBody, &res)

	return res, err
}

func (s *service) handle(w http.ResponseWriter, r *http.Request) {

	logger.Printf("Handling write %s %s", r.Host, r.UserAgent())

	res, err := unmarshal(r)

	// Spew(res)


	if err != nil  {
		logger.Printf("Error unmarshalling: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	logger.Printf("Messages topic=%s count=%d", res.Topic, len(res.Messages))

	for _, element := range res.Messages {

		logger.Printf("Processing %s %s", element.MessageType, element.MessageID)

		data, err1 := json.Marshal(element)

		if err1 != nil {
			logger.Printf("ERROR Failed to Marshal: %s", err1)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		partition, offset, err2 := s.kafkaClient.SendMessage(&sarama.ProducerMessage{
			Topic: res.Topic,
			Value: sarama.StringEncoder(data),
		})

		if err2 != nil {
			w.WriteHeader(http.StatusInternalServerError)
			logger.Printf("ERROR Failed to store your data error=%s", err2)
		} else {
			logger.Printf("Stored %s %s with partition=%d offset=%d", element.MessageType, element.MessageID, partition, offset)
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
