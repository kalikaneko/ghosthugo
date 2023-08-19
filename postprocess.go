package main

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// postProcessHTML will rewrite the inner content according to some extra rules
func postProcessHTML(htmlContent string, bundledir string) string {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		log.Fatal(err)
	}

	// Replace all <img> tags with their markdown equivalent,
	// which after a successful download should be a local file within
	// the bundle dir.
	doc.Find("img").Each(func(index int, item *goquery.Selection) {
		src, _ := item.Attr("src")
		//alt, _ := item.Attr("alt")

		if src == "" {
			// we consider this a malformed image, has no src.
			return
		}
		filename := filenameFromURL(src)
		fp := filepath.Join(bundledir, filename)
		err := downloadFile(fp, src)
		if err != nil {
			fmt.Println("Error downloading file:", err)
			return
		}

		item.SetAttr("src", filename)
	})

	// Find all anchor tags
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if exists && strings.HasPrefix(href, internetArchiveAudioPrefix) {
			// Construct the audio player element using the href as the audio source
			audioPlayer := fmt.Sprintf(`<audio controls>
  <source src="%s" type="audio/mpeg">
  Your browser does not support the audio element.
</audio>`, href)

			// Replace the link with the audio player
			s.ReplaceWithHtml(audioPlayer)
		}
	})
	// Extract the inner content of the <body> tag
	modifiedHtml, err := doc.Find("body").Html()
	if err != nil {
		fmt.Printf("err = %+v\n", err)
	}
	return modifiedHtml
}
