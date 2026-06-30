# Go Spaceship SDK

[![Go Reference](https://pkg.go.dev/badge/github.com/namecheap/go-spaceship-sdk.svg)](https://pkg.go.dev/github.com/namecheap/go-spaceship-sdk)
[![CI](https://github.com/namecheap/go-spaceship-sdk/actions/workflows/ci.yml/badge.svg)](https://github.com/namecheap/go-spaceship-sdk/actions/workflows/ci.yml)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)

Go client for the [Spaceship](https://www.spaceship.com/) domain & DNS API.

- [Spaceship API Documentation](https://docs.spaceship.dev/)
- [API Manager (create credentials)](https://www.spaceship.com/application/api-manager/)

## Install

```sh
go get github.com/namecheap/go-spaceship-sdk
```

## Usage

```go
import "github.com/namecheap/go-spaceship-sdk/client"

c, err := client.NewClient("https://spaceship.dev/api/v1", apiKey, apiSecret)
if err != nil {
    log.Fatal(err)
}

// List domains
domains, err := c.GetDomainList(ctx)

// Add a DNS record
err = c.CreateDNSRecord(ctx, "example.com", client.DNSRecord{
    Type:    "A",
    Name:    "www",
    Address: "11.12.13.14",
    TTL:     3600,
})
```

Authentication uses the `X-API-Key` / `X-API-Secret` headers; credentials are created in the Spaceship
[API Manager](https://www.spaceship.com/application/api-manager/). The library takes credentials explicitly — it does
**not** read environment variables.

## API coverage

The SDK implements Domains, DNS records, and Personal Nameservers. DNS records are managed in the `custom` group only
(records owned by Spaceship products are read-only and left untouched); the create/upsert/delete methods do per-type
validation for A, AAAA, ALIAS, CAA, CNAME, HTTPS, MX, NS, PTR, SRV, SVCB, TLSA, and TXT.

The table below maps every published Spaceship API endpoint to its status and the backing SDK method(s). Verified
against the Spaceship API as of 2026-06-30 (SDK v0.1.0). See the
[`client` package reference](https://pkg.go.dev/github.com/namecheap/go-spaceship-sdk/client) for full method
signatures and godoc.

Legend: ✅ implemented · 🚧 not yet implemented · ⛔ not available in API (endpoint returns HTTP 501).

**Domains**

| Endpoint | Status | SDK method(s) |
|---|---|---|
| `GET /domains` | ✅ | `GetDomainList` |
| `GET /domains/{domain}` | ✅ | `GetDomainInfo` (falls back to the list endpoint on HTTP 429) |
| `POST /domains/{domain}` (register) | 🚧 | — |
| `POST /domains/{domain}/renew` | 🚧 | — |
| `POST /domains/{domain}/restore` | 🚧 | — |
| `DELETE /domains/{domain}` | ⛔ | — |
| `PUT /domains/{domain}/autorenew` | ✅ | `UpdateAutoRenew` |
| `PUT /domains/{domain}/nameservers` | ✅ | `UpdateDomainNameServers` |
| `PUT /domains/{domain}/contacts` | 🚧 | — |
| `PUT /domains/{domain}/privacy/preference` | 🚧 | — |
| `PUT /domains/{domain}/privacy/email-protection-preference` | 🚧 | — |

**Domain availability**

| Endpoint | Status | SDK method(s) |
|---|---|---|
| `POST /domains/available` (batch) | 🚧 | — |
| `GET /domains/{domain}/available` | 🚧 | — |

**Domain transfer**

| Endpoint | Status | SDK method(s) |
|---|---|---|
| `POST /domains/{domain}/transfer` | 🚧 | — |
| `GET /domains/{domain}/transfer` | 🚧 | — |
| `GET /domains/{domain}/transfer/auth-code` | 🚧 | — |
| `PUT /domains/{domain}/transfer/lock` | 🚧 | — |

**Personal nameservers**

| Endpoint | Status | SDK method(s) |
|---|---|---|
| `GET /domains/{domain}/personal-nameservers` | ✅ | `ListPersonalNameservers`, `FindPersonalNameserver` |
| `PUT /domains/{domain}/personal-nameservers/{currentHost}` | ✅ | `UpsertPersonalNameserver` |
| `DELETE /domains/{domain}/personal-nameservers/{currentHost}` | ✅ | `DeletePersonalNameserver` |
| `GET /domains/{domain}/personal-nameservers/{currentHost}` | ⛔ | — |

**DNS records**

| Endpoint | Status | SDK method(s) |
|---|---|---|
| `GET /dns/records/{domain}` | ✅ | `GetDNSRecords`, `FindDNSRecord` |
| `PUT /dns/records/{domain}` | ✅ | `UpsertDNSRecords`, `CreateDNSRecord` |
| `DELETE /dns/records/{domain}` | ✅ | `DeleteDNSRecord`, `DeleteDNSRecords`, `ClearDNSRecords` |

**Contacts**

| Endpoint | Status | SDK method(s) |
|---|---|---|
| `PUT /contacts` | 🚧 | — |
| `GET /contacts/{contact}` | 🚧 | — |
| `PUT /contacts/attributes` | 🚧 | — |
| `GET /contacts/attributes/{contact}` | 🚧 | — |

**Async operations**

| Endpoint | Status | SDK method(s) |
|---|---|---|
| `GET /async-operations/{operationId}` | 🚧 | — |

**SellerHub**

| Endpoint | Status | SDK method(s) |
|---|---|---|
| `GET /sellerhub/domains` | 🚧 | — |
| `POST /sellerhub/domains` | 🚧 | — |
| `GET /sellerhub/domains/{domain}` | 🚧 | — |
| `PATCH /sellerhub/domains/{domain}` | 🚧 | — |
| `DELETE /sellerhub/domains/{domain}` | 🚧 | — |
| `GET /sellerhub/domains/reports/sold` | 🚧 | — |
| `GET /sellerhub/domains/verification-records` | 🚧 | — |
| `POST /sellerhub/checkout-links` | 🚧 | — |

## Testing

Unit tests run with no credentials and are what CI executes:

```sh
make test
```

The SDK also ships **live acceptance tests** (`TestAcc*`) that exercise the real Spaceship API. They are skipped
unless credentials are set, and the DNS tests **create and delete records on a real domain**. Copy `.env.example`
to `.env` and fill it in — `make testacc` loads `.env` automatically:

```sh
cp .env.example .env   # then set SPACESHIP_API_KEY, SPACESHIP_API_SECRET, SPACESHIP_TEST_DOMAIN
make testacc
```

Exporting the same variables in your shell works too. `SPACESHIP_TEST_DOMAIN` is required only for the DNS tests;
without it they skip rather than fail.

> ⚠️ Acceptance tests mutate real DNS records on `SPACESHIP_TEST_DOMAIN`. Point them at a throwaway domain you
> control, never a production domain.

See [CONTRIBUTING.md](CONTRIBUTING.md) for the full test strategy.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for the development workflow, test layers, commit conventions, and DCO sign-off.

## Security

See [SECURITY.md](SECURITY.md) for how to report a vulnerability.

## License

Apache 2.0 — see [LICENSE](LICENSE).
