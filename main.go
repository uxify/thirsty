package main

import (
	"fmt"
	"net/http"
	"net/url"
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

func filterDomainLink(crawlUrl string, link string) (ok bool) {
	u, er := url.Parse(crawlUrl)
	l, err := url.Parse(link)

	if er != nil || err != nil {
		fmt.Println("ERROR: Failed to compare domain")
		return
	}
	urlDomain := u.Hostname()
	linkDomain := l.Hostname()

	if urlDomain == linkDomain {
		ok = true
	}
	return ok
}

func crawl(crawlUrl string, ch chan string, chFinished chan bool) {
	resp, err := http.Get(crawlUrl)

	defer func() {
		chFinished <- true
	}()

	if err != nil {
		fmt.Println("ERROR: Failed to crawl \"" + crawlUrl + "\"")
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

			ok, link := getHref(t)
			if !ok {
				continue
			}

			hasProto := strings.Index(link, "http") == 0
			if hasProto {
				isSameDomain := filterDomainLink(crawlUrl, link)
				if isSameDomain {
					ch <- link
				}
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

	for _, crawlUrl := range pageUrls {
		go crawl(crawlUrl, chUrls, chFinished)
	}

	for c := 0; c < len(pageUrls); {
		select {
		case crawlUrl := <-chUrls:
			linkUrls[crawlUrl] = true
		case <-chFinished:
			c++
		}
	}

	fmt.Println("\nTotal ", len(linkUrls), " links on this page:")

	for crawlUrl, _ := range linkUrls {
		fmt.Println(" - " + crawlUrl)
	}

	close(chUrls)
}
