# Contributing to Go Spaceship SDK

You're welcome to start a discussion about required features, file an issue or submit a work in progress (WIP) pull
request. Feel free to ask us for help. We'll do our best to guide you and help you to get on it.

## Prerequisites

- Go 1.25+ (the toolchain version is pinned in [`go.mod`](go.mod))
- [golangci-lint](https://golangci-lint.run/usage/install/#local-installation) — install the same version used in CI

## Development workflow

Run all checks before pushing:

```shell
make fmt    # auto-fix formatting (golangci-lint fmt)
make vet    # go vet
make lint   # golangci-lint (must be 0 issues)
make test   # unit tests (excludes live/acceptance tests)
make build  # compile all packages
```

Run `make help` to see every available target.

## Tests

This project has two test layers — keep them distinct (see [`CLAUDE.md`](CLAUDE.md) for the full strategy):

- **Unit tests** (`*_test.go`) — API serialization, pagination, error mapping, and record validation against
  `httptest` mock servers. Fast, no credentials. These are what `make test` runs.
- **Live tests** (`*_acc_test.go`, `TestAcc*` prefix) — exercise the real Spaceship API. They are credential-gated
  and excluded from `make test`.

Add or update tests for any code you change. If a mock server can prove it, write a unit test; if only the real API
can, write a live test.

### Running unit tests

```shell
make test                                # all unit tests
go test -run TestFunctionName ./client   # a single test
make test-cover                          # with a coverage report
```

### Running live tests

Live tests hit the real Spaceship API and require credentials. Copy `.env.example` to `.env` and fill it in (or
export the variables in your shell); `make testacc` loads `.env` automatically. Create credentials in the Spaceship
[API Manager](https://www.spaceship.com/application/api-manager/).

| Variable | Required | Purpose |
| --- | --- | --- |
| `SPACESHIP_API_KEY` | yes | API credential. |
| `SPACESHIP_API_SECRET` | yes | API credential. |
| `SPACESHIP_TEST_DOMAIN` | DNS tests only | Domain the DNS tests create and delete records on. Without it those tests skip. Use a throwaway domain. |
| `SPACESHIP_TEST_RECORD_PREFIX` | no | Namespaces created record names (default `goacc`). |
| `SPACESHIP_BASE_URL` | no | Override the API base URL (default `https://spaceship.dev/api/v1`). |

```shell
make testacc                                            # all live tests
go test -run TestAccGetDomainListPagination ./client -v # a single live test
```

> ⚠️ Live tests mutate real DNS records and domain settings on the account tied to your credentials. Only run them
> against an account you control.

## Commit conventions

- **Format** — use [Conventional Commits](https://www.conventionalcommits.org/) (`feat`, `fix`, `docs`, `chore`,
  `refactor`, `test`, `ci`, `perf`, `build`, `revert`). Keep commit bodies short — a line or two stating the why.
- **Dependency bumps must use a releasing type** — release-please only bumps the version on `fix:` / `feat:` /
  breaking. A `go.mod` bump ships in the next tagged SDK release (consumers `go get` a tag), so commit it as
  `fix(deps):`, **not** `chore(deps):` (which never releases, leaving the update unpublished until an unrelated
  `fix:`/`feat:` sweeps it in). CI/action-only bumps stay non-releasing (`ci(deps):`). Dependabot applies these
  prefixes automatically via `commit-message.prefix` in `.github/dependabot.yml`.
- **Sign-off** — this project requires a [Developer Certificate of Origin](https://developercertificate.org/) sign-off
  on every commit. Add a `Signed-off-by:` trailer with `git commit -s` (or by hand):

  ```
  Signed-off-by: Your Name <your-email@example.com>
  ```

  Pull requests with commits missing the sign-off will fail the DCO check in CI.
- **No other trailers** — do not add `Co-Authored-By` lines.

## Release

We publish a new tagged release once significant changes accumulate. If you need a release with a specific fix,
open an issue or contact us.
