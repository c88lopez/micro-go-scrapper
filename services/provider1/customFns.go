package main

import (
	"log"
	"regexp"
	"strings"

	"sarasa/libs/providersCommon"
	"sarasa/schemas"

	"github.com/PuerkitoBio/goquery"
)

func getDetailLink(s *goquery.Selection) string {
	return s.Find("a").First().AttrOr("href", "invalid_link")
}

//noinspection SpellCheckingInspection
func getDetails(doc *goquery.Document, _ schemas.Source) schemas.Provider {
	providerData := regexp.MustCompile(`^([a-zA-Záéíóúñ\s.]+)\s[0-9]{2}<img[\w\W]*/>([a-zA-Záéíóúñ\s]+)[\w\W]*>([0-9\-]+)`)

	providerTitle := doc.Find("article h1.page-title").First().Text()
	providerDataMatch := providerData.FindStringSubmatch(providerTitle)

	if providerDataMatch == nil {
		log.Println("Ignore provider, no data match")
		return schemas.Provider{}
	}

	if providerDataMatch[2] == "Zona Norte" ||
		providerDataMatch[2] == "Zona Oeste" ||
		providerDataMatch[2] == "Zona Sur" {
		log.Printf("Ignore provider by zone")
		return schemas.Provider{}
	}

	return schemas.Provider{
		Name:  providerDataMatch[1],
		Phone: strings.Replace(providerDataMatch[3], "-", "", -1),
		Place: providerDataMatch[2],
		Pics: providersCommon.GetPics(
			doc, "div#galeria figure", "src=\"[a-zA-Z0-9:/\\-._]+\"", "urlSubString", true),
	}
}
