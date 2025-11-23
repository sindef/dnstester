# Gitea Actions Workflows

This repository includes three automated workflows:

## 1. Security Scan (trufflehog.yml)

**Triggers:** Pull requests (opened, synchronize, reopened)

**Purpose:** Scans code for secrets and credentials using TruffleHog

**Behavior:**
- Blocks PR merge if secrets are found
- Only checks verified findings (`--only-verified` flag)
- Compares PR head against base branch

## 2. Lint (lint.yml)

**Triggers:** Pull requests (opened, synchronize, reopened)

**Purpose:** Ensures code quality and formatting

**Checks:**
- `golangci-lint` with latest version
- `go vet` for static analysis
- `go fmt` formatting check

**Behavior:**
- Fails if code is not properly formatted
- Fails if linting errors are found

## 3. Docker Build and Push (docker-build.yml)

**Triggers:**
- Push to `main` or `master` branch
- Push of tags matching `v*` pattern (e.g., `v1.2.3`)

**Purpose:** Builds and pushes Docker images to registry

**Versioning:**
- **Tagged releases:** When pushing `v1.2.3`, creates tags: `1.2.3`, `1.2`, `1`, `latest`
- **Branch builds:** Generates version from `git describe`, tags as `latest` on main/master

**Required Secrets:**
- `REGISTRY_URL`: Container registry URL
- `REGISTRY_USERNAME`: Registry username
- `REGISTRY_PASSWORD`: Registry password/token
- `IMAGE_NAME`: Docker image name

**Features:**
- Multi-stage build caching
- Automatic semantic versioning
- Multiple tag support for easy deployment

## Setup

1. Configure secrets in Gitea repository settings
2. Ensure workflows are in `.gitea/workflows/` directory
3. Workflows will automatically run on PR and merge events

