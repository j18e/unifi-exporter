package main

import (
	"crypto/tls"
	"flag"
	"net/http"
	"net/http/cookiejar"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	log "github.com/sirupsen/logrus"
)

func main() {
	addr := flag.String("address", "", "Unifi controller's address")
	user := flag.String("user", "", "User to connect as")
	password := flag.String("password", "", "Given user's password")
	insecure := flag.Bool("insecure", false, "whether to accept invalid TLS certificates")
	listenAddr := flag.String("listen", ":8080", "local address on which to listen for connections")
	watchNets := flag.String("watch.networks", "", "comma separated list of Unifi networks to watch")

	flag.Parse()

	if *addr == "" {
		log.Fatal("required flag -address")
	} else if *user == "" {
		log.Fatal("required flag -user")
	} else if *password == "" {
		log.Fatal("required flag -password")
	} else if *watchNets == "" {
		log.Fatal("required flag -watch.networks")
	}

	networks := strings.Split(*watchNets, ",")

	// cookie jar carries auth tokens
	jar, err := cookiejar.New(nil)
	if err != nil {
		log.Fatalf("creating cookie jar: %w", err)
	}

	cli := &Client{
		address:  *addr,
		user:     *user,
		password: *password,
		client: &http.Client{
			Timeout:   time.Second * 5,
			Jar:       jar,
			Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: *insecure}},
		},
	}

	if err := cli.authenticate(); err != nil {
		log.Fatalf("connecting to Unifi controller: %v", err)
	}
	log.Infof("connected to Unifi controller at %s", cli.address)

	col := newCollector(cli, networks)
	prometheus.MustRegister(col)

	http.Handle("/metrics", promhttp.Handler())

	log.Infof("listening for connections on http://%s/metrics...", *listenAddr)
	log.Fatal(http.ListenAndServe(*listenAddr, nil))

}

func newCollector(cli *Client, nets []string) *Collector {
	stnLabels := []string{
		"mac",
		"hostname",
		"network",
		"manufacturer",
		"wired",
		"ip",
	}
	return &Collector{
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

type Collector struct {
	upMetric          *prometheus.Desc
	stnUptimeMetric   *prometheus.Desc
	stnLastSeenMetric *prometheus.Desc
	cli               *Client
	networks          []string
}

func (c Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.upMetric
	ch <- c.stnUptimeMetric
	ch <- c.stnLastSeenMetric
}

func (c Collector) Collect(ch chan<- prometheus.Metric) {
	if err := c.cli.authenticate(); err != nil {
		ch <- prometheus.MustNewConstMetric(c.upMetric, prometheus.GaugeValue, 0)
		log.Errorf("talking to unifi controller: %v", err)
		return
	}
	ch <- prometheus.MustNewConstMetric(c.upMetric, prometheus.GaugeValue, 1)

	stations, err := c.cli.getStations()
	if err != nil {
		log.Errorf("getting stations: %w", err)
		return
	}

	for _, s := range stations {
		if !stringInSlice(s.Network, c.networks) {
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

func stringInSlice(s string, sx []string) bool {
	for _, v := range sx {
		if v == s {
			return true
		}
	}
	return false
}
