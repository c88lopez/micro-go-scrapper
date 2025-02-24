package main

import (
	"sarasa/libs/providersCommon"
)

func main() {
	providersCommon.ProviderProcessor{
		ServiceName:            "provider4",
		Selector:               "div.cont div.thumb",
		CustomGetDetailsLinkFn: getDetailLink,
		CustomGetDetailsFn:     getDetails,
	}.Run()
}
