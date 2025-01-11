// Package main provides a simple example of how to use the libdns-gcore package.
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	gcore "git.mills.io/prologic/libdns-gcore"
	"github.com/libdns/libdns"
)

func main() {
	apiKey := os.Getenv("GCORE_API_KEY")
	if apiKey == "" {
		fmt.Printf("GCORE_API_KEY not set\n")
		return
	}

	if len(os.Args) < 2 {
		fmt.Printf("Usage: %s <zone>\n", filepath.Base(os.Args[0]))
		os.Exit(1)
	}

	zone := os.Args[1]

	provider := &gcore.Provider{
		APIKey: apiKey,
	}

	records, err := provider.GetRecords(context.Background(), zone)
	if err != nil {
		fmt.Printf("Error: %s", err.Error())
		return
	}
	for _, record := range records {
		fmt.Printf("%s %s %s\n", record.Name, record.Type, record.Value)
	}

	records, err = provider.AppendRecords(context.Background(), zone, []libdns.Record{
		{
			Name:  "0xffs.xyz",
			Type:  "TXT",
			TTL:   300,
			Value: "Hello World!",
		},
	})
	if err != nil {
		fmt.Printf("Error: %s", err.Error())
		return
	}

	time.Sleep(1)

	records, err = provider.GetRecords(context.Background(), zone)
	if err != nil {
		fmt.Printf("Error: %s", err.Error())
		return
	}
	for _, record := range records {
		fmt.Printf("%s %s %s\n", record.Name, record.Type, record.Value)
	}

	time.Sleep(1)

	records, err = provider.DeleteRecords(context.Background(), zone, []libdns.Record{
		{
			Name:  "0xffs.xyz",
			Type:  "TXT",
			Value: "Hello World!",
		},
	})
	if err != nil {
		fmt.Printf("Error: %s", err.Error())
		return
	}

	time.Sleep(1)

	records, err = provider.GetRecords(context.Background(), zone)
	if err != nil {
		fmt.Printf("Error: %s", err.Error())
		return
	}
	for _, record := range records {
		fmt.Printf("%s %s %s\n", record.Name, record.Type, record.Value)
	}
}
