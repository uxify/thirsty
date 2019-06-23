package main

import (
	"fmt"
	"net/http"
	"os"

	"golang.org/x/net/html"
)

func getMetaValue(t html.Token) (ok bool, metaName string, metaContent string) {
	for _, a := range t.Attr {
		if a.Key == "content" {
			metaContent = a.Val
		} else if a.Key == "name" {
			metaName = a.Val
		}
	}

	if len(metaName) > 0 {
		ok = true
	}

	return
}

func getMetaTags(t html.Token) (isMeta bool) {
	isMeta = t.Data == "meta"
	return isMeta
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
	defer b.Close() // close Body when the function returns

	z := html.NewTokenizer(b)

	for {
		tt := z.Next()

		switch tt {
		case html.ErrorToken:
			return
		case html.StartTagToken, html.SelfClosingTagToken:
			t := z.Token()
			metaTags := getMetaTags(t)
			if !metaTags {
				continue
			}
			ok, metaName, metaContent := getMetaValue(t)
			if !ok {
				continue
			}
			metaData := metaName + ": " + metaContent
			ch <- metaData
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

	fmt.Println("\nTotal Meta Values", len(linkUrls))

	for url, _ := range linkUrls {
		fmt.Println(" - " + url)
	}

	close(chUrls)
}
