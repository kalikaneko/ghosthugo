package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
)

func filenameFromURL(urlStr string) string {
	u, err := url.Parse(urlStr)
	if err != nil {
		fmt.Println("Error:", err)
		return ""
	}

	return path.Base(u.Path)
}

func downloadFile(fp string, url string) (err error) {
	out, err := os.Create(fp)
	if err != nil {
		return err
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}
