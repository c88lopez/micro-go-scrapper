package configHandling

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"sarasa/schemas"
)

func LoadConfig(c *schemas.Config, serviceName string) error {
	configServerSchema := os.Getenv("CONFIG_SERVER_SCHEMA")
	configServerHost := os.Getenv("CONFIG_SERVER_HOST")
	configServerPort := os.Getenv("CONFIG_SERVER_PORT")

	if configServerSchema == "" ||
		configServerHost == "" ||
		configServerPort == "" {
		return fmt.Errorf("need to define config server schema, host and port")
	}

	response, err := http.Get(
		fmt.Sprintf(
			"%s://%s:%s/?serviceName=%s",
			configServerSchema, configServerHost, configServerPort, serviceName),
	)
	if err != nil {
		return err
	}

	return json.NewDecoder(response.Body).Decode(c)
}
