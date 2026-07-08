---
name: code_review
description: Instructions and guidelines for conducting comprehensive code reviews in the Xynon project.
---
# Code Review Instructions for Autonomous Agents

As an autonomous agent, when you are tasked with reviewing code (e.g., Pull Requests, diffs, or newly implemented features), you must review the changes strictly against the project's documentation.

## 1. Documentation-Driven Review
Before and during the review, you MUST read and cross-reference the relevant documentation files outlined in `AGENTS.md`. Do NOT rely on generic code review standards; instead, verify against:

- **Coding & Review Guidelines:** `docs/architecture/coding-guidelines.md` (for performance, security, style, and testing requirements).
- **Architecture & System Design:** `docs/architecture/system-design.md` and `docs/architecture/api-spec.md`.
- **Product Requirements:** The relevant PRD in `docs/requirements/` (e.g., `PRD-proxy-engine.md`, `PRD-wasm-plugins.md`).

## 2. Testing & Definition of Done
As per the `AGENTS.md` guidelines, no review is complete until you have verified:
1. `make test` runs without errors (Unit Tests).
2. `make test-e2e` runs without errors (End-to-End Tests).
3. If new APIs, flags, or system behaviors are introduced, the corresponding `docs/` files have been updated.

## 3. Autonomous Remediation (Loop Engineering)
If the user's intent is for you to not just review but also resolve issues, use the **Loop Engineering** process:
1. Apply the necessary fixes based on your review findings.
2. Run tests.
3. Iterate and debug until all tests pass before returning the final result to the user.
