package main

import (
	"encoding/json"
	"log"
	"time"

	"sarasa/libs/signals"

	"sarasa/libs/configHandling"

	"github.com/streadway/amqp"

	"github.com/google/uuid"

	"sarasa/libs/errorHandling"
	"sarasa/libs/influxdb"
	"sarasa/libs/postgres"
	"sarasa/libs/rabbitMQ"
	"sarasa/schemas"
)

var configuration schemas.Config

var postgresSingleton postgres.Client
var influxSingleton influxdb.Client
var rabbitMQSingleton rabbitMQ.Client

var availableZones map[string]int
var availableSources map[string]int

func init() {
	var err error

	errorHandling.FailOnError(configHandling.LoadConfig(&configuration, "core"), "Failed to load configuration")

	/**
	 * Postgres Singleton
	 */
	errorHandling.FailOnError(
		postgresSingleton.Init(configuration.Postgres), "Could not initialize PostgresSingleton")

	availableZones, err = postgresSingleton.GetZones()
	errorHandling.FailOnError(err, "Could not load zones")

	availableSources, err = postgresSingleton.GetSources()
	errorHandling.FailOnError(err, "Could not load sources")

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
	signals.SignalHandler(rabbitMQSingleton, postgresSingleton, influxSingleton)
}

func main() {
	var err error

	forever := make(chan bool)

	influxTags := map[string]string{"run_uuid": uuid.New().String()}
	influxFields := map[string]interface{}{}

	/**
	 * Declaring queue to produce
	 */
	errorHandling.FailOnError(
		rabbitMQSingleton.Channel.ExchangeDeclare(
			"refresh", "fanout", true, false, false, false, nil),
		"Failed to declare a queue")

	errorHandling.FailOnError(
		rabbitMQSingleton.Channel.Publish(
			"refresh", "", false, false, amqp.Publishing{Body: []byte{}}),
		"Failed to publish refresh message",
	)

	/**
	 * Declaring queue to consume
	 */
	providersQueue, err := rabbitMQSingleton.Channel.QueueDeclare(
		"providers", false, false, false, false, nil,
	)
	errorHandling.FailOnError(err, "Failed to declare queue")

	messages, err := rabbitMQSingleton.Channel.Consume(
		providersQueue.Name, "", true, false, false, false, nil,
	)
	errorHandling.FailOnError(err, "Failed to start consumer")

	go func() {
		var err error
		var providers []schemas.Provider
		var sanitizedProviders []schemas.Provider

		for d := range messages {
			startTime := time.Now()

			err = json.Unmarshal(d.Body, &providers)
			if err != nil {
				errorHandling.LogOnError(err, "Fail to unmarshal message")
				continue
			}

			receivedProvidersCount := len(providers)

			log.Printf("Received %d providers", receivedProvidersCount)

			invalidProvidersCount := 0
			for i := 0; i < receivedProvidersCount; i++ {
				if isProviderInvalid(providers[i]) {
					invalidProvidersCount++
					continue
				}

				sanitizedProviders = append(sanitizedProviders, providers[i])
			}

			log.Printf("Detected %d invalid providers", invalidProvidersCount)

			if len(sanitizedProviders) == 0 {
				continue
			}

			influxTags["source"] = sanitizedProviders[0].Source
			influxFields["providersCount"] = len(sanitizedProviders)

			err = postgresSingleton.SaveProvidersList(sanitizedProviders, availableZones, availableSources)
			errorHandling.FailOnError(err, "Error saving providers")

			sanitizedProviders = nil

			availableZones, err = postgresSingleton.GetZones()
			errorHandling.FailOnError(err, "Could not load zones")

			availableSources, err = postgresSingleton.GetSources()
			errorHandling.FailOnError(err, "Could not load sources")

			influxFields = map[string]interface{}{
				"receivedProvidersCount": receivedProvidersCount,
				"elapsed":                time.Since(startTime).Milliseconds(),
				"success":                true,
			}

			go errorHandling.LogOnError(
				influxSingleton.Send("core_process", influxTags, influxFields),
				"Could not write to InfluxDB")

			log.Printf("End. Elapsed time %s\n", time.Since(startTime))
		}
	}()

	log.Printf("Waiting for messages. To exit press CTRL+C")
	<-forever
}

func isProviderInvalid(provider schemas.Provider) bool {
	status := false
	reason := ""

	if len(provider.Phone) != 10 {
		reason = "phone.length != 10"
		status = true
	}

	if provider.Phone == "vacaciones" {
		reason = "phone = 'vacaciones"
		status = true
	}

	if provider.Name == "" {
		reason = "name empty"
		status = true
	}

	if provider.Place == "" {
		reason = "place empty"
		status = true
	}

	if len(provider.Pics) == 0 {
		reason = "no pics"
		status = true
	}

	if status {
		log.Printf("invalid provider: %#v - %s", provider, reason)
	}

	return status
}
