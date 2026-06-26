package client_test

import (
	"context"
	"fmt"
	"log"

	"github.com/namecheap/go-spaceship-sdk/client"
)

// Example shows the typical lifecycle: construct a client with explicit
// credentials, then list the account's domains.
//
// Credentials are created in the Spaceship API Manager
// (https://www.spaceship.com/application/api-manager/). The library never reads
// them from the environment, so pass them in directly.
func Example() {
	c, err := client.NewClient("https://spaceship.dev/api/v1", "your-api-key", "your-api-secret")
	if err != nil {
		log.Fatal(err)
	}

	domains, err := c.GetDomainList(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	for _, d := range domains.Items {
		fmt.Printf("%s (auto-renew: %t)\n", d.Name, d.AutoRenew)
	}
}

// ExampleNewClient validates the base URL and returns a client ready to use.
// An invalid base URL is reported as an error rather than panicking later.
func ExampleNewClient() {
	c, err := client.NewClient("https://spaceship.dev/api/v1", "your-api-key", "your-api-secret")
	if err != nil {
		log.Fatal(err)
	}
	_ = c
}

// ExampleClient_CreateDNSRecord adds a single A record. Only the fields
// relevant to the record's Type need to be set; the server assigns new records
// to the "custom" group automatically.
func ExampleClient_CreateDNSRecord() {
	c, err := client.NewClient("https://spaceship.dev/api/v1", "your-api-key", "your-api-secret")
	if err != nil {
		log.Fatal(err)
	}

	err = c.CreateDNSRecord(context.Background(), "example.com", client.DNSRecord{
		Type:    "A",
		Name:    "www",
		Address: "11.12.13.14",
		TTL:     3600,
	})
	if err != nil {
		log.Fatal(err)
	}
}

// ExampleClient_GetDNSRecords lists the DNS records for a domain. Only records
// in the "custom" group are returned; records owned by Spaceship products are
// filtered out and left untouched.
func ExampleClient_GetDNSRecords() {
	c, err := client.NewClient("https://spaceship.dev/api/v1", "your-api-key", "your-api-secret")
	if err != nil {
		log.Fatal(err)
	}

	records, err := c.GetDNSRecords(context.Background(), "example.com")
	if err != nil {
		log.Fatal(err)
	}

	for _, r := range records {
		fmt.Printf("%s %s\n", r.Type, r.Name)
	}
}

// ExampleClient_UpsertDNSRecords creates or updates a batch of records in one
// call. Pass force=true to overwrite conflicting records that already exist.
func ExampleClient_UpsertDNSRecords() {
	c, err := client.NewClient("https://spaceship.dev/api/v1", "your-api-key", "your-api-secret")
	if err != nil {
		log.Fatal(err)
	}

	records := []client.DNSRecord{
		{Type: "A", Name: "@", Address: "11.12.13.14", TTL: 3600},
		{Type: "TXT", Name: "@", Value: "v=spf1 -all", TTL: 3600},
	}

	const force = true
	if err := c.UpsertDNSRecords(context.Background(), "example.com", force, records); err != nil {
		log.Fatal(err)
	}
}

// ExampleClient_GetDomainInfo fetches details for a single domain. Under heavy
// load the per-domain endpoint may return HTTP 429; this method transparently
// falls back to the higher-limit domain list endpoint, so callers rarely need
// to handle rate limiting themselves.
func ExampleClient_GetDomainInfo() {
	c, err := client.NewClient("https://spaceship.dev/api/v1", "your-api-key", "your-api-secret")
	if err != nil {
		log.Fatal(err)
	}

	info, err := c.GetDomainInfo(context.Background(), "example.com")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s expires %s\n", info.Name, info.ExpirationDate)
}

// ExampleClient_UpdateDomainNameServers points a domain at custom nameservers.
// Use BasicNameserverProvider with an empty Hosts slice to revert to
// Spaceship's defaults.
func ExampleClient_UpdateDomainNameServers() {
	c, err := client.NewClient("https://spaceship.dev/api/v1", "your-api-key", "your-api-secret")
	if err != nil {
		log.Fatal(err)
	}

	err = c.UpdateDomainNameServers(context.Background(), "example.com", client.UpdateNameserverRequest{
		Provider: client.CustomNameserverProvider,
		Hosts:    []string{"ns1.example.net", "ns2.example.net"},
	})
	if err != nil {
		log.Fatal(err)
	}
}

// ExampleClient_FindDNSRecord locates a record by type, name, and value
// signature, returning a not-found error when no match exists. Matching is
// case-insensitive for every field except TXT values, which are compared
// case-sensitively to mirror the API.
func ExampleClient_FindDNSRecord() {
	c, err := client.NewClient("https://spaceship.dev/api/v1", "your-api-key", "your-api-secret")
	if err != nil {
		log.Fatal(err)
	}

	record, err := c.FindDNSRecord(context.Background(), "example.com", "A", "www", "11.12.13.14")
	if err != nil {
		// IsNotFoundError distinguishes "no such record" from real failures.
		if client.IsNotFoundError(err) {
			fmt.Println("no matching record")
			return
		}
		log.Fatal(err)
	}

	fmt.Printf("found %s record for %s\n", record.Type, record.Name)
}
