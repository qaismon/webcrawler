package main

import (
	"fmt"
	"net/http"
	"net/url"
	"golang.org/x/net/html"
	"sync"
	"strings"
	"time"
	"sort"
)

type Job struct{
	URL string
	Depth int
}

type JobData struct {
	Title string
	Tags  []string
}

type SkillStat struct {
	Name  string
	Count int
}

var (
	links []string
	visited=make(map[string]bool)
	count int
	mu sync.Mutex
	wg sync.WaitGroup
	baseHost = "remoteok.com"
)

var skills = []string{
	"go", "react", "node", "docker",
	"python", "aws", "kubernetes", "mongodb",
}

var skillFreq = make(map[string]int)
var jobsData []JobData



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
		URL : "https://remoteok.com/remote-dev-jobs",
		Depth: 0,
	}

	wg.Wait()
	analyzeFromJobs()

	close(jobs)

fmt.Println("\n====== DONE ======")

// build stats slice
var stats []SkillStat
for k, v := range skillFreq {
	stats = append(stats, SkillStat{k, v})
}

sort.Slice(stats, func(i, j int) bool {
	return stats[i].Count > stats[j].Count
})

total := 0
for _, s := range stats {
	total += s.Count
}

fmt.Println("\n====== SKILL DEMAND RANKING ======")

for _, s := range stats {
	percent := float64(s.Count) / float64(total) * 100
	fmt.Printf("%s → %d (%.2f%%)\n", s.Name, s.Count, percent)
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
		extractJobs(doc)
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


func extractText(n *html.Node) string {
	if n.Type == html.TextNode {
		return n.Data
	}

	var result string

	for c:=n.FirstChild; c !=nil; c=c.NextSibling {
		result+=extractText(c)
	}
	return result
}

func analyzeSkills(text string){
	text=strings.ToLower(text)

	mu.Lock()
	defer mu.Unlock()

	for _, skill := range skills{
		if strings.Contains(text, skill){
			skillFreq[skill]++
		}
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

func extractJobs(n *html.Node) {
	if n.Type == html.ElementNode && n.Data == "tr" {
		for _, attr := range n.Attr {
			if attr.Key == "class" && strings.Contains(attr.Val, "job") {

				job := JobData{}

				var f func(*html.Node)
				f = func(node *html.Node) {

					if node.Type == html.ElementNode && node.Data == "h2" && node.FirstChild != nil {
						job.Title = node.FirstChild.Data
					}

					if node.Type == html.ElementNode && node.Data == "div" {
						for _, a := range node.Attr {
							if a.Key == "class" && strings.Contains(a.Val, "tags") {

								for c := node.FirstChild; c != nil; c = c.NextSibling {
									if c.Type == html.ElementNode && c.Data == "span" && c.FirstChild != nil {
										job.Tags = append(job.Tags, c.FirstChild.Data)
									}
								}
							}
						}
					}

					for c := node.FirstChild; c != nil; c = c.NextSibling {
						f(c)
					}
				}

				f(n)

				mu.Lock()
				jobsData = append(jobsData, job)
				mu.Unlock()
			}
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		extractJobs(c)
	}
}

func analyzeFromJobs() {
	mu.Lock()
	defer mu.Unlock()

	for _, job := range jobsData {
		for _, tag := range job.Tags {
			tag = strings.ToLower(tag)

			for _, skill := range skills {
				if strings.Contains(tag, skill) {
					skillFreq[skill]++
				}
			}
		}
	}
}