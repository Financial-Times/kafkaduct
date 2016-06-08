package main

import (
	"crypto/tls"
	"crypto/x509"
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
		TrustedCert   string `env:"KAFKA_TRUSTED_CERT"`
		ClientCertKey string `env:"KAFKA_CLIENT_CERT_KEY"`
		ClientCert    string `env:"KAFKA_CLIENT_CERT"`
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

func (ac *AppConfig) createKafkaProducer(appConfig *AppConfig, brokers []string, tc *tls.Config) sarama.SyncProducer {
	config := sarama.NewConfig()

	config.Net.TLS.Config = tc
	config.Net.TLS.Enable = tc != nil
	config.Producer.Return.Errors = true
	config.ClientID = strings.Join([]string{ConsumerIdBase, appConfig.Env, appConfig.Dyno}, "-")

	err := config.Validate()
	if err != nil {
		panic(err)
	}
	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		panic(err)
	}

	return producer
}

func (ac *AppConfig) createTlsConfig() *tls.Config {
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
		log.Log("broker", u.Host)
	}
	return addrs
}
