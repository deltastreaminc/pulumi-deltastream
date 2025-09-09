# Tasks: GitHub Actions CI and Release Workflows

**Input**: Design documents from `/specs/003-build-a-ci/`
**Prerequisites**: plan.md (required), research.md, data-model.md, contracts/

## Phase 1: Setup

- [x] T001 Create .github directory and workflows subdirectory
  ```bash
  mkdir -p .github/workflows
  ```

## Phase 2: CI Workflow Implementation

- [x] T002 Create CI workflow file with basic structure
  ```bash
  touch .github/workflows/ci.yml
  ```

- [x] T003 Implement fork detection in CI workflow
  ```yaml
  # In .github/workflows/ci.yml
  jobs:
    setup:
      runs-on: ubuntu-latest
      outputs:
        is_fork: ${{ steps.check.outputs.is_fork }}
      steps:
        - name: Check if PR is from fork
          id: check
          run: echo "is_fork=${{ github.event.pull_request.head.repo.full_name != github.repository }}" >> $GITHUB_OUTPUT
  ```

- [x] T004 Configure build job with tool setup in CI workflow
  ```yaml
  # In .github/workflows/ci.yml
  build:
    needs: setup
    # Build job can run for all PRs since it doesn't use secrets
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v3
      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.24.x'
      - name: Setup Pulumi
        uses: pulumi/actions@v4
        with:
          pulumi-version: '>=3.182.0'
      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '20.x'
          cache: 'yarn'
      - name: Install Yarn
        run: npm install -g yarn@1.22.22
      - name: Build
        run: make clean build schema generate build_sdks
      - name: Upload provider build artifacts
        uses: actions/upload-artifact@v4
        with:
          name: provider-build
          path: |
            bin/**
            schema.json
            sdk/**
      - name: Upload yarn.lock artifact
        uses: actions/upload-artifact@v4
        with:
          name: yarn-lock
          path: sdk/nodejs/yarn.lock
  ```

- [x] T005 Configure test job with credentials setup in CI workflow
  ```yaml
  # In .github/workflows/ci.yml
  test:
    needs: [setup, build]
    if: ${{ needs.setup.outputs.is_fork == 'false' || github.event_name == 'workflow_dispatch' }}
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v3
      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.24.x'
      - name: Setup Pulumi
        uses: pulumi/actions@v4
        with:
          pulumi-version: '>=3.182.0'
      - name: Download provider build artifacts
        uses: actions/download-artifact@v4
        with:
          name: provider-build
          path: .
      - name: Download yarn.lock artifact
        uses: actions/download-artifact@v4
        with:
          name: yarn-lock
          path: sdk/nodejs
      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '20.x'
          cache: 'yarn'
      - name: Install Yarn
        run: npm install -g yarn@1.22.22
      - name: Setup credentials
        run: |
          mkdir -p ~/.pulumi-deltastream
          echo "${{ secrets.CI_CREDENTIALS_YAML }}" > examples/credentials.yaml
      - name: Run tests
        run: make install_sdks test
  ```

- [x] T006 Configure workflow triggers for CI
  ```yaml
  # In .github/workflows/ci.yml
  name: CI
  on:
    pull_request:
      branches: [ main ]
    workflow_dispatch:
      inputs:
        pr_number:
          description: 'PR number to test'
          required: true
  ```

## Phase 3: Release Workflow Implementation (Updated for Package Publisher)

- [x] T007 Create Release workflow file with basic structure
  ```bash
  touch .github/workflows/release.yml
  ```

- [x] T008 Configure setup job for version extraction in Release workflow
  ```yaml
  # In .github/workflows/release.yml
  setup:
    runs-on: ubuntu-latest
    outputs:
      version: ${{ steps.get_version.outputs.version }}
    steps:
      - name: Get version from tag
        id: get_version
        run: echo "version=${GITHUB_REF#refs/tags/}" >> $GITHUB_OUTPUT
        if: github.event_name == 'push'
      - name: Get version from input
        id: get_input_version
        run: echo "version=${{ github.event.inputs.version }}" >> $GITHUB_OUTPUT
        if: github.event_name == 'workflow_dispatch'
  ```

