package main

import (
	"fmt"
	"strings"

	"github.com/biter777/countries"
	"github.com/gocolly/colly"
)

var currentCountry string

func crawlLaunchpad() (mirrors []Mirror, err error) {
	c := colly.NewCollector(
		colly.AllowedDomains("launchpad.net"), // only visit launchpad.net
		colly.MaxDepth(1),                     // only scrape the first page, dont recurse
	)
	c.OnError(func(r *colly.Response, e error) {
		err = e
	})
	c.OnHTML("table#mirrors_list > tbody", func(h *colly.HTMLElement) {
		h.ForEach("tr", func(_ int, row *colly.HTMLElement) {
			row.ForEach("*", func(_ int, cell *colly.HTMLElement) {
				if cell.Attr("colspan") == "2" {
					switch strings.TrimSpace(cell.Text) {
					case "":
						return
					case "Total":
						return
					default:
						country := countries.ByName(cell.Text)
						fmt.Printf("Country: %s (%s)\n", cell.Text, country.Alpha2())
					}
				}
			})
		})
	})
	c.OnScraped(func(r *colly.Response) {
		fmt.Println("Finished")
	})
	err = c.Visit("https://launchpad.net/ubuntu/+archivemirrors")
	return
}
