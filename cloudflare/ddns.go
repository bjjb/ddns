package cloudflare

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

// DDNS figures out the record type (A, AAAA or CNAME), then fetches the
// zones, finds the zone which is a suffix for the record name, and then
// creates or updates a record matching the name and type.
var DDNS = func(f ...func(*http.Request) *http.Request) func(string, string, time.Duration) error {
	return func(name, content string, ttl time.Duration) error {
		recordType := recordType(content)
		client := New(f...)
		zones, err := client.GetZones()
		if err != nil {
			return err
		}
		zone, err := func() (*Zone, error) {
			for _, zone := range zones {
				if strings.HasSuffix(name, zone.Name) {
					return zone, nil
				}
			}
			return nil, fmt.Errorf("no suitable cloudflare zone found for %q", name)
		}()
		if err != nil {
			return err
		}
		records, err := client.GetDNSRecords(zone.ID)
		if err != nil {
			return err
		}
		record := func() *Record {
			for _, record := range records {
				if record.Name == name && record.Type == recordType {
					return record
				}
			}
			return &Record{
				Name:   name,
				Type:   recordType,
				ZoneID: zone.ID,
			}
		}()
		if record != nil {
			record.Content = content
			record.TTL = ttl
		}
		if record.ID != "" {
			return client.PutDNSRecord(record)
		}
		_, err = client.PostDNSRecord(record)
		return err
	}
}
