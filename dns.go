package main

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

// updateDNS finds a provider which has a zone for the given domain record
// name, and attempts to create or updateDNS that record to have the given
// content, kind (i.e., type, e.g. A or AAAA) and TTL.
func updateDNS(name, kind, ip string, ttl time.Duration) error {
	var err error
	if ip == "" {
		ip, err = getIP()
		if err != nil {
			return err
		}
	}
	if kind == "" {
		kind = detectRecordType(ip)
	}
	for _, h := range dnsManagers {
		ok, err := h.ownsRecord(name)
		if err != nil {
			return err
		}
		if ok {
			if err := h.createOrUpdateRecord(name, kind, ip, ttl); err != nil {
				return err
			}
			return nil
		}
	}
	return fmt.Errorf("no records updated")
}

// A dnsManager has functions to applyToCmd, report whether it ownsRecord and
// createOrUpdateRecord.
type dnsManager interface {
	ownsRecord(string) (bool, error)
	createOrUpdateRecord(string, string, string, time.Duration) error
	applyToCmd(*cobra.Command)
}

// dnsManagers is a list of DNS managers.
var dnsManagers = []dnsManager{
	&cloudflare{auth: env("DDNS_CLOUDFLARE_AUTH", "")},
}
