package main

import (
	"log"
	"os"
)

const (
	ConsumerIDBase = "kafkaduct"
	ConsumerGroup  = "kafkaduct"
)

const logPattern = log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile | log.LUTC

var logger = log.New(os.Stdout, "", logPattern)

func main() {

	appconfig := InitAppConfig()

	StartServer(appconfig)
}
