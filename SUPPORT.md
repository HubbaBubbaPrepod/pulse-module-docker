[Русский](SUPPORT.ru.md)

# Support

Read the [installation and troubleshooting guide](README.en.md) first. For bugs and proposals, open an [Issue](https://github.com/HubbaBubbaPrepod/pulse-module-docker/issues) in English or Russian. For vulnerabilities, use [private reporting](SECURITY.md).

Include module, Pulse Monitor, and Pulse Agent versions; OS/architecture; Docker Engine version; socket or pipe type; relevant service permissions; expected and actual results; and minimal reproduction steps. Redact tokens, private hostnames, container logs, and credentials. Use sanitized examples instead of a real module configuration.

For `unsupported`, check the package's platform list. For `error`, check that Docker Engine runs and the service account can access its socket/pipe, that the log container ID is valid, and that collection fits the timeout. Logging driver limitations may prevent log retrieval. Unknown totals mean incomplete statistics, not zero usage; network counters are cumulative. Full log retention is outside the module's scope.

The [dashboard repository](https://github.com/HubbaBubbaPrepod/pulse-monitor) handles upload, assignments, and rendering; the [agent repository](https://github.com/HubbaBubbaPrepod/pulse-agent) handles download and execution. File the report where the failing behavior belongs, and link related reports. Support has no guaranteed response time or SLA.
