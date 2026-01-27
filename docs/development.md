# Development

## Requirements
- Go `1.24.9`
- `make`

## Build and test
```sh
make
```

`make` runs `fmt`, `vet`, `test`, and `build`.

To run checks without mutating formatting:
```sh
make check
```

### Install (local checkout)
```sh
go install ./cmd/asm
```

Ensure `$(go env GOPATH)/bin` (or `GOBIN`) is on your `PATH`.

## Common targets
- `make build` — build all packages
- `make bin` — build `./asm`
- `make test` — run all tests
- `make test-one PKG=./internal/cli` — run a single package
- `make test-one PKG=./internal/cli RUN=TestName` — run one test
- `make test-race` — race detector
- `make test-cover` — coverage report
- `make fmt` — apply gofmt
- `make fmt-check` — fail if gofmt would change files
- `make vet` — go vet
- `make tidy` — go mod tidy

## Architecture
- Design note: `context/specs/module-boundaries.md`
- Dependency direction: `cmd/asm` -> `internal/cli` -> `internal/asm` -> `internal/manifest`, `internal/source`, `internal/gitstore`, `internal/linker`
- Only `internal/gitstore` imports `go-git`
- Use-case functions return report structs; CLI formats output

## Test notes
- Tests use `t.TempDir()` for filesystem fixtures.
- CLI tests set `HOME` via `t.Setenv` to avoid writing to real user config.
- Avoid `t.Parallel()` in CLI tests (cobra globals).

## Storage locations
Repo-local scope:
- Manifest: `<repo>/skills.jsonc` (or `skills.json`)
- Sum: `<repo>/skills.sum`
- Store: `<repo>/.asm/store/`
- Cache: `<repo>/.asm/cache/`
- Install: `<repo>/skills/`

## Release tagging (SemVer)
- Use annotated tags: `git tag -a v0.1.0 -m "v0.1.0"`
- Run `make check` before tagging.
- Push tags: `git push origin v0.1.0`
- Prefer `v0.x.y` until CLI/config stabilizes.
