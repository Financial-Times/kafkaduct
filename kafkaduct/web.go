package main

import (
	"bytes"
	"encoding/json"
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

	protected := root.NewRoute().Headers("X-Api-Key", appConfig.Web.ApiKey).Subrouter()

	root.NewRoute().Headers("X-Api-Key", "").HandlerFunc(error(403, "Missing or Invalid Api Key")) // invalid api-key

	registerAPI(protected, appConfig)

	protected.NewRoute().HandlerFunc(http.NotFound) // valid api-key, but invalid URL

	http.Handle("/", root)

	log.Log("FATAL", http.ListenAndServe(":"+appConfig.Web.Port, root))
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

func registerAPI(router *mux.Router, appConfig *AppConfig) {

	kafkaClient := newKafkaClient(appConfig)

	router.HandleFunc("/write", func(w http.ResponseWriter, r *http.Request) {

		res := messageRequest{}

		buf := new(bytes.Buffer)
		buf.ReadFrom(r.Body)

		json.Unmarshal(buf.Bytes(), &res)

		log.Log("DEBUG", "Request", res)

		for _, element := range res.Messages {

			data, err1 := json.Marshal(element)

			if err1 != nil {
				w.WriteHeader(http.StatusInternalServerError)
				log.Log("FATAL", "Failed to Marshal", err1)
			}

			partition, offset, err := kafkaClient.SendMessage(&sarama.ProducerMessage{
				Topic: res.Topic,
				Value: sarama.StringEncoder(data),
			})

			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				log.Log("FATAL", "Failed to store your data", err)
			} else {
				// The tuple (topic, partition, offset) can be used as a unique identifier
				// for a message in a Kafka cluster.
				log.Log(w, "Your data is stored with unique identifier important", partition, offset)
			}
		}
	})
}

// Setup the Kafka client for producing and consumer messages.
// Use the specified configuration environment variables.
func newKafkaClient(appConfig *AppConfig) sarama.SyncProducer {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll // Wait for all in-sync replicas to ack the message
	config.Producer.Retry.Max = 10                   // Retry up to 10 times to produce the message
	tlsConfig := appConfig.createTlsConfig()

	if tlsConfig != nil {
		config.Net.TLS.Config = tlsConfig
		config.Net.TLS.Enable = true
	}

	// On the broker side, you may want to change the following settings to get
	// stronger consistency guarantees:
	// - For your broker, set `unclean.leader.election.enable` to false
	// - For the topic, you could increase `min.insync.replicas`.

	producer, err := sarama.NewSyncProducer(appConfig.brokerAddresses(), config)
	if err != nil {
		log.Log("FATAL", "Failed to start Sarama producer:", err)
	}

	return producer
}

func goodToGo() gtg.Status {
	return gtg.Status{GoodToGo: true}
}

func error(code int, message string) func(http.ResponseWriter, *http.Request) {
	return func(resp http.ResponseWriter, req *http.Request) { http.Error(resp, message, code) }
}
