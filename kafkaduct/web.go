package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Financial-Times/service-status-go/gtg"
	"github.com/Financial-Times/service-status-go/httphandlers"
	"github.com/Shopify/sarama"
	"github.com/gorilla/mux"
)

func StartServer(appConfig *AppConfig) {
	root := mux.NewRouter()

	root.HandleFunc(httphandlers.GTGPath, httphandlers.NewGoodToGoHandler(goodToGo))
	root.HandleFunc(httphandlers.BuildInfoPath, httphandlers.BuildInfoHandler)
	root.HandleFunc(httphandlers.BuildInfoPathDW, httphandlers.BuildInfoHandler)

	protected := root.NewRoute().Headers("X-Api-Key", appConfig.Web.APIKey).Subrouter()

	root.NewRoute().Headers("X-Api-Key", "").HandlerFunc(error(403, "Missing or Invalid Api Key")) // invalid api-key

	registerAPI(protected, appConfig)

	protected.NewRoute().HandlerFunc(http.NotFound) // valid api-key, but invalid URL

	http.Handle("/", root)

	fmt.Printf("starting server on %s", appConfig.Web.Port)

	logger.Fatalf("%s", http.ListenAndServe(":"+appConfig.Web.Port, root))
}

func registerAPI(router *mux.Router, appConfig *AppConfig) {

	var service service
	service.kafkaClient = newKafkaClient(appConfig)

	router.HandleFunc("/write", service.handle)
}

type service struct {
	kafkaClient sarama.SyncProducer
}

func unmarshal(r *http.Request) messageRequest {
	res := messageRequest{}

	buf := new(bytes.Buffer)
	buf.ReadFrom(r.Body)

	json.Unmarshal(buf.Bytes(), &res)

	return res
}

func (s *service) handle(w http.ResponseWriter, r *http.Request) {

	res := unmarshal(r)

	for _, element := range res.Messages {

		data, err1 := json.Marshal(element)

		if err1 != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Printf("ERROR Failed to Marshal: %s", err1)
		}

		partition, offset, err := s.kafkaClient.SendMessage(&sarama.ProducerMessage{
			Topic: res.Topic,
			Value: sarama.StringEncoder(data),
		})

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Printf("ERROR Failed to store your data error=%s", err)
		} else {
			fmt.Printf("Stored with partition=%d offset=%d", partition, offset)
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

func error(code int, message string) func(http.ResponseWriter, *http.Request) {
	return func(resp http.ResponseWriter, req *http.Request) { http.Error(resp, message, code) }
}
