# Release Process

This document outlines the process for creating new releases of the asq tool and updating the Homebrew formula.

## Creating a Release

1. Update Version
   - Update version number in `Formula/asq.rb`
   - Follow [Semantic Versioning](https://semver.org/) (MAJOR.MINOR.PATCH)
   - Example: `version "0.1.0"` → `version "0.1.1"`

2. Create Git Tag
   ```bash
   git tag -a v0.1.0 -m "Release v0.1.0"
   git push origin v0.1.0
   ```

3. Create GitHub Release
   - Go to GitHub Releases page
   - Click "Create a new release"
   - Choose the tag you just created
   - Title: "v0.1.0"
   - Description: Include changelog and any breaking changes
   - Mark as "pre-release" for testing
   - GitHub will automatically create source archives (.zip and .tar.gz)

4. Update Homebrew Formula
   - Update version in `Formula/asq.rb` to match the new release
   - Update URL to point to the new release tarball
   - Calculate and update SHA256 hash:
     ```bash
     # Download the tarball
     curl -L -o asq.tar.gz https://github.com/StCredZero/asq/archive/refs/tags/v0.1.0.tar.gz
     # Calculate SHA256
     sha256sum asq.tar.gz
     ```
   - Update `sha256` in `Formula/asq.rb` with the new hash
   - Remove the "TODO" comment

5. Test Installation
   ```bash
   # Test local installation
   brew install --build-from-source ./Formula/asq.rb
   
   # Verify installation
   asq --help
   asq tree-sitter testdata/example1.go
   asq query testdata/example1.go
   ```

6. Finalize Release
   - If all tests pass, remove "pre-release" tag on GitHub
   - Update README.md to remove "coming soon" note for Homebrew installation
   - Create PR with formula updates

## Version Numbering

- MAJOR version for incompatible API changes
- MINOR version for new functionality in a backward compatible manner
- PATCH version for backward compatible bug fixes

## Example Release Commit

```bash
git checkout -b release/v0.1.0
# Update version and SHA256 in Formula/asq.rb
git commit -am "release: prepare v0.1.0"
git push origin release/v0.1.0
# Create PR and wait for approval
```

## Troubleshooting

If the Homebrew formula fails to install:
1. Verify the SHA256 hash matches the downloaded tarball
2. Ensure the URL points to the correct release
3. Check that all dependencies are correctly specified
4. Test with `brew install --debug --verbose`
