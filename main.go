package main

import (
	"fmt"
	"golang.org/x/net/html"
	"html/template"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type AnalysisResult struct {
	URL           string
	HTMLVersion   string
	Title         string
	HeadingsCount map[string]int
	InternalLinks int
	ExternalLinks int
	BrokenLinks   int
	HasLoginForm  bool
	ErrorMessage  string
}

func main() {
	http.HandleFunc("/", formHandler)
	http.HandleFunc("/analyze", analyzeHandler)
	http.ListenAndServe(":8080", nil)
}

func formHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/index.html"))
	tmpl.Execute(w, nil)
}

func analyzeHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Form parse error", http.StatusBadRequest)
		return
	}
	urlStr := r.FormValue("url")
	if !strings.HasPrefix(urlStr, "http") {
		urlStr = "http://" + urlStr
	}

	res, err := http.Get(urlStr)
	if err != nil {
		showError(w, fmt.Sprintf("Failed to fetch URL: %v", err))
		return
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		showError(w, fmt.Sprintf("Received HTTP status: %d %s", res.StatusCode, res.Status))
		return
	}

	doc, err := html.Parse(res.Body)
	if err != nil {
		showError(w, fmt.Sprintf("Failed to parse HTML: %v", err))
		return
	}

	result := analyzePage(doc, urlStr)
	tmpl := template.Must(template.ParseFiles("templates/index.html"))
	tmpl.Execute(w, result)
}

func showError(w http.ResponseWriter, message string) {
	tmpl := template.Must(template.ParseFiles("templates/index.html"))
	tmpl.Execute(w, AnalysisResult{ErrorMessage: message})
}

func analyzePage(doc *html.Node, baseURL string) AnalysisResult {
	result := AnalysisResult{
		URL:           baseURL,
		HeadingsCount: make(map[string]int),
	}

	var f func(*html.Node)
	baseParsedURL, _ := url.Parse(baseURL)

	var links []string

	f = func(n *html.Node) {
		if n.Type == html.ElementNode {
			switch n.Data {
			case "title":
				if n.FirstChild != nil {
					result.Title = n.FirstChild.Data
					//titleFound = true
				}
			case "h1", "h2", "h3", "h4", "h5", "h6":
				result.HeadingsCount[n.Data]++
			case "a":
				for _, attr := range n.Attr {
					if attr.Key == "href" {
						links = append(links, attr.Val)
					}
				}
			case "form":
				for _, attr := range n.Attr {
					if attr.Key == "action" && strings.Contains(strings.ToLower(attr.Val), "login") {
						result.HasLoginForm = true
					}
				}
				// Check if there is a password field inside the form
				for c := n.FirstChild; c != nil; c = c.NextSibling {
					if c.Type == html.ElementNode && c.Data == "input" {
						for _, attr := range c.Attr {
							if attr.Key == "type" && attr.Val == "password" {
								result.HasLoginForm = true
							}
						}
					}
				}
			case "html":
				for _, attr := range n.Attr {
					if attr.Key == "xmlns" {
						// Assume HTML5 if xmlns is missing
						result.HTMLVersion = "HTML5"
					}
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}

	f(doc)

	// Analyze links
	for _, link := range links {
		u, err := url.Parse(link)
		if err != nil {
			continue
		}
		if u.Host == "" || u.Host == baseParsedURL.Host {
			result.InternalLinks++
		} else {
			result.ExternalLinks++
		}

		client := http.Client{
			Timeout: 5 * time.Second,
		}
		linkToCheck := link
		if u.Scheme == "" {
			linkToCheck = baseParsedURL.Scheme + "://" + baseParsedURL.Host + link
		}
		resp, err := client.Head(linkToCheck)
		if err != nil || resp.StatusCode >= 400 {
			result.BrokenLinks++
		}
	}

	// Default HTML version
	if result.HTMLVersion == "" {
		result.HTMLVersion = "HTML5"
	}

	return result
}
