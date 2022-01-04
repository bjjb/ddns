package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strconv"
	"time"

	"github.com/spf13/cobra"
)

// A cloudflare implements dnsManager. It requires an auth, which is either an
// API token or an email/API-key pair, but if that's blank, it will attempt to
// read its auth from the DDNS_CLOUDFLARE_AUTH environment variable.
type cloudflare struct {
	baseURL string
	auth    string
	http    *http.Client
	cmd     *cobra.Command
	verbose bool
	zones   []*struct{ id, name string }
	records map[string][]*struct {
		id, name, kind, content string
		ttl                     int
		proxied                 bool
	}
}

// ownsRecord returns true if the struct is configured, and the given name
// fits within one of its available zones.
func (c *cloudflare) ownsRecord(name string) (bool, error) {
	if c.getAuth() == "" {
		return false, nil
	}
	z, err := c.getZones()
	if err != nil {
		return false, err
	}
	for _, z := range z {
		if regexp.MustCompile(fmt.Sprintf("%s$", z.name)).MatchString(name) {
			return true, nil
		}
	}
	return false, nil
}

// createOrUpdateRecord creates or updates a record with the given name, kind
// (record-type), content and TTL value.
func (c *cloudflare) createOrUpdateRecord(
	name, kind, content string,
	ttl time.Duration,
) error {
	if c.records == nil {
		c.records = make(map[string][]*struct {
			id, name, kind, content string
			ttl                     int
			proxied                 bool
		})
	}
	if c.getAuth() == "" {
		return fmt.Errorf("cloudflare not configured")
	}
	zones, err := c.getZones()
	if err != nil {
		return err
	}
	for _, z := range zones {
		if regexp.MustCompile(fmt.Sprintf("%s$", z.name)).MatchString(name) {
			records, err := c.getRecords(z.id)
			if err != nil {
				return err
			}
			for _, r := range records {
				if r.name == name && r.kind == kind {
					if err := c.updateRecord(z.id, r.id, content, c.ttl(ttl)); err != nil {
						return err
					}
					return nil
				}
			}
			if err := c.createRecord(z.id, name, kind, content, c.ttl(ttl)); err != nil {
				return err
			}
			delete(c.records, z.id)
			return nil
		}
	}
	return fmt.Errorf("no zone found for %s", name)
}

func (c *cloudflare) updateRecord(zoneID, id, content string, ttl int) error {
	if c.cmd != nil && c.verbose {
		c.cmd.Printf(
			"cloudflare updating %s record %s with %s (ttl=%d)...\n",
			zoneID,
			id,
			content,
			ttl,
		)
	}
	record := &struct {
		Content string `json:"content"`
		TTL     int    `json:"ttl"`
		Proxied bool   `json:"proxied"`
	}{content, ttl, false}
	path := fmt.Sprintf("zones/%s/dns_records/%s", zoneID, id)
	resp, err := c.patch(path, record)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf(
			"%s %d - %s",
			resp.Request.URL.String(),
			resp.StatusCode,
			resp.Status,
		)
	}
	defer resp.Body.Close()
	if resp.Header.Get("Content-Type") != "application/json" {
		return fmt.Errorf(
			"%s Content-Type unexpected - %s",
			resp.Request.URL.String(),
			resp.Header.Get("Content-Type"),
		)
	}
	result := &struct {
		Success bool `json:"success"`
		Errors  []*struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"errors"`
	}{}
	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return err
	}
	if !result.Success {
		return fmt.Errorf(
			"errors from %s - %v",
			resp.Request.URL.String(),
			result.Errors,
		)
	}
	delete(c.records, zoneID)
	return nil
}

func (c *cloudflare) createRecord(
	zoneID, name, kind, content string,
	ttl int,
) error {
	if c.cmd != nil && c.verbose {
		c.cmd.Printf(
			"cloudflare creating %s record %s (in zone %s) with %s (ttl=%d)...\n",
			kind,
			name,
			zoneID,
			content,
			ttl,
		)
	}
	record := &struct {
		Name    string `json:"name"`
		Type    string `json:"type"`
		Content string `json:"content"`
		TTL     int    `json:"ttl"`
		Proxied bool   `json:"proxied"`
	}{name, kind, content, ttl, false}
	path := fmt.Sprintf("zones/%s/dns_records", zoneID)
	resp, err := c.post(path, record)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf(
			"%s %d - %s",
			resp.Request.URL.String(),
			resp.StatusCode,
			resp.Status,
		)
	}
	defer resp.Body.Close()
	if resp.Header.Get("Content-Type") != "application/json" {
		return fmt.Errorf(
			"%s Content-Type unexpected - %s",
			resp.Request.URL.String(),
			resp.Header.Get("Content-Type"),
		)
	}
	result := &struct {
		Success bool `json:"success"`
		Errors  []*struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"errors"`
	}{}
	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return err
	}
	if !result.Success {
		return fmt.Errorf(
			"errors from %s - %v",
			resp.Request.URL.String(),
			result.Errors,
		)
	}
	return nil
}

