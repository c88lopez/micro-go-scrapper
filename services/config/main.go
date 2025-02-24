package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"sarasa/libs/errorHandling"
)

var configurationMap map[string]map[string]interface{}

func init() {
	configurationMap = make(map[string]map[string]interface{})

	errorHandling.FailOnError(getConfig(), "Failed getting configuration")
}

func main() {
	r := gin.Default()
	r.GET("/", func(c *gin.Context) {
		serviceName := c.Query("serviceName")

		c.JSON(200, configurationMap[serviceName])
	})

	r.Run(":8090")
}

func index(w http.ResponseWriter, req *http.Request) {
	errorHandling.LogOnError(req.ParseForm(), "Failed to parse request form")

	serviceName := req.Form.Get("serviceName")

	configJSON, err := json.Marshal(configurationMap[serviceName])
	errorHandling.FailOnError(err, "Failed to write to response")

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	_, err = w.Write(configJSON)
	errorHandling.FailOnError(err, "Failed to write to response")
}

func getConfig() error {
	file, err := os.Open("config.json")
	if err != nil {
		return err
	}

	defer func() {
		errorHandling.LogOnError(
			file.Close(), "Error closing config file")
	}()

	var config struct {
		Services     []string            `json:"services"`
		Dependencies map[string][]string `json:"dependencies"`
		Aliases      map[string]string   `json:"aliases"`
	}

	err = json.NewDecoder(file).Decode(&config)
	if err != nil {
		return fmt.Errorf("error decoding config: %v", err)
	}

	configMap := make(map[string]interface{})
	for _, s := range config.Services {
		sConfig, err := getServiceConfig(s)
		errorHandling.FailOnError(err, fmt.Sprintf("Failed to retrieve service \"%s\" configuration", s))

		configMap[s] = sConfig
	}

	for s, di := range config.Dependencies {
		if c, ok := configMap[s]; ok && c != nil {
			if a, ok := config.Aliases[s]; ok {
				configurationMap[s] = map[string]interface{}{a: c}
			} else {
				configurationMap[s] = map[string]interface{}{s: c}
			}
		}

		for _, s2 := range di {
			if len(configurationMap[s]) == 0 {
				configurationMap[s] = map[string]interface{}{s2: configMap[s2]}
			} else {
				configurationMap[s][s2] = configMap[s2]
			}
		}
	}

	return nil
}

func getServiceConfig(serviceName string) (interface{}, error) {
	file, err := os.Open(fmt.Sprintf("services/%s.json", serviceName))
	if os.IsNotExist(err) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	defer func() {
		errorHandling.LogOnError(
			file.Close(), "Error closing config file")
	}()

	var serviceConfig interface{}

	err = json.NewDecoder(file).Decode(&serviceConfig)
	if err != nil {
		return nil, err
	}

	return serviceConfig, nil
}
