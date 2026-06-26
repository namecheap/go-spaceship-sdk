# Release Process

This document describes how changes merged to `main` become published versions
of the SDK, installable via `go get github.com/namecheap/go-spaceship-sdk@vX.Y.Z`.

This is a Go library: a release is a SemVer git tag plus a GitHub release and a
changelog entry. There are no binaries to build, sign, or upload — the Go module
proxy serves the tag directly.

## Overview

The project uses a **semi-automated, maintainer-gated** release flow. Changes
merged to `main` do **not** ship immediately. They accumulate in a long-lived
"release PR" maintained by [release-please](https://github.com/googleapis/release-please),
and a version is tagged only when a maintainer merges that PR.

Two workflows participate:

| Workflow | Trigger | Role |
|---|---|---|
| [`ci.yml`](.github/workflows/ci.yml) | push, PRs | Build, vet, lint, unit tests, gated acceptance tests, security scan |
| [`versioning.yml`](.github/workflows/versioning.yml) | `CI` success on `main`, manual dispatch | Runs release-please to open/update the release PR and cut the tag |

Supporting configuration:

- [`pr-title.yml`](.github/workflows/pr-title.yml) — enforces
  [Conventional Commits](https://www.conventionalcommits.org/) on PR titles;
  release-please consumes these to compute version bumps.
- [`.release-please-config.json`](.release-please-config.json) /
  [`.release-please-manifest.json`](.release-please-manifest.json) —
  release-please configuration and state. The manifest is the source of truth
  for the current version.

## Step-by-step flow

### 1. PR is opened and merged

- PR title must follow Conventional Commits; `pr-title.yml` fails the PR otherwise.
- PRs are squash-merged, so the PR title becomes the commit message that
  release-please classifies.
- `ci.yml` runs on the PR and again on the resulting merge commit.

At this point nothing is released — the commit simply sits on `main`.

### 2. release-please opens or updates the release PR

- `versioning.yml` triggers on a successful `CI` run against `main`
  (`workflow_run` with `conclusion == 'success'`).
- release-please walks commits since the last tag, classifies them, and:
  - computes the next SemVer bump (`fix:` → patch, `feat:` → minor; in 0.x a
    breaking change stays within 0.x rather than jumping to 1.0.0),
  - updates `CHANGELOG.md` with a generated entry grouped by type,
  - updates `.release-please-manifest.json` with the new version,
  - opens (or updates) a PR titled `chore(main): release X.Y.Z`.
- Authentication uses a dedicated GitHub App (`SPS_RELEASE_CLIENT_ID`,
  `SPS_RELEASE_PRIVATE_KEY`), **not** the default `GITHUB_TOKEN`: events authored
  by `GITHUB_TOKEN` do not re-trigger workflows.

The release PR is long-lived — as more PRs merge to `main`, release-please keeps
appending to the same PR.

### 3. Maintainer cuts the release

A release happens only when a maintainer reviews the computed version bump and
changelog in the release PR and merges it. Merging the release PR:

- commits the version bump and regenerated `CHANGELOG.md` to `main`,
- creates the `vX.Y.Z` git tag and the GitHub release.

There is no fixed cadence — merge when there is enough change to justify a new
version.

### 4. Availability

Once the tag is pushed, the version is immediately installable:

```
go get github.com/namecheap/go-spaceship-sdk@vX.Y.Z
```

The Go module proxy and pkg.go.dev pick up the tag automatically (the first
request for a new version warms the proxy cache).

## Versioning

- Current version lives in [`.release-please-manifest.json`](.release-please-manifest.json).
- [Semantic Versioning](https://semver.org/), derived automatically from
  Conventional Commit types since the previous tag.
- Pre-1.0.0: minor versions may contain breaking changes. After 1.0.0, standard
  SemVer applies. (When moving to v2+, the Go module path must gain a `/v2`
  suffix per Go's module rules.)

## First release

The first run is pinned to `v0.1.0` via `"release-as": "0.1.0"` in
`.release-please-config.json`. **Remove that line after `v0.1.0` is tagged** so
subsequent versions are computed from commits.

## Required configuration

| Name | Kind | Used by | Purpose |
|---|---|---|---|
| `SPS_RELEASE_CLIENT_ID` | variable | `versioning.yml` | GitHub App client ID for release-please |
| `SPS_RELEASE_PRIVATE_KEY` | secret | `versioning.yml` | GitHub App private key |

The org's release GitHub App must be installed on this repository for the above
to work.

## Manual / emergency release

Prefer the normal flow. In rare cases (release-please unavailable, out-of-band
hotfix), a release can be cut by hand:

1. Bump the version in `.release-please-manifest.json` and add the corresponding
   entry to `CHANGELOG.md`. Commit to `main`.
2. Tag the commit: `git tag -a vX.Y.Z -m "Release vX.Y.Z" && git push origin vX.Y.Z`.
3. The next release-please run reconciles its state with the updated manifest.
