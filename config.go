package main

import (
	"errors"
	"flag"
	"strings"

	log "github.com/sirupsen/logrus"
)

type Cmd int

const (
	PROMETHEUS Cmd = iota
	INFLUXDB
)

type Config struct {
	cmd Cmd

	unifiAddr     string
	unifiUser     string
	unifiPassword string
	unifiInsecure bool

	promListenAddr string

	influxAddr     string
	influxDB       string
	influxSyncFreq int

	nets map[string]bool
}

func getConfig() (Config, error) {
	var conf Config

	// define, parse flags
	unifiAddr := flag.String("unifi.address", "", "Unifi controller's address")
	unifiUser := flag.String("unifi.user", "", "User to connect as")
	unifiPassword := flag.String("unifi.password", "", "Given user's password")
	unifiInsecure := flag.Bool("unifi.insecure", false, "whether to accept invalid TLS certificates")
	watchNets := flag.String("watch-networks", "", "comma separated list of Unifi networks to watch")

	// influx specific
	influxAddr := flag.String("influx.address", "", "InfluxDB server's address")
	influxDB := flag.String("influx.db", "", "name of the database to connect to")
	syncFreq := flag.Int("influx.sync-frequency", 60, "seconds between fetching of metrics")

	// prometheus specific
	listenAddr := flag.String("prom.listen", "0.0.0.0:8080", "local address on which to listen for connections")

	flag.Parse()

	switch flag.Arg(0) {
	case "prometheus":
		log.Info("starting unifi controller prometheus exporter")
		conf.cmd = PROMETHEUS
		conf.promListenAddr = *listenAddr
	case "influxdb":
		log.Info("starting unifi controller influxdb exporter")
		if *influxAddr == "" {
			return conf, errors.New("required flag -influx.address")
		} else if *influxDB == "" {
			return conf, errors.New("required flag -influx.db")
		}
		conf.cmd = INFLUXDB
		conf.influxAddr = *influxAddr
		conf.influxDB = *influxDB
		conf.influxSyncFreq = *syncFreq
	default:
		return conf, errors.New("valid first arg [prometheus influxdb]")
	}

	// check common flags
	if *unifiAddr == "" {
		return conf, errors.New("required flag -unifi.address")
	} else if *unifiUser == "" {
		return conf, errors.New("required flag -unifi.user")
	} else if *unifiPassword == "" {
		return conf, errors.New("required flag -unifi.password")
	} else if *watchNets == "" {
		return conf, errors.New("required flag -watch-networks")
	}
	conf.unifiAddr = *unifiAddr
	conf.unifiUser = *unifiUser
	conf.unifiPassword = *unifiPassword
	conf.unifiInsecure = *unifiInsecure

	// enumerate networks
	nets := make(map[string]bool)
	for _, net := range strings.Split(*watchNets, ",") {
		nets[net] = true
	}
	conf.nets = nets

	return conf, nil
}
