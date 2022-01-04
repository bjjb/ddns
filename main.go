// Package main contains a command-line application which either starts a DDNS
// server or makes a DDNS request.
package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"mime"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"

	_ "embed"

	"github.com/spf13/cobra"
)

const (
	// errnoFailed is returned when something goes wrong.
	errnoFailed = 2
)

// name is the name of the programme.
var name = "ddns"

// version is the version of the programme.
//go:embed public/version.txt
var version string

// summary is a short description of the programme.
var summary = "dynamic domain name servicer"

// description is a longer description of the programme, used in help.
//go:embed public/description.txt
var description string

// exit exits the programme.
var exit = os.Exit

// getenv obtains some string by a key.
var getenv = os.Getenv

// stdin is the standard input stream.
var stdin io.Reader = os.Stdin

// stdout is the standard output stream.
var stdout io.Writer = os.Stdout

// stderr is the standard error stream.
var stderr io.Writer = os.Stderr

// args are the command-line arguments (excluding the programme name).
var args = os.Args[1:]

// dsn is the data-source name of the database
var dsn string = "file::memory:?cache=shared"

// ipServiceURL is the default URL of an IP Echoing service
var ipServiceURL = env("DDNS_IP_ECHO_URL", "https://icanhazip.com")

// kind is the record type to use. If blank, it'll be detected for new
// records.
var kind = ""

// ttl is the record TTL to use
var ttl = 5 * time.Minute

// main sets up the root command and Executes it.
func main() {
	cmd := &cobra.Command{
		Use:     name,
		Version: version,
		Short:   summary,
		Long:    description,
		Args:    cobra.ExactArgs(1),
		Run:     run,
	}
	cmd.SetIn(stdin)
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)
	cmd.SetArgs(args)
	cmd.AddCommand(ipCmd(), serverCmd())
	flags := cmd.LocalFlags()
	flags.DurationVarP(&ttl, "ttl", "t", ttl, "the record TTL (in seconds)")
	flags.StringVarP(&kind, "type", "k", kind, "the record type")
	flags.StringVarP(&ipServiceURL, "ip-service", "I", ipServiceURL, "IP echo service URL")
	flags.StringVarP(&dsn, "dsn", "D", dsn, "database name")
	for _, h := range dnsManagers {
		h.applyToCmd(cmd)
	}
	cmd.Execute()
}

// ipCmd builds a command which prints out the current public IP address by
// using an IP address echoing service.
var ipCmd = func() *cobra.Command {
	return &cobra.Command{
		Use:     "ip",
		Aliases: []string{"address", "addr"},
		Args:    cobra.NoArgs,
		Short:   "prints the current public IP address",
		Long: `
Calls a remote service to get the public IP address, which it then prints. The
service needs to return the address as plain text.`,
		Run: func(c *cobra.Command, args []string) {
			result, err := getIP()
			if err != nil {
				c.PrintErr(err)
			}
			c.Printf("%s", result)
		},
	}
}

// serverCmd builds a command which starts a server.
var serverCmd = func() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "server",
		Aliases: []string{"daemon", "start"},
		Short:   "starts a server",
		Long: `
Starts a HTTP server which includes an API and a small web-app to allow users
to manage and configure DDNS entries.
`,
		Args: cobra.NoArgs,
		Run: func(c *cobra.Command, args []string) {
			startServer()
		},
	}
	return cmd
}

// getIP returns the caller's IP address.
var getIP = func() (string, error) {
	_, err := url.Parse(ipServiceURL)
	if err != nil {
		return "", err
	}
	r, err := http.Get(ipServiceURL)
	if err != nil {
		return "", err
	}
	defer r.Body.Close()
	if r.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%d from %s - %s", r.StatusCode, ipServiceURL, r.Status)
	}
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil {
		return "", err
	}
	switch mediaType {
	case "text/plain":
		data, err := io.ReadAll(r.Body)
		if err != nil {
			return "", err
		}
		ip := string(bytes.TrimSpace(data))
		if net.ParseIP(ip) == nil {
			return "", fmt.Errorf(`failed to parse IP address "%s"`, data)
		}
		return ip, nil
	default:
		return "", fmt.Errorf(`content-type "%s" not supported`, mediaType)
	}
}

// startServer starts a server.
var startServer = func() {
	log.Printf("starting a server")
}

// run is the function run by the default command.
var run = func(c *cobra.Command, args []string) {
	ip, err := getIP()
	if err != nil {
		c.PrintErr(err)
		exit(errnoFailed)
		return
	}
	for _, name := range args {
		if err := updateDNS(name, kind, ip, ttl); err != nil {
			c.PrintErr(err)
			exit(errnoFailed)
			return
		}
	}
}
