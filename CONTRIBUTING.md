[Русский](CONTRIBUTING.ru.md)

# Contributing

## Setup and checks

Install Go 1.23+ and, for packaging, PowerShell 7.4+. Clone the repository and run from its root:

```powershell
git clone https://github.com/HubbaBubbaPrepod/pulse-module-docker.git
cd pulse-module-docker
go build ./...
go test ./...
go vet ./...
./scripts/build.ps1 -Version 1.0.2
```

The packaging script selects Go 1.27.2; use `-GoExe` if Go is outside PATH. CI runs vet and tests on Linux and Windows, and race tests on Linux. Cross-compilation does not establish runtime support: a runtime check needs a real Docker Engine and access to its Unix socket or Windows named pipe. Report the OS, architecture, Docker version, service account permissions, commands run, and limitations. Do not claim checks on an unavailable platform.

## Changes and pull requests

Keep changes focused on the reported problem. Explain the behavior before and after, update relevant tests and both language versions of affected docs, and record user-visible changes in [CHANGELOG.md](CHANGELOG.md) and [CHANGELOG.ru.md](CHANGELOG.ru.md).

Coordinate field, view, and protocol changes with [Pulse Monitor](https://github.com/HubbaBubbaPrepod/pulse-monitor) and [Pulse Agent](https://github.com/HubbaBubbaPrepod/pulse-agent); update the manifest and document compatibility. Publish binary changes under a new module version. Never overwrite an existing package or tag.

Use commit subjects in the form `type(scope): summary`, in imperative mood, at most 72 characters, without a trailing period. PRs should describe the problem, changes, actual validation, and risks. Separate compilation, automated tests, and checks against Docker Engine.

Follow the [community conduct rules](CODE_OF_CONDUCT.md). Use [Issues](https://github.com/HubbaBubbaPrepod/pulse-module-docker/issues) for bugs and proposals. Remove credentials and sensitive container logs from reports; use [private reporting](SECURITY.md) for vulnerabilities. Contributions are licensed under [MIT](LICENSE).
