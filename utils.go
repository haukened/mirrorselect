package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"mirrorselect/internal/llog"
	"net/http"
	"os"
	"strings"
)

// contains checks if a given string is present in a slice of strings.
// It returns true if the string is found, and false otherwise.
//
// Parameters:
//
//	slice []string - the slice of strings to search within
//	s string - the string to search for
//
// Returns:
//
//	bool - true if the string is found in the slice, false otherwise
func contains(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

// getDistribCodename reads the /etc/lab-release file to find and return the
// distribution codename. It looks for a line that starts with "DISTRIB_CODENAME="
// and returns the value after the equals sign. If the file cannot be opened or
// read, or if the codename is not found, it returns an error.
func getDistribCodename() (string, error) {
	file, err := os.Open("/etc/lsb-release")
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "DISTRIB_CODENAME=") {
			return strings.TrimPrefix(line, "DISTRIB_CODENAME="), nil
		}
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return "", nil
}

type geoIP struct {
	IP          string `json:"ip"`
	ASO         string `json:"aso"`
	ASN         string `json:"asn"`
	Continent   string `json:"continent"`
	CountryCode string `json:"cc"`
	CountryName string `json:"country"`
	City        string `json:"city"`
	PostalCode  string `json:"postal"`
	Latitude    string `json:"latitude"`
	Longitude   string `json:"longitude"`
	Timezone    string `json:"tz"`
}

// getGeoIP fetches the geoIP information from https://ident.me/json and parses it into a geoIP struct
func getGeoIP() (*geoIP, error) {
	llog.Debug("Fetching geoIP data")
	resp, err := http.Get("https://ident.me/json")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	llog.Debugf("Response status: %s", resp.Status)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch geoIP data: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var geo geoIP
	if err := json.Unmarshal(body, &geo); err != nil {
		return nil, err
	}
	llog.Debugf("GeoIP data: %+v", geo)
	return &geo, nil
}

func humanizeTransferSpeed(bytes int64, seconds float64) string {
	bits := bytes * 8
	if seconds == 0 {
		return "0 b/s"
	}
	speed := float64(bits) / seconds
	units := []string{"b/s", "Kbps", "Mbps", "Gbps", "Tbps"}
	for _, unit := range units {
		if speed < 1024 {
			return fmt.Sprintf("%4.2f %s", speed, unit)
		}
		speed /= 1024
	}
	return fmt.Sprintf("%.2f %s", speed, "Tbps")
}