// get makes a GET request to the given path.
func (c *cloudflare) get(resource string, page int) (*http.Response, error) {
	if c.baseURL == "" {
		c.baseURL = "https://api.cloudflare.com/client/v4"
	}
	u, err := url.Parse(c.baseURL)
	if err != nil {
		return nil, err
	}
	u.Path = path.Join(u.Path, resource)
	if page != 0 {
		u.RawQuery = url.Values{"page": []string{strconv.Itoa(page)}}.Encode()
	}
	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	if email, apiKey := c.email(), c.apiKey(); email != "" && apiKey != "" {
		req.SetBasicAuth(email, apiKey)
	}
	if token := c.token(); token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	}
	req.Header.Set("Accept", "application/json")
	return c.httpClient().Do(req)
}

// do makes a request to the given resources, serialising i as JSON.
func (c *cloudflare) post(resources string, i interface{}) (
	*http.Response, error,
) {
	if c.baseURL == "" {
		c.baseURL = "https://api.cloudflare.com/client/v4"
	}
	u, err := url.Parse(c.baseURL)
	if err != nil {
		return nil, err
	}
	u.Path = path.Join(u.Path, resources)
	b := new(bytes.Buffer)
	if err := json.NewEncoder(b).Encode(i); err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, u.String(), b)
	if err != nil {
		return nil, err
	}
	if email, apiKey := c.email(), c.apiKey(); email != "" && apiKey != "" {
		req.SetBasicAuth(email, apiKey)
	}
	if token := c.token(); token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	}
	req.Header.Set("Content-Type", "application/json")
	return c.httpClient().Do(req)
}

// patch makes a PATCH request to the given resources, serialising i as JSON.
func (c *cloudflare) patch(resources string, i interface{}) (
	*http.Response, error,
) {
	if c.baseURL == "" {
		c.baseURL = "https://api.cloudflare.com/client/v4"
	}
	u, err := url.Parse(c.baseURL)
	if err != nil {
		return nil, err
	}
	u.Path = path.Join(u.Path, resources)
	b := new(bytes.Buffer)
	if err := json.NewEncoder(b).Encode(i); err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPatch, u.String(), b)
	if err != nil {
		return nil, err
	}
	if email, apiKey := c.email(), c.apiKey(); email != "" && apiKey != "" {
		req.SetBasicAuth(email, apiKey)
	}
	if token := c.token(); token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	}
	req.Header.Set("Content-Type", "application/json")
	return c.httpClient().Do(req)
}

// getZones returns all zones for this instance.
func (c *cloudflare) getZones() ([]*struct{ id, name string }, error) {
	if c.zones != nil {
		return c.zones, nil
	}
	zones := []*struct{ id, name string }{}
	for page := 0; c.zones == nil; page++ {
		resp, err := c.get("zones", page)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf(
				"%s %d - %s",
				resp.Request.URL.String(),
				resp.StatusCode,
				resp.Status,
			)
		}
		defer resp.Body.Close()
		if resp.Header.Get("Content-Type") != "application/json" {
			return nil, fmt.Errorf(
				"%s Content-Type unexpected - %s",
				resp.Request.URL.String(),
				resp.Header.Get("Content-Type"),
			)
		}
		result := &struct {
			Success bool `json:"success"`
			Errors  []*struct {
				Code    int    `json:"code"`
				Message string `json:"message"`
			} `json:"errors"`
			ResultInfo *struct {
				Page       int `json:"page"`
				PerPage    int `json:"per_page"`
				Count      int `json:"count"`
				TotalCount int `json:"total_count"`
			} `json:"result_info"`
			Result []*struct {
				ID     string `json:"id"`
				Name   string `json:"name"`
				Status string `json:"status"`
			} `json:"result"`
		}{}
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return nil, err
		}
		if !result.Success {
			return nil, fmt.Errorf(
				"errors from %s - %v",
				resp.Request.URL.String(),
				result.Errors,
			)
		}
		for _, z := range result.Result {
			zones = append(zones, &struct{ id, name string }{z.ID, z.Name})
		}
		n, tc := result.ResultInfo.Count, result.ResultInfo.TotalCount
		p, pp := result.ResultInfo.Page, result.ResultInfo.PerPage
		if ((p-1)*pp)+n >= tc {
			c.zones = zones
		}
	}
	return c.zones, nil
}

