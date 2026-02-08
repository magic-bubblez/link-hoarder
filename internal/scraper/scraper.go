package scraper

import (
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/html"
)

// Main func to scrape title and open graph image from the url
func ScrapeTitle(url string) (*string, *string, error) {
	client := http.Client{
		Timeout: 5 * time.Second,
	}
	resp, err := client.Get(url) //computer opens a pipe to website's server
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close() //need to close the connection else "memory consumption and a file descriptor leak"

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return nil, nil, err
	}
	title := getTitle(doc)
	img := getImagePreview(doc)
	return title, img, nil
}

// Recursive tree walker for title
func getTitle(n *html.Node) *string {
	if n.Type == html.ElementNode && n.Data == "title" {
		// title text is child of <title> node
		if n.FirstChild != nil {
			t := strings.TrimSpace(n.FirstChild.Data)
			return &t
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		result := getTitle(c)
		if result != nil {
			return result
		}
	}
	return nil
}

// To find Open Graph image
func getImagePreview(n *html.Node) *string {
	if n.Type == html.ElementNode && n.Data == "meta" {
		var prop, content string
		// ex: <meta property="og:image" content="url">
		for _, attr := range n.Attr {
			if attr.Key == "property" || attr.Key == "name" {
				prop = attr.Val
			}
			if attr.Key == "content" {
				content = attr.Val
			}
		}
		if (prop == "og:image" || prop == "twitter:image") && content != "" {
			return &content
		}
	}
	//Recursive Search
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		result := getImagePreview(c)
		if result != nil {
			return result
		}
	}
	return nil
}
