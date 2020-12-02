package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/mmcloughlin/geohash"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type randomness struct {
	Randomness string `json:"randomness"`
}

type ipLocation struct {
	CountryCode string  `json:"countryCode"`
	Latitude    float64 `json:"lat"`
	Longitude   float64 `json:"lon"`
}

const defaultURL = "localhost"

var (
	drngStatus = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "drng_status",
			Help: "status of the drng instance",
		},
		[]string{"location", "geo"},
	)

	countryCode  string
	geohashValue string
	urlPtr       string
	drngPort     string
)

func recordMetrics() {

	if urlPtr != defaultURL {
		log.Printf("\tINFO\tARGS\t\tNon default url [%s] will be monitored", urlPtr)
	} else {
		log.Printf("\tINFO\tARGS\t\tDefault url [%s] will be monitored", defaultURL)
	}

	go func() {
		for {
			//Fetch url/ip location
			if urlPtr != defaultURL {
				countryCode, geohashValue, _ = getLocationFromIP(urlPtr)
			} else {
				countryCode, geohashValue, _ = getLocationFromIP("")
			}
			log.Printf("\tINFO\tLOCATION\tFetched country code [%s] and geohash [%s]", countryCode, geohashValue)

			//Fetch drng status and randomness
			message, err := http.Get("http://" + urlPtr + ":" + drngPort + "/public/latest")
			if err != nil {
				log.Println("\tINFO\tDRNG\t\tNode is offline")
				drngStatus.WithLabelValues(countryCode, geohashValue).Set(0)
			} else {
				body, err := ioutil.ReadAll(message.Body)
				if err != nil {
					log.Fatalln(err)
				}
				defer message.Body.Close()

				drngRandomness := randomness{}

				err = json.Unmarshal(body, &drngRandomness)
				if err != nil {
					log.Fatalln(err)
				}
				log.Printf("\tINFO\tDRNG\t\tFetched drng randomness [%s]", drngRandomness.Randomness)
				log.Println("\tINFO\tDRNG\t\tNode is online")
				drngStatus.WithLabelValues(countryCode, geohashValue).Set(1)
			}

			time.Sleep(3 * time.Second)
		}
	}()
}

func getLocationFromIP(ip string) (countryCode, geohashValue string, err error) {
	message, err := http.Get("http://ip-api.com/json/" + ip)
	if err != nil {
		log.Printf("\tERROR\tLOCATION\t%s", "err")
		return "", "", err
	}
	body, err := ioutil.ReadAll(message.Body)
	if err != nil {
		log.Fatalln("err")
	}

	location := ipLocation{}

	err = json.Unmarshal(body, &location)
	if err != nil {
		log.Printf("\tERROR\tLOCATION\t%s", "err")
		return "", "", err
	}

	geohashValue = geohash.Encode(location.Latitude, location.Longitude)
	countryCode = location.CountryCode
	return countryCode, geohashValue, nil
}

func main() {
	flag.StringVar(&urlPtr, "url", defaultURL, "the url to monitor")
	flag.StringVar(&drngPort, "drngPort", "8081", "the drng public-listen port")
	prometheusPort := flag.String("port", "2112", "prometheus metrics port")
	flag.Parse()
	recordMetrics()

	log.Printf("\tINFO\tPROMETHEUS\tExporting prometheus metrics on [localhost:%s]", *prometheusPort)
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":"+*prometheusPort, nil)
}
