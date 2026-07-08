---
name: feature
description: Follows a 5-step Loop Engineering process (Plan, Implement, Test, Fix, Complete) to implement a new feature.
---
# Feature Implementation Skill

When implementing a new feature in this codebase, you MUST follow the strict autonomous loop described below.

## Execution Steps

1. **Implementation Plan (計画を立てる)**
   Before writing any code, analyze the requirements and generate an Implementation Plan. Detail the files to be modified, the architecture changes, and the exact steps you will take.
   
2. **Implementation (実装を行う)**
   Write the necessary code based on your Implementation Plan.

3. **Run Tests (テストを実行する)**
   Run the project's test suites to verify your changes:
   - `make test` for unit tests.
   - `make test-e2e` for End-to-End tests.

4. **Self-Analyze and Fix (自己分析と修正)**
   If any tests fail or errors occur, DO NOT report back to the user or ask for help. Instead, read the test logs, self-analyze the root cause of the failure, implement a fix, and **return to Step 3**.

5. **Completion (完了)**
   You may only consider the task "done" and report back to the user when **ALL** tests run successfully without any errors.