- [x] T009 Configure matrix build job for multiple platforms in Release workflow
  ```yaml
  # In .github/workflows/release.yml
  build:
    needs: setup
    strategy:
      matrix:
        os: [ubuntu-latest, macos-latest]
        arch: [amd64, arm64]
        exclude:
          - os: macos-latest
            arch: amd64
    runs-on: ${{ matrix.os }}
    steps:
      - name: Checkout code
        uses: actions/checkout@v3
      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.24.x'
      - name: Setup Pulumi
        uses: pulumi/actions@v4
        with:
          pulumi-version: '>=3.182.0'
      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '20.x'
          cache: 'yarn'
      - name: Install Yarn
        run: npm install -g yarn@1.22.22
      - name: Build artifacts
        run: make build GOOS=${{ matrix.os == 'ubuntu-latest' && 'linux' || 'darwin' }} GOARCH=${{ matrix.arch }}
      - name: Upload provider build artifacts (matrix)
        uses: actions/upload-artifact@v4
        with:
          name: provider-build-${{ matrix.os == 'ubuntu-latest' && 'linux' || 'darwin' }}-${{ matrix.arch }}
          path: |
            bin/**
            schema.json
            sdk/**
      - name: Upload yarn.lock artifact
        if: matrix.os == 'ubuntu-latest' && matrix.arch == 'amd64'
        uses: actions/upload-artifact@v4
        with:
          name: yarn-lock
          path: sdk/nodejs/yarn.lock
      - name: Upload artifacts
        uses: actions/upload-artifact@v3
        with:
          name: pulumi-deltastream-${{ matrix.os == 'ubuntu-latest' && 'linux' || 'darwin' }}-${{ matrix.arch }}
          path: bin/
  ```

-- REMOVED: Previous macOS code signing task (T010) replaced by unsigned darwin builds per pulumi/pulumi-aws pattern.

- [x] T011 Configure test job for artifacts in Release workflow
  ```yaml
  # In .github/workflows/release.yml
  test:
    needs: [setup, build]
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v3
      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.24.x'
      - name: Setup Pulumi
        uses: pulumi/actions@v4
        with:
          pulumi-version: '>=3.182.0'
      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '20.x'
          cache: 'yarn'
      - name: Install Yarn
        run: npm install -g yarn@1.22.22
      - name: Download provider build artifacts
        uses: actions/download-artifact@v4
        with:
          path: artifacts
          pattern: provider-build-*
          merge-multiple: true
      - name: Download yarn.lock artifact
        uses: actions/download-artifact@v4
        with:
          name: yarn-lock
          path: sdk/nodejs
      - name: Download all artifacts
        uses: actions/download-artifact@v3
        with:
          path: artifacts
      - name: Move artifacts to bin directory
        run: |
          mkdir -p bin
          cp -r artifacts/pulumi-deltastream-*/bin/* bin/
      - name: Setup credentials
        run: |
          mkdir -p ~/.pulumi-deltastream
          echo "${{ secrets.CI_CREDENTIALS_YAML }}" > examples/credentials.yaml
      - name: Run tests
        run: make install_sdks test
  ```

-- REMOVED: Manual Node publish (replaced by package publisher composite action)

-- (Renumbered due to removals)
  ```yaml
  # In .github/workflows/release.yml, within publish job
  - name: Create GitHub Release
    uses: softprops/action-gh-release@v1
    with:
      name: Release ${{ needs.setup.outputs.version }}
      files: |
        artifacts/pulumi-deltastream-linux-amd64/bin/**
        artifacts/pulumi-deltastream-linux-arm64/bin/**
        artifacts/pulumi-deltastream-darwin-arm64/bin/**
      tag_name: ${{ needs.setup.outputs.version }}
    env:
      GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
  ```

