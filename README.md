# Prometheus exporter for the Unifi Controller

This Prometheus exporter exposes information about clients, wired or wireless,
connected to networks managed by a Ubiquiti Unifi controller.

## Running
Before using you'll need:
- the network address of the Unifi controller
- a user account on the controller with read permissions. This account will be
  used to talk to the controller's little known REST API.
- the names of the networks you want to monitor. These names are set up in the
  controller itself.

Here's an example which uses the `-insecure` option as we haven't set up a TLS
certificate for the controller.
```
./unifi-exporter \
    -address 192.168.1.10:8443 \
    -user unifi-exporter \
    -password secret \
    -insecure \
    -watch.networks LAN1,LAN2
```

### Running in Docker
```
docker run -it -p 8080:8080 j18e/unifi-exporter:latest \
    -address 192.168.1.10:8443 \
    -user unifi-exporter \
    -password secret \
    -insecure \
    -watch.networks LAN1,LAN2
```

## Unifi Controller API
The controller's stations endpoint returns a JSON formatted list of "stations",
ie devices connected to Unifi's LAN or WLAN. Each station has a number of useful
fields. Below is example response, with some fields omitted. Our exporter only
uses a few of the below fields, but it could be easily extended.

```json
{
  "meta": {
    "rc": "ok"
  },
  "data": [
    {
      "assoc_time": 1550017627,
      "latest_assoc_time": 1551865546,
      "oui": "Ubiquiti",
      "mac": "18:e8:29:44:44:44",
      "is_guest": false,
      "first_seen": 1543875005,
      "last_seen": 1551865586,
      "is_wired": true,
      "hostname": "unifi",
      "name": "UNIFI",
      "network": "LAN1",
      "ip": "10.0.1.10",
      "uptime": 1847959,
      "gw_mac": "b4:fb:e4:44:44:44",
      "tx_bytes": 238407808,
      "rx_bytes": 9125866,
      "tx_packets": 185367,
      "rx_packets": 97485,
      "bytes-r": 0,
      "tx_bytes-r": 0,
      "rx_bytes-r": 0,
      "authorized": true,
      "qos_policy_applied": true
    },
    {
      "assoc_time": 1550017629,
      "latest_assoc_time": 1551865167,
      "oui": "Apple",
      "mac": "a0:ed:cd:44:44:44",
      "is_guest": false,
      "first_seen": 1549126674,
      "last_seen": 1551865586,
      "is_wired": true,
      "hostname": "Apple-TV",
      "network": "LAN2",
      "ip": "10.0.10.50",
      "uptime": 1847957,
      "gw_mac": "b4:fb:e4:44:44:44",
      "tx_bytes": 166171,
      "rx_bytes": 140656,
      "tx_packets": 993,
      "rx_packets": 1284,
      "bytes-r": 0,
      "tx_bytes-r": 0,
      "rx_bytes-r": 0,
      "authorized": true,
      "qos_policy_applied": true
    }
  ]
}
```
