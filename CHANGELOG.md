[Русский](CHANGELOG.ru.md)

# Changelog

## Unreleased

- Add a broker-backed WASI guest and schema-2 v3 package. Keep native v1/v2 source and legacy v1 release artifacts. The guest has no direct Docker socket access; the Agent retains state/readiness verification.

- Update the existing Windows support dependency to fix Unicode length validation; source builds now require Go 1.25+, release builds use Go 1.27.3.

- Verify exact-container state after start/stop/restart; wait for configured healthchecks and confirm a new StartedAt after restart.
- Preserve unknown outcomes when readiness cannot be established within the 20-second budget; never repeat a mutation automatically.
- Build release archives with Go 1.27.3.

- Fix Windows named pipe transport with go-winio v0.6.2, context cancellation and the 20-second manual operation budget.

- Add schema2 selected-container logs and scoped start/restart/stop operations with bounded grace/deadlines and unknown outcomes. Retain schema1 collector behavior, add container IDs, and build separate v1 compatibility artifacts.

## 1.0.3

### Added

- Add the `pulse-module` discovery marker for Pulse Marketplace and include it in the source archive.
- Keep the native collector behavior and schema-1 installation protocol unchanged. Marketplace publication requires independent moderation; discovery is not approval.

## 1.0.2

### Fixed

- CPU is calculated from two current samples one second apart rather than the ratio of accumulated lifetime counters.
- Totals remain unknown when a running container's metrics are unavailable, instead of showing a partial sum.

## 1.0.1

### Added

- Standalone repository, MIT license, build and installation documentation.
- JSON packaging for Linux amd64/arm64 and Windows amd64, a source archive, and SHA-256 checksums.

### Fixed

- Windows Docker named-pipe connections use asynchronous handles and retry a busy pipe within the request deadline.

### Notes

- The module retains `id: docker`, `schemaVersion: 1`, and `protocol: 1`, with an independent package version.
- The first public set combines Pulse Monitor v0.4.0, Pulse Agent v0.4.0, and Docker module v1.0.1.
- v1.0.1 identifies new package content; an installed v1.0.0 remains unchanged.

## 1.0.0

### Added

- Original Docker collector: engine version, containers, resource metrics, network counters, and optional recent container logs.
- Native JSON stdin/stdout process for module protocol v1. This package predates the standalone repository.
