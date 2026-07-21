---
name: code_review
description: Xynonプロジェクトにおける包括的なコードレビューをGemini Pro 3.1モデルを使用して実施するための指示とガイドラインです。すべての観点（機能性、設計、可読性、セキュリティ等）で逐一レビューを行います。
---
# Code Review Instructions for Autonomous Agents

As an autonomous agent, when you are tasked with reviewing code (e.g., Pull Requests, diffs, or newly implemented features), you MUST systematically review the changes against the comprehensive criteria listed below.

## Model Specification
- **Code Review (コードレビュー):** Code review tasks MUST be executed using the **Gemini 3.1 Pro (Gemini Pro 3.1)** model (e.g., specify `Model: "pro"` when launching/delegating subagents for code review tasks).

## 1. Documentation-Driven Review
Before and during the review, you MUST read and cross-reference the relevant documentation files outlined in `AGENTS.md`:
- **Coding & Review Guidelines:** `docs/architecture/coding-guidelines.md`
- **Architecture & System Design:** `docs/architecture/system-design.md` and `docs/architecture/api-spec.md`
- **Product Requirements:** The relevant PRD in `docs/requirements/`

## 2. Comprehensive Review Perspectives
You MUST evaluate the code step-by-step using the following 7 perspectives. Do not skip any of these categories:

### A. Functionality & Correctness
- **Requirements Satisfaction:** Does the code correctly implement the requirements defined in the PRD or Issue?
- **Edge Cases:** Are boundary values and abnormal inputs handled gracefully without crashing?
- **Error Handling:** Are errors caught properly, with adequate logging, meaningful messages, and necessary retries?

### B. Architecture & Design
- **System Consistency:** Does the implementation align with the architecture outlined in `docs/architecture/system-design.md`?
- **Design Principles:** Are the components loosely coupled and highly cohesive? Are SOLID principles followed?
- **Reusability & Extensibility:** Is the code designed for future changes? Is there unnecessary duplication?

### C. Readability & Maintainability
- **Naming Conventions:** Do variables, functions, and classes have clear, descriptive names indicating their purpose?
- **Complexity:** Are functions reasonably sized? Is early return used to prevent deep nesting?
- **Comments & Context:** Do comments explain the "Why" (intent/rationale) rather than just the "What" (what the code does)?

### D. Security
- **Vulnerabilities:** Are there risks of Injection, XSS, CSRF, or other common vulnerabilities?
- **Input Validation:** Is external input properly sanitized and validated?
- **Secrets Management:** Are API keys, passwords, or tokens hardcoded? (They must not be).

### E. Performance & Efficiency
- **Algorithmic Efficiency:** Are there unnecessary loops, O(N^2) operations where O(N) is possible, or inefficient data structures?
- **Resource Management:** Are there memory leaks, N+1 queries, or unclosed file/network handlers?
- **Concurrency:** Is the code thread-safe? Are there potential deadlocks or race conditions?

### F. Testing
- **Test Coverage:** Are there sufficient Unit Tests and End-to-End (E2E) tests for the new/modified logic?
- **Positive & Negative Cases:** Do the tests cover both successful execution paths and failure/error scenarios?
- **Reliability:** Are the tests deterministic (not flaky) and independent of external environments?

### G. Documentation & Standards
- **Documentation Updates:** If APIs, CLI flags, or architectures changed, are the `README.md` and `docs/` files updated?
- **Coding Guidelines:** Does the code adhere to `docs/architecture/coding-guidelines.md` and project linters/formatters?

## 3. Testing & Definition of Done
As per the `AGENTS.md` guidelines, no review is complete until you have verified testing requirements:
- Refer to **`.agents/skills/testing/SKILL.md`** to verify that Unit Tests and E2E Tests pass successfully.

## 4. Autonomous Remediation (Loop Engineering)
If the user's intent is for you to not just review but also resolve issues, you must execute the **Loop Engineering** process (iterative fixing and testing) as defined in **`.agents/skills/testing/SKILL.md`**.
Do not consider the review complete until all identified issues are fixed and all tests pass.
