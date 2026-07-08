# Coding & Code Review Guidelines

This document outlines the standards for code quality, security, and performance expected in the Xynon project. All developers and autonomous agents must adhere to these guidelines when writing or reviewing code.

## 1. Architecture Alignment
- **System Boundaries:** Changes must adhere to the system boundaries described in `system-design.md` and `api-spec.md`.
- **Requirements:** Code must fulfill the intended functionality as defined in the respective PRD files within `docs/requirements/`.

## 2. Performance
- **Proxy Engine & Plugins:** Efficiency is critical for the core proxy engine and WASM/RPC plugins.
- **Memory Management:** Look for and eliminate unnecessary memory allocations.
- **Optimization:** Avoid unoptimized loops and blocking operations, especially in hot network paths.

## 3. Security
- **Input Sanitization:** Ensure proper sanitization of all incoming data.
- **Error Handling:** Handle errors gracefully without leaking sensitive internal system data.
- **Data Transmission:** Guarantee secure data transmission and prevent unauthorized access.

## 4. Test Verification
Code cannot be considered complete without adequate test coverage.
- **Unit Tests:** New logic must be covered by unit tests, including error paths.
- **E2E Verification:** Changes must not break the end-to-end proxy and plugin integration tests.
- **Edge Cases:** Actively account for and test unhandled edge cases (e.g., timeout scenarios, missing HTTP headers, malformed payloads).

## 5. Maintainability & Style
- **Go Standards:** Code must be formatted correctly and pass `gofmt` while adhering to standard Go conventions.
- **Readability:** Use meaningful names for variables and functions. Complex algorithms must include explanatory comments.
- **Documentation Updates:** Whenever APIs, schemas, or CLI flags are modified, the corresponding documentation in the `docs/` directory must be updated.