- [x] T014 Configure workflow triggers for Release
### New Packaging & Multi-Language Tasks

- [x] T015 Add packaging job to create provider tarballs per OS/ARCH
  ```yaml
  package:
    needs: build
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Download build artifacts
        uses: actions/download-artifact@v4
        with:
          pattern: provider-build-*
          path: dist/raw
          merge-multiple: true
      - name: Create tarballs & checksums
        run: |
          set -euo pipefail
          VERSION=${{ needs.setup.outputs.version }}
          mkdir -p dist/pkg
          for dir in dist/raw/bin/*; do
            # Expect architecture/OS encoded in matrix artifact name; adjust if layout differs
            :
          done
          # Fallback: scan raw artifacts for provider binaries by naming convention
          find dist/raw -type f -name 'pulumi-resource-deltastream' | while read -r bin; do
            osarch=$(echo "$bin" | sed -E 's/.*(linux|darwin)-(amd64|arm64).*/\1-\2/')
            name="pulumi-resource-deltastream-v${VERSION}-${osarch}.tar.gz"
            work=dist/pkg/work
            mkdir -p "$work"
            cp "$bin" "$work/pulumi-resource-deltastream"
            cp dist/raw/schema.json "$work/schema.json" 2>/dev/null || true
            [ -f LICENSE ] && cp LICENSE "$work/" || true
            [ -f README.md ] && cp README.md "$work/" || true
            tar -czf "dist/pkg/${name}" -C "$work" .
            rm -rf "$work"
          done
          (cd dist/pkg && shasum -a 256 pulumi-resource-deltastream-v*.tar.gz > pulumi-deltastream_${VERSION}_checksums.txt)
      - name: Upload packaged artifacts
        uses: actions/upload-artifact@v4
        with:
          name: provider-packages
          path: dist/pkg/*
  ```

- [x] T016 Generate schema diff summary before release
  ```yaml
  - name: Schema diff summary
    id: schema_diff
    run: |
      LAST=$(gh release view --json tagName -q .tagName || echo 'NONE')
      echo 'summary<<EOF' >> $GITHUB_OUTPUT
      if [ "$LAST" != 'NONE' ]; then
        schema-tools compare --provider deltastream --old-commit "$LAST" --repository github://api.github.com/deltastreaminc --new-commit --local-path=provider/cmd/pulumi-resource-deltastream/schema.json
      fi
      echo 'EOF' >> $GITHUB_OUTPUT
  ```

-- (Renumbered; remains required) Add Go SDK publish (curated) task
  ```yaml
  - name: Publish Go SDK
    uses: pulumi/publish-go-sdk-action@v1
    with:
      repository: ${{ github.repository }}
      base-ref: ${{ github.sha }}
      source: sdk
      path: sdk
      version: ${{ needs.setup.outputs.version }}
      additive: false
      files: |
        go.*
        go/**
        !*.tar.gz
  ```

-- (Renumbered) Add checksum + tarball assets to release
  ```yaml
  - name: Create GitHub Release (packaged)
    uses: softprops/action-gh-release@v2
    with:
      tag_name: ${{ needs.setup.outputs.version }}
      body: ${{ steps.schema_diff.outputs.summary }}
      files: |
        artifacts/provider-packages/pulumi-resource-deltastream-v*.tar.gz
        artifacts/provider-packages/pulumi-deltastream_*_checksums.txt
        artifacts/provider-packages/schema.json
    env:
      GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
  ```

