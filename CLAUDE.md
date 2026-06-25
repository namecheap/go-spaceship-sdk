Go client library (SDK) for the [Spaceship](https://www.spaceship.com/) domain & DNS API. API reference: https://docs.spaceship.dev/.

## Commands

```bash
make build                                   # Compile all packages (go build ./...)
make test                                    # Unit tests (excludes live/acceptance tests)
go test -run TestFunctionName ./client       # Single unit test
make lint                                    # Lint (linters + formatter check)
make fmt                                      # Auto-fix formatting
make vet                                      # go vet

# Live tests — only run when explicitly asked. These hit the real Spaceship API.
# Require SPACESHIP_API_KEY and SPACESHIP_API_SECRET (loaded from .env if present).
make testacc
go test -run TestAccGetDomainListPagination ./client -v   # Single live test
```

## Verification workflow

After making changes, follow this order:

1. **Unit tests** — `make test` (or `go test -run TestName ./client` for a specific test)
2. **Lint** — `make lint` (run `make fmt` to auto-fix formatting)
3. **Build** — `make build`
4. **Live tests** — run only when the user explicitly asks. They hit the real API.

## Architecture

A standalone Go client for protocol-level access to the Spaceship API. No Terraform or other framework dependencies — the only third-party dependency is `github.com/dlclark/regexp2`.

- `client/` — the HTTP client (`NewClient`), domains, DNS records, personal nameservers, request/error plumbing, and record matching.
- `client/records/` — one file per DNS record type (A, AAAA, ALIAS, CAA, CNAME, HTTPS, MX, NS, PTR, SRV, SVCB, TLSA, TXT), each with a struct and `Validate*` methods. Shared helpers live in `common.go`.

`NewClient(baseURL, apiKey, apiSecret)` takes credentials explicitly — the library does **not** read environment variables (only the live tests do). The production base URL is `https://spaceship.dev/api/v1`; auth is via `X-API-Key` / `X-API-Secret` headers.

## Testing strategy

Two layers — do not duplicate coverage between them.

- **Unit tests** (`*_test.go`) — API calls, HTTP serialization, pagination, error mapping, and record validation, using `httptest` mock servers. Fast, no credentials.
- **Live tests** (`*_acc_test.go`, `TestAcc*` prefix) — exercise the real Spaceship API. They are credential-gated (`t.Skip()` unless `SPACESHIP_API_KEY` / `SPACESHIP_API_SECRET` are set) and excluded from `make test` by the `TestAcc` prefix. These are the authoritative "does it really work" tests.

When adding a test, ask: "Is this testing something this layer alone can?" If a mock server proves it, it is a unit test; if only the real API can, it is a live test.

## Key design rules

- **DNS records are scoped to the custom group**: the API returns records across three groups — `custom`, `product`, `personalNS`. This client manages only `custom`-group records (see `filterCustomDNSRecords`). Do not touch the other groups.
- **DNS record matching mirrors the API**: records are matched by type + name + data using case-insensitive comparison, **except** TXT values, which are case-sensitive. `RecordKey` / `MatchDNSRecord` must follow these rules.
- **Rate-limit fallback**: `GetDomainInfo` falls back to the domain list endpoint on HTTP 429 (the list endpoint has far higher limits and returns equivalent data). Preserve this pattern.
- **Domain list pagination**: `GetDomainList` pages with `take`/`skip`; the API caps `take` at 100. `GetDomainInfo`'s 429 fallback depends on this, so keep the list complete.
- **Validators mirror the API, with verified exceptions noted in comments**: per-record `Validate*` methods enforce documented API constraints only. Where the live API diverges from the published spec — e.g. SVCB/HTTPS require a scheme when a port is set; TXT rejects whitespace-only values and caps length in **bytes**; ALIAS/CNAME/MX/SRV target hostname fields reject `@` and `*` — the real behavior is documented in a comment next to the check. Do not invent stricter rules; edge cases belong in live tests.

## Adding a DNS record type validator

Copy an existing type (A, AAAA, ALIAS, CAA, SRV) rather than inventing a shape:

1. **Struct + `Validate*` methods** in `client/records/<type>.go`, with unit tests in `<type>_test.go`. Mirror API-documented constraints only. Use the shared helpers `ValidateName` / `ValidateTTL` from `common.go`.
2. If the type is new to the wire format, wire its fields into `client/dns.go` (`DNSRecord`) and its data signature into `client/dns_match.go` (`RecordValueSignature`).

Live tests own API edge cases: unusual values, update/delete+upsert paths, case-sensitivity round-trips, same-host multiplicity.

## Git

Follow the commit conventions in [CONTRIBUTING.md](CONTRIBUTING.md#commit-conventions): Conventional Commits, a DCO `Signed-off-by:` trailer on every commit (`git commit -s`), and no `Co-Authored-By` lines.
