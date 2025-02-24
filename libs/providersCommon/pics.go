package providersCommon

import (
	"log"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func GetPics(doc *goquery.Document, selector, regexpString, urlSubString string, removeSrc bool) []string {
	var pics []string

	var imgSrcRegex *regexp.Regexp
	if regexpString != "" {
		imgSrcRegex = regexp.MustCompile(regexpString)
	}

	doc.Find(selector).Each(func(i int, s *goquery.Selection) {
		var match string
		if regexpString != "" {
			regexMatch := imgSrcRegex.FindStringSubmatch(s.Text())
			if len(regexMatch) != 0 {
				match = regexMatch[0]
			}
		} else {
			for _, attr := range s.Nodes[0].Attr {
				if attr.Key == "src" {
					match = attr.Val
				}
			}
		}

		if len(match) == 0 {
			log.Printf("imsSrc no match - rawText: %s", match)
			return
		}

		var pic string
		if removeSrc {
			pic = match[5 : len(match)-1]
		} else {
			pic = match
		}

		if strings.Contains(pic, urlSubString) {
			pics = append(
				pics,
				pic,
			)
		}
	})

	return pics
}
