package main

import (
	"strconv"

	"github.com/j18e/unifi-exporter/unifi"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

func newCollector(cli *unifi.Client, nets map[string]bool) *collector {
	stnLabels := []string{
		"mac",
		"hostname",
		"network",
		"manufacturer",
		"wired",
		"ip",
	}
	return &collector{
		upMetric: prometheus.NewDesc(
			"unifi_controller_up",
			"was talking to the Unifi controller successful",
			nil, nil,
		),
		stnUptimeMetric: prometheus.NewDesc(
			"unifi_station_uptime_seconds",
			"uptime of device connected to Unifi controller's network",
			stnLabels,
			nil,
		),
		stnLastSeenMetric: prometheus.NewDesc(
			"unifi_station_last_seen",
			"unix time when a device was last seen by the Unifi controller",
			stnLabels,
			nil,
		),
		cli:      cli,
		networks: nets,
	}
}

type collector struct {
	upMetric          *prometheus.Desc
	stnUptimeMetric   *prometheus.Desc
	stnLastSeenMetric *prometheus.Desc
	cli               *unifi.Client
	networks          map[string]bool
}

func (c collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.upMetric
	ch <- c.stnUptimeMetric
	ch <- c.stnLastSeenMetric
}

func (c collector) Collect(ch chan<- prometheus.Metric) {
	if err := c.cli.Authenticate(); err != nil {
		ch <- prometheus.MustNewConstMetric(c.upMetric, prometheus.GaugeValue, 0)
		log.Errorf("talking to unifi controller: %v", err)
		return
	}
	ch <- prometheus.MustNewConstMetric(c.upMetric, prometheus.GaugeValue, 1)

	stations, err := c.cli.GetStations()
	if err != nil {
		log.Errorf("getting stations: %w", err)
		return
	}

	for _, s := range stations {
		if !c.networks[s.Network] {
			continue
		}
		ch <- prometheus.MustNewConstMetric(
			c.stnUptimeMetric,
			prometheus.CounterValue,
			float64(s.Uptime),
			s.MAC,
			s.Hostname,
			s.Network,
			s.Manufacturer,
			strconv.FormatBool(s.Wired),
			s.IP,
		)
		ch <- prometheus.MustNewConstMetric(
			c.stnLastSeenMetric,
			prometheus.CounterValue,
			float64(s.LastSeen),
			s.MAC,
			s.Hostname,
			s.Network,
			s.Manufacturer,
			strconv.FormatBool(s.Wired),
			s.IP,
		)
	}
}
