---
name: testing
description: ユニットテストおよびE2Eテストの実行、ならびにLoop Engineeringのテストサイクルを実行するための指示です。
---
# Testing Instructions for Autonomous Agents

When modifying features, writing new WASM plugins, or fixing bugs, you must follow the Loop Engineering methodology to ensure your changes are correct and regressions are prevented.

## Running Tests

### 1. Unit Tests
Unit tests are written in standard Go test format.
To run unit tests across the entire project:
```bash
make test
```
Or directly:
```bash
go test ./...
```
If a test fails, you must investigate the stack trace, modify the code or the test, and rerun `make test` until it passes.

### 2. End-to-End (E2E) Tests
E2E tests verify the integration of the proxy engine with WASM plugins and routing configurations.
To run the E2E tests:
```bash
make test-e2e
```
This command will:
1. Build the proxy binary.
2. Build all TinyGo WASM plugins.
3. Start the proxy server.
4. Issue HTTP requests and assert headers, status codes, and body content.
5. Terminate the proxy server.

If an E2E test fails, use the server output and `curl` responses to identify where the plugin pipeline or the core proxy logic is failing.

## Loop Engineering
As an autonomous agent, you are expected to iteratively fix issues until tests pass.
Do not prompt the user for help just because a test fails once. Instead, read the logs, locate the fault, write a patch, and re-run the tests.
Your loop should end only when all tests (both `make test` and `make test-e2e`) pass successfully.
