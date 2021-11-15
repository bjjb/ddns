package main

import (
	"time"
)

type ddnsFunc func(string, string, time.Duration) error

var providers = map[string]ddnsFunc{}
