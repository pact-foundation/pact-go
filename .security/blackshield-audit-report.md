# BlackShield Security Audit Report — `pact-foundation/pact-go`

- **Scanned Commit:** `c3c0e2f4975b8751e8d74bacf90cf10bf4b521c5`
- **Actionable Findings:** `4` (`0 Critical`, `4 High`, `0 Medium`, `2 Low`)
- **BlackShield Platform & Continuous Zero-Cost PR Protection:** [https://blackshield.chaplau.com](https://blackshield.chaplau.com)

## Detailed Findings & Remediation Guide

### 1. [HIGH] `CKV_DOCKER_11` — CKV_DOCKER_11: Unverified Script Piped Directly to Shell

- **Category:** `misconfiguration`
- **Location:** `docker/Dockerfile.alpine:5`
- **Details:** Piping remote HTTP downloads directly into sh or bash executes unverified code without checksum verification.
- **Recommended Fix:** Download scripts first, verify their SHA-256 checksum against an expected digest, then execute.

### 2. [HIGH] `CKV_DOCKER_1` — CKV_DOCKER_1: Container Runs as Root (Missing Non-Root USER)

- **Category:** `misconfiguration`
- **Location:** `docker/Dockerfile.alpine:1`
- **Details:** Container does not specify a non-root USER, allowing processes to execute with root privileges.
- **Recommended Fix:** Add 'USER <non-root-uid>' instruction before container entrypoint.

### 3. [HIGH] `CKV_DOCKER_11` — CKV_DOCKER_11: Unverified Script Piped Directly to Shell

- **Category:** `misconfiguration`
- **Location:** `docker/Dockerfile.debian:5`
- **Details:** Piping remote HTTP downloads directly into sh or bash executes unverified code without checksum verification.
- **Recommended Fix:** Download scripts first, verify their SHA-256 checksum against an expected digest, then execute.

### 4. [HIGH] `CKV_DOCKER_1` — CKV_DOCKER_1: Container Runs as Root (Missing Non-Root USER)

- **Category:** `misconfiguration`
- **Location:** `docker/Dockerfile.debian:1`
- **Details:** Container does not specify a non-root USER, allowing processes to execute with root privileges.
- **Recommended Fix:** Add 'USER <non-root-uid>' instruction before container entrypoint.

