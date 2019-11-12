package main

import (
	"fmt"
	"strconv"
	"time"

	_ "github.com/influxdata/influxdb1-client" // this is important because of the bug in go mod
	influx "github.com/influxdata/influxdb1-client/v2"
	"github.com/j18e/unifi-exporter/unifi"
)

func influxLoop(uCli *unifi.Client, idb influx.Client, dbName string, nets map[string]bool) error {
	const METRIC_NAME = "unifi_client"

	// get clients connected to Unifi networks
	stations, err := uCli.GetStations()
	if err != nil {
		return fmt.Errorf("getting stations from unifi: %w", err)
	}

	bp, err := influx.NewBatchPoints(influx.BatchPointsConfig{Database: dbName, Precision: "s"})
	if err != nil {
		return fmt.Errorf("creating influxdb batch points: %w", err)
	}

	for _, sta := range stations {
		// only consider devices on included unifi networks
		if !nets[sta.Network] {
			continue
		}

		tags := map[string]string{
			"mac":      sta.MAC,
			"hostname": sta.Hostname,
			"wired":    strconv.FormatBool(sta.Wired),
		}
		fields := map[string]interface{}{
			"uptime":  sta.Uptime,
			"ip":      sta.IP,
			"network": sta.Network,
		}
		pt, err := influx.NewPoint(METRIC_NAME, tags, fields, time.Unix(int64(sta.LastSeen), 0))
		if err != nil {
			return fmt.Errorf("creating influxdb point: %w", err)
		}
		bp.AddPoint(pt)
	}

	if err := idb.Write(bp); err != nil {
		return fmt.Errorf("writing metrics to influxdb: %w", err)
	}
	return nil
}
