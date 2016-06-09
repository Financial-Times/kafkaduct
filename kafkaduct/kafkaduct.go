package main

import (
	"log"
	"os"
)

const (
	ConsumerIDBase = "kafkaduct"
	ConsumerGroup  = "kafkaduct"
)

func main() {
	appconfig := InitAppConfig()

	log.SetOutput(os.Stdout)

	StartServer(appconfig)
}