-- (Renumbered) Add post-publish verification job
  ```yaml
  verify:
    needs: publish
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Install Pulumi CLI
        uses: pulumi/actions@v5
      - name: Node smoke test
        run: |
          npm init -y >/dev/null
          npm install @deltastream/pulumi-deltastream@${{ needs.setup.outputs.version#v }}
          echo "import * as ds from '@deltastream/pulumi-deltastream';" > index.js
      - name: Go mod init smoke test
        run: |
          go mod init verify && go get github.com/deltastreaminc/pulumi-deltastream/sdk/go/pulumi-deltastream@${{ needs.setup.outputs.version }}
  ```

- [x] T021 Introduce prerelease flag handling (`isPrerelease` input) controlling draft releases & schema diff inclusion.

- [ ] T022 Document new secrets (PYPI_API_TOKEN) and optional future (.NET/Java) secrets in README. (Deferred / not in current scope)

## Adjusted Dependencies
- Packaging (T015) depends on build matrix (T009) completion.
- Schema diff (T016) depends on packaging artifacts availability (schema included).
- Go SDK publishing task (T018) depends on packaging or at least build artifacts.
- Release creation (T019) depends on checksum + tarball generation (T015) and schema diff (T016).
- Verification (T020) depends on publish completion (Node/Go). 

  ```yaml
  # In .github/workflows/release.yml
  name: Release
  on:
    push:
      tags:
        - 'v*'
    workflow_dispatch:
      inputs:
        version:
          description: 'Version to release (e.g., v1.2.3)'
          required: true
  ```

## Phase 4: Final Integration (Updated)

- [x] T015 Create complete CI workflow file
  ```bash
  cat > .github/workflows/ci.yml << 'EOF'
  name: CI
  
  on:
    pull_request:
      branches: [ main ]
    workflow_dispatch:
      inputs:
        pr_number:
          description: 'PR number to test'
          required: true
  
  jobs:
    setup:
      runs-on: ubuntu-latest
      outputs:
        is_fork: ${{ steps.check.outputs.is_fork }}
      steps:
        - name: Check if PR is from fork
          id: check
          run: echo "is_fork=${{ github.event.pull_request.head.repo.full_name != github.repository }}" >> $GITHUB_OUTPUT
  
    build:
      needs: setup
      # Build job can run for all PRs since it doesn't use secrets
      runs-on: ubuntu-latest
      steps:
        - name: Checkout code
          uses: actions/checkout@v3
        - name: Setup Go
          uses: actions/setup-go@v4
          with:
            go-version: '1.24.x'
        - name: Setup Pulumi
          uses: pulumi/actions@v4
          with:
            pulumi-version: '>=3.182.0'
        - name: Build
          run: make clean build schema generate build_sdks
        - name: Setup Node.js
          uses: actions/setup-node@v4
          with:
            node-version: '20.x'
            cache: 'yarn'
        - name: Install Yarn
          run: npm install -g yarn@1.22.22
  
    test:
      needs: [setup, build]
      # Only run automatically if PR is not from a fork, otherwise needs manual approval
      if: ${{ needs.setup.outputs.is_fork == 'false' || github.event_name == 'workflow_dispatch' }}
      runs-on: ubuntu-latest
      steps:
        - name: Checkout code
          uses: actions/checkout@v3
        - name: Setup Go
          uses: actions/setup-go@v4
          with:
            go-version: '1.24.x'
        - name: Setup Pulumi
          uses: pulumi/actions@v4
          with:
            pulumi-version: '>=3.182.0'
        - name: Setup Node.js
          uses: actions/setup-node@v4
          with:
            node-version: '20.x'
            cache: 'yarn'
        - name: Install Yarn
          run: npm install -g yarn@1.22.22
        - name: Setup credentials
          run: |
            mkdir -p ~/.pulumi-deltastream
            echo "${{ secrets.CI_CREDENTIALS_YAML }}" > examples/credentials.yaml
        - name: Run tests
          run: make install_sdks test
  EOF
  ```

