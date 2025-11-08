# Version Guide

## Current Version
**v1.0.0**

## Versioning System
This project follows [Semantic Versioning](https://semver.org/) (SemVer):

```
MAJOR.MINOR.PATCH (v1.2.3)
```

- **MAJOR** (v**X**.0.0): Incompatible API changes or major feature overhauls
- **MINOR** (vX.**Y**.0): New features added in a backward-compatible manner
- **PATCH** (vX.Y.**Z**): Backward-compatible bug fixes

## How to Update Version

### 1. Update version.go
Edit `version.go` and change the Version constant:

```go
const (
    Version = "vX.Y.Z"  // Update this
)
```

### 2. Update server_frontend.go
Edit `server_frontend.go` and change the Version constant (duplicate due to build ignore):

```go
// Version information (duplicated from version.go due to build ignore)
const Version = "vX.Y.Z"  // Update this
```

### 3. Commit Message Format
All commit messages **MUST** contain the version number. Format:

```
vX.Y.Z: <commit message>
```

**Examples:**
- `v1.0.1: Fix TTS audio cache memory leak`
- `v1.1.0: Add support for Italian language`
- `v2.0.0: Complete UI redesign with new translation engine`

### 4. Git Tag (Optional but Recommended)
After committing, create a git tag for the version:

```bash
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0
```

## Version Display
The version is displayed in the bottom-right corner of the UI:
- Desktop: 12px font
- Mobile: 10px font
- Styled with monospace font, low opacity
- Non-interactive (user-select: none, pointer-events: none)

## Release Checklist

When preparing a new release:

1. [ ] Update `version.go`
2. [ ] Update `server_frontend.go` (if applicable)
3. [ ] Test build: `go build .`
4. [ ] Run all tests: `go test ./...`
5. [ ] Run fuzz tests: `.\scripts\test\run-fuzz-tests.ps1`
6. [ ] Run build combinations: `.\scripts\test\test-build-combinations.ps1`
7. [ ] Update CHANGELOG.md (if exists)
8. [ ] Commit with version: `git commit -m "vX.Y.Z: <description>"`
9. [ ] Create git tag: `git tag -a vX.Y.Z -m "Release vX.Y.Z"`
10. [ ] Push with tags: `git push && git push --tags`

## Version History

### v1.0.0 (Initial Release)
- Initial versioning system implementation
- Version display in UI (bottom-right corner)
- Semantic versioning adopted
