package main

import (
	"fmt"
	"log"

	"github.com/playwright-community/playwright-go"
)

func main() {
	pw, err := playwright.Run()
	if err != nil {
		log.Fatal(err)
	}

	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(true),
	})
	if err != nil {
		log.Fatal(err)
	}

	page, err := browser.NewPage()
	if err != nil {
		log.Fatal(err)
	}

	_, err = page.Goto("https://remoteok.com/remote-dev-jobs")
	if err != nil {
		log.Fatal(err)
	}

	// wait for jobs to load
	_, err = page.WaitForSelector("tr.job")
	if err != nil {
		log.Fatal(err)
	}

	// 🔥 EXTRACTION STARTS HERE
	rows, err := page.QuerySelectorAll("tr.job")
	if err != nil {
		log.Fatal(err)
	}

	for _, row := range rows {

		titleEl, _ := row.QuerySelector("h2")
		tagEls, _ := row.QuerySelectorAll(".tags span")

		title := ""
		if titleEl != nil {
			title, _ = titleEl.InnerText()
		}

		var tags []string
		for _, t := range tagEls {
			txt, _ := t.InnerText()
			tags = append(tags, txt)
		}

		fmt.Println("Title:", title)
		fmt.Println("Tags:", tags)
		fmt.Println("------")
	}

	browser.Close()
	pw.Stop()
}