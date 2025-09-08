# Tasks: GitHub Actions CI and Release Workflows

**Input**: Design documents from `/specs/003-build-a-ci/`
**Prerequisites**: plan.md (required), research.md, data-model.md, contracts/

## Phase 1: Setup

- [ ] T001 Create .github directory and workflows subdirectory
  ```bash
  mkdir -p .github/workflows
  ```

## Phase 2: CI Workflow Implementation

- [ ] T002 Create CI workflow file with basic structure
  ```bash
  touch .github/workflows/ci.yml
  ```

- [ ] T003 Implement fork detection in CI workflow
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

- [ ] T004 Configure build job with tool setup in CI workflow
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

- [ ] T005 Configure test job with credentials setup in CI workflow
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
          name: provider-build
          path: .
      - name: Download yarn.lock artifact
        uses: actions/download-artifact@v4
        with:
          name: yarn-lock
          path: sdk/nodejs
      - name: Setup credentials
        run: |
          mkdir -p ~/.pulumi-deltastream
          echo "${{ secrets.CI_CREDENTIALS_YAML }}" > examples/credentials.yaml
      - name: Run tests
        run: make install_sdks test
  ```

- [ ] T006 Configure workflow triggers for CI
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

## Phase 3: Release Workflow Implementation

- [ ] T007 Create Release workflow file with basic structure
  ```bash
  touch .github/workflows/release.yml
  ```

- [ ] T008 Configure setup job for version extraction in Release workflow
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

- [ ] T009 Configure matrix build job for multiple platforms in Release workflow
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

- [ ] T010 [P] Implement code signing for macOS builds in Release workflow
  ```yaml
  # In .github/workflows/release.yml, within build job for macOS
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
  
  - name: Sign macOS binaries
    if: matrix.os == 'macos-latest'
    run: |
      find bin -type f -name "pulumi-resource-deltastream" -exec codesign --force --sign "${{ secrets.APPLE_SIGNATURE_IDENTITY }}" --options runtime {} \;
  ```

- [ ] T011 Configure test job for artifacts in Release workflow
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

- [ ] T012 Configure npm publishing job using yarn in Release workflow
  ```yaml
  # In .github/workflows/release.yml
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
      - name: Publish to npm registry using yarn
        run: |
          cd sdk/nodejs
          yarn version --new-version ${{ needs.setup.outputs.version }} --no-git-tag-version
          yarn publish --access public
        env:
          NODE_AUTH_TOKEN: ${{ secrets.NPM_TOKEN }}
  ```

- [ ] T013 Configure GitHub release creation with artifacts
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

- [ ] T014 Configure workflow triggers for Release
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

## Phase 4: Final Integration

- [ ] T015 Create complete CI workflow file
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

- [ ] T016 Create complete Release workflow file
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
        - name: Publish to npm registry using yarn
          run: |
            cd sdk/nodejs
            yarn version --new-version ${{ needs.setup.outputs.version }} --no-git-tag-version
            yarn publish --access public
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

## Phase 5: Documentation and Verification

- [ ] T017 [P] Verify workflow files with GitHub Actions Lint
  ```bash
  # If actionlint is installed, or alternatively through GitHub UI
  actionlint .github/workflows/ci.yml
  actionlint .github/workflows/release.yml
  ```

- [ ] T018 [P] Ensure all required secrets are documented in README
  ```markdown
  # Required GitHub Secrets for CI/Release Workflows
  
  The following secrets need to be set in your GitHub repository settings:
  
  - `CI_CREDENTIALS_YAML`: YAML file containing credentials for integration tests
  - `NPM_TOKEN`: Token for publishing to npm registry
  - `APPLE_DEVELOPER_CERTIFICATE_P12_BASE64`: Base64-encoded Apple developer certificate
  - `APPLE_DEVELOPER_CERTIFICATE_PASSWORD`: Password for the Apple developer certificate
  - `APPLE_SIGNATURE_IDENTITY`: Identity used for code signing macOS binaries
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