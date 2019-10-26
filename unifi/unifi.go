package unifi

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"time"
)

// Client provides a means of authenticating against and communicating with a
// Unifi controller.
type Client struct {
	address  string
	user     string
	password string
	insecure bool
	client   *http.Client
}

// NewClient configures and tests a Client object.
func NewClient(addr, user, pass string, insecure bool) (*Client, error) {
	// cookie jar carries auth tokens
	jar, err := cookiejar.New(nil)
	if err != nil {
		return &Client{}, fmt.Errorf("creating cookie jar: %w", err)
	}

	cli := &Client{
		address:  addr,
		user:     user,
		password: pass,
		client: &http.Client{
			Timeout:   time.Second * 5,
			Jar:       jar,
			Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: insecure}},
		},
	}

	if err := cli.Authenticate(); err != nil {
		return cli, fmt.Errorf("connecting to Unifi controller: %w", err)
	}

	return cli, nil
}

// Station is a result object coming from Unifi's API. Many of the available
// fields returned in the JSON have been excluded from this implementation.
type Station struct {
	MAC          string `json:"mac"`
	IP           string `json:"ip"`
	Hostname     string `json:"hostname"`
	Uptime       int    `json:"uptime"`
	Network      string `json:"network"`
	LastSeen     int    `json:"last_seen"`
	Manufacturer string `json:"oui"`
	Wired        bool   `json:"is_wired"`
}

// GetStations calls the Unifi controller's sta API endpoint and marshals the
// response into a slice of Stations.
func (c *Client) GetStations() ([]*Station, error) {
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

	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		return data.Stations, fmt.Errorf("decoding response: %w", err)
	}

	return data.Stations, err
}

// Authenticate authenticates against the Unifi controller using credentials
// found in the Client object and stores the resulting token in the Client's
// cookie jar.
func (c *Client) Authenticate() error {
	url := c.address + "/api/login"

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
