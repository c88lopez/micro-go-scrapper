package providersCommon

import (
	"encoding/json"
	"log"
	"time"

	"sarasa/libs/signals"

	"sarasa/libs/configHandling"

	"sarasa/libs/influxdb"
	"sarasa/libs/rabbitMQ"
	"sarasa/schemas"

	"github.com/PuerkitoBio/goquery"
	"github.com/google/uuid"
	"github.com/streadway/amqp"
	"sarasa/libs/errorHandling"
)

var configuration schemas.Config

var influxSingleton influxdb.Client
var rabbitMQSingleton rabbitMQ.Client

var providerDetailsLink chan string
var results chan schemas.Provider

type CustomGetDetailsFn func(doc *goquery.Document, source schemas.Source) schemas.Provider
type CustomGetDetailsLinkFn func(s *goquery.Selection) string

type ProviderProcessor struct {
	ServiceName string
	Selector    string
	CustomGetDetailsFn
	CustomGetDetailsLinkFn
}

func (pd ProviderProcessor) scrapWorker() {
	for link := range providerDetailsLink {
		details, err := GetDetails(link, configuration.Provider.Source, pd.CustomGetDetailsFn)

		if err != nil {
			log.Printf("Failed getting details - error: %s", err)
			continue
		}

		details.Source = configuration.Provider.Source.Url

		results <- details
	}
}

func (pd ProviderProcessor) initialize() {
	errorHandling.FailOnError(
		configHandling.LoadConfig(&configuration, pd.ServiceName), "Could not get configuration")

	log.Printf("Configuration fetched %v", configuration)

	/**
	 * RabbitMQ Singleton
	 */
	errorHandling.FailOnError(
		rabbitMQSingleton.Init(configuration.RabbitMQ), "Could not initialize rabbitMQSingleton")

	/**
	 * InfluxDB Singleton
	 */
	errorHandling.LogOnError(
		influxSingleton.Init(configuration.Influx), "Could not initialize InfluxSingleton")

	/**
	 * Signal handling
	 */
	signals.SignalHandler(rabbitMQSingleton, influxSingleton)

	/*
	 * Workers pool
	 */
	providerDetailsLink = make(chan string, 200)
	results = make(chan schemas.Provider, 200)

	for w := 1; w <= configuration.Provider.ScrapWorkersCount; w++ {
		go pd.scrapWorker()
	}
}

func (pd ProviderProcessor) Run() {

	pd.initialize()

	defer func() {
		errorHandling.LogOnError(
			rabbitMQSingleton.Connection.Close(), "Error closing rabbitmq connection - error: %s")
	}()
	defer func() {
		errorHandling.LogOnError(
			rabbitMQSingleton.Channel.Close(), "Error closing rabbitmq channel")
	}()

	var influxTags map[string]string
	var influxFields map[string]interface{}

	influxTags = map[string]string{"run_uuid": uuid.New().String(), "source": configuration.Provider.Source.Url}

	/**
	 * Declaring queue to produce
	 */
	providersQueue, err := rabbitMQSingleton.Channel.QueueDeclare(
		"providers", false, false, false, false, nil)
	errorHandling.FailOnError(err, "Failed to declare a queue")

	/**
	 * Declaring queue to consume
	 */
	errorHandling.FailOnError(
		rabbitMQSingleton.Channel.ExchangeDeclare(
			"refresh", "fanout", true, false, false, false, nil),
		"Failed to declare a queue")

	refreshQueue, err := rabbitMQSingleton.Channel.QueueDeclare(
		"", false, false, false, false, nil)
	errorHandling.FailOnError(err, "Failed to declare a queue")

	errorHandling.FailOnError(
		rabbitMQSingleton.Channel.QueueBind(refreshQueue.Name, "", "refresh", false, nil),
		"Failed to bind queue")

	refreshQueueMessages, err := rabbitMQSingleton.Channel.Consume(
		refreshQueue.Name, "", true, false, false, false, nil,
	)
	errorHandling.FailOnError(err, "Failed to create consumer")

	forever := make(chan bool)

	go func() {
		for range refreshQueueMessages {
			startTime := time.Now()

			providers, err := GetElements(
				configuration.Provider.Source.Url, pd.Selector,
				configuration.Provider.ProvidersCount,
				func(s *goquery.Selection) string {
					return pd.CustomGetDetailsLinkFn(s)
				},
				providerDetailsLink, results)
			errorHandling.FailOnError(err, "Failed to get providers")

			body, err := json.Marshal(providers)
			errorHandling.FailOnError(err, "Failed to marshal providers")

			errorHandling.FailOnError(
				rabbitMQSingleton.Channel.Publish(
					"", providersQueue.Name, false, false, amqp.Publishing{
						ContentType: "text/json",
						Body:        body,
					}), "Failed to publish provider message")

			providersCount := len(providers)

			influxFields = map[string]interface{}{
				"elapsed":      time.Since(startTime).Milliseconds(),
				"providerSent": providersCount,
				"success":      true,
			}

			go errorHandling.LogOnError(
				influxSingleton.Send("provider_process", influxTags, influxFields),
				"Could not write to InfluxDB")

			log.Printf("End. Got %d providers in %s\n", providersCount, time.Since(startTime))
		}
	}()

	log.Printf("Waiting for messages. To exit press CTRL+C")
	<-forever
}
