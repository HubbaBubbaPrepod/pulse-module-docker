# Pulse Module Docker

[Pulse website](https://pulse.hubba-cloud.ru/en/) · [Practical guides](https://pulse.hubba-cloud.ru/en/guides/)

[Русский](README.md) · **English**

> A native Docker Engine collector for Pulse Monitor and Pulse Agent.

[![CI](https://github.com/HubbaBubbaPrepod/pulse-module-docker/actions/workflows/ci.yml/badge.svg)](https://github.com/HubbaBubbaPrepod/pulse-module-docker/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/HubbaBubbaPrepod/pulse-module-docker)](https://github.com/HubbaBubbaPrepod/pulse-module-docker/releases)
[![MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

[Installation](#installation) · [Configuration](#configuration) · [Metrics](#metrics-and-limits) · [Security](SECURITY.md) · [Support](SUPPORT.md)

The module collects the engine version, containers, CPU, memory, network counters, and recent logs from a selected container. The dashboard renders cards, a table, charts, and logs from the manifest. The module has its own releases as an uploadable JSON package; it runs through the agent rather than as a separate service or Docker image.

```text
Pulse Monitor → server module assignment → Pulse Agent → collector → Docker Engine
Pulse Monitor ← cards, table, charts and logs ← JSON result
```

## Installation

1. Install [Pulse Monitor](https://github.com/HubbaBubbaPrepod/pulse-monitor) and [Pulse Agent](https://github.com/HubbaBubbaPrepod/pulse-agent), then connect a real server. The first public set uses dashboard and agent **v0.4.0**.
2. Download [docker-module.json v1.0.3](https://github.com/HubbaBubbaPrepod/pulse-module-docker/releases/download/v1.0.3/docker-module.json) and [checksums.txt](https://github.com/HubbaBubbaPrepod/pulse-module-docker/releases/download/v1.0.3/checksums.txt). Compare the file's SHA-256 with its entry in `checksums.txt`:

   ```bash
   sha256sum docker-module.json
   ```

   In PowerShell:

   ```powershell
   Get-FileHash ./docker-module.json -Algorithm SHA256
   ```

3. Upload the JSON in the dashboard's **Modules** section. In the target server's settings, assign **Docker**, fill in the fields, and enable it.
4. The agent downloads and runs the collector for its platform. Open the module page in the dashboard. Create a separate assignment for each additional server.

The package includes **Linux amd64/arm64** and **Windows amd64** binaries. Go and Node.js are unnecessary on the target server. A running Docker Engine and access from the agent service are required. SHA-256 verifies file integrity; also verify the package's origin.

## Configuration

| Field          | Default      | Purpose                                                               |
| -------------- | ------------ | --------------------------------------------------------------------- |
| `socket`       | Empty string | Unix socket or Windows named pipe; empty selects the platform default |
| `logContainer` | Empty string | Optional container ID for logs, at least 12 characters                |

Linux defaults to `/var/run/docker.sock`; Windows to `\\.\pipe\docker_engine`. For rootless Docker, specify the user's socket and grant the agent service access. On Windows, check whether the service account can access the selected engine's named pipe: installing Docker Desktop alone does not guarantee access.

## Metrics and limits

| Metric     | Meaning                                                                                                        |
| ---------- | -------------------------------------------------------------------------------------------------------------- |
| Containers | At most 100 entries; `truncated` indicates a limited list, and the container count refers to that list         |
| CPU        | Two batched passes of current statistics one second apart; 100% means one fully used CPU core                  |
| Memory     | Container memory usage excluding Linux file cache; Memory total sums running containers rather than server RAM |
| Network    | Cumulative received and sent counters, not transfer rates                                                      |
| Logs       | The selected container's last 50 lines; the module does not retain a full archive                              |
| Charts     | Recent measurements within the current dashboard view                                                          |

Statistics are collected for at most 100 containers. If a running container's metrics are unavailable, the affected total stays unknown (`—`) instead of displaying a partial sum. Unknown does not mean zero.

## Permissions and security

The collector sends read requests to the Docker API, but access to its socket or pipe commonly grants root-level capabilities on the host. Treat it as privileged Docker daemon access. Do not expose the Docker API over the network for this module.

The native process runs with the agent's privileges without a sandbox. The `permissions` field describes access; it does not enforce OS restrictions. Install trusted packages only. Container logs may contain secrets. See [SECURITY.md](SECURITY.md) for details and private vulnerability reporting.

## Updates and removal

Package v1.0.3 uses `id: docker`, `schemaVersion: 1`, and `protocol: 1`. Its version is independent of the dashboard and agent. Save the current package and settings before updating, upload the new version with the same ID, then check assignments and results. Do not overwrite a released package or tag: conflicting content for an existing version is rejected.

Disable the assignment on a server to stop collection. Remove the module in dashboard settings to delete it; the removed configuration is no longer sent to agents. Cached binaries may remain in the agent's state directory. Remove them only with the agent stopped and after verifying the paths.

## Build and protocol

Requirements: Go 1.23+ and PowerShell 7.4+ for packaging. Go must be in PATH; use `-GoExe` for another path.

```powershell
go test ./...
go vet ./...
./scripts/build.ps1 -Version 1.0.3
```

The script uses the Go 1.27.2 toolchain. Outputs in `artifacts/releases/`: `docker-module.json`, `pulse-module-docker_v1.0.3_source.zip`, and `checksums.txt`.

The process receives one JSON document, `{"protocol":1,"config":{}}`, on stdin and returns one JSON object on stdout with exit code 0. Diagnostics go to stderr; the agent does not forward stderr to the dashboard. stdout is limited to **512 KiB**; download and execution share a **5-second** budget. See the [module protocol](https://github.com/HubbaBubbaPrepod/pulse-monitor/blob/main/docs/MODULES.md).

CI runs `go vet` and tests on Linux and Windows, plus the race detector on Linux. These checks do not demonstrate access to a real Docker Engine on every platform.

## Troubleshooting and contributions

`unsupported` means no binary is available for the platform. For `error`, check Docker Engine, socket/pipe, service permissions, container ID, and timeout. Unavailable logs may reflect logging driver limitations. Do not run two agents with the same ID/token.

[Contributing](CONTRIBUTING.md) · [Changelog](CHANGELOG.md) · [Support](SUPPORT.md) · [Issues](https://github.com/HubbaBubbaPrepod/pulse-module-docker/issues)

## Pulse ecosystem

| Repository                                                                     | Purpose                                                  |
| ------------------------------------------------------------------------------ | -------------------------------------------------------- |
| [Pulse Monitor](https://github.com/HubbaBubbaPrepod/pulse-monitor)             | Dashboard, servers, module packages and views            |
| [Pulse Agent](https://github.com/HubbaBubbaPrepod/pulse-agent)                 | Host metric collection and execution of assigned modules |
| [Pulse Module Docker](https://github.com/HubbaBubbaPrepod/pulse-module-docker) | This Docker Engine collector                             |

## License

[MIT](LICENSE). Contributions use the same license.

## Marketplace

The repository includes `pulse-marketplace.json` and the `pulse-module` topic. The marker must also exist at the release tag's commit; `docker-module.json` from release assets remains the installable package. Marketplace validates the exact version and SHA-256, then queues it for independent moderation. A GitHub release does not imply catalog approval. Manual JSON installation remains available.

## Interactive v2 (unreleased)

The development build emits `docker-module.json` (2.0.0-alpha1, schema2) and `docker-module-v1.json` (1.0.4, compatibility schema1). Existing published1.0.3 remains unchanged. Interactive v2 requires the unreleased module-runtime builds of both Pulse Monitor and Agent; published0.4.0 does not support it. No minimum stable release is claimed yet.

V2 shows/copies full container IDs, uses standard Pulse resource charts, and opens the clicked container's last200 log lines without changing shared configuration. Start/restart/stop require explicit `actions.execute` per node, confirmation and a full64-character ID. Stop/restart use a5-second grace period within a20-second request budget. Lost responses may produce an unknown outcome; inspect actual state before retrying. Commands are never automatically replayed. Collection continues with no management grant. Container logs may contain secrets.

Build v2 with `./scripts/build.ps1 -GoExe D:/Go/bin/go.exe -GoToolchain local -Version 2.0.0-alpha1`. For an explicitly legacy release, use `-Version 1.0.4 -LegacyOnly`; its primary package remains schema1. Uploading v2 disables component settings/grants until explicit configuration. Remove the active v1 installation first when migrating the same ID; saved configuration is not migrated automatically.
