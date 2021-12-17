// Package main contains a command-line application which either starts a DDNS
// server or makes a DDNS request.
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"regexp"

	"github.com/spf13/cobra"
)

var name = "ddns"
var version = "0.0.1"

var defaults = map[string]string{
	"IP_PROVIDER":    "https://icanhazip.com",
	"CLOUDFLARE_API": "https://api.cloudflare.com/client/v4",
	"ADDR":           ":8053",
}

func cfg(key string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaults[key]
}

// main builds a command and calls its Execute function.
func main() {
	if err := cmd(os.Args[1:], os.Stdout, os.Stderr).Execute(); err != nil {
		log.Fatal(err)
	}
}

type executor interface{ Execute() error }

type runE func(*cobra.Command, []string) error

var exit = os.Exit

// cmd returns a new cmd for parsing cmd-line arguments.
var cmd = func(
	args []string,
	out, err io.Writer,
) executor {
	use := name
	short := "Dynamic DNS client/server"
	long := `Each argument is treated as a name, which is used as a DNS record
value. The program attempts to create or update a record with each name. If no
type is specified as an option, then it will choose A, AAAA or CNAME,
depending on the value. If no value is given, then it attempts to determine
its remote IP address by calling a provider. If no TTL is specified, it will
use 1h for new records, or leave existing record untouched.

If no arguments are given, it simply prints out its remote address (if it can
be determined).`
	example := `$ ddns
10.1.2.3
2001::dead:beef

$ ddns -4
10.1.2.3

$ ddns -6
2001::dead:beef

$ printf 'HTTP/1.1 200 OK\nContent-Type: text/plain\nContent-Length: 7\n\n1.2.3.4' | nc -l 8000 &
$ ddns -I http://localhost:8000
1.2.3.4
`
	ipp := firstString(os.Getenv("IP_PROVIDER"), "https://icanhazip.com")
	cmd := &cobra.Command{
		Use:     use,
		Version: version,
		Short:   short,
		Long:    long,
		Args:    cobra.NoArgs,
		Example: example,
		Run:     runify(runIP(), exit),
	}
	cmd.SetOut(out)
	cmd.SetErr(err)
	cmd.PersistentFlags().StringVarP(
		&ipp,
		"ip-provider",
		"I",
		ipp,
		"IP lookup service URL",
	)
	cmd.SetArgs(args)
	cmd.AddCommand(commands()...)
	return cmd
}

func runify(r runE, exit func(int)) func(*cobra.Command, []string) {
	return func(c *cobra.Command, args []string) {
		if err := r(c, args); err != nil {
			c.PrintErr(err)
			exit(1)
		}
	}
}

func runIP() runE {
	return func(c *cobra.Command, _ []string) error {
		service, err := c.Root().PersistentFlags().GetString("ip-provider")
		if err != nil {
			return err
		}
		ips, err := publicIPs(service)
		if err != nil {
			return err
		}
		for _, ip := range ips {
			c.Println(ip)
		}
		return nil
	}
}

func clients() []*http.Client {
	// TODO return an IPv4 and an IPv6 client here.
	return []*http.Client{&http.Client{}}
}

func publicIPs(service string) ([]string, error) {
	ips := []string{}
	for _, c := range clients() {
		r, err := http.NewRequest(http.MethodGet, service, nil)
		if err != nil {
			return nil, err
		}
		resp, err := c.Do(r)
		if err != nil {
			log.Print(err)
			err = nil
			continue
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("%d %s", resp.StatusCode, resp.Status)
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		ips = append(ips, string(bytes.TrimSpace(body)))
	}
	return ips, nil
}

func commands() []*cobra.Command {
	cmds := []*cobra.Command{}
	cmds = append(cmds, cmdCloudflare())
	return cmds
}

func cmdCloudflare() *cobra.Command {
	t := os.Getenv("CLOUDFLARE_TOKEN")
	u := os.Getenv("CLOUDFLARE_USERNAME")
	p := os.Getenv("CLOUDFLARE_PASSWORD")
	f := func(r *http.Request) {
		if t != "" {
			r.Header.Add("Authorization", fmt.Sprintf("Bearer %s", t))
		} else if u != "" && p != "" {
			r.SetBasicAuth(u, p)
		}
	}
	cmd := &cobra.Command{
		Use:     "cloudflare",
		Aliases: []string{"cf"},
		Short:   "set a CloudFlare DNS record",
		Long: `
Creates or updates a new DNS record on CloudFlare.`,
		RunE: runCloudflare(f),
	}
	cmd.PersistentFlags().StringVarP(&t, "token", "t", t, "cloudflare token")
	cmd.PersistentFlags().StringVarP(&u, "user", "u", u, "cloudflare email")
	cmd.PersistentFlags().StringVarP(&p, "pass", "p", p, "cloudflare token")
	return cmd
}

func runCloudflare(f ...func(*http.Request)) runE {
	return func(c *cobra.Command, args []string) error {
		ipService, err := c.Root().PersistentFlags().GetString("ip-provider")
		if err != nil {
			return err
		}
		ips, err := publicIPs(ipService)
		if err != nil {
			return err
		}
		for _, name := range args {
			if err := updateCloudflare(name, ips, f...); err != nil {
				return err
			}
		}
		return nil
	}
}

func updateCloudflare(name string, ips []string, f ...func(r *http.Request)) error {
	api := firstString(os.Getenv("CLOUDFLARE_API"), defaults["CLOUDFLARE_API"])
	u, err := url.Parse(fmt.Sprintf("%s/zones", api))
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return err
	}
	for _, f := range f {
		f(req)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%d %s", resp.StatusCode, resp.Status)
	}
	defer resp.Body.Close()
	response := &struct {
		Result []*struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"result"`
		Success bool `json:"success"`
		Errors  []*struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"errors"`
		Messages   []string `json:"messages"`
		ResultInfo *struct {
			Page       int `json:"page"`
			PerPage    int `json:"per_page"`
			Count      int `json:"count"`
			TotalCount int `json:"total_count"`
		} `json:"result_info"`
	}{}
	if err := json.NewDecoder(resp.Body).Decode(response); err != nil {
		return err
	}
	for _, z := range response.Result {
		rex, err := regexp.Compile(z.Name + "$")
		if err != nil {
			return err
		}
		if rex.MatchString(name) {
			return updateCloudflareRecords(z.ID, name, ips, f...)
		}
	}
	return fmt.Errorf("zone not found for %s", name)
}

