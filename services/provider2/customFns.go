package main

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"sarasa/libs/providersCommon"
	"sarasa/schemas"

	"github.com/PuerkitoBio/goquery"
)

func getDetailLink(s *goquery.Selection) string {
	parent := s.Parent()
	if parent == nil {
		return "invalid_link"
	}

	providerLink, attrExists := parent.Attr("href")

	if !attrExists {
		return "invalid_link"
	}

	if !strings.Contains(providerLink, fmt.Sprintf(".com/perl/site2/individual")) {
		return "invalid_link"
	}

	return providerLink
}

func getDetails(doc *goquery.Document, _ schemas.Source) schemas.Provider {
	place, err := getPlace(doc)
	if err != nil {
		log.Printf("%s", err)
		return schemas.Provider{}
	}

	//noinspection ALL
	return schemas.Provider{
		Name:  getName(doc),
		Phone: getPhone(doc),
		Place: place,
		Pics: providersCommon.GetPics(
			doc, "center > img", "", "urlSubString", false),
	}
}

func getName(doc *goquery.Document) string {
	return doc.Find("td:nth-child(2) > center > table > tbody > tr > td > table > tbody > tr:nth-child(3) > td > center > b > div").First().Text()
}

func getPhone(doc *goquery.Document) string {
	providerPhoneRaw := doc.Find("tr:nth-child(7) > td > center > div").First().Text()

	providerPhoneRaw = strings.Replace(providerPhoneRaw, "Cel  ", "", 1)
	providerPhoneRaw = strings.Replace(providerPhoneRaw, "+549", "", 1)
	providerPhoneRaw = strings.Replace(providerPhoneRaw, "-", "", -1)

	return providerPhoneRaw
}

//noinspection ALL
func getPlace(doc *goquery.Document) (string, error) {
	providerPlaceRaw := doc.Find("td:nth-child(2) > center > table > tbody > tr > td > table > tbody > tr:nth-child(5) > td > center > div").First().Text()

	if strings.Contains(providerPlaceRaw, "Zona de ") {
		providerZone := regexp.MustCompile(`Zona de ([a-zA-Zñ\s]*)([a-zA-Zñ\s]*)`)
		providerDataMatch := providerZone.FindStringSubmatch(providerPlaceRaw)

		if providerDataMatch == nil {
			return "", fmt.Errorf("ignore provider, no data match")
		}

		providerPlaceRaw = providerDataMatch[1]
	}

	if strings.Contains(providerPlaceRaw, "Zona ") {
		providerZone := regexp.MustCompile(`Zona ([a-zA-Zñ\s]*)`)
		providerDataMatch := providerZone.FindStringSubmatch(providerPlaceRaw)

		if providerDataMatch == nil {
			return "", fmt.Errorf("ignore provider, no data match")
		}

		providerPlaceRaw = providerDataMatch[1]
	}

	if providerPlaceRaw == "norte" {
		return "", fmt.Errorf("Ignored zone \"%s\"", providerPlaceRaw)
	}

	if providerPlaceRaw == "Oeste y CABA" {
		return "", fmt.Errorf("Ignored zone \"%s\"", providerPlaceRaw)
	}

	providerPlaceRaw = strings.Replace(providerPlaceRaw, " CABA", "", 1)
	providerPlaceRaw = strings.Replace(providerPlaceRaw, "las cañitas", "Palermo", 1)
	providerPlaceRaw = strings.Replace(providerPlaceRaw, "Las Cañitas", "Palermo", 1)
	providerPlaceRaw = strings.Replace(providerPlaceRaw, "Palermo -", "Palermo", 1)
	providerPlaceRaw = strings.Replace(providerPlaceRaw, "Belgrano R", "Belgrano", 1)
	providerPlaceRaw = strings.Replace(providerPlaceRaw, "Belgrano ", "Belgrano", 1)

	return providerPlaceRaw, nil
}
