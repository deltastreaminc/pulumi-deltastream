# CI Workflow Contract

## Event Triggers
- Pull Request: 
  ```yaml
  on:
    pull_request:
      branches: [ main ]
  ```
- Manual Trigger:
  ```yaml
  on:
    workflow_dispatch:
      inputs:
        pr_number:
          description: 'PR number to test'
          required: true
  ```

## Jobs Structure
```yaml
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
```

## Security Considerations
- Secrets are not accessible in workflows triggered by pull requests from forked repositories
- Integration tests requiring credentials only run for trusted PRs or with manual approval
- No environment secrets are exposed in logs or outputs
- CI_CREDENTIALS_YAML secret is only used when running tests from trusted sources
- All tests requiring secrets only run after manual approval for forked PRs