---
name: issue_implementation
description: GitHub Issueを取得して要件を分析し、解決のための設計と実装を行い、Pull Requestを作成します。
---

# Issue Implementation Skill

This Skill is a workflow for GitHub integration that fetches a registered GitHub Issue, analyzes the stated requirements, and handles everything up to creating a Pull Request.
The actual implementation (the coding and testing loop) is completely delegated to the existing `feature` Skill.

## Prerequisites
- The `gh` (GitHub CLI) command is available and authentication is complete.
- Understand the basic repository structure (e.g., `docs/architecture/`).

## Execution Steps

### 1. Fetch and Understand the Issue (Fetch)
- If the target Issue is not specified, run the `gh issue list` command to get a list of Issues to address and decide on the target with the user.
- Once the target Issue number is decided, use the `gh issue view <Issue Number>` command to get the detailed contents of the Issue (summary, reproduction steps, task list, etc.).

### 2. Requirements Analysis and Design (Design)
- Analyze the requirements from the fetched Issue contents and confirm the files that need to be changed and the impact on the project architecture.
- Create a new working branch (e.g., `git checkout -b fix/issue-<number>` or `feature/issue-<number>`).

### 3. Delegate Implementation to Subagent (Subagent 委譲)
- Launch a subagent via `invoke_subagent` tool for the actual code implementation and testing loop.
  - **Role:** `"Feature Implementer"`
  - **Model:** `"Opus 4.6"` (Claude 4.6 Opus)
  - **Prompt:** Instruct the subagent to read and execute **`.agents/skills/feature/SKILL.md`** to implement the required changes, run tests, and perform code review.
- Wait for the subagent to complete its task. You must not consider implementation complete until the subagent reports success and all criteria in the `feature` Skill are satisfied.

### 4. Create Pull Request (Submit)
After the implementation and testing based on the `feature` Skill are complete, submit the work to the repository.
**IMPORTANT: You MUST require explicit user approval immediately before all remote or repository mutations including git add, commit, push, and gh pr create.**
1. Inspect the change set and create a commit using `git add <explicit paths>` and `git commit -m "Fix: resolve issue #<number>"`.
2. Push to remote with `git push origin <branch name>`.
3. Create a Pull Request using the `gh pr create` command.
   - In the PR body, be sure to include keywords like `Closes #<Issue Number>` or `Resolves #<Issue Number>` to automatically close the Issue.

**Example PR Creation Command:**
```bash
gh pr create --title "Fix: <Issue Summary>" --body-file - << 'MYEOF'
Closes #123

## Changes
- Implementation details 1
- Implementation details 2
MYEOF
```

### 5. Completion Reporting (Report)
Collect the URL of the created Pull Request, report it to the user, and request a review.
