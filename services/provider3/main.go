package main

import (
	"sarasa/libs/providersCommon"
)

func main() {
	providersCommon.ProviderProcessor{
		ServiceName:            "provider3",
		Selector:               "div.card-image",
		CustomGetDetailsLinkFn: getDetailLink,
		CustomGetDetailsFn:     getDetails,
	}.Run()
}
