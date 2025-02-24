package influxdb

import (
	"log"

	"sarasa/schemas"

	influxDbClient "github.com/influxdata/influxdb1-client/v2"
)

type Client struct {
	httpClient influxDbClient.Client
}

var enabled bool
var database string

func (influxDB *Client) Init(ic schemas.InfluxConfig) error {
	enabled = true
	if !ic.Enabled {
		log.Println("InfluxDB client disabled.")

		enabled = false
		return nil
	}

	log.Println("Initializing InfluxDB client...")

	var err error

	influxDB.httpClient, err = influxDbClient.NewHTTPClient(influxDbClient.HTTPConfig{
		Addr: ic.Url,
	})

	if err != nil {
		return err
	}

	database = ic.Database

	return nil
}

func (influxDb Client) Close() error {
	if influxDb.httpClient == nil {
		return nil
	}

	return influxDb.httpClient.Close()
}

func (rabbitmq Client) String() string {
	return "InfluxDB"
}

func (influxDB *Client) Send(pointName string, tags map[string]string, fields map[string]interface{}) error {
	if !enabled {
		return nil
	}

	log.Println("Sending to InfluxDB...")

	p, err := influxDbClient.NewPoint(pointName, tags, fields)
	if err != nil {
		return err
	}

	bp, err := influxDbClient.NewBatchPoints(influxDbClient.BatchPointsConfig{Database: database})
	if err != nil {
		return err
	}

	bp.AddPoint(p)

	err = influxDB.httpClient.Write(bp)
	if err != nil {
		return err
	}

	return nil
}
