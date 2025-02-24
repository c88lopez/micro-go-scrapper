package retryHandling

import (
	"fmt"
	"log"
	"time"

	"sarasa/libs/errorHandling"
)

type TryFunctionFn func() error
type PreFunctionFn func() error

func Try(f TryFunctionFn, retryConditions []error, preFn PreFunctionFn, retryCount int, sleepBetweenRetry time.Duration) error {
	var err error

	if retryCount <= 0 {
		return nil
	}

	if err = f(); nil != err {
		for ; retryCount > 0; retryCount-- {
			for _, e := range retryConditions {
				if err == e {
					log.Printf("Retry number: %d - Waiting: %d ms", retryCount, sleepBetweenRetry.Milliseconds())

					time.Sleep(sleepBetweenRetry)

					errorHandling.LogOnError(preFn(), "Failed to restart ")

					errorHandling.FailOnError(f(),
						fmt.Sprintf("After retry ExchangeDeclare"))
				}
			}
		}
	}

	return err
}
