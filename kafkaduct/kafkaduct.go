package main

const (
	ConsumerIDBase = "kafkaduct"
	ConsumerGroup  = "kafkaduct"
)

func main() {
	appconfig := InitAppConfig()

	StartServer(appconfig)
}
