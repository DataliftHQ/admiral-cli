# Admiral CLI

The official command-line interface for [Admiral](https://admiral.io), the platform orchestrator by [DataliftHQ](https://github.com/DataliftHQ).

Admiral unifies infrastructure provisioning and application deployment into a single, dependency-aware control plane. It connects the tools you already use — Terraform, Helm, Kustomize, and any CI/CD system — while maintaining the dependency graph across your full stack. No proprietary formats, no lock-in. If you stop using Admiral, you keep all your manifests and modules.

The CLI provides direct access to the Admiral API for managing clusters, runners, service accounts, and deployments from your terminal or CI/CD pipelines.

## Installation

### Homebrew (macOS/Linux)

```bash
brew install DataliftHQ/tap/admiral
```

### Scoop (Windows)

```powershell
scoop bucket add admiral https://github.com/DataliftHQ/scoop-bucket
scoop install admiral
```

### Go

```bash
go install go.admiral.io/cli@latest
```

### Docker

```bash
docker run --rm ghcr.io/datalifthq/admiral:latest
```

Pre-built binaries for Linux, macOS, and Windows are available on the [Releases](https://github.com/DataliftHQ/admiral-cli/releases) page.

## Quick Start

```bash
# Authenticate with your Admiral instance
admiral login

# List your clusters
admiral cluster list
```

## Documentation

Full documentation is available at [admiral.io/docs](https://admiral.io/docs).

## Community & Feedback

- [GitHub Issues](https://github.com/DataliftHQ/admiral-cli/issues) — Bug reports and feature requests
- [GitHub Discussions](https://github.com/DataliftHQ/admiral-cli/discussions) — Questions and community conversation
- [Admiral Community](https://github.com/DataliftHQ/admiral-community) — Join the broader Admiral community

## License

Admiral CLI is licensed under the [Mozilla Public License 2.0](LICENSE).