// getRecords gets all the records for a given zone.
func (c *cloudflare) getRecords(zone string) ([]*struct {
	id, name, kind, content string
	ttl                     int
	proxied                 bool
}, error) {
	if c.records[zone] != nil {
		return c.records[zone], nil
	}
	records := []*struct {
		id, name, kind, content string
		ttl                     int
		proxied                 bool
	}{}
	for page := 0; c.records[zone] == nil; page++ {
		resp, err := c.get(fmt.Sprintf("zones/%s/dns_records", zone), page)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf(
				"%s %d - %s",
				resp.Request.URL.String(),
				resp.StatusCode,
				resp.Status,
			)
		}
		defer resp.Body.Close()
		if resp.Header.Get("Content-Type") != "application/json" {
			return nil, fmt.Errorf(
				"%s Content-Type unexpected - %s",
				resp.Request.URL.String(),
				resp.Header.Get("Content-Type"),
			)
		}
		result := &struct {
			Success bool `json:"success"`
			Errors  []*struct {
				Code    int    `json:"code"`
				Message string `json:"message"`
			} `json:"errors"`
			ResultInfo *struct {
				Page       int `json:"page"`
				PerPage    int `json:"per_page"`
				Count      int `json:"count"`
				TotalCount int `json:"total_count"`
			} `json:"result_info"`
			Result []*struct {
				ID      string `json:"id"`
				Name    string `json:"name"`
				Type    string `json:"type"`
				Content string `json:"content"`
				TTL     int    `json:"ttl"`
				Proxied bool   `json:"proxied"`
				Status  string `json:"status"`
			} `json:"result"`
		}{}
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return nil, err
		}
		if !result.Success {
			return nil, fmt.Errorf(
				"errors from %s - %v",
				resp.Request.URL.String(),
				result.Errors,
			)
		}
		for _, r := range result.Result {
			records = append(records, &struct {
				id, name, kind, content string
				ttl                     int
				proxied                 bool
			}{r.ID, r.Name, r.Type, r.Content, r.TTL, r.Proxied})
		}
		n, tc := result.ResultInfo.Count, result.ResultInfo.TotalCount
		p, pp := result.ResultInfo.Page, result.ResultInfo.PerPage
		if ((p-1)*pp)+n >= tc {
			c.records[zone] = records
		}
	}
	return c.records[zone], nil
}

// applyToCmd adds new flags to the command's persistent flag-set.
func (c *cloudflare) applyToCmd(cmd *cobra.Command) {
	c.cmd = cmd
	flags := cmd.LocalFlags()
	flags.StringVarP(
		&c.auth,
		"cloudflare-auth",
		"",
		c.getAuth(),
		"CloudFlare authorization token/email:key",
	)
}

// email gets the email from the auth, if it can.
func (c *cloudflare) email() string {
	email, _ := c.basic()
	return email
}

// apiKey gets the apiKey from the auth, if it can.
func (c *cloudflare) apiKey() string {
	_, apiKey := c.basic()
	return apiKey
}

// token gets the token from the auth, if it can.
func (c *cloudflare) token() string {
	if email, apiKey := c.basic(); email != "" || apiKey != "" {
		return ""
	}
	return c.auth
}

// basic gets the email and apiKey from the auth, if it can.
func (c *cloudflare) basic() (string, string) {
	re := regexp.MustCompile(`^([^:]+):([^:]+)$`)
	matches := re.FindStringSubmatch(c.getAuth())
	if len(matches) != 3 {
		return "", ""
	}
	return matches[1], matches[2]
}

// getAuth gets the authorization from the struct or from the environment.
func (c *cloudflare) getAuth() string {
	if c.auth == "" {
		c.auth = env("DDNS_CLOUDFLARE_AUTH", "")
	}
	return c.auth
}

// httpClient gets a http.Client.
func (c *cloudflare) httpClient() *http.Client {
	if c.http == nil {
		c.http = &http.Client{}
	}
	return c.http
}

// ttl converts a time to live time.Duration to seconds, handling the special
// case of 0 being converted to 1 (which in CloudFlare means "automatic").
func (c *cloudflare) ttl(ttl time.Duration) int {
	seconds := int(ttl.Round(time.Second).Seconds())
	if seconds == 0 {
		return 1
	}
	return seconds
}
