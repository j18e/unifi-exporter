package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Client struct {
	address  string
	user     string
	password string
	insecure bool
	client   *http.Client
}

type Station struct {
	MAC          string `json:"mac"`
	Hostname     string `json:"hostname"`
	Uptime       int    `json:"uptime"`
	Network      string `json:"network"`
	LastSeen     int    `json:"last_seen"`
	Manufacturer string `json:"oui"`
	Wired        bool   `json:"is_wired"`
}

func (c *Client) getStations() ([]*Station, error) {
	url := c.address + "/api/s/default/stat/sta"

	var data struct {
		Stations []*Station `json:"data"`
	}

	res, err := c.client.Get(url)
	if err != nil {
		return data.Stations, fmt.Errorf("getting %s: %w", url, err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return data.Stations, fmt.Errorf("got status %d", res.StatusCode)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return data.Stations, fmt.Errorf("reading response: %w", err)
	}

	if err := json.Unmarshal(body, &data); err != nil {
		return data.Stations, fmt.Errorf("unmarshaling response: %w", err)
	}
	return data.Stations, err
}

func (c *Client) authenticate() error {
	url := fmt.Sprintf("https://%s/api/login", c.address)

	auth := map[string]string{
		"username": c.user,
		"password": c.password,
	}
	buf := new(bytes.Buffer)
	json.NewEncoder(buf).Encode(auth)

	res, err := c.client.Post(url, "application/json", buf)
	if err != nil {
		return fmt.Errorf("contacting server: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return fmt.Errorf("got status %d", res.StatusCode)
	}

	return nil
}