- [x] T016 Create complete Release workflow file (to include new packaging, schema diff, multi-language publish, verification)
  ```bash
  cat > .github/workflows/release.yml << 'EOF'
  name: Release
  
  on:
    push:
      tags:
        - 'v*'
    workflow_dispatch:
      inputs:
        version:
          description: 'Version to release (e.g., v1.2.3)'
          required: true
  
  jobs:
    setup:
      runs-on: ubuntu-latest
      outputs:
        version: ${{ steps.get_version.outputs.version }}
      steps:
        - name: Get version from tag
          id: get_version
          run: echo "version=${GITHUB_REF#refs/tags/}" >> $GITHUB_OUTPUT
          if: github.event_name == 'push'
        - name: Get version from input
          id: get_input_version
          run: echo "version=${{ github.event.inputs.version }}" >> $GITHUB_OUTPUT
          if: github.event_name == 'workflow_dispatch'
  
    build:
      needs: setup
      strategy:
        matrix:
          os: [ubuntu-latest, macos-latest]
          arch: [amd64, arm64]
          exclude:
            - os: macos-latest
              arch: amd64
      runs-on: ${{ matrix.os }}
      steps:
        - name: Checkout code
          uses: actions/checkout@v3
        - name: Setup Go
          uses: actions/setup-go@v4
          with:
            go-version: '1.24.x'
        - name: Setup Pulumi
          uses: pulumi/actions@v4
          with:
            pulumi-version: '>=3.182.0'
        - name: Setup Node.js
          uses: actions/setup-node@v4
          with:
            node-version: '20.x'
            cache: 'yarn'
        - name: Install Yarn
          run: npm install -g yarn@1.22.22
        
        # Apple code signing for macOS builds
        - name: Install Apple certificate
          if: matrix.os == 'macos-latest'
          run: |
            echo "${{ secrets.APPLE_DEVELOPER_CERTIFICATE_P12_BASE64 }}" | base64 --decode > certificate.p12
            security create-keychain -p "${{ github.run_id }}" build.keychain
            security default-keychain -s build.keychain
            security unlock-keychain -p "${{ github.run_id }}" build.keychain
            security import certificate.p12 -k build.keychain -P "${{ secrets.APPLE_DEVELOPER_CERTIFICATE_PASSWORD }}" -T /usr/bin/codesign
            security set-key-partition-list -S apple-tool:,apple: -s -k "${{ github.run_id }}" build.keychain
            rm certificate.p12
        
        - name: Build artifacts
          run: make build GOOS=${{ matrix.os == 'ubuntu-latest' && 'linux' || 'darwin' }} GOARCH=${{ matrix.arch }}
        
        - name: Sign macOS binaries
          if: matrix.os == 'macos-latest'
          run: |
            find bin -type f -name "pulumi-resource-deltastream" -exec codesign --force --sign "${{ secrets.APPLE_SIGNATURE_IDENTITY }}" --options runtime {} \;
        
        - name: Upload artifacts
          uses: actions/upload-artifact@v3
          with:
            name: pulumi-deltastream-${{ matrix.os == 'ubuntu-latest' && 'linux' || 'darwin' }}-${{ matrix.arch }}
            path: bin/
  
    test:
      needs: [setup, build]
      runs-on: ubuntu-latest
      steps:
        - name: Checkout code
          uses: actions/checkout@v3
        - name: Setup Go
          uses: actions/setup-go@v4
          with:
            go-version: '1.24.x'
        - name: Setup Pulumi
          uses: pulumi/actions@v4
          with:
            pulumi-version: '>=3.182.0'
        - name: Setup Node.js
          uses: actions/setup-node@v4
          with:
            node-version: '20.x'
            cache: 'yarn'
        - name: Install Yarn
          run: npm install -g yarn@1.22.22
        - name: Download all artifacts
          uses: actions/download-artifact@v3
          with:
            path: artifacts
        - name: Move artifacts to bin directory
          run: |
            mkdir -p bin
            cp -r artifacts/pulumi-deltastream-*/bin/* bin/
        - name: Setup credentials
          run: |
            mkdir -p ~/.pulumi-deltastream
            echo "${{ secrets.CI_CREDENTIALS_YAML }}" > examples/credentials.yaml
        - name: Run tests
          run: make install_sdks test
  
    publish:
      needs: [setup, test]
      runs-on: ubuntu-latest
      steps:
        - name: Checkout code
          uses: actions/checkout@v3
        - name: Setup Node.js with yarn
          uses: actions/setup-node@v4
          with:
            node-version: '20.x'
            registry-url: 'https://registry.npmjs.org'
            scope: '@deltastream'
            cache: 'yarn'
        - name: Install Yarn
          run: npm install -g yarn@1.22.22
        - name: Download all artifacts
          uses: actions/download-artifact@v3
          with:
            path: artifacts
        - name: Move artifacts to bin directory
          run: |
            mkdir -p bin
            cp -r artifacts/pulumi-deltastream-*/bin/* bin/
        - name: Publish Node SDK (package publisher)
          uses: pulumi/pulumi-package-publisher@v0.0.22
          with:
            sdk: nodejs
            nodejs-path: sdk/nodejs
          env:
            NODE_AUTH_TOKEN: ${{ secrets.NPM_TOKEN }}
        
        - name: Create GitHub Release
          uses: softprops/action-gh-release@v1
          with:
            name: Release ${{ needs.setup.outputs.version }}
            files: |
              artifacts/pulumi-deltastream-linux-amd64/bin/**
              artifacts/pulumi-deltastream-linux-arm64/bin/**
              artifacts/pulumi-deltastream-darwin-arm64/bin/**
            tag_name: ${{ needs.setup.outputs.version }}
          env:
            GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
  EOF
  ```

