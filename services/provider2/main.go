package main

import (
	"sarasa/libs/providersCommon"
)

func main() {
	providersCommon.ProviderProcessor{
		ServiceName:            "provider2",
		Selector:               "tbody > tr:nth-child(1) > td > center > a > img",
		CustomGetDetailsLinkFn: getDetailLink,
		CustomGetDetailsFn:     getDetails,
	}.Run()
}
