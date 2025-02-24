package main

import (
	"sarasa/libs/providersCommon"
)

func main() {
	providersCommon.ProviderProcessor{
		ServiceName:            "provider1",
		Selector:               "article",
		CustomGetDetailsLinkFn: getDetailLink,
		CustomGetDetailsFn:     getDetails,
	}.Run()
}
