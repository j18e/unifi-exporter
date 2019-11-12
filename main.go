package main

import (
	"net/http"
	"time"

	_ "github.com/influxdata/influxdb1-client" // this is important because of the bug in go mod
	influx "github.com/influxdata/influxdb1-client/v2"
	"github.com/j18e/unifi-exporter/unifi"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	log "github.com/sirupsen/logrus"
)

func main() {
	conf, err := getConfig()
	if err != nil {
		log.Fatal(err)
	}

	// connect to unifi controller
	log.Infof("connecting to Unifi controller at %s with insecure=%t", conf.unifiAddr, conf.unifiInsecure)
	unifiCli, err := unifi.NewClient(conf.unifiAddr, conf.unifiUser, conf.unifiPassword, conf.unifiInsecure)
	if err != nil {
		log.Fatalf("creating unifi client: %v", err)
	}
	log.Info("success")

	switch conf.cmd {
	case PROMETHEUS:
		// initialize metrics
		col := newCollector(unifiCli, conf.nets)
		prometheus.MustRegister(col)

		// expose metrics and start server
		log.Infof("listening for connections on http://%s/metrics...", conf.promListenAddr)
		http.Handle("/metrics", promhttp.Handler())
		log.Fatal(http.ListenAndServe(conf.promListenAddr, nil))

	case INFLUXDB:
		// connect to influxdb
		influxCli, err := influx.NewHTTPClient(influx.HTTPConfig{Addr: conf.influxAddr})
		if err != nil {
			log.Fatalf("connecting to influxdb: %v", err)
		}
		defer influxCli.Close()
		if _, _, err := influxCli.Ping(time.Second * 5); err != nil {
			log.Fatalf("pinging influxdb: %v", err)
		}

		log.Infof("running sync every %d seconds", conf.influxSyncFreq)
		// allow for number of failed syncs every hour
		failures := 0
		hourlyFailureLimit := 10
		lastReset := time.Now()
		for {
			if time.Since(lastReset) > time.Hour {
				failures = 0
				lastReset = time.Now()
			}
			if failures > hourlyFailureLimit {
				log.Fatal("out of retries")
			}
			if err := influxLoop(unifiCli, influxCli, conf.influxDB, conf.nets); err != nil {
				log.Errorf("running sync: %v", err)
				failures++
			}
			time.Sleep(time.Second * time.Duration(conf.influxSyncFreq))
		}
	}
}
