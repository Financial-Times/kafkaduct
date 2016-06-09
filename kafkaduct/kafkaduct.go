package main

import (
	"log"
	"os"
)

const (
	ConsumerIDBase = "kafkaduct"
	ConsumerGroup  = "kafkaduct"
)

var logger = log.New(os.Stdout, "", 0)

func main() {
	appconfig := InitAppConfig()

	logger.SetOutput(os.Stdout)

	StartServer(appconfig)
}
