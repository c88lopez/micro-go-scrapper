package main

import (
	"github.com/gin-gonic/gin"
	"sarasa/libs/postgres"
	"sarasa/libs/signals"

	"sarasa/libs/configHandling"

	"sarasa/libs/errorHandling"
	"sarasa/libs/influxdb"
	"sarasa/libs/rabbitMQ"
	"sarasa/schemas"
)

var configuration schemas.Config

var influxSingleton influxdb.Client
var rabbitMQSingleton rabbitMQ.Client
var postgresSingleton postgres.Client

func init() {
	errorHandling.FailOnError(configHandling.LoadConfig(&configuration, "showcase_server"), "Failed to load configuration")

	/**
	 * InfluxDB Singleton
	 */
	errorHandling.LogOnError(
		influxSingleton.Init(configuration.Influx), "Could not initialize InfluxSingleton")

	/**
	* Postgres Singleton
	 */
	errorHandling.FailOnError(
		postgresSingleton.Init(configuration.Postgres), "Could not initialize PostgresSingleton")

	/**
	 * Signal handling
	 */
	signals.SignalHandler(rabbitMQSingleton, influxSingleton)
}

func main() {

	r := gin.Default()
	r.GET("/providers", func(c *gin.Context) {
		providers, err := postgresSingleton.GetProviders()
		errorHandling.FailOnError(err, "Could not get providers")

		c.JSON(200, providers)
	})

	r.Run()
}
