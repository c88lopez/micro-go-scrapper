package errorHandling

import (
	"log"
)

func FailOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s - error: %s", msg, err)
	}
}

func LogOnError(err error, msg string) {
	if err != nil {
		log.Printf("%s - error: %s", msg, err)
	}
}
