# Security Policy

## Supported versions

Security fixes are applied to the latest `2.x.x` release, the current major
version. The previous major, `1.x.x`, is maintained on the [`release/1.x.x`
branch](https://github.com/pact-foundation/pact-go/tree/release/1.x.x) and,
per
[`DEVELOPER.md`](https://github.com/pact-foundation/pact-go/blob/master/DEVELOPER.md#1xx),
bug fixes and security updates are still considered there. The `0.x.x`
series is no longer maintained.

## Reporting a vulnerability

Report security issues through [GitHub's private vulnerability
reporting](https://github.com/pact-foundation/pact-go/security/advisories/new)
rather than a public issue. Do not report suspected vulnerabilities in a
public issue, discussion, or chat channel.

Please include the affected version, a description of the impact, and steps
to reproduce. You can expect an acknowledgement within a week.

pact-go downloads the Pact FFI library at install time. Vulnerabilities in
that library belong in
[pact-reference](https://github.com/pact-foundation/pact-reference).
