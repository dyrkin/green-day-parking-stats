package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
)

const (
	universityParkingURL = "https://gd.zaparkuj.pl/api/freegroupcountervalue-green.json"
	greenDayParkingURL   = "https://gd.zaparkuj.pl/api/freegroupcountervalue.json"
	port                 = ":2112"
)

func main() {
	registry := prometheus.NewRegistry()

	registerGauge(registry, "gd_zaparkuj_pl_parking_green_day", "The total number of available parking spaces in Green Day parking lot", greenDayParkingURL)
	registerGauge(registry, "gd_zaparkuj_pl_parking_university", "The total number of available parking spaces in University parking lot", universityParkingURL)

	handler := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	http.Handle("/metrics", handler)
	log.Printf("Listening on port %s", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

func registerGauge(registry *prometheus.Registry, name, help, url string) {
	gaugeFunc := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: name,
		Help: help,
	}, func() float64 {
		res, err := fetchParkingData(url)
		if err != nil {
			log.Printf("Error fetching parking data for %s: %v", name, err)
			return 0
		}
		return res
	})
	registry.MustRegister(gaugeFunc)
}

func fetchParkingData(url string) (float64, error) {
	resp, err := http.Get(url)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch parking data from %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("failed to fetch parking data from %s: non-200 status code: %s", url, resp.Status)
	}

	var fields map[string]interface{}
	if err = json.NewDecoder(resp.Body).Decode(&fields); err != nil {
		return 0, fmt.Errorf("failed to decode parking data from %s: %w", url, err)
	}

	value, ok := fields["CurrentFreeGroupCounterValue"].(float64)
	if !ok {
		return 0, errors.New("invalid data format: 'CurrentFreeGroupCounterValue' is not a float64")
	}

	return value, nil
}
