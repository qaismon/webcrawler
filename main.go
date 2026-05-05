package main

import (
	"fmt"
	"net/http"
	"net/url"
	"golang.org/x/net/html"
	"sync"
)

type Job struct{
	URL string
	Depth int
}

var (
	links []string
	visited=make(map[string]bool)
	count int
	mu sync.Mutex
	baseHost = "news.ycombinator.com"
)

const maxDepth=2


func main() {
	jobs:=make(chan Job, 100)
	
	for i:=0;i<5;i++{
		go worker(jobs)
	}

	jobs <-Job{
		URL : "https://news.ycombinator.com",
		Depth: 0,
	}

	select {}
}


func worker(jobs chan Job){
	for job:=range jobs{
		if job.Depth>maxDepth{
			continue
		}
		fmt.Println("Crawling:", job.URL, "Depth:", job.Depth)

		resp, err:= http.Get(job.URL)
		if(err!=nil){
			continue
		}

		doc, err:= html.Parse(resp.Body)
		resp.Body.Close()

		if(err!=nil){
			continue
		}
		extractLinks(doc,jobs, job.Depth)
	}
}


func extractLinks(n *html.Node, jobs chan Job, depth int) {
	if n.Type == html.ElementNode && n.Data == "a" {
		for _, attr := range n.Attr {
			if attr.Key == "href" {
				link:=attr.Val

				u, err :=url.Parse(link)
				if err != nil{
					continue
				}

				if u.Host != "" && u.Host!=baseHost{
					continue
				}

				if u.Host==""{
					u.Scheme="https"
					u.Host= baseHost
					link=u.String()
				}

					mu.Lock()

					if !visited[link]{
						visited[link]=true
						links= append(links,link)
						count++
						fmt.Println("Found:", link)
						mu.Unlock()

						select{
						case jobs<- Job{URL: link, Depth: depth+1}:
						default:
						}
					}else{
						mu.Unlock()
					}
				} 

			}
		}
	

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		extractLinks(c, jobs, depth)
	}
}