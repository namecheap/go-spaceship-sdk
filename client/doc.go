// Package client is a Go SDK for the Spaceship domain & DNS API.
//
// It provides protocol-level access to the Spaceship API
// (https://docs.spaceship.dev/) for managing domains, DNS records, and
// personal nameservers. The client speaks the HTTP API directly and has no
// Terraform or other framework dependencies.
//
// # Authentication
//
// Credentials are created in the Spaceship API Manager
// (https://www.spaceship.com/application/api-manager/) and sent on every
// request via the X-API-Key and X-API-Secret headers. They are passed
// explicitly to NewClient; the library never reads them from the environment.
//
//	c, err := client.NewClient("https://spaceship.dev/api/v1", apiKey, apiSecret)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	domains, err := c.GetDomainList(ctx)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	err = c.CreateDNSRecord(ctx, "example.com", client.DNSRecord{
//		Type:    "A",
//		Name:    "www",
//		Address: "11.12.13.14",
//		TTL:     3600,
//	})
//
// # What it covers
//
//   - Domains — list (GetDomainList), get info (GetDomainInfo), toggle
//     auto-renew (UpdateAutoRenew), and update nameservers
//     (UpdateDomainNameServers).
//   - DNS records — list, find, create, upsert, and delete, plus a bulk clear.
//   - Personal nameservers — list, find, upsert, and delete.
//
// # DNS record groups
//
// The Spaceship API returns DNS records across three groups: custom, product,
// and personalNS. This client manages only the custom group — records owned by
// Spaceship products are read-only and are left untouched. Every DNS read is
// filtered to the custom group, and writes operate on it exclusively.
//
// Records are matched by type, name, and data using case-insensitive
// comparison, with one exception: TXT record values are compared
// case-sensitively, mirroring the API.
//
// # Rate limiting
//
// GetDomainInfo falls back to the domain list endpoint when the per-domain
// endpoint returns HTTP 429. The list endpoint has far higher rate limits and
// returns equivalent data, so callers transparently get domain info even under
// throttling.
//
// # Record types and validation
//
// The records subpackage defines one struct per supported DNS record type — A,
// AAAA, ALIAS, CAA, CNAME, HTTPS, MX, NS, PTR, SRV, SVCB, TLSA, and TXT — each
// with Validate methods enforcing the constraints documented by the API.
package client
