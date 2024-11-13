package main

import (
	"mirrorselect/internal/llog"
	"strings"

	"github.com/biter777/countries"
	"github.com/gocolly/colly"
)

var currentCC string

func crawlLaunchpad(desiredCC string) (mirrors []Mirror, err error) {
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
					cName := strings.TrimSpace(cell.Text)
					switch cName {
					case "":
						return
					case "Total":
						return
					default:
						country := parseCountry(cName)
						currentCC = country.Alpha2()
						llog.Debugf("Updated country to %s (%s)", country.Info().Name, country.Alpha2())
					}
				} else if cell.Attr("href") != "" {
					link := cell.Attr("href")
					if strings.HasPrefix(link, "http") && currentCC == desiredCC {
						mirror, ok := NewMirror(link)
						if ok {
							mirrors = append(mirrors, mirror)
						}
					}
				}
			})
		})
	})
	c.OnScraped(func(r *colly.Response) {
		llog.Debug("Finished scraping launchpad.net")
	})
	err = c.Visit("https://launchpad.net/ubuntu/+archivemirrors")
	return
}

func parseCountry(c string) countries.CountryCode {
	cS := strings.Split(c, ",")
	if len(cS) == 0 {
		return countries.Unknown
	}
	return countries.ByName(cS[0])
}
