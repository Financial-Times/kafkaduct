package main

import (
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

	log.Log("Message", "Created client")

	StartServer(appconfig)
}

func main() {
	Run()
}
