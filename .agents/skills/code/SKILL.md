---
name: code
description: GitHub Issueからの新規機能実装・Pull Request作成、および既存PRのコードレビューコメント取得と修正対応を行います。
---

# Code Skill

This Skill integrates the workflow for GitHub issue resolution, feature implementation, Pull Request creation, and addressing PR review feedback.

## Model Specification
- **Implementation & Test-Fix Loop (実装・テスト修正・レビュー対応サイクル):** MUST be executed using the **Claude 4.6 Opus (Opus 4.6)** model.
- **Code Review (コードレビュー):** MUST be delegated to a subagent using the **Gemini Pro 3.1 (pro)** model.

## Prerequisites
- The `gh` (GitHub CLI) command is available and authentication is complete.
- Understand the basic repository structure (e.g., `docs/architecture/`).

## Execution Modes
Determine the user's intent and proceed with either **Mode A (New Issue Implementation)** or **Mode B (Review Fix)**.

---

### Mode A: New Issue Implementation (新規実装フロー)
Use this mode when starting a new task from a GitHub Issue.

#### 1. Fetch and Understand the Issue (Issue取得と要件定義)
- If the target Issue is not specified, run `gh issue list` to get a list of Issues to address and decide on the target with the user.
- Once the target Issue number is decided, use `gh issue view <Issue Number>` to get the detailed contents of the Issue (summary, reproduction steps, task list, etc.).
- Analyze the requirements and confirm the files that need to be changed.
- Create a new working branch (e.g., `git checkout -b fix/issue-<number>` or `feature/issue-<number>`).

#### 2. Implementation Plan (実装計画)
Before writing any code, analyze the requirements and generate an Implementation Plan. Detail the files to be modified and the architecture changes.

#### 3. Implementation (実装)
Write the necessary code based on your Implementation Plan using **Claude 4.6 Opus**.
- **Plugin Scaffolding:** If adding a new WASM/RPC plugin, follow **`.agents/skills/plugin_scaffold/SKILL.md`**.
- **Documentation Update:** Update `docs/` or `README.md` to reflect new specifications, API changes, etc.

#### 4. Test & Fix (テストと修正 / Loop Engineering)
Execute tests and self-correcting loop using **Claude 4.6 Opus** according to **`.agents/skills/testing/SKILL.md`**.

#### 5. Code Review (ローカルでのコードレビュー)
Once all tests pass, launch a subagent via `invoke_subagent` to perform a code review following **`.agents/skills/code_review/SKILL.md`**:
- **Role:** `"Code Reviewer"`, **Model:** `"pro"`
If any issues are found, return to Step 3 or 4 to resolve them before proceeding.

#### 6. Create Pull Request (Submit)
1. Commit changes: `git add .` and `git commit -m "Fix: resolve issue #<number>"`.
2. Push to remote: `git push origin <branch name>`.
3. Create a Pull Request using `gh pr create`. (e.g., `gh pr create --title "Fix: <Issue Summary>" --body "Closes #<Issue Number>"`)

#### 7. Completion Reporting & Learn
Collect the PR URL, report to the user, and request a review. Recommend `/learn` if you acquired valuable insights.

---

### Mode B: PR Review Fix (レビュー対応フロー)
Use this mode when responding to feedback or code review comments on an existing Pull Request.

#### 1. Fetch Review Comments (レビューの取得)
- Checkout the existing PR branch (e.g., `gh pr checkout <PR Number>`).
- Use the `gh pr view <PR Number> --comments` command (or view specific review threads) to fetch the reviewer's feedback.
- Analyze the feedback and understand the required fixes.

#### 2. Fix Implementation (修正の実装)
- Modify the code to address the reviewer's comments using **Claude 4.6 Opus**.
- If the reviewer requests documentation updates, apply them to the `docs/` folder.

#### 3. Test (テストによる検証)
- Run tests (`make test` and `make test-e2e`) to ensure the fixes haven't broken any existing functionality, following **`.agents/skills/testing/SKILL.md`**.

#### 4. Push Updates (修正のプッシュ)
1. Commit the fixes: `git add .` and `git commit -m "Fix: address review comments"`.
2. Push to the remote branch: `git push origin <branch name>`.
3. (Optional) Comment on the PR to notify the reviewer that fixes have been applied: `gh pr comment <PR Number> --body "Review comments have been addressed."`

#### 5. Completion Reporting
Report to the user that the review feedback has been addressed and pushed.
