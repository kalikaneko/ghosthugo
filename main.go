package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

var (
	defaultLang  = "en"
	postEndpoint = "/ghost/api/content/posts/"
	postsPath    = "content/posts"

	internetArchiveAudioPrefix = "https://archive.org/download/"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Err: %s", err)
	}

	baseURL := os.Getenv("GHOST_URL")
	key := os.Getenv("GHOST_KEY")

	if baseURL == "" || key == "" {
		fmt.Println("Need GHOST_URL and GHOST_KEY")
		os.Exit(1)
	}

	url := baseURL + postEndpoint + "?key=" + key

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("New Request Error", err)
	}
	req.Header.Set("Acept-Version", "v5.0")

	client := &http.Client{}
	resp, err := client.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		log.Fatal(err)
	}

	body, readErr := ioutil.ReadAll(resp.Body)
	if readErr != nil {
		fmt.Println("body error:", readErr)
		return
	}

	// TODO need to check the meta fields and loop
	/*
		"meta": {
		      "pagination": {
		      "page": 1,
		      "limit": 15,
		      "pages": 1,
		      "total": 4,
		      "next": null,
		      "prev": null
		}
	*/

	var obj map[string]interface{}
	json.Unmarshal(body, &obj)

	posts := obj["posts"].([]any)

	for _, post := range posts {
		if err := processPost(post); err != nil {
			fmt.Println("error:", err)
		}
	}
}
