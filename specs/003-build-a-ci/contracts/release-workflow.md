# Release Workflow Contract

## Event Triggers
- Tag Push:
  ```yaml
  on:
    push:
      tags:
        - 'v*'
  ```
- Manual Trigger:
  ```yaml
  on:
    workflow_dispatch:
      inputs:
        version:
          description: 'Version to release (e.g., v1.2.3)'
          required: true
  ```

## Jobs Structure
```yaml
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
        uses: actions/setup-go@v3
        with:
          go-version: '1.21'
      - name: Setup Node.js
        uses: actions/setup-node@v3
        with:
          node-version: '18'
      - name: Build artifacts
        run: make build GOOS=${{ matrix.os == 'ubuntu-latest' && 'linux' || 'darwin' }} GOARCH=${{ matrix.arch }}
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
      - name: Setup credentials
        run: |
          mkdir -p ~/.pulumi-deltastream
          echo "${{ secrets.CI_CREDENTIALS_YAML }}" > examples/credentials.yaml
      - name: Test artifacts
        run: make install_sdks test

  publish:
    needs: [setup, test]
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
      - name: Download all artifacts
        uses: actions/download-artifact@v3
      - name: Setup Node.js with yarn
        uses: actions/setup-node@v4
        with:
          node-version: '20.x'
          registry-url: 'https://registry.npmjs.org'
          scope: '@deltastream'
          cache: 'yarn'
      - name: Install Yarn
        run: npm install -g yarn@1.22.22
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
            pulumi-deltastream-linux-amd64/**
            pulumi-deltastream-linux-arm64/**
            pulumi-deltastream-darwin-arm64/**
```

## Security Considerations
- NPM_TOKEN secret used only in the publish step
- GITHUB_TOKEN used with minimal required permissions
- No secrets exposed in build artifacts or logs
- Apple signing secrets (APPLE_DEVELOPER_CERTIFICATE_P12_BASE64, APPLE_DEVELOPER_CERTIFICATE_PASSWORD, APPLE_SIGNATURE_IDENTITY) used only for macOS builds
- CI_CREDENTIALS_YAML secret used only for running tests