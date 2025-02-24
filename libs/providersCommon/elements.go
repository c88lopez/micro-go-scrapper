package providersCommon

import (
	"log"
	"net/http"

	"github.com/PuerkitoBio/goquery"
	"sarasa/schemas"
)

type getElementLinkFn func(s *goquery.Selection) string

func GetElements(url, selector string, limit int, getElementLinkFn getElementLinkFn, providerDetailsLink chan string, results chan schemas.Provider) ([]schemas.Provider, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	var providers []schemas.Provider

	log.Println("Getting providers' details...")

	imagesSelector := doc.Find(selector)

	if imagesSelector.Size() > limit {
		imagesSelector = imagesSelector.Slice(0, limit)
	}

	initCounter := 1
	imagesSelector.Each(func(i int, s *goquery.Selection) {
		link := getElementLinkFn(s)

		if link == "invalid_link" {
			initCounter++
			return
		}

		log.Printf("Fetching link: %s", link)
		if link[0] == '/' {
			link = configuration.Provider.Source.Domain + link
		}

		providerDetailsLink <- link
	})

	for a := initCounter; a <= imagesSelector.Size(); a++ {
		providers = append(providers, <-results)
	}

	return providers, nil
}
