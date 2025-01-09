// Package main provides a simple example of how to use the libdns-gcore package.
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	gcore "git.mills.io/prologic/libdns-gcore"
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

	fmt.Println(records)
}
