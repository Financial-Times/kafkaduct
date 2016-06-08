package main

import (
	"crypto/tls"
	"os"

	"github.com/Shopify/sarama"

	kitlog "github.com/go-kit/kit/log"
)

type KafkaClient struct {
	producer sarama.SyncProducer
}

const (
	ConsumerIdBase = "kafkaduct"
	ConsumerGroup  = "kafkaduct"
)

var KafkaTopics = []string{"membership_users_v1", "commerce_domain_events"}

var log = kitlog.NewJSONLogger(os.Stdout)

func Run() {

	appconfig := InitAppConfig()

	// client := newKafkaClient(appconfig)
	// defer client.producer.Close()

	log.Log("Message", "Created client")

	StartServer(appconfig)
}

// Setup the Kafka client for producing and consumer messages.
// Use the specified configuration environment variables.
func newKafkaClient(config *AppConfig) *KafkaClient {
	var tlsConfig *tls.Config = nil
	if len(config.Kafka.ClientCert) != 0 {
		tlsConfig = config.createTlsConfig()
	}

	brokerAddrs := config.brokerAddresses()
	return &KafkaClient{
		producer: config.createKafkaProducer(config, brokerAddrs, tlsConfig),
	}
}

func main() {
	Run()
}
