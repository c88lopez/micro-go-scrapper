package signals

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"sarasa/libs/errorHandling"

	"sarasa/schemas"
)

func SignalHandler(services ...schemas.Service) {
	s := make(chan os.Signal, 1)

	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)

	go func() {
		<-s

		for _, service := range services {
			log.Printf("Closing service %s", service)
			errorHandling.LogOnError(service.Close(), "Failed closing service "+service.String())
		}

		os.Exit(0)
	}()
}
