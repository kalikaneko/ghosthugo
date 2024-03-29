package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/pelletier/go-toml"
)

func getString(v any) string {
	switch {
	case v != nil:
		return v.(string)
	default:
		return "null"
	}
}

func getFloat(v any) float64 {
	switch {
	case v != nil:
		return v.(float64)
	default:
		return 0
	}
}

// timeLayout as returned by ghost json
const timeLayout = "2006-01-02T15:04:05.000-07:00"

func parseTimestamp(ts any) time.Time {
	t, err := time.Parse(timeLayout, ts.(string))
	if err != nil {
		return time.Time{}
	}
	return t
}

// processPost is the main action for any raw post. It will create a Post struct,
// download the feature image, and dump the serialized version into the current
// hugo content folder.
func processPost(data itemJSON) error {
	post := &Post{
		Title:         getString(data["title"]),
		Slug:          getString(data["slug"]),
		Description:   getString(data["excerpt"]),
		Summary:       getString(data["excerpt"]),
		Created:       parseTimestamp(data["created_at"]),
		LastModified:  parseTimestamp(data["updated_at"]),
		Published:     parseTimestamp(data["published_at"]),
		Lang:          defaultLang,
		content:       getString(data["html"]),
		FeaturedImage: getString(data["feature_image"]),
		ReadingTime:   getFloat(data["reading_time"]),
		Images:        []string{},
		Audio:         []string{},
		Videos:        []string{},
	}

	bundlePath := post.getBundlePath()
	os.MkdirAll(bundlePath, 0777)

	// downloading for now. we need a toggle switch to avoid saving assets that we already have.

	if post.FeaturedImage != "null" {
		fn := filenameFromURL(post.FeaturedImage)
		err := downloadFile(filepath.Join(bundlePath, fn), post.FeaturedImage)
		if err != nil {
			fmt.Println("Error:", err)
		}

		// store the local path, once that we have downloaded it to the bundle
		post.FeaturedImage = fn
		post.Images = append(post.Images, fn)
	}

	post.content = postProcessHTML(post.content, bundlePath)
	if err := post.Dump(); err != nil {
		return err
	}
	return nil
}

// Post represents a single post that we will serialize to markdown.
type Post struct {
	Title       string `toml:"title"`
	Slug        string `toml:"slug"`
	Description string `toml:"description"`
	Summary     string `toml:"summary"`

	Created      time.Time `toml:"created"`
	LastModified time.Time `toml:"lastmod"`
	Published    time.Time `toml:"date"`

	Lang    string
	content string

	FeaturedImage string `toml:"featured_image"`

	Images []string
	Audio  []string
	Videos []string

	ReadingTime float64
}

func (p *Post) Dump() error {
	out, err := os.Create(filepath.Join(p.getBundlePath(), "index.md"))
	if err != nil {
		return err
	}
	defer out.Close()
	out.Write([]byte("+++\n"))

	var buf bytes.Buffer
	if err := toml.NewEncoder(&buf).Order(toml.OrderPreserve).Encode(p); err != nil {
		panic(err)
	}
	out.Write(buf.Bytes())
	out.Write([]byte("+++\n"))
	out.Write([]byte(p.content))
	return nil
}

func (p *Post) getBundlePath() string {
	contentBase := os.Getenv("HUGO_CONTENT")
	if contentBase == "" {
		contentBase = postsPath
	}
	return filepath.Join(contentBase, fmt.Sprintf("%d-%d-%s", p.Created.Year(), p.Created.Month(), p.Slug))
}
