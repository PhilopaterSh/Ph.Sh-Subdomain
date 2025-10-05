package main

import (
	"os"
)

var apiKeys = map[string]string{
	"urlscan":     os.Getenv("URLSCAN_API_KEY"),
	"dnsdumpster": os.Getenv("DNSDUMPSTER_API_KEY"),
	"vt":          os.Getenv("VT_API_KEY"),
	"securitytrails": os.Getenv("SECURITYTRAILS_API_KEY"),
	"shodan":         os.Getenv("SHODAN_API_KEY"),
}
