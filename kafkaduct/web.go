package main

import (
	"bytes"
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

	protected := root.NewRoute().Headers("X-Api-Key", appConfig.Web.ApiKey).Subrouter()

	root.NewRoute().Headers("X-Api-Key", "").HandlerFunc(error(403, "Missing or Invalid Api Key")) // invalid api-key

	registerApi(protected, appConfig)

	protected.NewRoute().HandlerFunc(http.NotFound) // valid api-key, but invalid URL

	http.Handle("/", root)

	log.Log("FATAL", http.ListenAndServe(":"+appConfig.Web.Port, root))
}

func registerApi(router *mux.Router, appConfig *AppConfig) {

	kafkaClient := newKafkaClient(appConfig)

	router.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {

		buf := new(bytes.Buffer)
		buf.ReadFrom(r.Body)
		s := buf.String()

		partition, offset, err := kafkaClient.SendMessage(&sarama.ProducerMessage{
			Topic: "test-write",
			Value: sarama.StringEncoder(s),
		})

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Failed to store your data:, %s", err)
		} else {
			// The tuple (topic, partition, offset) can be used as a unique identifier
			// for a message in a Kafka cluster.
			fmt.Fprintf(w, "Your data is stored with unique identifier important/%d/%d", partition, offset)
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
