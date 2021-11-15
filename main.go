// Package main contains a command-line application which either starts a DDNS
// server or makes a DDNS request.
package main

import (
	"bytes"
	"io"
	"net/http"
	"os"

	"github.com/spf13/cobra"
)

var name = "ddns"
var version = "0.0.1"

var env = map[string]string{
	"IP_PROVIDER": "https://icanhazip.com",
}

// main calls parse with configuration from the OS.
func main() {
	command(
		os.Args[1:],
		os.Stdout,
		os.Stderr,
		envFunc(os.Getenv, getenv(env)),
		os.Exit,
	).Execute()
}

// command returns a new command for parsing command-line arguments.
var command = func(
	args []string,
	out, err io.Writer,
	env func(string) string,
	exit func(int),
) *cobra.Command {
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
10.0.0.1

$ ddns -6
2001::dead:beef

$ ddns -d &
ddns v0.1.0 listening on :5380

$ curl -Ls localhost
127.0.0.1

$ curl -Ls6 localhost
::1
`
	validArgsFunction := func(
		cmd *cobra.Command,
		args []string,
		toComplete string,
	) ([]string, cobra.ShellCompDirective) {
		return []string{}, cobra.ShellCompDirectiveDefault
	}
	ipProvider := env("IP_PROVIDER")
	run := func(c *cobra.Command, args []string) {
		resp, err := http.Get(ipProvider)
		if err != nil {
			c.PrintErr(err)
			exit(1)
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			c.PrintErr(resp.Status)
			exit(resp.StatusCode)
			return
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			c.PrintErr(err)
			exit(2)
			return
		}
		c.Println(string(bytes.TrimSpace(body)))
	}
	cmd := &cobra.Command{
		Use:               use,
		Version:           version,
		Short:             short,
		Long:              long,
		Example:           example,
		ValidArgsFunction: validArgsFunction,
		Run:               run,
	}
	cmd.SetOut(out)
	cmd.SetErr(err)
	cmd.PersistentFlags().StringVarP(&ipProvider, "ip", "I", ipProvider, "IP lookup service URL")
	cmd.SetArgs(args)
	return cmd
}

func envFunc(f ...func(string) string) func(string) string {
	return func(k string) string {
		for _, f := range f {
			if v := f(k); v != "" {
				return v
			}
		}
		return ""
	}
}

func getenv(map[string]string) func(string) string {
	return func(k string) string {
		v, _ := env[k]
		return v
	}
}
