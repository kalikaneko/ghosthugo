package main

import (
	"fmt"
	"log"
	"path"
	"path/filepath"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/samber/lo"
)

// imagePathWithSize returns a modified image filename that adds
// a suffix with the size information after the base name, before
// the extension.
func imagePathWithSize(filename, size string) string {
	_parts := strings.Split(filename, ".")
	name, ext := _parts[0], _parts[1]
	return name + "_" + size + "." + ext
}

// postProcessHTML will rewrite the inner content according to some extra rules
func postProcessHTML(htmlContent string, bundledir string) string {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		log.Fatal(err)
	}

	// Traverse all <img> tags. Download to a file within
	// the bundle dir. Also changes the sourceset to point to the scaled images in the bundle.
	doc.Find("img").Each(func(index int, item *goquery.Selection) {
		src, _ := item.Attr("src")

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

		// we rely on the scaling being properly done, otherwise it stretches the image.
		item.RemoveAttr("height")
		item.SetAttr("width", "100%")

		// split the filepaths in the source set
		srcset, _ := item.Attr("srcset")

		// map size => remote path
		imgs := make(map[string]string)
		srcsetArr := strings.Split(srcset, ",")
		lo.ForEach(srcsetArr, func(str string, _ int) {
			_parts := strings.Split(strings.Trim(str, " "), " ")
			path, size := _parts[0], _parts[1]
			imgs[size] = path
		})

		srcsetNew := []string{}
		for size, remote := range imgs {
			fn := path.Base(remote)
			if strings.Contains(remote, "/size/") {
				fn = imagePathWithSize(fn, size)
			}

			fp := filepath.Join(bundledir, fn)
			err := downloadFile(fp, remote)
			if err != nil {
				fmt.Println("Error downloading file:", err)
				continue
			}
			srcsetNew = append(srcsetNew, fmt.Sprintf("%s %s", fn, size))
		}
		// replace the srcset attribute with the newly formatted paths, local to the bundle
		item.SetAttr("srcset", strings.Join(srcsetNew, ", "))
	})

	// Find all anchor tags
	// TODO: only if enabled from config
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if exists && strings.HasPrefix(href, internetArchiveAudioPrefix) {
			fmt.Println("[+] Replacing audio player with source:", href)
			// Construct the audio player element using the href as the audio source
			audioPlayer := fmt.Sprintf(`<div class="audioplayer"><audio controls style="width:100%%;">
  <source src="%s" type="audio/mpeg">
  Your browser does not support the audio element.
</audio><span><a class="audio-download-link" href="%s" rel="noreferrer" download="fuga">Descarga</a></div>`, href, href)

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
