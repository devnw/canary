# Development Environment Setup

## Requirements

- **Go 1.20+** (1.24+ recommended; see `go.mod`)
- Git

## Build

```bash
# From repository root
go build ./cmd/canary
# Binary: ./canary (or go build -o bin/canary ./cmd/canary)
```

## Test

```bash
go test ./...
# Acceptance tests (scanner self-canary, update-stale, etc.):
go test ./tools/canary/...
```

## Lint / format

```bash
go fmt ./...
# Optional: goimports, staticcheck per project standards
```

## Scan / verify (self-canary)

```bash
./canary scan --root . --out status.json --csv status.csv
./canary scan --verify docs/GAP_ANALYSIS.md --strict
```

## Contributing workflow

1. Create a branch, make changes, run `go test ./...`.
2. Run `./canary scan --verify docs/GAP_ANALYSIS.md` if you touch requirements.
3. Open a merge request; ensure CI (if configured) passes.

## CI/CD (GitLab)

The project uses GitLab CI/CD following the same pattern as [void](https://gitlab.com/devnw/codepros/oss/void) (devnw/ci-catalog):

- **Pipeline**: `.gitlab-ci.yml` — stages: build → test → lint → security → release → publish.
- **Release**: On tag push, [goreleaser](https://goreleaser.com/) produces binaries, archives, and **.deb** packages under `dist/`.
- **APT deployment**: The **apt-publish** job publishes `dist/*.deb` to **apt.codepros.org** (signed with `canary.gpg`). Vault at `devnw/codepros/oss/apt` is used for signing and publish credentials.
- **Docker**: Images are pushed to the GitLab Container Registry; manifests are built for amd64/arm64.

To publish a release: push a version tag (e.g. `v1.2.3`). The release job runs on **saas-linux-medium-amd64** runners; the apt-publish job runs after goreleaser and uploads the `.deb` packages.

## Optional tooling

- **Nix / direnv**: Not required. Build and test use the system Go toolchain only. If your team uses Nix for other projects, you may use it here too; it is not in the critical path for building or testing canary.
