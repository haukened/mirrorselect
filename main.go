package main

import (
	"fmt"
	"log"
	"mirrorselect/internal/llog"
	"os"
	"runtime"
	"sort"
	"strings"

	"github.com/urfave/cli/v2"
)

var FinalMirrors []Mirror

func main() {
	app := &cli.App{
		Name:  "mirrorselect",
		Usage: "An application to select fastest mirrors for Ubuntu",
		Authors: []*cli.Author{{
			Name:  "David Haukeness",
			Email: "david@hauken.us",
		}},
		Copyright: "(C) 2024 David Haukeness, distributed under the GNU GPL v3 license",
		Action:    run,
		Before:    before,
		After:     after,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "arch",
				Aliases:     []string{"a"},
				Usage:       "Architecture to select mirrors for (amd64, i386, arm64, armhf, ppc64el, riscv64, s390x)",
				Value:       runtime.GOARCH,
				DefaultText: runtime.GOARCH,
				Category:    "System Options",
				Action: func(c *cli.Context, v string) error {
					allowedArchs := []string{"amd64", "arm64", "armhf", "ppc64el", "riscv64", "s390x"}
					value := strings.ToLower(v)
					if !contains(allowedArchs, value) {
						return fmt.Errorf("invalid architecture: %s", v)
					}
					return nil
				},
			},
			&cli.StringFlag{
				Name:        "country",
				Aliases:     []string{"c"},
				Usage:       "Country to select mirrors from (ISO 3166-1 alpha-2 country code)",
				Value:       "",
				DefaultText: "Auto Select Based on IP address",
				Category:    "Mirror Options",
				Action: func(c *cli.Context, v string) error {
					// ensure the country code is two characters
					if len(v) != 2 {
						return fmt.Errorf("invalid country code: %s", v)
					}
					// ensure the country code is uppercase
					c.Set("country", strings.ToUpper(v))
					return nil
				},
			},
			&cli.IntFlag{
				Name:        "max",
				Aliases:     []string{"m"},
				Usage:       "Maximum number of mirrors to test (if available)",
				Value:       5,
				DefaultText: "5",
				Category:    "Mirror Options",
			},
			&cli.StringFlag{
				Name:        "protocol",
				Aliases:     []string{"p"},
				Usage:       "Protocol to select mirrors for (http, https, any)",
				Value:       "any",
				DefaultText: "any",
				Category:    "Mirror Options",
			},
			&cli.StringFlag{
				Name:        "release",
				Aliases:     []string{"r"},
				Usage:       "Release to select mirrors for",
				DefaultText: "The current system release",
				Category:    "System Options",
			},
			&cli.IntFlag{
				Name:        "timeout",
				Aliases:     []string{"t"},
				Usage:       "Timeout for testing mirrors in milliseconds",
				Value:       500,
				DefaultText: "500",
				Category:    "Mirror Options",
			},
			&cli.StringFlag{
				Name:        "verbosity",
				Aliases:     []string{"v"},
				Usage:       "Set the log verbosity level (DEBUG, INFO, WARN, ERROR)",
				Value:       "INFO",
				DefaultText: "INFO",
				Category:    "System Options",
				Action: func(c *cli.Context, v string) error {
					allowedLevels := []string{"DEBUG", "INFO", "WARN", "ERROR"}
					value := strings.ToUpper(v)
					if !contains(allowedLevels, value) {
						return fmt.Errorf("invalid log level: %s", v)
					}
					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

// after is a function that is called after the main action is executed.
// context is available to the function and can be used to access the flags
// usefull for cleanup or finalization tasks
// after runs even if the main action or before action fails
func after(c *cli.Context) error {
	if len(FinalMirrors) == 0 {
		llog.Info("Flag options resulted in no mirrors being selected")
	} else {
		for i, mirror := range FinalMirrors {
			fmt.Printf("%d. %s %s\n", i+1, humanizeTransferSpeed(mirror.Size, mirror.Time), mirror.URL)
		}
	}
	return nil
}

// before is a function that is called before the main action is executed.
// context is available to the function and can be used to access the flags
// usefull for setup tasks, the main action will not run if before returns an error
func before(c *cli.Context) error {
	// set the logging level
	err := llog.SetLogLevel(c.String("verbosity"))
	if err != nil {
		return err
	}
	// get the distribution codename
	if !c.IsSet("release") {
		codename, err := getDistribCodename()
		if err != nil {
			return err
		}
		llog.Infof("Detected distribution codename %s", codename)
		c.Set("release", codename)
	}
	// ensure we have a country code
	if !c.IsSet("country") {
		// get the country code from the geoIP
		geo, err := getGeoIP()
		if err != nil {
			llog.Error("Unable to auto-detect country code, please specify one manually using --country")
			return err
		}
		llog.Infof("Using public IP address %s", geo.IP)
		llog.Infof("Detected country %s (%s)", geo.CountryName, geo.CountryCode)
		c.Set("country", geo.CountryCode)
	}
	return nil
}

// run is the main action that is executed when the application is run
// context is available to the function and can be used to access the flags
func run(c *cli.Context) error {
	// get the mirrors
	mirrors, err := getMirrors(c.String("country"), c.String("protocol"), c.String("arch"))
	if err != nil {
		return err
	}

	// test the mirrors for latency
	llog.Infof("Testing %d mirrors", len(mirrors))
	timeout := c.Int("timeout")
	for idx := range len(mirrors) {
		// grab the pointer to the mirror so it can self-update
		mirror := &mirrors[idx]
		// test the latency and validity of the mirror
		mirror.TestLatency(timeout, c.String("release"))
	}

	// filter out the invalid mirrors
	mirrors = filterInvalidMirrors(mirrors)

	// get the top N mirrors
	mirrors = TopNByLatency(mirrors, c.Int("max"))

	// then test the mirrors for download speed
	for idx := range len(mirrors) {
		// grab the pointer to the mirror so it can self-update
		mirror := &mirrors[idx]
		// test the download speed of the mirror
		mirror.TestDownload(c.String("release"))
	}

	sort.Sort(ByTransferSpeed(mirrors))
	FinalMirrors = mirrors

	return nil
}
