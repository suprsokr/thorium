# Releasing Thorium

This document describes how to create a new Thorium release.

## Automated Release (Recommended)

Releases are automatically built and published via GitHub Actions when you push a version tag.

### Steps

1. **Update version** in `cmd/thorium/main.go`:
   ```go
   const version = "X.Y.Z"
   ```

2. **Commit the version bump**:
   ```bash
   git add cmd/thorium/main.go
   git commit -m "Bump version to X.Y.Z"
   git push
   ```

3. **Create and push a tag**:
   ```bash
   git tag vX.Y.Z
   git push origin vX.Y.Z
   ```

4. **GitHub Actions automatically**:
   - Builds binaries for all platforms (linux/darwin/windows, amd64/arm64)
   - Creates a GitHub Release with the binaries attached
   - Generates release notes from commit history

5. **Verify** the release at: https://github.com/suprsokr/thorium/releases

## Manual Release (If Needed)

If you need to build releases manually:

```bash
# Build all platform binaries
make build-all

# Binaries are created in build/:
ls -la build/
# thorium-linux-amd64
# thorium-linux-arm64
# thorium-darwin-amd64
# thorium-darwin-arm64
# thorium-windows-amd64.exe
```

Then manually create a GitHub Release and upload the binaries.

## Version Scheme

We use semantic versioning: `MAJOR.MINOR.PATCH`

- **MAJOR**: Breaking changes to CLI interface or config format
- **MINOR**: New features, backwards compatible
- **PATCH**: Bug fixes, backwards compatible

## Testing Before Release

```bash
# Run tests
make test

# Build and verify it works
make build
./build/thorium version
./build/thorium help
```
