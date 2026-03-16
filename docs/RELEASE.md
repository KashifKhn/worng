# Release Guide

This project publishes cross-platform binaries via GitHub Actions when a version tag is pushed.

## Supported binaries

- Linux: `amd64`, `arm64`
- macOS: `amd64`, `arm64`
- Windows: `amd64`

## How to release `v0.1.0`

1. Ensure local branch is clean and up to date.
2. Run validation gate:

   ```bash
   make test
   make test-golden
   make test-fuzz
   make test-coverage
   make lint
   ```

3. Create and push tag:

   ```bash
   git tag v0.1.0
   git push origin v0.1.0
   ```

4. Open GitHub Actions and watch the `Release` workflow.
5. After success, binaries and `checksums.txt` appear in the GitHub Release page.

## Artifact naming

`worng_<version>_<os>_<arch>[.exe]`

Examples:

- `worng_0.1.0_linux_amd64`
- `worng_0.1.0_darwin_arm64`
- `worng_0.1.0_windows_amd64.exe`
