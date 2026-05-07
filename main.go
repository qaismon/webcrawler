package main

import (
	"fmt"
"strings"
	"github.com/playwright-community/playwright-go"
	"net/http"
	"encoding/json"
)

var skillFreq = make(map[string]int)

type SkillStat struct {
	Name string `json:"name"`
	Count int `json:"count"`
	Percent float64 `json: "percent"`
}

var ignoreWords = map[string]bool{
	"senior": true,
	"lead": true,
	"leader": true,
	"engineer": true,
	"executive": true,
	"manager": true,
	"director": true,
	"sales": true,
	"technical": true,
}

func main() {
    http.HandleFunc("/skills", skillHandler)

    fmt.Println("Server running on :8080")
    http.ListenAndServe(":8080", nil)
}

func getSkillStats() []SkillStat {

    skillFreq = make(map[string]int) // reset

    pw, _ := playwright.Run()
    browser, _ := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
        Headless: playwright.Bool(true),
    })

    page, _ := browser.NewPage()
    page.Goto("https://remoteok.com/remote-dev-jobs")
    page.WaitForSelector("tr.job:not(.placeholder)")

    rows, _ := page.QuerySelectorAll("tr.job:not(.placeholder)")

    for _, row := range rows {

        tagEls, _ := row.QuerySelectorAll(".tag")

        seen := make(map[string]bool)

        for _, t := range tagEls {
            txt, _ := t.InnerText()

            txt = strings.TrimSpace(txt)
            txt = strings.ToLower(txt)
            txt = normalizeSkill(txt)

            if ignoreWords[txt] || txt == "" {
                continue
            }

            if seen[txt] {
                continue
            }
            seen[txt] = true

            skillFreq[txt]++
        }
    }

    browser.Close()
    pw.Stop()

    var stats []SkillStat

    total := 0
    for _, v := range skillFreq {
        total += v
    }

    for k, v := range skillFreq {
        stats = append(stats, SkillStat{
            Name:    k,
            Count:   v,
            Percent: float64(v) / float64(total) * 100,
        })
    }

    return stats
}

func skillHandler(w http.ResponseWriter, r *http.Request) {
	stats := getSkillStats()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func normalizeSkill(s string) string {
	switch s {
	case "golang":
		return "go"
	case "js":
		return "javascript"
	default:
		return s
	}
}