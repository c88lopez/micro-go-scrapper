package main

import (
	"sarasa/libs/providersCommon"
)

func main() {
	providersCommon.ProviderProcessor{
		ServiceName:            "provider5",
		Selector:               "div.text-thumb",
		CustomGetDetailsLinkFn: getDetailLink,
		CustomGetDetailsFn:     getDetails,
	}.Run()
}