func updateCloudflareRecords(z, name string, ips []string, f ...func(*http.Request)) error {
	for _, ip := range ips {
		if err := updateCloudflareRecord(z, name, ip, f...); err != nil {
			return err
		}
	}
	return nil
}

func updateCloudflareRecord(z, name, ip string, f ...func(*http.Request)) error {
	api := firstString(os.Getenv("CLOUDFLARE_API"), defaults["CLOUDFLARE_API"])
	t := typeForIP(ip)
	s := fmt.Sprintf("%s/zones/%s/dns_records?type=%s&name=%s", api, z, t, name)
	u, err := url.Parse(s)
	if err != nil {
		return err
	}
	r, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return err
	}
	for _, f := range f {
		f(r)
	}
	c := &http.Client{}
	resp, err := c.Do(r)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%d %s", resp.StatusCode, resp.Status)
	}
	defer resp.Body.Close()
	response := &struct {
		Result []*struct {
			ID        string `json:"id"`
			Name      string `json:"name"`
			Type      string `json:"type"`
			Content   string `json:"content"`
			TTL       int    `json:"ttl"`
			ZoneID    string `json:"zone_id"`
			ZoneName  string `json:"zone_name"`
			Proxiable bool   `json:"proxiable"`
			Proxied   bool   `json:"proxied"`
		} `json:"result"`
		Success bool `json:"success"`
		Errors  []*struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"errors"`
		Messages   []string `json:"messages"`
		ResultInfo *struct {
			Page       int `json:"page"`
			PerPage    int `json:"per_page"`
			Count      int `json:"count"`
			TotalCount int `json:"total_count"`
		} `json:"result_info"`
	}{}
	if err := json.NewDecoder(resp.Body).Decode(response); err != nil {
		return err
	}
	if len(response.Result) == 0 {
		return postCloudflareRecord(z, name, ip, t, f...)
	}
	for _, r := range response.Result {
		if err := patchCloudflareRecord(r.ZoneID, r.ID, ip, f...); err != nil {
			return err
		}
	}
	return nil
}

func postCloudflareRecord(z, name, ip, t string, f ...func(*http.Request)) error {
	api := firstString(os.Getenv("CLOUDFLARE_API"), defaults["CLOUDFLARE_API"])
	s := fmt.Sprintf("%s/zones/%s/dns_records", api, z)
	u, err := url.Parse(s)
	if err != nil {
		return err
	}
	payload := &struct {
		Name    string `json:"name"`
		Type    string `json:"type"`
		Content string `json:"content"`
		TTL     int    `json:"ttl"`
		Proxied bool   `json:"proxied"`
	}{name, t, ip, 300, false}
	b := new(bytes.Buffer)
	if err := json.NewEncoder(b).Encode(payload); err != nil {
		return err
	}
	r, err := http.NewRequest(http.MethodPost, u.String(), b)
	if err != nil {
		return err
	}
	for _, f := range f {
		f(r)
	}
	c := &http.Client{}
	resp, err := c.Do(r)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%d %s", resp.StatusCode, resp.Status)
	}
	return nil
}

func patchCloudflareRecord(z, r, ip string, f ...func(*http.Request)) error {
	api := firstString(os.Getenv("CLOUDFLARE_API"), defaults["CLOUDFLARE_API"])
	s := fmt.Sprintf("%s/zones/%s/dns_records/%s", api, z, r)
	u, err := url.Parse(s)
	if err != nil {
		return err
	}
	payload := &struct {
		Content string `json:"content"`
	}{ip}
	b := new(bytes.Buffer)
	if err := json.NewEncoder(b).Encode(payload); err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPatch, u.String(), b)
	if err != nil {
		return err
	}
	for _, f := range f {
		f(req)
	}
	c := &http.Client{}
	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%d %s", resp.StatusCode, resp.Status)
	}
	return nil
}

func typeForIP(ip string) string {
	a := net.ParseIP(ip)
	if a == nil {
		return "CNAME"
	}
	if a.To4() == nil {
		return "AAAA"
	}
	return "A"
}

func firstString(s ...string) string {
	for _, s := range s {
		if s != "" {
			return s
		}
	}
	return ""
}
