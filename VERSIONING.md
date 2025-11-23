# Versioning

This project uses semantic versioning (SemVer) for Docker image tags.

## Automatic Versioning

### Tagged Releases
When you push a git tag in the format `v1.2.3`, the Docker build workflow will:
- Create image tags: `1.2.3`, `1.2`, `1`, and `latest` (on main/master branch)
- Example: Tagging `v1.2.3` creates tags: `1.2.3`, `1.2`, `1`, `latest`

### Branch Builds
When pushing to `main` or `master` branch (without a tag):
- Version is generated from `git describe --tags --always`
- Creates a single tag with the generated version
- Example: `1.2.3-5-gabc123` or `dev-abc123`

## Creating a Release

To create a new release:

```bash
# Create and push a version tag
git tag v1.2.3
git push
```

When a PR is merged, Docker build workflow will automatically:
1. Build the container image
2. Tag it with the semantic version
3. Push to the configured registry

## Required Secrets

Configure these secrets in Gitea:
- `REGISTRY_URL`: Container registry URL (e.g., `registry.example.com` or `docker.io`)
- `REGISTRY_USERNAME`: Registry username
- `REGISTRY_PASSWORD`: Registry password or token
- `IMAGE_NAME`: Image name (e.g., `dnstester`)

