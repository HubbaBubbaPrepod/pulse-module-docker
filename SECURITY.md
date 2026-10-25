[Русский](SECURITY.ru.md)

# Security policy

## Reporting a vulnerability

Use [GitHub Private Vulnerability Reporting](https://github.com/HubbaBubbaPrepod/pulse-module-docker/security/advisories/new). Do not post an exploit, credentials, private server addresses, or sensitive container logs in public Issues. Include the module, dashboard and agent versions, OS and architecture, affected behavior, reproduction steps, impact, and any mitigation. Test only on systems you own or administer. Response times and an SLA are not promised.

Report issues against the latest release; consult [release notes](CHANGELOG.md) when using an earlier version. There is no published backport schedule.

## Trust and privileges

The native collector runs with the agent's account privileges and has no sandbox. Its manifest `permissions` is descriptive, not an enforcement mechanism. The collector makes read requests, but access to the Docker socket or Windows named pipe commonly provides root-level host capabilities. Only grant access to a trusted agent and install module packages from a trusted source. Do not expose the Docker API over the network to make this module work.

Download tagged [release assets](https://github.com/HubbaBubbaPrepod/pulse-module-docker/releases), verify `docker-module.json` against `checksums.txt`, and inspect the source where appropriate. SHA-256 establishes integrity, not publisher trust. Never replace published binary content under the same version.

Logs may reveal application secrets or personal data. Select a log container deliberately and restrict dashboard access. If a package or account is compromised, disable the module assignment, stop the affected agent when needed, preserve evidence, rotate exposed credentials, and replace the package with a trusted release before resuming collection.

## Isolated components

WASI components use an embedded interpreter without host filesystem mounts, inherited secrets, raw network or native process access. Explicit broker grants authorize scoped Docker operations or configured notifications. The host runtime, broker and Docker daemon remain trusted; 128 MiB limits guest linear memory, not total host RSS. Native components remain full trust. See [the WASI contract](https://github.com/HubbaBubbaPrepod/pulse-monitor/blob/main/docs/MODULES-WASI.en.md).
