package main

import (
	"fmt"
	"net/http"
	"net/url"
	"golang.org/x/net/html"
	"sync"
	"time"
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
	wg sync.WaitGroup
	baseHost = "github.com"
)

const (
	maxDepth=2
	maxPages=100
)



func main() {
	jobs:=make(chan Job, 20)
	ticker:= time.NewTicker(300 * time.Millisecond)
	defer ticker.Stop()
	rate:= ticker.C
	
	for i:=0;i<5;i++{
		go worker(jobs, rate)
	}

	wg.Add(1)
	jobs <-Job{
		URL : "https://github.com/qaismon",
		Depth: 0,
	}

	wg.Wait()

	close(jobs)
	fmt.Println("\n====== DONE ======")
	fmt.Println("Total unique links:", count)

	for _,l := range links{
		fmt.Println(l)
	}
}


func worker(jobs chan Job, rate <-chan time.Time){
	for job:=range jobs{

		if job.Depth>=maxDepth{
			wg.Done()
			continue
		}

		<-rate

		fmt.Println("Crawling:", job.URL, "Depth:", job.Depth)

		resp, err:= fetchWithRetry(job.URL, 3)
		if err != nil {
			fmt.Println("Failed:", job.URL)
			wg.Done()
			continue
		}

		doc, err:= html.Parse(resp.Body)
		resp.Body.Close()

		if(err!=nil){
			wg.Done()
			continue
		}
		extractLinks(doc,jobs, job.Depth)

		wg.Done()
	}
}


func fetchWithRetry(url string, maxRetries int) (*http.Response, error) {
	var resp *http.Response
	var err error

	for i:=0;i<maxRetries;i++{

		resp, err = http.Get(url)
		if err == nil{
			return resp, nil
		}

		wait:= time.Duration(200*(1<<i)) * time.Millisecond
		fmt.Println("Retrying:", url, "in", wait)
		time.Sleep(wait)
	}

	return nil, err
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

				u.Fragment=""
				u.RawQuery=""

				if u.Host != "" && u.Host!=baseHost{
					continue
				}

				if u.Host==""{
					u.Scheme="https"
					u.Host= baseHost
					
				}

				link=u.String()

					mu.Lock()

					if count >= maxPages {
					mu.Unlock()
					return
				}

					if !visited[link]{
						visited[link]=true
						links= append(links,link)
						count++
						fmt.Println("Found:", link)
						mu.Unlock()

						wg.Add(1)
						select{
						case jobs <- Job{
							URL: link, 
							Depth: depth+1,
							}:
						default:
							wg.Done()

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