package cloudflare

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
)

// A Client is a CloudFlare API v4 client.
type Client struct {
	*http.Client
	f []func(*http.Request) *http.Request
}

// New returns a new Client. f will be run on all requests.
func New(f ...func(r *http.Request) *http.Request) *Client {
	return &Client{
		&http.Client{},
		append([]func(*http.Request) *http.Request{defaultEndpoint}, f...),
	}
}

// Get returns a new GET request
func (c *Client) Get(path string, f ...func(*http.Request) *http.Request) *http.Request {
	r := NewRequest(http.MethodGet, nil)
	r = apply(r, c.f...)
	r = apply(r, URL(Path(path)))
	r = apply(r, f...)
	return r
}

// Do configures and performs the request.
func (c *Client) Do(r *http.Request) (*http.Response, error) {
	client, f := c.Client, c.f
	if client == nil {
		client = &http.Client{}
	}
	return client.Do(apply(r, f...))
}

// GetZones gets zones.
func (c *Client) GetZones(f ...func(*http.Request) *http.Request) ([]*Zone, error) {
	req := c.Get("zones")
	req = apply(req, Header("Accept")("application/json"))
	req = apply(req, Header("Content-Type")("application/json"))
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s responded with %d %s", req.URL, resp.StatusCode, resp.Status)
	}
	zr := &struct {
		*metadata
		Info   *ResultInfo `json:"result_info"`
		Result []*Zone     `json:"result"`
	}{}
	if err := json.NewDecoder(resp.Body).Decode(&zr); err != nil {
		return nil, err
	}
	if success := zr.Success; !success {
		return nil, zr.metadata
	}
	zones := zr.Result
	info := zr.Info
	if (info.Page-1)*info.PerPage+info.Count >= info.TotalCount {
		return zones, nil
	}
	next, err := c.GetZones(URL(Query("page", strconv.Itoa(info.Page+1))))
	if err != nil {
		return nil, err
	}
	return append(zones, next...), nil
}

// GetDNSRecords gets DNS records for the zone.
func (c *Client) GetDNSRecords(zoneID string) ([]*Record, error) {
	return nil, errors.New("cloudflare.Client.GetDNSRecords not implemented")
}

// PostDNSRecord creates a new DNS record.
func (c *Client) PostDNSRecord(record *Record) (*Record, error) {
	return nil, errors.New("cloudflare.Client.GetDNSRecords not implemented")
}

// PutDNSRecord replaces a DNS record.
func (c *Client) PutDNSRecord(record *Record) error {
	return errors.New("cloudflare.Client.GetDNSRecords not implemented")
}

// DefaultClient is the default CloudFlare client.
var DefaultClient = &Client{
	&http.Client{},
	[]func(*http.Request) *http.Request{
		defaultEndpoint,
		Token(os.Getenv("CLOUDFLARE_TOKEN")),
	},
}

// A Response wraps a http.Response and provides methods for parsing the body
// and pagination.
type Response struct {
	*http.Response
	*Result
}
