[Русский](CHANGELOG.ru.md)

# Changelog

## Unreleased

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
