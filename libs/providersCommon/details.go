package providersCommon

import (
	"fmt"
	"log"
	"net/http"

	"github.com/PuerkitoBio/goquery"
	"sarasa/schemas"
)

func GetDetails(providerLink string, source schemas.Source, getDetailsFn CustomGetDetailsFn) (schemas.Provider, error) {
	resp, err := http.Get(providerLink)
	if err != nil {
		fmt.Printf("Error on request: %s", err)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return schemas.Provider{}, err
	}

	log.Printf("Getting details: %s", providerLink)
	provider := getDetailsFn(doc, source)
	log.Printf("Details: %v", provider)

	provider.Link = providerLink

	return provider, nil
}