## Phase 5: Documentation and Verification (Updated)

- [ ] T017 [P] Verify workflow files with GitHub Actions Lint (pending optional lint execution)
  ```bash
  # If actionlint is installed, or alternatively through GitHub UI
  actionlint .github/workflows/ci.yml
  actionlint .github/workflows/release.yml
  ```

-- [x] T018 [P] Ensure all required secrets are documented in README (remove Apple signing secrets; keep NPM_TOKEN, CI_CREDENTIALS_YAML)
  ```markdown
  # Required GitHub Secrets for CI/Release Workflows
  
  The following secrets need to be set in your GitHub repository settings:
  
  - `CI_CREDENTIALS_YAML`: YAML file containing credentials for integration tests
  - `NPM_TOKEN`: Token for publishing to npm registry
    (Removed Apple signing secrets; no longer required)
  ```

## Dependencies
- T001 (Create directories) blocks all other tasks
- T003-T006 (CI workflow components) block T015 (Complete CI workflow)
- T008-T014 (Release workflow components) block T016 (Complete Release workflow)
- T015 and T016 are independent (parallel)
- T017 and T018 depend on T015 and T016

## Parallel Example
```
# Launch these tasks in parallel:
T001: "Create .github directory and workflows subdirectory"
T017: "Verify workflow files with GitHub Actions Lint"
T018: "Ensure all required secrets are documented in README"
```

## Notes
- [P] tasks are marked where different files are modified and can be run in parallel
- Each task includes specific file paths or code snippets for clarity
- The complete workflow files in T015 and T016 serve as the final implementation
- Security considerations are integrated throughout the tasks, especially in T003 and T005
- Artifact strategy: Every job runs in a clean VM; required build outputs (`bin/**`, `schema.json`, generated `sdk/**`, and `sdk/nodejs/yarn.lock`) are transferred via upload/download artifacts. Release build uploads matrix-qualified names; `yarn.lock` only uploaded once.
- Ordering & Executable Restoration: Download provider and yarn-lock artifacts BEFORE `actions/setup-node` so caching sees `sdk/nodejs/yarn.lock`. Immediately after download, restore execute bits with `chmod +x bin/* || true` (or path-qualified for release) before tests or publishing.