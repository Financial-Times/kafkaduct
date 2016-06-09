package main

import (
	"crypto/tls"
	"crypto/x509"
	"log"
	"net/url"
	"strings"

	"github.com/Shopify/sarama"
	"github.com/joeshaw/envdecode"
)

type AppConfig struct {
	Env  string `env:"ENV,default=localhost"`
	Dyno string `env:"DYNO,required"`

	Kafka struct {
		Url           string `env:"KAFKA_URL,required"`
		TrustedCert   string `env:"KAFKA_TRUSTED_CERT,required"`
		ClientCertKey string `env:"KAFKA_CLIENT_CERT_KEY,required"`
		ClientCert    string `env:"KAFKA_CLIENT_CERT,required"`
	}

	Web struct {
		Port   string `env:"PORT,required"`
		ApiKey string `env:"API_KEY,required"`
	}
}

func InitAppConfig() *AppConfig {
	appconfig := AppConfig{}
	envdecode.MustDecode(&appconfig)
	return &appconfig
}

type KafkaClient struct {
	producer sarama.SyncProducer
}

func newKafkaClient(appConfig *AppConfig) sarama.SyncProducer {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 10
	tlsConfig := appConfig.createTLSConfig()

	if tlsConfig != nil {
		config.Net.TLS.Config = tlsConfig
		config.Net.TLS.Enable = true
	}

	producer, err := sarama.NewSyncProducer(appConfig.brokerAddresses(), config)
	if err != nil {
		log.Fatalf("Failed to start Sarama producer error=%s", err)
	}

	return producer
}

func (ac *AppConfig) createTLSConfig() *tls.Config {
	cert, err := tls.X509KeyPair([]byte(ac.Kafka.ClientCert), []byte(ac.Kafka.ClientCertKey))
	if err != nil {
		panic(err)
	}
	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM([]byte(ac.Kafka.TrustedCert))

	tlsConfig := &tls.Config{
		Certificates:       []tls.Certificate{cert},
		InsecureSkipVerify: true,
		RootCAs:            certPool,
	}
	tlsConfig.BuildNameToCertificate()
	return tlsConfig
}

func (ac *AppConfig) brokerAddresses() []string {
	urls := strings.Split(ac.Kafka.Url, ",")
	addrs := make([]string, len(urls))
	for i, v := range urls {
		u, err := url.Parse(v)
		if err != nil {
			panic(err)
		}
		addrs[i] = u.Host
		log.Printf("broker=%s", u.Host)
	}
	return addrs
}
