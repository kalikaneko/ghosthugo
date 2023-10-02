package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/joho/godotenv"
	toml "github.com/pelletier/go-toml"
	"gopkg.in/yaml.v3"
)

var (
	defaultLang   = "en"
	api           = "/ghost/api/content"
	pagesEndpoint = api + "/pages/"
	postEndpoint  = api + "/posts/"
	postsPath     = "content/posts"

	internetArchiveAudioPrefix = "https://archive.org/download/"
)

func getFromConfig(section string) (string, error) {
	config, err := toml.LoadFile("hugo.toml")
	if err != nil {
		return "", err
	}
	_section := config.Get(section)
	if _section == nil {
		return "", nil
	}
	return config.Get(section).(string), nil

}

type itemJSON map[string]interface{}

func doRequest(endpoint string, itemType string, ch chan itemJSON) {
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

	page := 1

	for {
		url := baseURL + endpoint + "?key=" + key + "&page=" + strconv.Itoa(page) + "&include=tags"

		fmt.Println("GET", url)

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			fmt.Println("New Request Error", err)
		}
		req.Header.Set("Acept-Version", "v5.0")

		err = func() error {
			client := &http.Client{}
			resp, err := client.Do(req)
			if resp != nil {
				defer resp.Body.Close()
			}
			if err != nil {
				log.Fatal(err)
			}
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return err
			}
			var obj map[string]interface{}
			err = json.Unmarshal(body, &obj)
			if err != nil {
				return err
			}

			items := obj[itemType].([]interface{})
			for _, item := range items {
				ch <- item.(map[string]interface{})
			}

			// pagination check: do we need to request next page?
			meta := obj["meta"].(map[string]any)
			pages := meta["pagination"].(map[string]any)

			current := pages["page"].(float64)
			total := int(pages["total"].(float64)) / int(pages["limit"].(float64))
			next := pages["next"]
			fmt.Println("next", next)
			if next == nil {
				close(ch)
				return errors.New("lastPage")
			}
			fmt.Println("Page", current, "of", total)

			return nil
		}()
		// TODO catch error == last page only
		if err != nil {
			break
		}
		page += 1
		fmt.Println("request next page")
	}
}

type Event struct {
	Title   string
	Content string
}

func dumpEventAsDataFile(name string, event *Event) {
	out, err := yaml.Marshal(event)
	if err != nil {
		panic(err)
	}
	path := filepath.Join("data", "events", name+".yaml")
	f, err := os.Create(path)
	if err != nil {
		fmt.Println(err)
		return
	}
	f.Write(out)
}

// processEventsFromPages iterates through all pages. In the case that the page has the tag
// configured as target, we will dump a yaml file under hugo's data folder with the contents
// of the page.
func processEventsFromPages() {
	fmt.Println("[+] Fetching all pages")

	eventTag, err := getFromConfig("ghosthugo.events.tag")
	if err != nil {
		panic(err)
	}
	dataFile, err := getFromConfig("ghosthugo.events.data")
	if err != nil {
		panic(err)
	}

	pages := make(chan itemJSON)
	go doRequest(pagesEndpoint, "pages", pages)

	for page := range pages {
		title := page["title"]
		fmt.Println("title:", title)

		tags := page["tags"]
		if tags != nil {
			for _, _tag := range tags.([]interface{}) {
				tag := _tag.(map[string]any)["slug"]

				if tag == eventTag {
					event := &Event{
						Title:   page["title"].(string),
						Content: page["html"].(string),
					}
					dumpEventAsDataFile(dataFile, event)
					return
				}
			}
		}
	}
	fmt.Println("no matching tag")
}

func processAllPosts() {
	fmt.Println("[+] Processing all posts")

	posts := make(chan itemJSON)
	go doRequest(postEndpoint, "posts", posts)

	for post := range posts {
		if err := processPost(post); err != nil {
			fmt.Println("error:", err)
		}
	}
}

func main() {
	eventTag, err := getFromConfig("ghosthugo.events.tag")
	if err != nil {
		panic(err)
	}
	if eventTag != "" {
		processEventsFromPages()
	}

	processAllPosts()

}
