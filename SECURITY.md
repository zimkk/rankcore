# Security Policy

## Supported Versions
Only the latest major version is supported with security updates.

## Reporting a Vulnerability
Please do not open a public issue for security vulnerabilities. Instead, contact the maintainers privately or via security@rankcore.dev.

## Threat Model
RankCore is designed to run locally in your environment or via an agent. It does not phone home, does not send your code to external servers (unless your agent does), and operates deterministically.
Please note that `rankcore audit` makes HTTP requests. It includes SSRF protection to avoid fetching private network addresses when auditing remote public sites.
