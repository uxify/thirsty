package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"golang.org/x/net/html"
)

func getHref(t html.Token) (ok bool, href string) {
	for _, a := range t.Attr {
		if a.Key == "href" {
			href = a.Val
			ok = true
		}
	}
	return
}

func getAnchorTags(t html.Token) (isAnchor bool) {
	isAnchor = t.Data == "a"
	return isAnchor
}

func crawl(url string, ch chan string, chFinished chan bool) {
	resp, err := http.Get(url)

	defer func() {
		chFinished <- true
	}()

	if err != nil {
		fmt.Println("ERROR: Failed to crawl \"" + url + "\"")
		return
	}

	b := resp.Body
	defer b.Close()

	z := html.NewTokenizer(b)

	for {
		tt := z.Next()

		switch {
		case tt == html.ErrorToken:
			return
		case tt == html.StartTagToken:
			t := z.Token()

			anchorTags := getAnchorTags(t)

			if !anchorTags {
				continue
			}

			ok, url := getHref(t)
			if !ok {
				continue
			}

			hasProto := strings.Index(url, "http") == 0
			if hasProto {
				ch <- url
			}
		}
	}
}

func main() {
	linkUrls := make(map[string]bool)
	pageUrls := os.Args[1:]

	fmt.Println("\nPage", pageUrls)

	chUrls := make(chan string)
	chFinished := make(chan bool)

	for _, url := range pageUrls {
		go crawl(url, chUrls, chFinished)
	}

	for c := 0; c < len(pageUrls); {
		select {
		case url := <-chUrls:
			linkUrls[url] = true
		case <-chFinished:
			c++
		}
	}

	fmt.Println("\nTotal ", len(linkUrls), "links on this page:\n")

	for url, _ := range linkUrls {
		fmt.Println(" - " + url)
	}

	close(chUrls)
}
